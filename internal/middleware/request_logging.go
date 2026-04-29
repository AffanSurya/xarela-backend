package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

func RequestLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			startedAt := time.Now()
			err := next(c)

			request := c.Request()
			response := c.Response()
			fields := []any{
				"method", request.Method,
				"path", c.Path(),
				"status", response.Status,
				"latency_ms", time.Since(startedAt).Milliseconds(),
				"remote_ip", c.RealIP(),
				"user_agent", request.UserAgent(),
			}

			if err != nil {
				fields = append(fields, "error", err.Error())
				logger.Error("request completed with error", fields...)
				return err
			}

			logger.Info("request completed", fields...)
			return nil
		}
	}
}
