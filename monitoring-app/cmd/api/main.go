package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"monitoring-platform/internal/httpx"
	chstorage "monitoring-platform/internal/storage/clickhouse"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	// projectID from env for now; will come from auth token in phase 3.
	projectID := os.Getenv("DEFAULT_PROJECT_ID")
	if projectID == "" {
		projectID = "default"
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	chCfg := chstorage.ConfigFromEnv()
	chConn, err := chstorage.Open(ctx, chCfg)
	if err != nil {
		logger.Warn("clickhouse unavailable — reads will return empty", "err", err)
	}

	var reader *chstorage.Reader
	if chConn != nil {
		reader = chstorage.NewReader(chConn)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", httpx.Health("api"))

	mux.HandleFunc("/v1/runs", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if reader == nil {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"runs":[]}`))
			return
		}
		runs, err := reader.ListRuns(r.Context(), chstorage.RunFilter{
			ProjectID: projectID,
		})
		if err != nil {
			logger.Error("list runs", "err", err)
			http.Error(w, "storage error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"runs": runs})
	})

	mux.HandleFunc("/v1/runs/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		traceID := r.URL.Path[len("/v1/runs/"):]
		if traceID == "" {
			http.Error(w, "trace_id required", http.StatusBadRequest)
			return
		}
		if reader == nil {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"spans":[]}`))
			return
		}
		spans, err := reader.SpansByTrace(r.Context(), projectID, traceID)
		if err != nil {
			logger.Error("spans by trace", "err", err)
			http.Error(w, "storage error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"spans": spans})
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("api listening", "addr", addr)
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
