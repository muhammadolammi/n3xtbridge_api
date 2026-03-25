package handlers

import (
	"errors"
	"log"
	"net/http"

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
	user, err := cfg.DB.GetUserByID(r.Context(), parsedID)
	if err != nil {
		log.Println("DB ERROR error getting invoice: " + err.Error())
		return database.User{}, http.StatusUnauthorized, errors.New("user not authenticated")

	}
	return user, http.StatusOK, nil
}
