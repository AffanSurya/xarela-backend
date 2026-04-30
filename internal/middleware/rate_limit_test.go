package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
)

func TestRateLimitBlocksAfterThreshold(t *testing.T) {
	e := echo.New()
	handler := RateLimit(1, time.Hour)(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	req.RemoteAddr = "127.0.0.1:1234"

	rec1 := httptest.NewRecorder()
	c1 := e.NewContext(req, rec1)
	c1.SetPath("/api/v1/auth/login")
	if err := handler(c1); err != nil {
		t.Fatalf("first request failed: %v", err)
	}

	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req, rec2)
	c2.SetPath("/api/v1/auth/login")
	if err := handler(c2); err != nil {
		t.Fatalf("second request failed: %v", err)
	}

	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 status, got %d", rec2.Code)
	}
}
