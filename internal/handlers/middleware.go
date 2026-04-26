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

// RateLimiter is a custom Chi middleware (FAIL-OPEN SAFE VERSION)
func (cfg *Config) RateLimiter(limit int, window time.Duration) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Always allow OPTIONS (CORS preflight)
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// If Redis is not available → FAIL OPEN
			if cfg.RedisClient == nil {
				log.Println("⚠️ Redis not available, skipping rate limiter")
				next.ServeHTTP(w, r)
				return
			}

			// Safe recovery in case redis_rate panics internally
			defer func() {
				if rec := recover(); rec != nil {
					log.Println("⚠️ Rate limiter panic recovered:", rec)
					next.ServeHTTP(w, r)
				}
			}()

			limiter := redis_rate.NewLimiter(cfg.RedisClient)

			// Extract client IP
			clientIP := r.RemoteAddr
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				clientIP = xff
			} else if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				clientIP = host
			}

			requestPath := r.URL.Path
			key := fmt.Sprintf("ratelimit:%s:%s", clientIP, requestPath)

			// Try rate limit check
			res, err := limiter.Allow(r.Context(), key, redis_rate.Limit{
				Rate:   limit,
				Period: window,
				Burst:  limit,
			})

			// If Redis fails → FAIL OPEN
			if err != nil {
				log.Println("⚠️ Redis rate limit error, allowing request:", err)
				next.ServeHTTP(w, r)
				return
			}

			// If allowed → continue
			if res.Allowed <= 0 {
				log.Println("🚫 Rate limit exceeded:", key)
				helpers.RespondWithError(w, http.StatusTooManyRequests, "Too many requests.")
				return
			}
			// log.Println("rate limit working....")

			next.ServeHTTP(w, r)
		})
	}
}
