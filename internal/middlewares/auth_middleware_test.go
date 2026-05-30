package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"roly-poly/internal/constants"
	"roly-poly/internal/helpers"
	"roly-poly/internal/models"
)

type mockUserRepo struct {
	findByApiKeyFunc func(ctx context.Context, apiKey string) (*models.UserModel, error)
}

func (m *mockUserRepo) Create(ctx context.Context, user *models.UserModel) error { return nil }
func (m *mockUserRepo) FindByApiKey(ctx context.Context, apiKey string) (*models.UserModel, error) {
	return m.findByApiKeyFunc(ctx, apiKey)
}

func TestAuthMiddlewareWhitelist(t *testing.T) {
	paths := []string{
		constants.HealthCheckEndpoint,
		constants.ReadinessEndpoint,
		constants.OnboardUserEndpoint,
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			var called bool
			middleware := NewAuthMiddleware(nil)
			middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
			})).ServeHTTP(rec, req)

			if !called {
				t.Errorf("whitelisted path %s was blocked", path)
			}
			if rec.Code != http.StatusOK {
				t.Errorf("status = %d for whitelisted path %s", rec.Code, path)
			}
		})
	}
}

func TestAuthMiddlewareMissingKey(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/polls", nil)
	rec := httptest.NewRecorder()

	middleware := NewAuthMiddleware(nil)
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var resp models.Response
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ErrorCode != "RP-401" {
		t.Errorf("error_code = %s, want RP-401", resp.ErrorCode)
	}
}

func TestAuthMiddlewareValidKeyUserEnabled(t *testing.T) {
	userID := uuid.New()
	repo := &mockUserRepo{
		findByApiKeyFunc: func(ctx context.Context, apiKey string) (*models.UserModel, error) {
			return &models.UserModel{ID: userID, Enabled: true}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/polls", nil)
	req.Header.Set(constants.AuthorizationHeader, "rp_validkey")
	rec := httptest.NewRecorder()

	var nextCalled bool
	middleware := NewAuthMiddleware(repo)
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		if got := helpers.GetUserId(r); got != userID {
			t.Errorf("user ID in context = %v, want %v", got, userID)
		}
		if got := helpers.GetApiKey(r); got != "rp_validkey" {
			t.Errorf("api key in context = %q, want %q", got, "rp_validkey")
		}
	})).ServeHTTP(rec, req)

	if !nextCalled {
		t.Error("next handler was not called for valid key")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAuthMiddlewareUserDisabled(t *testing.T) {
	repo := &mockUserRepo{
		findByApiKeyFunc: func(ctx context.Context, apiKey string) (*models.UserModel, error) {
			return &models.UserModel{Enabled: false}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/polls", nil)
	req.Header.Set(constants.AuthorizationHeader, "rp_disabledkey")
	rec := httptest.NewRecorder()

	middleware := NewAuthMiddleware(repo)
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for disabled user")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}

	var resp models.Response
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.ErrorCode != constants.Forbidden {
		t.Errorf("error_code = %s, want %s", resp.ErrorCode, constants.Forbidden)
	}
}

func TestAuthMiddlewareDBError(t *testing.T) {
	repo := &mockUserRepo{
		findByApiKeyFunc: func(ctx context.Context, apiKey string) (*models.UserModel, error) {
			return nil, fmt.Errorf("db error")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/polls", nil)
	req.Header.Set(constants.AuthorizationHeader, "rp_key")
	rec := httptest.NewRecorder()

	middleware := NewAuthMiddleware(repo)
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called on db error")
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
