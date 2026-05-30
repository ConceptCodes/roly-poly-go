package middlewares

import (
	"context"
	"net/http"
	"strings"

	"roly-poly/internal/constants"
	"roly-poly/internal/helpers"
)

type AllowChecker interface {
	Allow(ctx context.Context, ip string) (bool, error)
}

func NewRateLimitMiddleware(limiter AllowChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r)

			allowed, err := limiter.Allow(r.Context(), ip)
			if err != nil {
				helpers.SendErrorResponse(w, "Rate limit check failed", constants.InternalServerError, err)
				return
			}

			if !allowed {
				helpers.SendErrorResponse(w, "Too many requests. Please try again later", constants.TooManyRequests, nil)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func extractIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.SplitN(fwd, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	addr := r.RemoteAddr
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}
