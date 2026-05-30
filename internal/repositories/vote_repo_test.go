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

func TestGormVoteRepository_Create(t *testing.T) {
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
			repo := &GormVoteRepository{db: db}

			vote := &models.VoteModel{
				UserID:   uuid.New(),
				OptionID: uuid.New(),
				PollID:   uuid.New(),
			}

			mock.ExpectBegin()

			if tt.wantErr {
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "vote_models" ("created_at","updated_at","deleted_at","user_id","option_id","poll_id") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`)).
					WillReturnError(fmt.Errorf("insert error"))
				mock.ExpectRollback()
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "vote_models" ("created_at","updated_at","deleted_at","user_id","option_id","poll_id") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
				mock.ExpectCommit()
			}

			err := repo.Create(ctx, vote)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGormVoteRepository_FindByUserAndPoll(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	pollID := uuid.New()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{name: "found", wantErr: false},
		{name: "not found", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := &GormVoteRepository{db: db}

			if !tt.wantErr && tt.name == "found" {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "vote_models" WHERE (user_id = $1 AND poll_id = $2) AND "vote_models"."deleted_at" IS NULL`)).
					WithArgs(userID, pollID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "poll_id", "option_id"}).
						AddRow(uuid.New(), userID, pollID, uuid.New()))
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "vote_models" WHERE (user_id = $1 AND poll_id = $2) AND "vote_models"."deleted_at" IS NULL`)).
					WithArgs(userID, pollID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			}

			votes, err := repo.FindByUserAndPoll(ctx, userID, pollID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && tt.name == "found" && len(votes) != 1 {
				t.Errorf("len(votes) = %d, want 1", len(votes))
			}
			if !tt.wantErr && tt.name == "not found" && len(votes) != 0 {
				t.Errorf("len(votes) = %d, want 0", len(votes))
			}
		})
	}
}
