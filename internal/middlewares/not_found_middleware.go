package middlewares

import (
	"net/http"

	"roly-poly/internal/constants"
	"roly-poly/internal/helpers"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	helpers.SendErrorResponse(w, "Not Found", constants.NotFound, nil)
}
