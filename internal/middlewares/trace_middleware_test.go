package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"roly-poly/internal/constants"
	"roly-poly/internal/helpers"
)

func TestTraceRequest_WithExistingHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(constants.TraceIdHeader, "existing-trace-id")
	rec := httptest.NewRecorder()

	TraceRequest(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := helpers.GetRequestId(r); got != "existing-trace-id" {
			t.Errorf("request ID in context = %q, want %q", got, "existing-trace-id")
		}
	})).ServeHTTP(rec, req)

	if got := rec.Header().Get(constants.TraceIdHeader); got != "existing-trace-id" {
		t.Errorf("response header = %q, want %q", got, "existing-trace-id")
	}
}

func TestTraceRequest_WithoutHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	TraceRequest(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestId := helpers.GetRequestId(r)
		if requestId == "" {
			t.Error("request ID in context is empty, expected a generated UUID")
		}
	})).ServeHTTP(rec, req)

	responseId := rec.Header().Get(constants.TraceIdHeader)
	if responseId == "" {
		t.Error("response header x-trace-id is empty, expected a UUID")
	}
}
