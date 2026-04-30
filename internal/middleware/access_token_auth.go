package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/AffanSurya/xarela-backend/internal/service"
	"github.com/labstack/echo/v4"
)

type AccessTokenAuthenticator interface {
	AuthenticateAccessToken(ctx context.Context, accessToken string) (service.AuthUser, error)
}

func AccessTokenAuth(authenticator interface {
	AuthenticateAccessToken(ctx context.Context, accessToken string) (service.AuthUser, error)
}) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			accessToken := bearerToken(c.Request().Header.Get("Authorization"))
			if accessToken == "" {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"error": map[string]any{
						"code":    "unauthorized",
						"message": "Missing or invalid access token",
					},
				})
			}

			user, err := authenticator.AuthenticateAccessToken(c.Request().Context(), accessToken)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"error": map[string]any{
						"code":    "unauthorized",
						"message": "Missing or invalid access token",
					},
				})
			}

			c.Set("auth_user", user)
			return next(c)
		}
	}
}

func bearerToken(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(trimmed, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
}
