package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AffanSurya/xarela-backend/internal/service"
	"github.com/labstack/echo/v4"
)

func TestRequireVerifiedUserBlocksUnverified(t *testing.T) {
	e := echo.New()
	handler := RequireVerifiedUser()(func(c echo.Context) error {
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
	handler := RequireVerifiedUser()(func(c echo.Context) error {
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
