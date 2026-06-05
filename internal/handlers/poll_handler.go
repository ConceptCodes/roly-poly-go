package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"roly-poly/internal/constants"
	"roly-poly/internal/helpers"
	"roly-poly/internal/models"
	repository "roly-poly/internal/repositories"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type PollHandler struct {
	pollRepo   repository.PollRepository
	optionRepo repository.OptionRepository
}

func NewPollHandler(pollRepo repository.PollRepository, optionRepo repository.OptionRepository) *PollHandler {
	return &PollHandler{pollRepo: pollRepo, optionRepo: optionRepo}
}

// CreatePoll godoc
// @Summary      Create a poll with options
// @Description  Create a new poll owned by the authenticated user with one or more options.
// @Tags         Polls
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request  body      models.CreatePollRequestDto  true  "Poll details"
// @Success      201     {object}  models.Response  "Poll created successfully"
// @Failure      400     {object}  models.Response  "Invalid request body"
// @Failure      500     {object}  models.Response  "Error while creating poll"
// @Router       /polls [post]
func (h *PollHandler) CreatePoll(w http.ResponseWriter, r *http.Request) {
	var data models.CreatePollRequestDto

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&data)

	if err != nil {
		helpers.SendErrorResponse(w, "Invalid request body", constants.BadRequest, err)
		return
	}

	if !helpers.ValidateStruct(w, &data) {
		return
	}

	userId := helpers.GetUserId(r)

	poll := &models.PollModel{
		Title:       data.Title,
		Description: data.Description,
		Public:      data.Public,
		UserID:      userId,
	}

	var options []*models.OptionModel
	for _, option := range data.Options {
		options = append(options, &models.OptionModel{
			Label: option,
		})
	}

	err = h.pollRepo.CreateWithOptions(r.Context(), poll, options)

	if err != nil {
		helpers.SendErrorResponse(w, "Error while creating poll", constants.InternalServerError, err)
		return
	}

	helpers.SendCreatedResponse(w, "Poll created successfully", poll)
}

// GetPolls godoc
// @Summary      List polls
// @Description  List polls owned by the authenticated user. Add ?public=true to list all public polls.
// @Tags         Polls
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        public  query     bool  false  "Set to true to list public polls"
// @Success      200    {object}  models.Response  "Polls fetched successfully"
// @Failure      500    {object}  models.Response  "Error while fetching polls"
// @Router       /polls [get]
func (h *PollHandler) GetPolls(w http.ResponseWriter, r *http.Request) {
	publicOnly := r.URL.Query().Get("public") == "true"

	userId := helpers.GetUserId(r)

	polls, err := h.pollRepo.FindAll(r.Context(), userId, publicOnly)

	if err != nil {
		helpers.SendErrorResponse(w, "Error while fetching polls", constants.InternalServerError, err)
		return
	}

	helpers.SendSuccessResponse(w, "Polls fetched successfully", polls)
}

// ClosePoll godoc
// @Summary      Close a poll
// @Description  Close a poll owned by the authenticated user. Once closed, no more votes are accepted.
// @Tags         Polls
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id    path      string  true  "Poll ID"
// @Success      200  {object}  models.Response  "Poll closed successfully"
// @Failure      403  {object}  models.Response  "User does not own the poll"
// @Failure      404  {object}  models.Response  "Poll not found"
// @Router       /polls/{id}/close [patch]
func (h *PollHandler) ClosePoll(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])

	if err != nil {
		helpers.SendErrorResponse(w, "Invalid poll id", constants.BadRequest, err)
		return
	}

	owns, err := h.pollRepo.OwnsPoll(r.Context(), helpers.GetUserId(r), id)
	if err != nil {
		helpers.SendErrorResponse(w, "Error checking poll ownership", constants.InternalServerError, err)
		return
	}
	if !owns {
		helpers.SendErrorResponse(w, fmt.Sprintf("User does not own the poll with id %s", id.String()), constants.Forbidden, nil)
		return
	}

	now := time.Now()
	poll := &models.PollModel{
		ID:     id,
		Closed: &now,
	}

	err = h.pollRepo.Update(r.Context(), poll)

	if err != nil {
		helpers.SendErrorResponse(w, "Error while closing poll", constants.InternalServerError, err)
		return
	}

	helpers.SendSuccessResponse(w, "Poll closed successfully", nil)
}

