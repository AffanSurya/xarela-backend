package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AffanSurya/xarela-backend/internal/handler"
	"github.com/AffanSurya/xarela-backend/internal/service"
	"github.com/labstack/echo/v4"
)

func TestHealthHandlerReturnsOk(t *testing.T) {
	e := echo.New()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	ctx := e.NewContext(req, recorder)

	h := handler.NewHealthHandler(service.NewHealthService())

	if err := h.Get(ctx); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, recorder.Code)
	}

	if body := recorder.Body.String(); body != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected body: %q", body)
	}
}