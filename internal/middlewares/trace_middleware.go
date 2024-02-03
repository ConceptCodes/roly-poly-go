package middlewares

import (
	"net/http"

	"github.com/google/uuid"

	"roly-poly/internal/constants"
	"roly-poly/internal/helpers"
)

func TraceRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := r.Header.Get(constants.TraceIdHeader)

		if requestId == "" {
			requestId = uuid.New().String()
		}

		r = helpers.SetRequestId(r, requestId)

		w.Header().Add(constants.TraceIdHeader, requestId)
		next.ServeHTTP(w, r)
	})
}
