package handler

import (
	"net/http"

	"github.com/AffanSurya/xarela-backend/internal/service"
	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	service service.HealthService
}

func NewHealthHandler(service service.HealthService) HealthHandler {
	return HealthHandler{service: service}
}

func (h HealthHandler) Register(group *echo.Group) {
	group.GET("/health", h.Get)
}

func (h HealthHandler) Get(c echo.Context) error {
	return c.JSON(http.StatusOK, h.service.Status())
}