// UpdatePoll godoc
// @Summary      Update a poll
// @Description  Update the title and/or description of a poll owned by the authenticated user.
// @Tags         Polls
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id       path    string                    true  "Poll ID"
// @Param        request  body    models.PollRequestDto      true  "Poll updates"
// @Success      200     {object}  models.Response  "Poll updated successfully"
// @Failure      400     {object}  models.Response  "Invalid request body"
// @Failure      403     {object}  models.Response  "User does not own the poll"
// @Router       /polls/{id} [patch]
func (h *PollHandler) UpdatePoll(w http.ResponseWriter, r *http.Request) {
	var data models.PollRequestDto

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&data)

	if err != nil {
		helpers.SendErrorResponse(w, "Invalid request body", constants.BadRequest, err)
		return
	}

	if !helpers.ValidateStruct(w, &data) {
		return
	}

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])

	if err != nil {
		helpers.SendErrorResponse(w, "Invalid poll id", constants.BadRequest, err)
		return
	}

	owns, err := h.pollRepo.OwnsPoll(r.Context(), helpers.GetUserId(r), id)
	if err != nil {
		helpers.SendErrorResponse(w, "Error checking poll ownership", constants.InternalServerError, err)
		return
	}
	if !owns {
		helpers.SendErrorResponse(w, fmt.Sprintf("User does not own the poll with id %s", id.String()), constants.Forbidden, nil)
		return
	}

	poll := &models.PollModel{
		ID:          id,
		Title:       data.Title,
		Description: data.Description,
	}

	err = h.pollRepo.Update(r.Context(), poll)

	if err != nil {
		helpers.SendErrorResponse(w, "Error while updating poll", constants.InternalServerError, err)
		return
	}

	helpers.SendSuccessResponse(w, "Poll updated successfully", poll)
}

// DeletePoll godoc
// @Summary      Delete a poll
// @Description  Delete a poll, its options, and all associated votes. Only the owner can delete.
// @Tags         Polls
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Poll ID"
// @Success      200  {object}  models.Response  "Poll deleted successfully"
// @Failure      403  {object}  models.Response  "User does not own the poll"
// @Failure      404  {object}  models.Response  "Poll not found"
// @Router       /polls/{id} [delete]
func (h *PollHandler) DeletePoll(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])

	if err != nil {
		helpers.SendErrorResponse(w, "Invalid poll id", constants.BadRequest, err)
		return
	}

	owns, err := h.pollRepo.OwnsPoll(r.Context(), helpers.GetUserId(r), id)
	if err != nil {
		helpers.SendErrorResponse(w, "Error checking poll ownership", constants.InternalServerError, err)
		return
	}
	if !owns {
		helpers.SendErrorResponse(w, fmt.Sprintf("User does not own the poll with id %s", id.String()), constants.Forbidden, nil)
		return
	}

	err = h.pollRepo.DeleteWithOptions(r.Context(), id)

	if err != nil {
		helpers.SendErrorResponse(w, "Error while deleting poll", constants.InternalServerError, err)
		return
	}

	helpers.SendSuccessResponse(w, "Poll deleted successfully", nil)
}

// GetPollReport godoc
// @Summary      Get poll report
// @Description  Get vote totals and percentages for each option in a poll.
// @Tags         Polls
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Poll ID"
// @Success      200  {object}  models.PollReportDto  "Poll report fetched successfully"
// @Failure      404  {object}  models.Response       "Poll not found"
// @Router       /polls/{id}/report [get]
func (h *PollHandler) GetPollReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])

	if err != nil {
		helpers.SendErrorResponse(w, "Invalid poll id", constants.BadRequest, err)
		return
	}

	poll, err := h.pollRepo.FindByIDForUser(r.Context(), id, helpers.GetUserId(r))

	if err != nil {
		helpers.SendErrorResponse(w, "Poll not found", constants.NotFound, err)
		return
	}

	var totalVotes uint
	optionSummaries := make([]models.OptionSummaryDto, len(poll.Options))

	for i, opt := range poll.Options {
		totalVotes += opt.Votes
		optionSummaries[i] = models.OptionSummaryDto{
			OptionID: opt.ID,
			Label:    opt.Label,
			Votes:    opt.Votes,
		}
	}

	for i := range optionSummaries {
		if totalVotes > 0 {
			optionSummaries[i].Percentage = float64(optionSummaries[i].Votes) / float64(totalVotes) * 100
		} else {
			optionSummaries[i].Percentage = 0
		}
	}

	report := models.PollReportDto{
		PollID:     poll.ID,
		Title:      poll.Title,
		TotalVotes: totalVotes,
		Options:    optionSummaries,
	}

	helpers.SendSuccessResponse(w, "Poll report fetched successfully", report)
}

// GetPollById godoc
// @Summary      Get a single poll
// @Description  Fetch a poll by ID. The authenticated user can see their own polls and public polls.
// @Tags         Polls
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id   path      string  true  "Poll ID"
// @Success      200  {object}  models.Response  "Poll fetched successfully"
// @Failure      404  {object}  models.Response  "Poll not found"
// @Router       /polls/{id} [get]
func (h *PollHandler) GetPollById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])

	if err != nil {
		helpers.SendErrorResponse(w, "Invalid poll id", constants.BadRequest, err)
		return
	}

	poll, err := h.pollRepo.FindByIDForUser(r.Context(), id, helpers.GetUserId(r))

	if err != nil {
		helpers.SendErrorResponse(w, "Error while fetching poll", constants.NotFound, err)
		return
	}

	helpers.SendSuccessResponse(w, "Poll fetched successfully", poll)
}
