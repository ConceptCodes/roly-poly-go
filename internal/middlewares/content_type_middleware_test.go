package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContentTypeJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	ContentTypeJSON(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	want := "application/json;charset=utf8"
	if ct != want {
		t.Errorf("Content-Type = %q, want %q", ct, want)
	}
}
