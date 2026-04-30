package server

import (
	"fmt"
	"log/slog"

	"github.com/AffanSurya/xarela-backend/internal/config"
	"github.com/AffanSurya/xarela-backend/internal/handler"
	appmiddleware "github.com/AffanSurya/xarela-backend/internal/middleware"
	"github.com/AffanSurya/xarela-backend/internal/service"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

type Server struct {
	cfg    config.Config
	logger *slog.Logger
	echo   *echo.Echo
}

func New(cfg config.Config, logger *slog.Logger) *Server {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(echoMiddleware.Recover())
	e.Use(appmiddleware.RequestLogger(logger))

	healthService := service.NewHealthService()
	healthHandler := handler.NewHealthHandler(healthService)
	healthHandler.Register(e.Group(""))

	authService := service.NewAuthService(cfg.DatabaseDSN)
	authHandler := handler.NewAuthHandler(authService)
	authHandler.Register(e.Group("/api/v1/auth"))

	return &Server{
		cfg:    cfg,
		logger: logger,
		echo:   e,
	}
}

func (s *Server) Start() error {
	address := fmt.Sprintf(":%s", s.cfg.Port)
	s.logger.Info("starting server", "address", address, "log_level", s.cfg.LogLevel)
	return s.echo.Start(address)
}
