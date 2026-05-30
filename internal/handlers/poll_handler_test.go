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
	"github.com/gorilla/mux"

	"roly-poly/internal/helpers"
	"roly-poly/internal/models"
)

type mockPollRepo struct {
	findAllFunc            func(ctx context.Context, userID uuid.UUID, publicOnly bool) ([]*models.PollModel, error)
	findByIDFunc           func(ctx context.Context, id uuid.UUID) (*models.PollModel, error)
	findByIDForUserFunc    func(ctx context.Context, id uuid.UUID, requesterID uuid.UUID) (*models.PollModel, error)
	createWithOptionsFunc  func(ctx context.Context, poll *models.PollModel, options []*models.OptionModel) error
	updateFunc             func(ctx context.Context, poll *models.PollModel) error
	deleteWithOptionsFunc  func(ctx context.Context, pollID uuid.UUID) error
	ownsPollFunc           func(ctx context.Context, userId, pollId uuid.UUID) (bool, error)
	countVotesFunc         func(ctx context.Context, userID, pollID uuid.UUID) (int64, error)
	castVotesFunc          func(ctx context.Context, pollID, userID uuid.UUID, optionIDs []uuid.UUID) error
	createFunc             func(ctx context.Context, poll *models.PollModel) error
	deleteFunc             func(ctx context.Context, poll *models.PollModel) error
}

func (m *mockPollRepo) FindAll(ctx context.Context, userID uuid.UUID, publicOnly bool) ([]*models.PollModel, error) {
	return m.findAllFunc(ctx, userID, publicOnly)
}
func (m *mockPollRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.PollModel, error) {
	return m.findByIDFunc(ctx, id)
}
func (m *mockPollRepo) FindByIDForUser(ctx context.Context, id uuid.UUID, requesterID uuid.UUID) (*models.PollModel, error) {
	return m.findByIDForUserFunc(ctx, id, requesterID)
}
func (m *mockPollRepo) Create(ctx context.Context, poll *models.PollModel) error {
	return m.createFunc(ctx, poll)
}
func (m *mockPollRepo) CreateWithOptions(ctx context.Context, poll *models.PollModel, options []*models.OptionModel) error {
	return m.createWithOptionsFunc(ctx, poll, options)
}
func (m *mockPollRepo) Update(ctx context.Context, poll *models.PollModel) error {
	return m.updateFunc(ctx, poll)
}
func (m *mockPollRepo) Delete(ctx context.Context, poll *models.PollModel) error {
	return m.deleteFunc(ctx, poll)
}
func (m *mockPollRepo) DeleteWithOptions(ctx context.Context, pollID uuid.UUID) error {
	return m.deleteWithOptionsFunc(ctx, pollID)
}
func (m *mockPollRepo) CastVotes(ctx context.Context, pollID, userID uuid.UUID, optionIDs []uuid.UUID) error {
	return m.castVotesFunc(ctx, pollID, userID, optionIDs)
}
func (m *mockPollRepo) OwnsPoll(ctx context.Context, userId, pollId uuid.UUID) (bool, error) {
	return m.ownsPollFunc(ctx, userId, pollId)
}
func (m *mockPollRepo) CountVotesByUserAndPoll(ctx context.Context, userID, pollID uuid.UUID) (int64, error) {
	return m.countVotesFunc(ctx, userID, pollID)
}

type mockOptionRepo struct{}

func (m *mockOptionRepo) FindAll(ctx context.Context) ([]*models.OptionModel, error) { return nil, nil }
func (m *mockOptionRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.OptionModel, error) {
	return nil, nil
}
func (m *mockOptionRepo) CreateMany(ctx context.Context, options []*models.OptionModel) error { return nil }
func (m *mockOptionRepo) Update(ctx context.Context, poll *models.OptionModel) error          { return nil }
func (m *mockOptionRepo) Delete(ctx context.Context, poll *models.OptionModel) error          { return nil }

func setupPollRequest(method, path string, body interface{}, userID uuid.UUID) *http.Request {
	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(b))
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	req = helpers.SetUserId(req, userID)
	return req
}

