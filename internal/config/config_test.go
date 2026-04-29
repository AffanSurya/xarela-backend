package config

import "testing"

func TestLoadUsesDefaultsWhenEnvironmentIsEmpty(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("DATABASE_DSN", "")

	cfg := Load()

	if cfg.Port != defaultPort {
		t.Fatalf("expected default port %q, got %q", defaultPort, cfg.Port)
	}

	if cfg.LogLevel != defaultLogLevel {
		t.Fatalf("expected default log level %q, got %q", defaultLogLevel, cfg.LogLevel)
	}

	if cfg.DatabaseDSN != defaultDatabaseDSN {
		t.Fatalf("expected default database dsn %q, got %q", defaultDatabaseDSN, cfg.DatabaseDSN)
	}
}

func TestLoadReadsEnvironmentOverrides(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("DATABASE_DSN", "postgres://user:pass@localhost:5432/xarela?sslmode=disable")

	cfg := Load()

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
