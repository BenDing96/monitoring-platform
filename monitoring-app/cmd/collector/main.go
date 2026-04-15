package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	coltracev1 "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/protobuf/proto"

	"monitoring-platform/internal/httpx"
	"monitoring-platform/internal/ingest"
	"monitoring-platform/internal/otelconv"
	chstorage "monitoring-platform/internal/storage/clickhouse"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	addr := os.Getenv("COLLECTOR_ADDR")
	if addr == "" {
		addr = ":4318"
	}

	// projectID from env for now; will come from API key in phase 3.
	projectID := os.Getenv("DEFAULT_PROJECT_ID")
	if projectID == "" {
		projectID = "default"
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var sink ingest.SpanSink
	chCfg := chstorage.ConfigFromEnv()
	chConn, err := chstorage.Open(ctx, chCfg)
	if err != nil {
		logger.Warn("clickhouse unavailable — spans will be dropped", "err", err)
	} else {
		if err := chstorage.Migrate(ctx, chConn); err != nil {
			logger.Warn("clickhouse migrate failed", "err", err)
		}
		sink = chstorage.NewWriter(chConn)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", httpx.Health("collector"))
	mux.HandleFunc("/v1/traces", otlpHandler(logger, projectID, sink))

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("collector listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func otlpHandler(logger *slog.Logger, projectID string, sink ingest.SpanSink) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
		if err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}

		var req coltracev1.ExportTraceServiceRequest
		if err := proto.Unmarshal(body, &req); err != nil {
			http.Error(w, "invalid protobuf", http.StatusBadRequest)
			return
		}

		spans := otelconv.SpansFromRequest(&req, projectID)
		logger.Info("received spans", "count", len(spans))

		if sink != nil {
			if err := sink.WriteSpans(r.Context(), spans); err != nil {
				logger.Error("write spans failed", "err", err)
				http.Error(w, "storage error", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