func TestCreatePoll(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		body       interface{}
		mockSetup  func(repo *mockPollRepo)
		wantStatus int
		wantErr    string
	}{
		{
			name: "successful create",
			body: map[string]interface{}{
				"title":   "Test Poll",
				"options": []string{"Option A", "Option B", "Option C"},
			},
			mockSetup: func(repo *mockPollRepo) {
				repo.createWithOptionsFunc = func(ctx context.Context, poll *models.PollModel, options []*models.OptionModel) error {
					poll.ID = uuid.New()
					return nil
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "validation error - empty title",
			body: map[string]interface{}{
				"title":   "",
				"options": []string{"A"},
			},
			mockSetup:  func(repo *mockPollRepo) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    "RP-400",
		},
		{
			name: "validation error - short option",
			body: map[string]interface{}{
				"title":   "Test",
				"options": []string{"X"},
			},
			mockSetup:  func(repo *mockPollRepo) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    "RP-400",
		},
		{
			name: "create error",
			body: map[string]interface{}{
				"title":   "Test Poll",
				"options": []string{"Option A", "Option B"},
			},
			mockSetup: func(repo *mockPollRepo) {
				repo.createWithOptionsFunc = func(ctx context.Context, poll *models.PollModel, options []*models.OptionModel) error {
					return fmt.Errorf("db error")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    "RP-500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPollRepo{}
			tt.mockSetup(repo)
			handler := NewPollHandler(repo, &mockOptionRepo{})

			req := setupPollRequest(http.MethodPost, "/api/polls", tt.body, userID)
			rec := httptest.NewRecorder()

			handler.CreatePoll(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestClosePoll(t *testing.T) {
	userID := uuid.New()
	pollID := uuid.New()

	tests := []struct {
		name       string
		pollID     string
		mockSetup  func(repo *mockPollRepo)
		wantStatus int
		wantErr    string
	}{
		{
			name:   "successful close",
			pollID: pollID.String(),
			mockSetup: func(repo *mockPollRepo) {
				repo.ownsPollFunc = func(ctx context.Context, uid, pid uuid.UUID) (bool, error) {
					return true, nil
				}
				repo.updateFunc = func(ctx context.Context, poll *models.PollModel) error {
					return nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "invalid poll id",
			pollID: "not-a-uuid",
			mockSetup: func(repo *mockPollRepo) {
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    "RP-400",
		},
		{
			name:   "not owner",
			pollID: pollID.String(),
			mockSetup: func(repo *mockPollRepo) {
				repo.ownsPollFunc = func(ctx context.Context, uid, pid uuid.UUID) (bool, error) {
					return false, nil
				}
			},
			wantStatus: http.StatusForbidden,
			wantErr:    "RP-403",
		},
		{
			name:   "owns poll returns error",
			pollID: pollID.String(),
			mockSetup: func(repo *mockPollRepo) {
				repo.ownsPollFunc = func(ctx context.Context, uid, pid uuid.UUID) (bool, error) {
					return false, fmt.Errorf("db error")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    "RP-500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPollRepo{}
			tt.mockSetup(repo)
			handler := NewPollHandler(repo, &mockOptionRepo{})

			req := setupPollRequest(http.MethodPatch, "/api/polls/"+tt.pollID+"/close", nil, userID)
			req = mux.SetURLVars(req, map[string]string{"id": tt.pollID})
			rec := httptest.NewRecorder()

			handler.ClosePoll(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestGetPollById(t *testing.T) {
	userID := uuid.New()
	pollID := uuid.New()

	tests := []struct {
		name       string
		mockSetup  func(repo *mockPollRepo)
		wantStatus int
		wantErr    string
	}{
		{
			name: "successful fetch",
			mockSetup: func(repo *mockPollRepo) {
				repo.findByIDForUserFunc = func(ctx context.Context, id, uid uuid.UUID) (*models.PollModel, error) {
					return &models.PollModel{ID: id, Title: "Test", UserID: uid}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "poll not found",
			mockSetup: func(repo *mockPollRepo) {
				repo.findByIDForUserFunc = func(ctx context.Context, id, uid uuid.UUID) (*models.PollModel, error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantStatus: http.StatusNotFound,
			wantErr:    "RP-404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPollRepo{}
			tt.mockSetup(repo)
			handler := NewPollHandler(repo, &mockOptionRepo{})

			req := setupPollRequest(http.MethodGet, "/api/polls/"+pollID.String(), nil, userID)
			req = mux.SetURLVars(req, map[string]string{"id": pollID.String()})
			rec := httptest.NewRecorder()

			handler.GetPollById(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestGetPolls(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		public     string
		mockSetup  func(repo *mockPollRepo)
		wantStatus int
		wantErr    string
	}{
		{
			name:   "successful fetch own polls",
			public: "false",
			mockSetup: func(repo *mockPollRepo) {
				repo.findAllFunc = func(ctx context.Context, uid uuid.UUID, publicOnly bool) ([]*models.PollModel, error) {
					return []*models.PollModel{{ID: uuid.New(), Title: "Poll 1"}}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "successful fetch public polls",
			public: "true",
			mockSetup: func(repo *mockPollRepo) {
				repo.findAllFunc = func(ctx context.Context, uid uuid.UUID, publicOnly bool) ([]*models.PollModel, error) {
					return []*models.PollModel{{ID: uuid.New(), Title: "Public Poll", Public: true}}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "repo error",
			public: "false",
			mockSetup: func(repo *mockPollRepo) {
				repo.findAllFunc = func(ctx context.Context, uid uuid.UUID, publicOnly bool) ([]*models.PollModel, error) {
					return nil, fmt.Errorf("db error")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    "RP-500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPollRepo{}
			tt.mockSetup(repo)
			handler := NewPollHandler(repo, &mockOptionRepo{})

			req := setupPollRequest(http.MethodGet, "/api/polls?public="+tt.public, nil, userID)
			rec := httptest.NewRecorder()

			handler.GetPolls(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestUpdatePoll(t *testing.T) {
	userID := uuid.New()
	pollID := uuid.New()

	tests := []struct {
		name       string
		pollID     string
		body       interface{}
		mockSetup  func(repo *mockPollRepo)
		wantStatus int
		wantErr    string
	}{
		{
			name:   "successful update",
			pollID: pollID.String(),
			body: map[string]string{
				"title": "Updated Title",
			},
			mockSetup: func(repo *mockPollRepo) {
				repo.ownsPollFunc = func(ctx context.Context, uid, pid uuid.UUID) (bool, error) {
					return true, nil
				}
				repo.updateFunc = func(ctx context.Context, poll *models.PollModel) error {
					return nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "invalid poll id",
			pollID: "not-a-uuid",
			body: map[string]string{
				"title": "Test",
			},
			mockSetup:  func(repo *mockPollRepo) {},
			wantStatus: http.StatusBadRequest,
			wantErr:    "RP-400",
		},
		{
			name:   "not owner",
			pollID: pollID.String(),
			body: map[string]string{
				"title": "Test",
			},
			mockSetup: func(repo *mockPollRepo) {
				repo.ownsPollFunc = func(ctx context.Context, uid, pid uuid.UUID) (bool, error) {
					return false, nil
				}
			},
			wantStatus: http.StatusForbidden,
			wantErr:    "RP-403",
		},
		{
			name:   "owns poll returns error",
			pollID: pollID.String(),
			body: map[string]string{
				"title": "Test",
			},
			mockSetup: func(repo *mockPollRepo) {
				repo.ownsPollFunc = func(ctx context.Context, uid, pid uuid.UUID) (bool, error) {
					return false, fmt.Errorf("db error")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    "RP-500",
		},
		{
			name:   "update error",
			pollID: pollID.String(),
			body: map[string]string{
				"title": "Test",
			},
			mockSetup: func(repo *mockPollRepo) {
				repo.ownsPollFunc = func(ctx context.Context, uid, pid uuid.UUID) (bool, error) {
					return true, nil
				}
				repo.updateFunc = func(ctx context.Context, poll *models.PollModel) error {
					return fmt.Errorf("db error")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    "RP-500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPollRepo{}
			tt.mockSetup(repo)
			handler := NewPollHandler(repo, &mockOptionRepo{})

			req := setupPollRequest(http.MethodPatch, "/api/polls/"+tt.pollID, tt.body, userID)
			req = mux.SetURLVars(req, map[string]string{"id": tt.pollID})
			rec := httptest.NewRecorder()

			handler.UpdatePoll(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestDeletePoll(t *testing.T) {
	userID := uuid.New()
	pollID := uuid.New()

	tests := []struct {
		name       string
		pollID     string
		mockSetup  func(repo *mockPollRepo)
		wantStatus int
		wantErr    string
	}{
		{
			name:   "successful delete",
			pollID: pollID.String(),
			mockSetup: func(repo *mockPollRepo) {
				repo.ownsPollFunc = func(ctx context.Context, uid, pid uuid.UUID) (bool, error) {
					return true, nil
				}
				repo.deleteWithOptionsFunc = func(ctx context.Context, pid uuid.UUID) error {
					return nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:   "invalid poll id",
			pollID: "not-a-uuid",
			mockSetup: func(repo *mockPollRepo) {
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    "RP-400",
		},
		{
			name:   "not owner",
			pollID: pollID.String(),
			mockSetup: func(repo *mockPollRepo) {
				repo.ownsPollFunc = func(ctx context.Context, uid, pid uuid.UUID) (bool, error) {
					return false, nil
				}
			},
			wantStatus: http.StatusForbidden,
			wantErr:    "RP-403",
		},
		{
			name:   "owns poll returns error",
			pollID: pollID.String(),
			mockSetup: func(repo *mockPollRepo) {
				repo.ownsPollFunc = func(ctx context.Context, uid, pid uuid.UUID) (bool, error) {
					return false, fmt.Errorf("db error")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    "RP-500",
		},
		{
			name:   "delete error",
			pollID: pollID.String(),
			mockSetup: func(repo *mockPollRepo) {
				repo.ownsPollFunc = func(ctx context.Context, uid, pid uuid.UUID) (bool, error) {
					return true, nil
				}
				repo.deleteWithOptionsFunc = func(ctx context.Context, pid uuid.UUID) error {
					return fmt.Errorf("db error")
				}
			},
			wantStatus: http.StatusInternalServerError,
			wantErr:    "RP-500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPollRepo{}
			tt.mockSetup(repo)
			handler := NewPollHandler(repo, &mockOptionRepo{})

			req := setupPollRequest(http.MethodDelete, "/api/polls/"+tt.pollID, nil, userID)
			req = mux.SetURLVars(req, map[string]string{"id": tt.pollID})
			rec := httptest.NewRecorder()

			handler.DeletePoll(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestGetPollReport(t *testing.T) {
	userID := uuid.New()
	pollID := uuid.New()
	opt1ID := uuid.New()
	opt2ID := uuid.New()

	tests := []struct {
		name       string
		mockSetup  func(repo *mockPollRepo)
		wantStatus int
		wantErr    string
	}{
		{
			name: "successful report",
			mockSetup: func(repo *mockPollRepo) {
				repo.findByIDForUserFunc = func(ctx context.Context, id, uid uuid.UUID) (*models.PollModel, error) {
					return &models.PollModel{
						ID:    id,
						Title: "Test",
						Options: []models.OptionModel{
							{ID: opt1ID, Label: "A", Votes: 10},
							{ID: opt2ID, Label: "B", Votes: 20},
						},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "poll not found",
			mockSetup: func(repo *mockPollRepo) {
				repo.findByIDForUserFunc = func(ctx context.Context, id, uid uuid.UUID) (*models.PollModel, error) {
					return nil, fmt.Errorf("not found")
				}
			},
			wantStatus: http.StatusNotFound,
			wantErr:    "RP-404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockPollRepo{}
			tt.mockSetup(repo)
			handler := NewPollHandler(repo, &mockOptionRepo{})

			req := setupPollRequest(http.MethodGet, "/api/polls/"+pollID.String()+"/report", nil, userID)
			req = mux.SetURLVars(req, map[string]string{"id": pollID.String()})
			rec := httptest.NewRecorder()

			handler.GetPollReport(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}
