package handlers

import (
	"encoding/json"
	"net/http"

	"roly-poly/internal/constants"
	"roly-poly/internal/helpers"
	"roly-poly/internal/models"
	repository "roly-poly/internal/repositories"
)

type AdminHandler struct {
	userRepo repository.UserRepository
}

func NewAdminHandler(userRepo repository.UserRepository) *AdminHandler {
	return &AdminHandler{
		userRepo: userRepo,
	}
}

func (h *AdminHandler) OnboardUser(w http.ResponseWriter, r *http.Request) {
	var data models.OnboardUserRequestDto

	err := json.NewDecoder(r.Body).Decode(&data)

	if err != nil {
		helpers.SendErrorResponse(w, err.Error(), constants.BadRequest, err)
		return
	}

	helpers.ValidateStruct(w, &data)

	user := &models.UserModel{
		FirstName: data.FirstName,
		LastName:  data.LastName,
	}

	err = h.userRepo.Create(user)

	if err != nil {
		helpers.SendErrorResponse(w, "Error while creating user", constants.InternalServerError, err)
		return
	}

	helpers.SendSuccessResponse(w, "User onboarded successfully", user.Simple())
	return

}
