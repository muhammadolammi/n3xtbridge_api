package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/muhammadolammi/n3xtbridge_api/internal/auth"
	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
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

type SignupInput struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	Address     string `json:"address"`
	Country     string `json:"country"`
	State       string `json:"state"`
}

func (cfg *Config) SignupHandler(w http.ResponseWriter, r *http.Request) {
	var input SignupInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, err: "+err.Error())
		return
	}

	// Check if user already exists
	existingUser, _ := cfg.DBQueries.GetUserByEmail(context.Background(), input.Email)
	if existingUser.ID != uuid.Nil {
		helpers.RespondWithError(w, http.StatusConflict, "user with this email already exists")
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// Create user
	params := database.CreateUserParams{
		Email:        strings.ToLower(strings.TrimSpace(input.Email)),
		PasswordHash: passwordHash,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		PhoneNumber:  sql.NullString{String: input.PhoneNumber, Valid: input.PhoneNumber != ""},
		Address:      input.Address,
		Country:      input.Country,
		State:        input.State,
		Role:         "user", // default role
	}

	_, err = cfg.DBQueries.CreateUser(context.Background(), params)
	if err != nil {
		log.Println("failed to create user: " + err.Error())
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to create user: ")
		return
	}

	helpers.RespondWithJson(w, 200, "signup successful")
}

func (cfg *Config) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	user, httpstatus, err := cfg.getUserFromReq(r)
	if err != nil {
		helpers.RespondWithError(w, httpstatus, err.Error())
		return

	}
	helpers.RespondWithJson(w, http.StatusOK, dbUserToUser(user))
}
