package middlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"roly-poly/internal/helpers"
)

func TestRequestLogger_CallsNextHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	var called bool
	RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if !called {
		t.Error("next handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestRequestLogger_PassesBodyThrough(t *testing.T) {
	body := []byte(`{"key":"value"}`)
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write(body)
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if rec.Body.String() != string(body) {
		t.Errorf("body = %q, want %q", rec.Body.String(), string(body))
	}
}

func TestResponseWriter_CapturesStatusCode(t *testing.T) {
	rw := &responseWriter{ResponseWriter: httptest.NewRecorder(), body: &bytes.Buffer{}}

	rw.WriteHeader(http.StatusNotFound)
	if rw.statusCode != http.StatusNotFound {
		t.Errorf("statusCode = %d, want %d", rw.statusCode, http.StatusNotFound)
	}
}

func TestRequestLogger_HandlesContextValues(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = helpers.SetRequestId(req, "test-request-id")
	req = helpers.SetApiKey(req, "rp_testkey")
	req = helpers.SetUserId(req, uuid.New())
	rec := httptest.NewRecorder()

	var called bool
	RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if got := helpers.GetRequestId(r); got != "test-request-id" {
			t.Errorf("request_id = %q, want %q", got, "test-request-id")
		}
		if got := helpers.GetApiKey(r); got != "rp_testkey" {
			t.Errorf("api_key = %q, want %q", got, "rp_testkey")
		}
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if !called {
		t.Error("next handler was not called")
	}
}
