package handlers

import (
	"log"
	"net/http"
	"slices"

	"github.com/muhammadolammi/n3xtbridge_api/internal/helpers"
)

// Middleware to check for the API key in the authorization header
func (cfg *Config) ClientAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			clientApiKey := r.Header.Get("client-api-key")
			if clientApiKey == "" {
				log.Println("empty client api key in request.")
				helpers.RespondWithError(w, http.StatusUnauthorized, "empty client api key in request.")
				return
			}
			if clientApiKey != cfg.ClientApiKey {
				log.Println("invalid client api key in request.")
				helpers.RespondWithError(w, http.StatusUnauthorized, "invalid client api key in request.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole middleware checks if the user has one of the required roles
func (cfg *Config) RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user from context (set by AuthMiddleware)
			user, httpstatus, err := cfg.getUserFromReq(r)
			if err != nil {
				helpers.RespondWithError(w, httpstatus, err.Error())

			}
			// Check if user role is in allowed roles
			hasAccess := slices.Contains(allowedRoles, user.Role)

			if !hasAccess {
				helpers.RespondWithError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
