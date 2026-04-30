package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

type rateLimitState struct {
	windowStart time.Time
	count       int
}

type rateLimiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	seen   map[string]*rateLimitState
}

func RateLimit(limit int, window time.Duration) echo.MiddlewareFunc {
	if limit <= 0 || window <= 0 {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}

	limiter := &rateLimiter{
		limit:  limit,
		window: window,
		seen:   make(map[string]*rateLimitState),
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !limiter.allow(c) {
				return c.JSON(http.StatusTooManyRequests, map[string]any{
					"error": map[string]any{
						"code":    "rate_limit_exceeded",
						"message": "Too many requests",
					},
				})
			}
			return next(c)
		}
	}
}

func (l *rateLimiter) allow(c echo.Context) bool {
	key := c.RealIP() + "|" + c.Request().Method + "|" + requestPath(c)
	now := time.Now()

	l.mu.Lock()
	defer l.mu.Unlock()

	state, ok := l.seen[key]
	if !ok || now.Sub(state.windowStart) >= l.window {
		l.seen[key] = &rateLimitState{windowStart: now, count: 1}
		return true
	}

	if state.count >= l.limit {
		return false
	}

	state.count++
	return true
}

func requestPath(c echo.Context) string {
	if path := c.Path(); path != "" {
		return path
	}
	return c.Request().URL.Path
}
