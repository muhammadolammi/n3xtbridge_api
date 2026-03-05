package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/muhammadolammi/n3xtbridge_api/internal/handlers"
)

func server(apiConfig *handlers.Config) {

	// Define CORS options
	corsOptions := cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:8081", "https://gojobmatch.com", "https://jobmatch-backend-755404739186.us-east1.run.app"}, // You can customize this based on your needs

		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"client-api-key",
			"X-CSRF-Token",
			"x-paystack-signature",
			"Accept",
		},
		AllowCredentials: true,
		MaxAge:           300,
	}
	router := chi.NewRouter()
	apiRoute := chi.NewRouter()
	router.Use(cors.Handler(corsOptions))
	// ADD MIDDLREWARE
	// A good base middleware stack
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(apiConfig.ClientAuth())

	// ADD ROUTES
	apiRoute.Get("/hello", handlers.HelloReady)
	apiRoute.Get("/error", handlers.ErrorReady)

	router.Mount("/api", apiRoute)
	srv := &http.Server{
		Addr:              ":" + apiConfig.Port,
		Handler:           router,
		ReadHeaderTimeout: time.Minute,
	}

	log.Printf("Serving on port: %s\n", apiConfig.Port)
	log.Fatal(srv.ListenAndServe())
}
