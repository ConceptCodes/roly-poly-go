//go:build integration

package repository

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"roly-poly/internal/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("roly_poly_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*1000*1000*1000),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %v", err)
		}
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}

	migrationsDir, err := filepath.Abs("../../infra/migrations")
	if err != nil {
		t.Fatalf("failed to get migrations dir: %v", err)
	}

	if err := goose.Up(sqlDB, migrationsDir); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

func TestIntegrationCreateUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormUserRepository(db)

	user := &models.UserModel{
		FirstName: "Ada",
		LastName:  "Lovelace",
	}

	err := repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if user.ID == uuid.Nil {
		t.Error("user ID should not be nil after creation")
	}
	if user.ApiKey == "" {
		t.Error("API key should not be empty after creation")
	}

	found, err := repo.FindByApiKey(context.Background(), user.ApiKey)
	if err != nil {
		t.Fatalf("FindByApiKey: %v", err)
	}
	if found.ID != user.ID {
		t.Errorf("found user ID = %v, want %v", found.ID, user.ID)
	}
	if found.FirstName != "Ada" {
		t.Errorf("first name = %q, want %q", found.FirstName, "Ada")
	}
}

func TestIntegrationCreateUserDuplicateApiKey(t *testing.T) {
	db := setupTestDB(t)
	repo := NewGormUserRepository(db)

	user := &models.UserModel{
		FirstName: "Ada",
		LastName:  "Lovelace",
	}

	if err := repo.Create(context.Background(), user); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	duplicate := &models.UserModel{
		FirstName: "Charles",
		LastName:  "Babbage",
	}

	if err := repo.Create(context.Background(), duplicate); err != nil {
		t.Fatalf("second Create: %v", err)
	}

	if user.ApiKey == duplicate.ApiKey {
		t.Error("duplicate user got the same API key")
	}
}

func TestIntegrationCreatePollWithOptions(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewGormUserRepository(db)
	pollRepo := NewGormPollRepository(db)
	optionRepo := NewGormOptionRepository(db)

	user := &models.UserModel{FirstName: "Ada", LastName: "Lovelace"}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	poll := &models.PollModel{
		Title:  "Test Poll",
		Public: true,
		UserID: user.ID,
	}
	options := []*models.OptionModel{
		{Label: "Option A"},
		{Label: "Option B"},
		{Label: "Option C"},
	}

	err := pollRepo.CreateWithOptions(context.Background(), poll, options)
	if err != nil {
		t.Fatalf("CreateWithOptions: %v", err)
	}

	if poll.ID == uuid.Nil {
		t.Error("poll ID should not be nil")
	}

	for _, opt := range options {
		if opt.ID == uuid.Nil {
			t.Error("option ID should not be nil")
		}
		if opt.PollID != poll.ID {
			t.Errorf("option poll_id = %v, want %v", opt.PollID, poll.ID)
		}
	}

	found, err := optionRepo.FindByPollID(context.Background(), poll.ID)
	if err != nil {
		t.Fatalf("FindByPollID: %v", err)
	}
	if len(found) != 3 {
		t.Fatalf("len(options) = %d, want 3", len(found))
	}

	labels := make(map[string]bool)
	for _, opt := range found {
		labels[opt.Label] = true
	}
	if !labels["Option A"] || !labels["Option B"] || !labels["Option C"] {
		t.Error("missing expected option labels")
	}
}

