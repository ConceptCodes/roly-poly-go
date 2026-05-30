package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"roly-poly/internal/models"
)

type mockUserRepo struct {
	createFunc      func(ctx context.Context, user *models.UserModel) error
	findByApiKeyFunc func(ctx context.Context, apiKey string) (*models.UserModel, error)
}

func (m *mockUserRepo) Create(ctx context.Context, user *models.UserModel) error {
	return m.createFunc(ctx, user)
}

func (m *mockUserRepo) FindByApiKey(ctx context.Context, apiKey string) (*models.UserModel, error) {
	return m.findByApiKeyFunc(ctx, apiKey)
}

func TestOnboardUser(t *testing.T) {
	tests := []struct {
		name       string
		body       interface{}
		createErr  error
		wantStatus int
		wantErr    string
	}{
		{
			name: "successful onboard",
			body: map[string]string{
				"first_name": "John",
				"last_name":  "Doe",
			},
			wantStatus: http.StatusCreated,
			wantErr:    "",
		},
		{
			name:       "missing body fields",
			body:       map[string]string{},
			wantStatus: http.StatusBadRequest,
			wantErr:    "RP-400",
		},
		{
			name: "create error",
			body: map[string]string{
				"first_name": "John",
				"last_name":  "Doe",
			},
			createErr:  fmt.Errorf("db error"),
			wantStatus: http.StatusInternalServerError,
			wantErr:    "RP-500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{
				createFunc: func(ctx context.Context, user *models.UserModel) error {
					user.ID = uuid.New()
					user.ApiKey = "rp_testkey123"
					return tt.createErr
				},
			}

			handler := NewAdminHandler(repo)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/onboard", bytes.NewReader(body))
			rec := httptest.NewRecorder()

			handler.OnboardUser(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			var resp models.Response
			json.NewDecoder(rec.Body).Decode(&resp)
			if resp.ErrorCode != tt.wantErr {
				t.Logf("got error_code = %q, want %q (message: %s)", resp.ErrorCode, tt.wantErr, resp.Message)
			}

			if tt.wantStatus == http.StatusCreated {
				data, ok := resp.Data.(map[string]interface{})
				if !ok {
					t.Fatal("response data is not an object")
				}
				if data["api_key"] == nil {
					t.Error("api_key should not be nil in creation response")
				}
			}
		})
	}
}
