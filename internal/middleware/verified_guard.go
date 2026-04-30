package middleware

import (
	"net/http"

	"github.com/AffanSurya/xarela-backend/internal/service"
	"github.com/labstack/echo/v4"
)

func RequireVerifiedUser() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			value := c.Get("auth_user")
			if value == nil {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"error": map[string]any{
						"code":    "unauthorized",
						"message": "Authentication required",
					},
				})
			}

			verified, ok := isVerifiedAuthUser(value)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]any{
					"error": map[string]any{
						"code":    "unauthorized",
						"message": "Authentication required",
					},
				})
			}

			if !verified {
				return c.JSON(http.StatusForbidden, map[string]any{
					"error": map[string]any{
						"code":    "email_not_verified",
						"message": "Email verification required",
					},
				})
			}

			return next(c)
		}
	}
}

func isVerifiedAuthUser(value any) (bool, bool) {
	switch user := value.(type) {
	case service.AuthUser:
		return user.IsEmailVerified, true
	case *service.AuthUser:
		if user == nil {
			return false, false
		}
		return user.IsEmailVerified, true
	default:
		return false, false
	}
}
