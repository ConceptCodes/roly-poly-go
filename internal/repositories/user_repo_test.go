package repository

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"roly-poly/internal/models"
)

func TestGormUserRepository_FindByApiKey(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "found", wantErr: false},
		{name: "not found", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := &GormUserRepository{db: db}

			if !tt.wantErr {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_models" WHERE api_key = $1 AND "user_models"."deleted_at" IS NULL ORDER BY "user_models"."id" LIMIT 1`)).
					WithArgs("rp_testkey").
					WillReturnRows(sqlmock.NewRows([]string{"id", "api_key"}).AddRow(userID, "rp_testkey"))
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_models" WHERE api_key = $1 AND "user_models"."deleted_at" IS NULL ORDER BY "user_models"."id" LIMIT 1`)).
					WithArgs("rp_testkey").
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			}

			user, err := repo.FindByApiKey(ctx, "rp_testkey")

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && user.ID != userID {
				t.Errorf("user.ID = %v, want %v", user.ID, userID)
			}
		})
	}
}

func TestGormUserRepository_Create(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "success", wantErr: false},
		{name: "db error", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := &GormUserRepository{db: db}

			user := &models.UserModel{
				FirstName: "John",
				LastName:  "Doe",
			}

			mock.ExpectBegin()

			if tt.wantErr {
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "user_models" ("created_at","updated_at","deleted_at","api_key","enabled","first_name","last_name") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`)).
					WillReturnError(fmt.Errorf("insert error"))
				mock.ExpectRollback()
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "user_models" ("created_at","updated_at","deleted_at","api_key","enabled","first_name","last_name") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
				mock.ExpectCommit()
			}

			err := repo.Create(ctx, user)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && user.ApiKey == "" {
				t.Error("expected api_key to be set")
			}
		})
	}
}
