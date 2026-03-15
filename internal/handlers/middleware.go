package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/muhammadolammi/n3xtbridge_api/internal/auth"
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

// AuthMiddleware checks for JWT token and validates it
func (cfg *Config) AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				helpers.RespondWithError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			bearerToken := strings.Split(authHeader, " ")
			if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
				helpers.RespondWithError(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}

			claims, err := auth.ValidateToken(bearerToken[1], cfg.JwtSecret)
			if err != nil {
				helpers.RespondWithError(w, http.StatusUnauthorized, "invalid token: "+err.Error())
				return
			}

			// Attach user info to context (optional, can be used in handlers)
			ctx := context.WithValue(r.Context(), "user", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole middleware checks if the user has one of the required roles
func (cfg *Config) RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user from context (set by AuthMiddleware)
			claimsVal := r.Context().Value("user")
			if claimsVal == nil {
				helpers.RespondWithError(w, http.StatusUnauthorized, "user not authenticated")
				return
			}

			claims, ok := claimsVal.(*auth.Claims)
			if !ok {
				helpers.RespondWithError(w, http.StatusInternalServerError, "invalid user context")
				return
			}

			// Check if user role is in allowed roles
			hasAccess := false
			for _, role := range allowedRoles {
				if claims.Role == role {
					hasAccess = true
					break
				}
			}

			if !hasAccess {
				helpers.RespondWithError(w, http.StatusForbidden, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
