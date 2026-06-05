package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"roly-poly/internal/constants"
	"roly-poly/internal/helpers"
	"roly-poly/internal/models"
	repository "roly-poly/internal/repositories"
)

type VoteHandler struct {
	pollRepo repository.PollRepository
}

func NewVoteHandler(pollRepo repository.PollRepository) *VoteHandler {
	return &VoteHandler{pollRepo: pollRepo}
}

func (h *VoteHandler) CastVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pollID, err := uuid.Parse(vars["id"])
	if err != nil {
		helpers.SendErrorResponse(w, "Invalid poll id", constants.BadRequest, err)
		return
	}

	var data models.CastVoteRequestDto
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&data); err != nil {
		helpers.SendErrorResponse(w, "Invalid request body", constants.BadRequest, err)
		return
	}

	if !helpers.ValidateStruct(w, &data) {
		return
	}

	userID := helpers.GetUserId(r)

	poll, err := h.pollRepo.FindByID(r.Context(), pollID)
	if err != nil {
		helpers.SendErrorResponse(w, "Poll not found", constants.NotFound, err)
		return
	}

	if poll.Closed != nil {
		helpers.SendErrorResponse(w, "Poll is closed", constants.Forbidden, nil)
		return
	}

	if !poll.AllowMultipleVotes && len(data.OptionIDs) > 1 {
		helpers.SendErrorResponse(w, "Multiple votes not allowed on this poll", constants.BadRequest, nil)
		return
	}

	for _, optionID := range data.OptionIDs {
		valid := false
		for _, opt := range poll.Options {
			if opt.ID == optionID {
				valid = true
				break
			}
		}
		if !valid {
			helpers.SendErrorResponse(w, fmt.Sprintf("Option %s does not belong to this poll", optionID.String()), constants.BadRequest, nil)
			return
		}
	}

	if !poll.AllowMultipleVotes {
		count, err := h.pollRepo.CountVotesByUserAndPoll(r.Context(), userID, pollID)
		if err != nil {
			helpers.SendErrorResponse(w, "Error checking vote eligibility", constants.InternalServerError, err)
			return
		}
		if count > 0 {
			helpers.SendErrorResponse(w, "You have already voted on this poll", constants.BadRequest, nil)
			return
		}
	}

	if err := h.pollRepo.CastVotes(r.Context(), pollID, userID, data.OptionIDs); err != nil {
		helpers.SendErrorResponse(w, "Error while casting vote", constants.InternalServerError, err)
		return
	}

	helpers.SendSuccessResponse(w, "Vote cast successfully", nil)
}
