package handlers

import (
	"log"
	"net/http"

	"github.com/muhammadolammi/n3xtbridge_api/internal/helpers"
)

// Middleware to check for the API key in the authorization header for all POST, PUT, DELETE, and OPTIONS requests
func (cfg *Config) ClientAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientApiKey := r.Header.Get("client-api-key")
			if clientApiKey == "" {
				log.Println("empty client api key  in request.")
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
