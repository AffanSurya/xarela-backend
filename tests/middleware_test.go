package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appmiddleware "github.com/AffanSurya/xarela-backend/internal/middleware"
	"github.com/AffanSurya/xarela-backend/internal/service"
	"github.com/labstack/echo/v4"
)

func TestRateLimitBlocksAfterThreshold(t *testing.T) {
	e := echo.New()
	handler := appmiddleware.RateLimit(1, time.Hour)(func(c echo.Context) error {
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

func TestRequireVerifiedUserBlocksUnverified(t *testing.T) {
	e := echo.New()
	handler := appmiddleware.RequireVerifiedUser()(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("auth_user", service.AuthUser{IsEmailVerified: false})

	if err := handler(c); err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 status, got %d", rec.Code)
	}
}

func TestRequireVerifiedUserAllowsVerified(t *testing.T) {
	e := echo.New()
	called := false
	handler := appmiddleware.RequireVerifiedUser()(func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("auth_user", service.AuthUser{IsEmailVerified: true})

	if err := handler(c); err != nil {
		t.Fatalf("request failed: %v", err)
	}

	if !called {
		t.Fatal("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 status, got %d", rec.Code)
	}
}