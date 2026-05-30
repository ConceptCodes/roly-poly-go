package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"roly-poly/internal/constants"
	"roly-poly/internal/models"
)

func TestNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()

	NotFound(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var resp models.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal("could not decode response body")
	}
	if resp.ErrorCode != constants.NotFound {
		t.Errorf("error_code = %s, want %s", resp.ErrorCode, constants.NotFound)
	}
}
