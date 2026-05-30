package middlewares

import (
	"net/http"
	"runtime/debug"

	"roly-poly/internal/constants"
	"roly-poly/internal/helpers"
	"roly-poly/pkg/logger"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log := logger.New()
				log.Error().
					Interface("panic", rec).
					Str("stack", string(debug.Stack())).
					Msg("Panic recovered")

				helpers.SendErrorResponse(w, "Internal server error", constants.InternalServerError, nil)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
