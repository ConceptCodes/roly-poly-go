package repository

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm: %v", err)
	}

	return gormDB, mock
}

func TestGormPollRepository_OwnsPoll(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	pollID := uuid.New()

	tests := []struct {
		name     string
		owns     bool
		wantErr  bool
	}{
		{
			name: "user owns poll",
			owns: true,
		},
		{
			name: "user does not own poll",
			owns: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := &GormPollRepository{db: db}

			mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "poll_models" WHERE (user_id = $1 AND id = $2) AND "poll_models"."deleted_at" IS NULL`)).
				WithArgs(userID, pollID).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(map[bool]int64{true: 1, false: 0}[tt.owns]))

			owns, err := repo.OwnsPoll(ctx, userID, pollID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if owns != tt.owns {
				t.Errorf("owns = %v, want %v", owns, tt.owns)
			}
		})
	}
}

func TestGormPollRepository_FindAll(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name       string
		publicOnly bool
	}{
		{
			name:       "find own polls",
			publicOnly: false,
		},
		{
			name:       "find public polls",
			publicOnly: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := &GormPollRepository{db: db}

			mock.ExpectQuery(`SELECT \* FROM "poll_models" WHERE .+ AND "poll_models"\."deleted_at" IS NULL`).
				WillReturnRows(sqlmock.NewRows([]string{"id", "title"}).
					AddRow(uuid.New(), "Poll 1"))

			mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "option_models" WHERE "option_models"."poll_id" = $1 AND "option_models"."deleted_at" IS NULL`)).
				WillReturnRows(sqlmock.NewRows([]string{"id", "label", "poll_id"}))

			polls, err := repo.FindAll(ctx, userID, tt.publicOnly)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if len(polls) != 1 {
				t.Errorf("len(polls) = %d, want 1", len(polls))
			}
		})
	}
}

func TestGormPollRepository_CountVotesByUserAndPoll(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	pollID := uuid.New()

	tests := []struct {
		name      string
		mockCount int64
	}{
		{
			name:      "no votes",
			mockCount: 0,
		},
		{
			name:      "has votes",
			mockCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := &GormPollRepository{db: db}

			mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "vote_models" WHERE (user_id = $1 AND poll_id = $2) AND "vote_models"."deleted_at" IS NULL`)).
				WithArgs(userID, pollID).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tt.mockCount))

			count, err := repo.CountVotesByUserAndPoll(ctx, userID, pollID)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if count != tt.mockCount {
				t.Errorf("count = %d, want %d", count, tt.mockCount)
			}
		})
	}
}
