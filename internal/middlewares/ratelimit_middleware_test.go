package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"roly-poly/internal/constants"
	"roly-poly/internal/models"
)

type mockAllowChecker struct {
	allowFunc func(ctx context.Context, ip string) (bool, error)
}

func (m *mockAllowChecker) Allow(ctx context.Context, ip string) (bool, error) {
	return m.allowFunc(ctx, ip)
}

func TestRateLimitMiddlewareAllowed(t *testing.T) {
	checker := &mockAllowChecker{
		allowFunc: func(ctx context.Context, ip string) (bool, error) {
			return true, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	var called bool
	NewRateLimitMiddleware(checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestRateLimitMiddlewareDenied(t *testing.T) {
	checker := &mockAllowChecker{
		allowFunc: func(ctx context.Context, ip string) (bool, error) {
			return false, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	NewRateLimitMiddleware(checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusTooManyRequests)
	}

	var resp models.Response
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ErrorCode != constants.TooManyRequests {
		t.Errorf("error_code = %s, want %s", resp.ErrorCode, constants.TooManyRequests)
	}
}

func TestRateLimitMiddlewareError(t *testing.T) {
	checker := &mockAllowChecker{
		allowFunc: func(ctx context.Context, ip string) (bool, error) {
			return false, fmt.Errorf("rate limit error")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	NewRateLimitMiddleware(checker)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var resp models.Response
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ErrorCode != constants.InternalServerError {
		t.Errorf("error_code = %s, want %s", resp.ErrorCode, constants.InternalServerError)
	}
}

func TestExtractIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")

	ip := extractIP(req)
	if ip != "1.2.3.4" {
		t.Errorf("ip = %q, want %q", ip, "1.2.3.4")
	}
}

func TestExtractIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "9.10.11.12")

	ip := extractIP(req)
	if ip != "9.10.11.12" {
		t.Errorf("ip = %q, want %q", ip, "9.10.11.12")
	}
}

func TestExtractIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:5678"

	ip := extractIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("ip = %q, want %q", ip, "192.168.1.1")
	}
}

func TestExtractIP_XForwardedForTakesPriority(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	req.Header.Set("X-Real-IP", "10.0.0.2")
	req.RemoteAddr = "10.0.0.3:9999"

	ip := extractIP(req)
	if ip != "10.0.0.1" {
		t.Errorf("ip = %q, want %q", ip, "10.0.0.1")
	}
}

func TestExtractIP_NoPortInRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1"

	ip := extractIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("ip = %q, want %q", ip, "192.168.1.1")
	}
}
