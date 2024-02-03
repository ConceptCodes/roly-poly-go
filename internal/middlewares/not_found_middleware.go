package middlewares

import (
	"net/http"

	"roly-poly/internal/constants"
	"roly-poly/internal/helpers"
)

func NotFound(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		helpers.SendErrorResponse(w, "Not Found", constants.NotFound, nil)
		return
	})
}
