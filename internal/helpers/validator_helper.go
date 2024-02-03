package helpers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"

	"roly-poly/pkg/logger"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterValidation("noSQLKeywords", noSQLKeywords)
}

func noSQLKeywords(fl validator.FieldLevel) bool {
	sqlKeywords := []string{"SELECT", "FROM", "WHERE", "DELETE", "UPDATE", "INSERT", "DROP", "CREATE", "ALTER", "TRUNCATE"}

	value := fl.Field().String()
	for _, keyword := range sqlKeywords {
		if strings.Contains(strings.ToUpper(value), keyword) {
			return false
		}
	}
	return true
}

func ValidateStruct(w http.ResponseWriter, s interface{}) error {
	log := logger.New()
	log.Debug().Interface("data", s).Msg("Validating Request Data")

	err := validate.Struct(s)
	if err != nil {
		var errMsgs []string
		for _, err := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", err.Field(), err.Tag()))
		}
		return errors.New(strings.Join(errMsgs, ", "))
	}
	return nil
}
