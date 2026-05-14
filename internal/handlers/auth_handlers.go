package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/markbates/goth/gothic"
	"github.com/muhammadolammi/n3xtbridge_api/internal/helpers"
)

// func GenerateVerificationToken() (string, error) {
// 	bytes := make([]byte, 32)

// 	_, err := rand.Read(bytes)
// 	if err != nil {
// 		return "", err
// 	}

//		return hex.EncodeToString(bytes), nil
//	}

func (cfg *Config) BeginAuthHandler(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	r = r.WithContext(context.WithValue(r.Context(), "provider", provider))

	gothic.BeginAuthHandler(w, r)

}
func (cfg *Config) GoogleAuthCallback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	cfg.AuthService.GoogleAuthCallback(w, r, provider)
}

func (cfg *Config) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	user, httpstatus, err := cfg.getUserFromReq(r)
	if err != nil {
		helpers.RespondWithError(w, httpstatus, err.Error())
		return

	}
	helpers.RespondWithJson(w, http.StatusOK, dbUserToUser(user))
}

func (cfg *Config) CheckLeadHandler(w http.ResponseWriter, r *http.Request) {
	input := struct {
		Email string `json:"email"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, err: "+err.Error())
		return
	}

	// Check if user already exists
	existingUser, err := cfg.DBQueries.GetUserByEmail(context.Background(), input.Email)
	res := struct {
		Exist bool `json:"exists"`
	}{}
	if existingUser.ID != uuid.Nil {
		// existing user
		res.Exist = true
		helpers.RespondWithJson(w, http.StatusOK, res)
		return
	}
	// check for other error

	if err != nil {
		if strings.Contains(err.Error(), "sql: no rows in result set") {
			res.Exist = false
			helpers.RespondWithJson(w, http.StatusOK, res)
		}
		helpers.RespondWithError(w, http.StatusInternalServerError, "Error checking lead")
		return
	}
	// just to be sure the default is already false
	res.Exist = false

	helpers.RespondWithJson(w, http.StatusOK, res)
}
