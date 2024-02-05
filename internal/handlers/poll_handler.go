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

func (h *PollHandler) CreatePoll(w http.ResponseWriter, r *http.Request) {
	var data models.CreatePollRequestDto

	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		helpers.SendErrorResponse(w, err.Error(), constants.BadRequest, err)
		return
	}

	helpers.ValidateStruct(w, &data)

	userId := helpers.GetUserId(r)

	poll := &models.PollModel{
		Title:       data.Title,
		Description: data.Description,
		UserID:      userId,
	}

	err = h.pollRepo.Create(poll)

	if err != nil {
		helpers.SendErrorResponse(w, "Error while creating poll", constants.InternalServerError, err)
		return
	}

	var options []*models.OptionModel

	for _, option := range data.Options {
		optionModel := &models.OptionModel{
			PollID: poll.ID,
			Label:  option,
		}
		options = append(options, optionModel)
	}

	err = h.optionRepo.CreateMany(options)

	if err != nil {
		helpers.SendErrorResponse(w, "Error while creating poll options", constants.InternalServerError, err)
		return
	}

	helpers.SendSuccessResponse(w, "Poll created successfully", poll.Simple())
}

func (h *PollHandler) GetPolls(w http.ResponseWriter, r *http.Request) {
	polls, err := h.pollRepo.FindAll()

	if err != nil {
		helpers.SendErrorResponse(w, "Error while fetching polls", constants.InternalServerError, err)
		return
	}

	helpers.SendSuccessResponse(w, "Polls fetched successfully", polls)
}

func (h *PollHandler) ClosePoll(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])

	if err != nil {
		helpers.SendErrorResponse(w, "Invalid poll id", constants.BadRequest, err)
		return
	}

	userOwnsPoll, err := h.pollRepo.OwnsPoll(helpers.GetUserId(r), id)

	if err != nil || !userOwnsPoll {
		helpers.SendErrorResponse(w, fmt.Sprintf("User does not own the poll with id %s", id.String()), constants.InternalServerError, err)
		return
	}

	poll := &models.PollModel{
		ID:     id,
		Closed: time.Now(),
	}

	err = h.pollRepo.Update(poll)

	if err != nil {
		helpers.SendErrorResponse(w, "Error while closing poll", constants.InternalServerError, err)
		return
	}

	helpers.SendSuccessResponse(w, "Poll closed successfully", nil)
}

func (h *PollHandler) UpdatePoll(w http.ResponseWriter, r *http.Request) {
	var data models.PollRequestDto

	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		helpers.SendErrorResponse(w, err.Error(), constants.BadRequest, err)
		return
	}

	helpers.ValidateStruct(w, &data)

	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])

	if err != nil {
		helpers.SendErrorResponse(w, "Invalid poll id", constants.BadRequest, err)
		return
	}

	userOwnsPoll, err := h.pollRepo.OwnsPoll(helpers.GetUserId(r), id)

	if err != nil || !userOwnsPoll {
		helpers.SendErrorResponse(w, fmt.Sprintf("User does not own the poll with id %s", id.String()), constants.InternalServerError, err)
		return
	}

	poll := &models.PollModel{
		ID:          id,
		Title:       data.Title,
		Description: data.Description,
	}

	err = h.pollRepo.Update(poll)

	if err != nil {
		helpers.SendErrorResponse(w, "Error while updating poll", constants.InternalServerError, err)
		return
	}

	helpers.SendSuccessResponse(w, "Poll updated successfully", poll)
}

func (h *PollHandler) GetPollById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := uuid.Parse(vars["id"])

	if err != nil {
		helpers.SendErrorResponse(w, "Invalid poll id", constants.BadRequest, err)
		return
	}

	poll, err := h.pollRepo.FindByID(id)

	if err != nil {
		helpers.SendErrorResponse(w, "Error while fetching poll", constants.InternalServerError, err)
		return
	}

	helpers.SendSuccessResponse(w, "Poll fetched successfully", poll.Simple())
}
