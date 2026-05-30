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

func TestGormOptionRepository_FindAll(t *testing.T) {
	ctx := context.Background()
	optID := uuid.New()
	pollID := uuid.New()

	db, mock := setupTestDB(t)
	repo := &GormOptionRepository{db: db}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "option_models" WHERE "option_models"."deleted_at" IS NULL`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "label", "poll_id"}).
			AddRow(optID, "Option A", pollID))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "poll_models" WHERE "poll_models"."id" = $1 AND "poll_models"."deleted_at" IS NULL`)).
		WithArgs(pollID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pollID))

	options, err := repo.FindAll(ctx)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if len(options) != 1 {
		t.Errorf("len(options) = %d, want 1", len(options))
	}
}

func TestGormOptionRepository_FindByID(t *testing.T) {
	ctx := context.Background()
	optID := uuid.New()
	pollID := uuid.New()

	db, mock := setupTestDB(t)
	repo := &GormOptionRepository{db: db}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "option_models" WHERE "option_models"."id" = $1 AND "option_models"."deleted_at" IS NULL ORDER BY "option_models"."id" LIMIT 1`)).
		WithArgs(optID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "label", "poll_id"}).AddRow(optID, "Option A", pollID))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "poll_models" WHERE "poll_models"."id" = $1 AND "poll_models"."deleted_at" IS NULL`)).
		WithArgs(pollID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pollID))

	opt, err := repo.FindByID(ctx, optID)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	if opt.ID != optID {
		t.Errorf("opt.ID = %v, want %v", opt.ID, optID)
	}
}

func TestGormOptionRepository_CreateMany(t *testing.T) {
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
			repo := &GormOptionRepository{db: db}

			options := []*models.OptionModel{
				{Label: "A", PollID: pollID},
				{Label: "B", PollID: pollID},
			}

			mock.ExpectBegin()

			if tt.wantErr {
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "option_models" ("created_at","updated_at","deleted_at","label","poll_id","votes") VALUES ($1,$2,$3,$4,$5,$6),($7,$8,$9,$10,$11,$12) RETURNING "id"`)).
					WillReturnError(fmt.Errorf("insert error"))
				mock.ExpectRollback()
			} else {
				mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "option_models" ("created_at","updated_at","deleted_at","label","poll_id","votes") VALUES ($1,$2,$3,$4,$5,$6),($7,$8,$9,$10,$11,$12) RETURNING "id"`)).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()).AddRow(uuid.New().String()))
				mock.ExpectCommit()
			}

			err := repo.CreateMany(ctx, options)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGormOptionRepository_Update(t *testing.T) {
	ctx := context.Background()
	optID := uuid.New()

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
			repo := &GormOptionRepository{db: db}

			option := &models.OptionModel{ID: optID, Label: "Updated", PollID: pollID}

			mock.ExpectBegin()
			exec := mock.ExpectExec(regexp.QuoteMeta(`UPDATE "option_models" SET "id"=$1,"updated_at"=$2,"label"=$3,"poll_id"=$4 WHERE id = $5 AND "option_models"."deleted_at" IS NULL`)).
				WithArgs(optID, sqlmock.AnyArg(), "Updated", pollID, optID)

			if tt.wantErr {
				exec.WillReturnError(fmt.Errorf("update error"))
				mock.ExpectRollback()
			} else {
				exec.WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			}

			err := repo.Update(ctx, option)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestGormOptionRepository_Delete(t *testing.T) {
	ctx := context.Background()
	optID := uuid.New()

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
			repo := &GormOptionRepository{db: db}

			option := &models.OptionModel{ID: optID}

			mock.ExpectBegin()
			exec := mock.ExpectExec(regexp.QuoteMeta(`UPDATE "option_models" SET "deleted_at"=$1 WHERE "option_models"."id" = $2 AND "option_models"."deleted_at" IS NULL`)).
				WithArgs(sqlmock.AnyArg(), optID)

			if tt.wantErr {
				exec.WillReturnError(fmt.Errorf("delete error"))
				mock.ExpectRollback()
			} else {
				exec.WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			}

			err := repo.Delete(ctx, option)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
