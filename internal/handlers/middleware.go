package handlers

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"slices"
	"time"

	"github.com/go-redis/redis_rate/v10"
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

// RateLimiter is a custom Chi middleware
func (cfg *Config) RateLimiter(limit int, window time.Duration) func(http.Handler) http.Handler {
	if cfg.RedisClient == nil {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		}
	}
	// log.Println("redis worked")
	limiter := redis_rate.NewLimiter(cfg.RedisClient)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			// Use Client IP as the unique key
			clientIP := r.RemoteAddr
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				clientIP = xff
			} else {
				if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
					clientIP = host
				}
			}

			// create a unique key for that endpoint, since diffrent endpoint have diffrent limit
			// log.Println("lets see url", r.URL)
			// log.Println("rate limit called.  url path", r.URL.Path)

			requestPath := r.URL.Path
			key := fmt.Sprintf("ratelimit:%s:%s", clientIP, requestPath)

			// Perform the check
			res, err := limiter.Allow(r.Context(), key, redis_rate.Limit{
				Rate:   limit,
				Period: window,
				Burst:  limit,
			})

			if err != nil {
				// If Redis is down, decide if you want to fail-open or fail-closed
				// Fail-open (allow request) is usually safer for UX
				next.ServeHTTP(w, r)
				return
			}

			// log.Printf("[RATE LIMIT] Key: %s | Limit: %d | Window: %v | Remaining: %d", key, limit, window, res.Allowed)

			if res.Allowed <= 0 {
				log.Println(res)
				helpers.RespondWithError(w, http.StatusTooManyRequests, "Too many requests.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