func TestIntegrationCastVote(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewGormUserRepository(db)
	pollRepo := NewGormPollRepository(db)

	voter := &models.UserModel{FirstName: "Ada", LastName: "Lovelace"}
	if err := userRepo.Create(context.Background(), voter); err != nil {
		t.Fatalf("create voter: %v", err)
	}

	poll := &models.PollModel{
		Title:  "Vote Test",
		Public: false,
		UserID: voter.ID,
	}
	options := []*models.OptionModel{
		{Label: "A"},
		{Label: "B"},
	}
	if err := pollRepo.CreateWithOptions(context.Background(), poll, options); err != nil {
		t.Fatalf("create poll: %v", err)
	}

	loaded, err := pollRepo.FindByID(context.Background(), poll.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}

	optionIDs := []uuid.UUID{loaded.Options[0].ID}
	if err := pollRepo.CastVotes(context.Background(), poll.ID, voter.ID, optionIDs); err != nil {
		t.Fatalf("CastVotes: %v", err)
	}

	count, err := pollRepo.CountVotesByUserAndPoll(context.Background(), voter.ID, poll.ID)
	if err != nil {
		t.Fatalf("CountVotesByUserAndPoll: %v", err)
	}
	if count != 1 {
		t.Errorf("vote count = %d, want 1", count)
	}

	err = pollRepo.CastVotes(context.Background(), poll.ID, voter.ID, optionIDs)
	if err == nil {
		t.Error("expected duplicate vote to fail but it succeeded")
	}
}

func TestIntegrationDeletePollCascade(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewGormUserRepository(db)
	pollRepo := NewGormPollRepository(db)

	user := &models.UserModel{FirstName: "Ada", LastName: "Lovelace"}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	poll := &models.PollModel{
		Title:  "Delete Test",
		Public: true,
		UserID: user.ID,
	}
	options := []*models.OptionModel{
		{Label: "X"},
	}
	if err := pollRepo.CreateWithOptions(context.Background(), poll, options); err != nil {
		t.Fatalf("create poll: %v", err)
	}

	if err := pollRepo.DeleteWithOptions(context.Background(), poll.ID); err != nil {
		t.Fatalf("DeleteWithOptions: %v", err)
	}

	_, err := pollRepo.FindByID(context.Background(), poll.ID)
	if err == nil {
		t.Error("expected error fetching deleted poll, got nil")
	}
}

func TestIntegrationVoteReport(t *testing.T) {
	db := setupTestDB(t)
	userRepo := NewGormUserRepository(db)
	pollRepo := NewGormPollRepository(db)

	user := &models.UserModel{FirstName: "Ada", LastName: "Lovelace"}
	if err := userRepo.Create(context.Background(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	poll := &models.PollModel{
		Title:  "Report Test",
		Public: true,
		UserID: user.ID,
	}
	options := []*models.OptionModel{
		{Label: "Option 1"},
		{Label: "Option 2"},
	}
	if err := pollRepo.CreateWithOptions(context.Background(), poll, options); err != nil {
		t.Fatalf("create poll: %v", err)
	}

	loaded, err := pollRepo.FindByID(context.Background(), poll.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}

	option1 := loaded.Options[0].ID

	otherUser := &models.UserModel{FirstName: "Charles", LastName: "Babbage"}
	if err := userRepo.Create(context.Background(), otherUser); err != nil {
		t.Fatalf("create other user: %v", err)
	}

	if err := pollRepo.CastVotes(context.Background(), poll.ID, user.ID, []uuid.UUID{option1}); err != nil {
		t.Fatalf("first vote: %v", err)
	}
	if err := pollRepo.CastVotes(context.Background(), poll.ID, otherUser.ID, []uuid.UUID{option1}); err != nil {
		t.Fatalf("second vote: %v", err)
	}

	loadedAgain, err := pollRepo.FindByIDForUser(context.Background(), poll.ID, user.ID)
	if err != nil {
		t.Fatalf("FindByIDForUser: %v", err)
	}

	for _, opt := range loadedAgain.Options {
		fmt.Printf("option %s: %d votes\n", opt.Label, opt.Votes)
	}

	if len(loadedAgain.Options) != 2 {
		t.Fatalf("len(options) = %d, want 2", len(loadedAgain.Options))
	}

	for _, opt := range loadedAgain.Options {
		switch opt.Label {
		case "Option 1":
			if opt.Votes != 2 {
				t.Errorf("Option 1 votes = %d, want 2", opt.Votes)
			}
		case "Option 2":
			if opt.Votes != 0 {
				t.Errorf("Option 2 votes = %d, want 0", opt.Votes)
			}
		}
	}
}
