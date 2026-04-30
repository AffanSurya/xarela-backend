package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	appmiddleware "github.com/AffanSurya/xarela-backend/internal/middleware"
	"github.com/AffanSurya/xarela-backend/internal/service"
	"github.com/labstack/echo/v4"
)

type fakeAccessTokenAuthenticator struct {
	user   service.AuthUser
	err    error
	called bool
	token  string
}

func (f *fakeAccessTokenAuthenticator) AuthenticateAccessToken(ctx context.Context, accessToken string) (service.AuthUser, error) {
	f.called = true
	f.token = accessToken
	if f.err != nil {
		return service.AuthUser{}, f.err
	}
	return f.user, nil
}

func TestAccessTokenAuthRejectsMissingHeader(t *testing.T) {
	e := echo.New()
	handler := appmiddleware.AccessTokenAuth(&fakeAccessTokenAuthenticator{})(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/settings", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := handler(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAccessTokenAuthSetsAuthUser(t *testing.T) {
	e := echo.New()
	fake := &fakeAccessTokenAuthenticator{user: service.AuthUser{ID: "user-1", Email: "user@example.com", FullName: "User", IsEmailVerified: true}}
	handler := appmiddleware.AccessTokenAuth(fake)(func(c echo.Context) error {
		stored := c.Get("auth_user")
		if stored == nil {
			t.Fatal("expected auth user in context")
		}
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/settings", nil)
	req.Header.Set("Authorization", "Bearer access-token")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := handler(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !fake.called || fake.token != "access-token" {
		t.Fatalf("unexpected authenticator call %#v", fake)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
