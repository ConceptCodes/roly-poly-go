package helpers

import (
	"encoding/json"
	"net/http"

	"roly-poly/internal/constants"
	"roly-poly/internal/models"
	"roly-poly/pkg/logger"
)

func SendSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	response := models.Response{
		Message: message,
		Data:    data,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func SendErrorResponse(w http.ResponseWriter, message string, errorCode string, err error) {
	log := logger.New()
	log.
		Error().
		Str("error_code", errorCode).
		Err(err).
		Msg(message)

	response := models.Response{
		Message:   message,
		ErrorCode: errorCode,
	}

	switch errorCode {
	case constants.NotFound:
		w.WriteHeader(http.StatusNotFound)
	case constants.Unauthorized:
		w.WriteHeader(http.StatusUnauthorized)
	case constants.InternalServerError:
		w.WriteHeader(http.StatusInternalServerError)
	case constants.Forbidden:
		w.WriteHeader(http.StatusForbidden)
	case constants.BadRequest:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(response)
}
