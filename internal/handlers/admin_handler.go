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

// OnboardUser godoc
// @Summary      Create a user and API key
// @Description  Creates a new user and returns an API key for authenticating subsequent requests.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.OnboardUserRequestDto  true  "User details"
// @Success      201     {object}  models.Response  "User onboarded successfully"
// @Failure      400     {object}  models.Response  "Invalid request body"
// @Failure      500     {object}  models.Response  "Error while creating user"
// @Router       /onboard [post]
func (h *AdminHandler) OnboardUser(w http.ResponseWriter, r *http.Request) {
	var data models.OnboardUserRequestDto

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

	user := &models.UserModel{
		FirstName: data.FirstName,
		LastName:  data.LastName,
	}

	err = h.userRepo.Create(r.Context(), user)

	if err != nil {
		helpers.SendErrorResponse(w, "Error while creating user", constants.InternalServerError, err)
		return
	}

	resp := models.OnboardUserResponseDto{
		ID:        user.ID,
		ApiKey:    user.ApiKey,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	helpers.SendCreatedResponse(w, "User onboarded successfully", resp)
}
