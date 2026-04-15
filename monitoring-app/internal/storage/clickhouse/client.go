package clickhouse

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// Config holds connection parameters read from environment variables.
type Config struct {
	Addr     string // e.g. "clickhouse:9000"
	Database string
	Username string
	Password string
	TLS      bool
}

// ConfigFromEnv reads CH connection settings from the environment.
func ConfigFromEnv() Config {
	return Config{
		Addr:     envOr("CLICKHOUSE_ADDR", "clickhouse:9000"),
		Database: envOr("CLICKHOUSE_DB", "monitoring"),
		Username: envOr("CLICKHOUSE_USER", "default"),
		Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		TLS:      os.Getenv("CLICKHOUSE_TLS") == "true",
	}
}

// Open opens a ClickHouse connection and verifies it with a ping.
func Open(ctx context.Context, cfg Config) (driver.Conn, error) {
	opts := &clickhouse.Options{
		Addr: []string{cfg.Addr},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		Compression: &clickhouse.Compression{Method: clickhouse.CompressionLZ4},
	}
	if cfg.TLS {
		opts.TLS = &tls.Config{MinVersion: tls.VersionTLS12}
	}

	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("clickhouse open: %w", err)
	}
	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("clickhouse ping: %w", err)
	}
	return conn, nil
}

// Migrate runs the schema DDL against the open connection.
func Migrate(ctx context.Context, conn driver.Conn) error {
	ddl, err := os.ReadFile("internal/storage/clickhouse/schema.sql")
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}
	for _, stmt := range splitStatements(string(ddl)) {
		if err := conn.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("exec ddl: %w", err)
		}
	}
	return nil
}

// splitStatements splits a SQL file on semicolons, skipping comments and blanks.
func splitStatements(sql string) []string {
	var stmts []string
	var cur []byte
	for i := 0; i < len(sql); i++ {
		c := sql[i]
		if c == '-' && i+1 < len(sql) && sql[i+1] == '-' {
			// skip line comment
			for i < len(sql) && sql[i] != '\n' {
				i++
			}
			continue
		}
		if c == ';' {
			s := string(cur)
			cur = cur[:0]
			if trimmed := trimSpace(s); trimmed != "" {
				stmts = append(stmts, trimmed)
			}
			continue
		}
		cur = append(cur, c)
	}
	return stmts
}

func trimSpace(s string) string {
	start, end := 0, len(s)-1
	for start <= end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end >= start && (s[end] == ' ' || s[end] == '\t' || s[end] == '\n' || s[end] == '\r') {
		end--
	}
	if start > end {
		return ""
	}
	return s[start : end+1]
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
