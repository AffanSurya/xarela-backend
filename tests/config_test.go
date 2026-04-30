package tests

import (
	"testing"

	"github.com/AffanSurya/xarela-backend/internal/config"
)

func TestConfigLoadUsesDefaultsWhenEnvironmentIsEmpty(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("DATABASE_DSN", "")

	cfg := config.Load()

	if cfg.Port != "8080" {
		t.Fatalf("expected default port %q, got %q", "8080", cfg.Port)
	}

	if cfg.LogLevel != "info" {
		t.Fatalf("expected default log level %q, got %q", "info", cfg.LogLevel)
	}

	if cfg.DatabaseDSN != "" {
		t.Fatalf("expected default database dsn to be empty, got %q", cfg.DatabaseDSN)
	}
}

func TestConfigLoadReadsEnvironmentOverrides(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/xarela?sslmode=disable")

	cfg := config.Load()

	if cfg.Port != "9090" {
		t.Fatalf("expected port override %q, got %q", "9090", cfg.Port)
	}

	if cfg.LogLevel != "debug" {
		t.Fatalf("expected log level override %q, got %q", "debug", cfg.LogLevel)
	}

	if cfg.DatabaseDSN != "postgres://user:pass@localhost:5432/xarela?sslmode=disable" {
		t.Fatalf("expected database dsn override to be used")
	}
}