package middlewares

import (
	"net/http"
	"regexp"

	"github.com/google/uuid"

	"roly-poly/internal/constants"
	"roly-poly/internal/helpers"
)

var validTraceID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

func TraceRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := r.Header.Get(constants.TraceIdHeader)

		if requestId == "" || !validTraceID.MatchString(requestId) {
			requestId = uuid.New().String()
		}

		r = helpers.SetRequestId(r, requestId)

		w.Header().Add(constants.TraceIdHeader, requestId)
		next.ServeHTTP(w, r)
	})
}
