package helpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"roly-poly/internal/constants"
	"roly-poly/internal/models"
)

func TestSendCreatedResponse(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}
	SendCreatedResponse(w, "created", data)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}

	var resp models.Response
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Message != "created" {
		t.Errorf("message = %q, want %q", resp.Message, "created")
	}
	if resp.ErrorCode != "" {
		t.Errorf("error_code = %q, want empty", resp.ErrorCode)
	}
}

func TestSendSuccessResponse(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}
	SendSuccessResponse(w, "success", data)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp models.Response
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Message != "success" {
		t.Errorf("message = %q, want %q", resp.Message, "success")
	}
}

func TestSendErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		errorCode  string
		wantStatus int
	}{
		{"bad request", constants.BadRequest, http.StatusBadRequest},
		{"unauthorized", constants.Unauthorized, http.StatusUnauthorized},
		{"forbidden", constants.Forbidden, http.StatusForbidden},
		{"not found", constants.NotFound, http.StatusNotFound},
		{"too many requests", constants.TooManyRequests, http.StatusTooManyRequests},
		{"internal server error", constants.InternalServerError, http.StatusInternalServerError},
		{"unknown code", "RP-999", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			SendErrorResponse(w, "error msg", tt.errorCode, nil)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			var resp models.Response
			json.NewDecoder(w.Body).Decode(&resp)
			if resp.ErrorCode != tt.errorCode {
				t.Errorf("error_code = %q, want %q", resp.ErrorCode, tt.errorCode)
			}
			if resp.Message != "error msg" {
				t.Errorf("message = %q, want %q", resp.Message, "error msg")
			}
		})
	}
}

func TestSetGetRequestId(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if got := GetRequestId(req); got != "" {
		t.Errorf("GetRequestId on fresh request = %q, want %q", got, "")
	}

	req = SetRequestId(req, "test-id-123")
	if got := GetRequestId(req); got != "test-id-123" {
		t.Errorf("GetRequestId = %q, want %q", got, "test-id-123")
	}
}

func TestSetGetApiKey(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if got := GetApiKey(req); got != "" {
		t.Errorf("GetApiKey on fresh request = %q, want %q", got, "")
	}

	req = SetApiKey(req, "rp_testkey")
	if got := GetApiKey(req); got != "rp_testkey" {
		t.Errorf("GetApiKey = %q, want %q", got, "rp_testkey")
	}
}

func TestSetGetUserId(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if got := GetUserId(req); got != uuid.Nil {
		t.Errorf("GetUserId on fresh request = %v, want %v", got, uuid.Nil)
	}

	id := uuid.New()
	req = SetUserId(req, id)
	if got := GetUserId(req); got != id {
		t.Errorf("GetUserId = %v, want %v", got, id)
	}
}

func TestValidateStruct_Valid(t *testing.T) {
	w := httptest.NewRecorder()

	data := &struct {
		Name string `validate:"required,min=2"`
	}{Name: "John"}

	result := ValidateStruct(w, data)

	if !result {
		t.Error("ValidateStruct returned false, want true")
	}
	if w.Code != http.StatusOK {
		t.Errorf("response was written (status %d) for valid struct", w.Code)
	}
}

func TestValidateStruct_Invalid(t *testing.T) {
	w := httptest.NewRecorder()

	data := &struct {
		Name string `validate:"required,min=2"`
	}{Name: ""}

	result := ValidateStruct(w, data)

	if result {
		t.Error("ValidateStruct returned true, want false")
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp models.Response
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.ErrorCode != constants.BadRequest {
		t.Errorf("error_code = %q, want %q", resp.ErrorCode, constants.BadRequest)
	}
}
