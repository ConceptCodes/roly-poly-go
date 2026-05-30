package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"roly-poly/internal/helpers"
	"roly-poly/internal/models"
)

func TestCastVote(t *testing.T) {
	userID := uuid.New()
	pollID := uuid.New()
	optID := uuid.New()
	closed := time.Now()

	tests := []struct {
		name        string
		pollID      string
		body        interface{}
		mockSetup   func(repo *mockPollRepo)
		wantStatus  int
		wantErrCode string
	}{
		{
			name:   "successful vote",
			pollID: pollID.String(),
			body: map[string]interface{}{
				"option_ids": []string{optID.String()},
			},
			mockSetup: func(repo *mockPollRepo) {
				repo.findByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.PollModel, error) {
					return &models.PollModel{
						ID:                 id,
						AllowMultipleVotes: true,
						Options:            []models.OptionModel{{ID: optID, Label: "A"}},
					}, nil
				}
				repo.castVotesFunc = func(ctx context.Context, pollID, uid uuid.UUID, opts []uuid.UUID) error {
					return nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "invalid poll id",
			pollID: "not-a-uuid",
			body: map[string]interface{}{
				"option_ids": []string{optID.String()},
			},
			mockSetup:   func(repo *mockPollRepo) {},
			wantStatus:  http.StatusBadRequest,
			wantErrCode: "RP-400",
		},
		{
			name:   "poll closed",
			pollID: pollID.String(),
			body: map[string]interface{}{
				"option_ids": []string{optID.String()},
			},
			mockSetup: func(repo *mockPollRepo) {
				repo.findByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.PollModel, error) {
					return &models.PollModel{
						ID:                 id,
						Closed:             &closed,
						AllowMultipleVotes: true,
						Options:            []models.OptionModel{{ID: optID, Label: "A"}},
					}, nil
				}
			},
			wantStatus:  http.StatusForbidden,
			wantErrCode: "RP-403",
		},
		{
			name:   "multiple votes not allowed",
			pollID: pollID.String(),
			body: map[string]interface{}{
				"option_ids": []string{optID.String(), uuid.New().String()},
			},
			mockSetup: func(repo *mockPollRepo) {
				repo.findByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.PollModel, error) {
					return &models.PollModel{
						ID:                 id,
						AllowMultipleVotes: false,
						Options:            []models.OptionModel{{ID: optID, Label: "A"}},
					}, nil
				}
			},
			wantStatus:  http.StatusBadRequest,
			wantErrCode: "RP-400",
		},
		{
			name:   "duplicate vote blocked",
			pollID: pollID.String(),
			body: map[string]interface{}{
				"option_ids": []string{optID.String()},
			},
			mockSetup: func(repo *mockPollRepo) {
				repo.findByIDFunc = func(ctx context.Context, id uuid.UUID) (*models.PollModel, error) {
					return &models.PollModel{
						ID:                 id,
						AllowMultipleVotes: false,
						Options:            []models.OptionModel{{ID: optID, Label: "A"}},
					}, nil
				}
				repo.countVotesFunc = func(ctx context.Context, uid, pid uuid.UUID) (int64, error) {
					return 1, nil
				}
			},
			wantStatus:  http.StatusBadRequest,
			wantErrCode: "RP-400",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPollRepo{}
			tt.mockSetup(repo)
			handler := NewVoteHandler(repo)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/polls/"+tt.pollID+"/vote", bytes.NewReader(body))
			req = mux.SetURLVars(req, map[string]string{"id": tt.pollID})
			req = helpers.SetUserId(req, userID)
			rec := httptest.NewRecorder()

			handler.CastVote(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
