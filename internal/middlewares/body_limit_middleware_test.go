package middlewares

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBodyLimit_UnderLimit(t *testing.T) {
	body := strings.Repeat("a", 100)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()

	var called bool
	BodyLimit(1024)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		read, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("error reading body: %v", err)
		}
		if string(read) != body {
			t.Errorf("body = %q, want %q", string(read), body)
		}
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if !called {
		t.Error("next handler was not called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestBodyLimit_OverLimit(t *testing.T) {
	body := strings.Repeat("a", 200)
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	rec := httptest.NewRecorder()

	var called bool
	BodyLimit(100)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		_, err := io.ReadAll(r.Body)
		if err == nil {
			t.Error("expected error reading oversized body, got nil")
		}
	})).ServeHTTP(rec, req)

	if !called {
		t.Error("next handler was not called")
	}
}
