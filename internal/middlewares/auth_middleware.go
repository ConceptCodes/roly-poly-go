package middlewares

import (
	"net/http"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"roly-poly/internal/constants"
	"roly-poly/internal/helpers"
	repository "roly-poly/internal/repositories"
	"roly-poly/pkg/storage/postgres"
)

var (
	db        *gorm.DB
	whitelist = []string{
		constants.HealthCheckEndpoint,
		constants.ReadinessEndpoint,
		constants.OnboardUserEndpoint,
	}
)

func init() {
	_db, err := postgres.New()

	if err != nil {
		log.Fatal().Err(err).Msg("Error while connecting to database")
	}

	db = _db
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		for _, path := range whitelist {
			if r.URL.Path == path {
				next.ServeHTTP(w, r)
				return
			}
		}

		apiKey := r.Header.Get(constants.AuthorizationHeader)

		if apiKey != "" {
			log.Error().Msg("Api key not found")
			helpers.SendErrorResponse(w, "Authorization token not found", constants.Unauthorized, nil)
			return
		}

		userRepo := repository.NewGormUserRepository(db)

		user, err := userRepo.FindByApiKey(apiKey)

		if err != nil {
			log.Error().Err(err).Msg("Error while fetching user by api key")
			helpers.SendErrorResponse(w, "Error while fetching user by api key", constants.InternalServerError, nil)
			return
		}

		r = helpers.SetApiKey(r, apiKey)
		r = helpers.SetUserId(r, user.ID)

		next.ServeHTTP(w, r)
	})
}
