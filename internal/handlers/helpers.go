package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"unicode"

	"github.com/google/uuid"
	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
)

func (cfg *Config) getUserFromReq(r *http.Request) (database.User, int, error) {

	// authorization
	userIDStr, ok := r.Context().Value("user_id").(string)
	if !ok {
		return database.User{}, http.StatusUnauthorized, errors.New("user not found in context")

	}
	parsedID, err := uuid.Parse(userIDStr)
	if err != nil {
		return database.User{}, http.StatusInternalServerError, errors.New("error parsing user id")

	}
	user, err := cfg.DBQueries.GetUserByID(r.Context(), parsedID)
	if err != nil {
		log.Println("DB ERROR error getting invoice: " + err.Error())
		return database.User{}, http.StatusUnauthorized, errors.New("user not authenticated")

	}
	return user, http.StatusOK, nil
}

func validatePassword(password string) error {
	var (
		hasUpper  bool
		hasLower  bool
		hasSymbol bool
	)

	if len(password) < 10 {
		return fmt.Errorf("password must be at least 10 characters long")
	}

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSymbol = true
		}
	}

	if !hasUpper || !hasLower || !hasSymbol {
		return fmt.Errorf("password must include uppercase, lowercase, and a symbol")
	}

	return nil
}
