package middlewares

import (
	"net/http"

	"roly-poly/internal/constants"
	"roly-poly/internal/helpers"
	repository "roly-poly/internal/repositories"
	"roly-poly/pkg/logger"
)

var whitelist = []string{
	constants.HealthCheckEndpoint,
	constants.ReadinessEndpoint,
	constants.OnboardUserEndpoint,
}

func NewAuthMiddleware(userRepo repository.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			for _, path := range whitelist {
				if r.URL.Path == path {
					next.ServeHTTP(w, r)
					return
				}
			}

			apiKey := r.Header.Get(constants.AuthorizationHeader)

			if apiKey == "" {
				logger.New().Error().Msg("Api key not found")
				helpers.SendErrorResponse(w, "Authorization token not found", constants.Unauthorized, nil)
				return
			}

			user, err := userRepo.FindByApiKey(r.Context(), apiKey)

			if err != nil {
				logger.New().Error().Err(err).Msg("Error while fetching user by api key")
				helpers.SendErrorResponse(w, "Error while fetching user by api key", constants.InternalServerError, nil)
				return
			}

			if !user.Enabled {
				helpers.SendErrorResponse(w, "User account is disabled", constants.Forbidden, nil)
				return
			}

			r = helpers.SetApiKey(r, apiKey)
			r = helpers.SetUserId(r, user.ID)

			next.ServeHTTP(w, r)
		})
	}
}
