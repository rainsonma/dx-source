package middleware

import (
	"fmt"

	"dx-api/app/helpers"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/facades"
)

// RateLimit creates a rate limiting middleware with configurable limit and window.
// It uses the sliding window algorithm via Redis sorted sets (helpers.CheckRateLimit).
// The rate limit key is built from the authenticated user ID (falling back to IP)
// combined with the HTTP method and path.
func RateLimit(limit int, windowSeconds int) contractshttp.Middleware {
	return func(ctx contractshttp.Context) {
		identifier := ""

		// Try to get the authenticated user ID from the "user" guard.
		if id, err := facades.Auth(ctx).Guard("user").ID(); err == nil && id != "" {
			identifier = id
		}

		// Fall back to client IP if no authenticated user.
		if identifier == "" {
			identifier = ctx.Request().Ip()
		}

		key := fmt.Sprintf("rate_limit:%s:%s:%s", ctx.Request().Method(), ctx.Request().Path(), identifier)

		allowed, err := helpers.CheckRateLimit(key, limit, windowSeconds)
		if err != nil {
			// Log the error but do not block the request on rate limit infrastructure failure.
			facades.Log().Errorf("rate limit check failed: %v", err)
			ctx.Request().Next()
			return
		}

		if !allowed {
			_ = ctx.Response().Json(429, helpers.Response{
				Code:    42900,
				Message: "too many requests",
			}).Abort()
			return
		}

		ctx.Request().Next()
	}
}
