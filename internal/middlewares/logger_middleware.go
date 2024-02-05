package middlewares

import (
	"bytes"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"roly-poly/internal/constants"
	"roly-poly/internal/helpers"
	"roly-poly/pkg/logger"
)

type responseWriter struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log := logger.New()

		requestId := helpers.GetRequestId(r)
		apiKey := helpers.GetApiKey(r)
		userId := helpers.GetUserId(r)

		log.UpdateContext(func(c zerolog.Context) zerolog.Context {
			return c.
				Str(constants.RequestIdCtxKey, requestId).
				Str(constants.ApiKeyCtxKey, apiKey).
				Str(constants.UserIdCtxKey, userId.String())
		})

		rw := &responseWriter{ResponseWriter: w, body: &bytes.Buffer{}}

		log.
			Info().
			Str("method", r.Method).
			Str("url", r.URL.RequestURI()).
			Str("user_agent", r.UserAgent()).
			Dur("elapsed_ms", time.Since(start)).
			Int("status_code", rw.statusCode).
			Str("remote_addr", r.RemoteAddr).
			Msgf("%s request", r.Method)

		next.ServeHTTP(w, r)

	})
}

func (lrw *responseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
