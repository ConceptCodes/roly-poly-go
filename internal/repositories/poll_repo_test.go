package repository

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"roly-poly/internal/models"
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

func TestGormPollRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	pollID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "found",
			wantErr: false,
		},
		{
			name:    "not found",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := &GormPollRepository{db: db}

			if !tt.wantErr {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "poll_models" WHERE "poll_models"."id" = $1 AND "poll_models"."deleted_at" IS NULL ORDER BY "poll_models"."id" LIMIT 1`)).
					WithArgs(pollID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "user_id"}).AddRow(pollID, "Test Poll", userID))

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "option_models" WHERE "option_models"."poll_id" = $1 AND "option_models"."deleted_at" IS NULL`)).
					WithArgs(pollID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "label", "poll_id"}))

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_models" WHERE "user_models"."id" = $1 AND "user_models"."deleted_at" IS NULL`)).
					WithArgs(userID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "poll_models" WHERE "poll_models"."id" = $1 AND "poll_models"."deleted_at" IS NULL ORDER BY "poll_models"."id" LIMIT 1`)).
					WithArgs(pollID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			}

			poll, err := repo.FindByID(ctx, pollID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if poll.ID != pollID {
				t.Errorf("poll.ID = %v, want %v", poll.ID, pollID)
			}
		})
	}
}

func TestGormPollRepository_FindByIDForUser(t *testing.T) {
	ctx := context.Background()
	pollID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "found",
			wantErr: false,
		},
		{
			name:    "not found",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := &GormPollRepository{db: db}

			if !tt.wantErr {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "poll_models" WHERE (id = $1 AND (public = $2 OR user_id = $3)) AND "poll_models"."deleted_at" IS NULL ORDER BY "poll_models"."id" LIMIT 1`)).
					WithArgs(pollID, true, userID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "title", "user_id"}).AddRow(pollID, "Test", userID))

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "option_models" WHERE "option_models"."poll_id" = $1 AND "option_models"."deleted_at" IS NULL`)).
					WithArgs(pollID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "label", "poll_id"}))

				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "user_models" WHERE "user_models"."id" = $1 AND "user_models"."deleted_at" IS NULL`)).
					WithArgs(userID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "poll_models" WHERE (id = $1 AND (public = $2 OR user_id = $3)) AND "poll_models"."deleted_at" IS NULL ORDER BY "poll_models"."id" LIMIT 1`)).
					WithArgs(pollID, true, userID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))
			}

			poll, err := repo.FindByIDForUser(ctx, pollID, userID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if poll.ID != pollID {
				t.Errorf("poll.ID = %v, want %v", poll.ID, pollID)
			}
		})
	}
}

func TestGormPollRepository_Update(t *testing.T) {
	ctx := context.Background()
	pollID := uuid.New()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "success",
			wantErr: false,
		},
		{
			name:    "db error",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := &GormPollRepository{db: db}

			poll := &models.PollModel{
				ID:    pollID,
				Title: "Updated Title",
			}

			mock.ExpectBegin()
			exec := mock.ExpectExec(regexp.QuoteMeta(`UPDATE "poll_models" SET "id"=$1,"updated_at"=$2,"title"=$3 WHERE id = $4 AND "poll_models"."deleted_at" IS NULL`)).
				WithArgs(pollID, sqlmock.AnyArg(), "Updated Title", pollID)

			if tt.wantErr {
				exec.WillReturnError(fmt.Errorf("update error"))
				mock.ExpectRollback()
			} else {
				exec.WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			err := repo.Update(ctx, poll)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGormPollRepository_DeleteWithOptions(t *testing.T) {
	ctx := context.Background()
	pollID := uuid.New()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "success",
			wantErr: false,
		},
		{
			name:    "db error on votes delete",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := &GormPollRepository{db: db}

			mock.ExpectBegin()

			if tt.wantErr {
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "vote_models" SET "deleted_at"=$1 WHERE poll_id = $2 AND "vote_models"."deleted_at" IS NULL`)).
					WithArgs(sqlmock.AnyArg(), pollID).
					WillReturnError(fmt.Errorf("delete votes error"))
				mock.ExpectRollback()
			} else {
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "vote_models" SET "deleted_at"=$1 WHERE poll_id = $2 AND "vote_models"."deleted_at" IS NULL`)).
					WithArgs(sqlmock.AnyArg(), pollID).
					WillReturnResult(sqlmock.NewResult(0, 0))

				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "option_models" SET "deleted_at"=$1 WHERE poll_id = $2 AND "option_models"."deleted_at" IS NULL`)).
					WithArgs(sqlmock.AnyArg(), pollID).
					WillReturnResult(sqlmock.NewResult(0, 0))

				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "poll_models" SET "deleted_at"=$1 WHERE id = $2 AND "poll_models"."deleted_at" IS NULL`)).
					WithArgs(sqlmock.AnyArg(), pollID).
					WillReturnResult(sqlmock.NewResult(0, 1))

				mock.ExpectCommit()
			}

			err := repo.DeleteWithOptions(ctx, pollID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGormPollRepository_CastVotes(t *testing.T) {
	ctx := context.Background()
	pollID := uuid.New()
	userID := uuid.New()
	optID := uuid.New()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "success",
			wantErr: false,
		},
		{
			name:    "lock error",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := setupTestDB(t)
			repo := &GormPollRepository{db: db}

			mock.ExpectBegin()

			if tt.wantErr {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "option_models" WHERE id = $1 AND "option_models"."deleted_at" IS NULL ORDER BY "option_models"."id" LIMIT 1 FOR UPDATE`)).
					WithArgs(optID).
					WillReturnError(fmt.Errorf("lock error"))
				mock.ExpectRollback()
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "option_models" WHERE id = $1 AND "option_models"."deleted_at" IS NULL ORDER BY "option_models"."id" LIMIT 1 FOR UPDATE`)).
					WithArgs(optID).
					WillReturnRows(sqlmock.NewRows([]string{"id", "label", "poll_id"}).AddRow(optID, "Option A", pollID))

				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "vote_models" ("created_at","updated_at","deleted_at","user_id","option_id","poll_id") VALUES ($1,$2,$3,$4,$5,$6) RETURNING "id"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))

				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "option_models" SET "votes"=votes + 1 WHERE id = $1 AND "option_models"."deleted_at" IS NULL`)).
					WithArgs(optID).
					WillReturnResult(sqlmock.NewResult(0, 1))

				mock.ExpectCommit()
			}

			err := repo.CastVotes(ctx, pollID, userID, []uuid.UUID{optID})

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGormPollRepository_Create(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

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
			repo := &GormPollRepository{db: db}

			poll := &models.PollModel{
				Title:  "Test Poll",
				UserID: userID,
			}

			mock.ExpectBegin()

			if tt.wantErr {
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "poll_models" ("created_at","updated_at","deleted_at","title","description","closed","public","allow_multiple_votes","user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "id"`)).
					WillReturnError(fmt.Errorf("insert error"))
				mock.ExpectRollback()
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "poll_models" ("created_at","updated_at","deleted_at","title","description","closed","public","allow_multiple_votes","user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "id"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
				mock.ExpectCommit()
			}

			err := repo.Create(ctx, poll)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGormPollRepository_CreateWithOptions(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	pollID := uuid.New()

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
			repo := &GormPollRepository{db: db}

			poll := &models.PollModel{
				Title:  "Test Poll",
				UserID: userID,
			}
			options := []*models.OptionModel{
				{Label: "Option A"},
				{Label: "Option B"},
			}

			mock.ExpectBegin()

			if tt.wantErr {
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "poll_models" ("created_at","updated_at","deleted_at","title","description","closed","public","allow_multiple_votes","user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "id"`)).
					WillReturnError(fmt.Errorf("insert error"))
				mock.ExpectRollback()
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "poll_models" ("created_at","updated_at","deleted_at","title","description","closed","public","allow_multiple_votes","user_id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "id"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pollID.String()))

				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "option_models" ("created_at","updated_at","deleted_at","label","poll_id","votes") VALUES ($1,$2,$3,$4,$5,$6),($7,$8,$9,$10,$11,$12) RETURNING "id"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()).AddRow(uuid.New().String()))

				mock.ExpectCommit()
			}

			err := repo.CreateWithOptions(ctx, poll, options)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGormPollRepository_Delete(t *testing.T) {
	ctx := context.Background()
	pollID := uuid.New()

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
			repo := &GormPollRepository{db: db}

			poll := &models.PollModel{ID: pollID}

			mock.ExpectBegin()

			if tt.wantErr {
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "poll_models" SET "deleted_at"=$1 WHERE "poll_models"."id" = $2 AND "poll_models"."deleted_at" IS NULL`)).
					WithArgs(sqlmock.AnyArg(), pollID).
					WillReturnError(fmt.Errorf("delete error"))
				mock.ExpectRollback()
			} else {
				mock.ExpectExec(regexp.QuoteMeta(`UPDATE "poll_models" SET "deleted_at"=$1 WHERE "poll_models"."id" = $2 AND "poll_models"."deleted_at" IS NULL`)).
					WithArgs(sqlmock.AnyArg(), pollID).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			}

			err := repo.Delete(ctx, poll)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
