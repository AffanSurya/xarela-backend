package main

import (
	"os"

	"github.com/AffanSurya/xarela-backend/internal/config"
	"github.com/AffanSurya/xarela-backend/internal/logger"
	"github.com/AffanSurya/xarela-backend/internal/server"
)

func main() {
	cfg := config.Load()
	appLogger := logger.New(cfg.LogLevel)
	app := server.New(cfg, appLogger)

	if err := app.Start(); err != nil {
		appLogger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
