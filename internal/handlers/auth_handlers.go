package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/muhammadolammi/n3xtbridge_api/internal/auth"
	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
	"github.com/muhammadolammi/n3xtbridge_api/internal/helpers"
)

type SignupInput struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	Address     string `json:"address"`
}

type SigninInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32)

	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}
func (cfg *Config) SignupHandler(w http.ResponseWriter, r *http.Request) {
	var input SignupInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, err: "+err.Error())
		return
	}

	// Check if user already exists
	existingUser, _ := cfg.DB.GetUserByEmail(context.Background(), input.Email)
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
		Email:        input.Email,
		PasswordHash: passwordHash,
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		PhoneNumber:  sql.NullString{String: input.PhoneNumber, Valid: input.PhoneNumber != ""},
		Address:      sql.NullString{String: input.Address, Valid: input.Address != ""},
		Role:         "user", // default role
	}

	user, err := cfg.DB.CreateUser(context.Background(), params)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to create user: "+err.Error())
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID.String(), user.Email, user.Role, cfg.JwtSecret)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	response := AuthResponse{
		Token: token,
		User:  dbUserToUser(user),
	}

	helpers.RespondWithJson(w, http.StatusCreated, response)
}

func (cfg *Config) SigninHandler(w http.ResponseWriter, r *http.Request) {
	var input SigninInput

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request, err: "+err.Error())
		return
	}

	// Get user by email
	user, err := cfg.DB.GetUserByEmail(context.Background(), input.Email)
	if err != nil {
		helpers.RespondWithError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Check password
	if !auth.CheckPasswordHash(input.Password, user.PasswordHash) {
		helpers.RespondWithError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID.String(), user.Email, user.Role, cfg.JwtSecret)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	response := AuthResponse{
		Token: token,
		User:  dbUserToUser(user),
	}

	helpers.RespondWithJson(w, http.StatusOK, response)
}
