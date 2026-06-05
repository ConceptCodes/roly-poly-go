package middlewares

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"roly-poly/internal/constants"
	"roly-poly/internal/helpers"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusRecorder) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func Tracing(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tracer := otel.Tracer(serviceName)
			ctx, span := tracer.Start(r.Context(), r.Method+" "+r.URL.Path,
				trace.WithAttributes(
					attribute.String("http.method", r.Method),
					attribute.String("http.url", r.URL.String()),
					attribute.String("http.user_agent", r.UserAgent()),
				),
			)
			defer span.End()

			requestId := helpers.GetRequestId(r)
			if requestId != "" {
				span.SetAttributes(attribute.String(constants.RequestIdCtxKey, requestId))
			}

			sw := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			r = r.WithContext(ctx)

			next.ServeHTTP(sw, r)

			route := mux.CurrentRoute(r)
			if route != nil {
				if tmpl, err := route.GetPathTemplate(); err == nil {
					span.SetAttributes(attribute.String("http.route", tmpl))
				}
			}

			span.SetAttributes(attribute.Int("http.status_code", sw.statusCode))
		})
	}
}
