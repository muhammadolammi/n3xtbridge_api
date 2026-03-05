package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/muhammadolammi/n3xtbridge_api/internal/helpers"
)

// Middleware to check for the API key in the authorization header for all POST, PUT, DELETE, and OPTIONS requests
func (cfg *Config) ClientAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// if r.Method == http.MethodOptions {
			// 	next.ServeHTTP(w, r)
			// 	return
			// }
			// Bypass SSE endpoint and inject Authorization header
			if strings.HasPrefix(r.URL.Path, "/api/sessions/sse") {
				// Get token from query parameter
				accessToken := r.URL.Query().Get("access_token")
				if accessToken == "" {
					helpers.RespondWithError(w, http.StatusUnauthorized, "missing access token")
					return
				}

				// Inject Authorization header for downstream AuthMiddleware
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

				// Continue to next handler
				next.ServeHTTP(w, r)
				return
			}
			// Bypass Paystack webhook
			if strings.HasPrefix(r.URL.Path, "/api/webhook/paystack") {
				// TODO handle paystack athorization
				// Continue to next handler
				next.ServeHTTP(w, r)
				return
			}

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
