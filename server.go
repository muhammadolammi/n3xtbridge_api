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

	// Auth routes (public, no auth required)
	apiRoute.Post("/auth/signup", apiConfig.SignupHandler)
	apiRoute.Post("/auth/signin", apiConfig.SigninHandler)

	// Invoice routes (requires JWT auth + staff or admin role)
	apiRoute.With(apiConfig.AuthMiddleware(), apiConfig.RequireRole("admin", "staff")).Post("/invoice", apiConfig.CreateInvoiceHandler)

	// ORDER ROUTES
	apiRoute.With(apiConfig.AuthMiddleware()).Post("/service_order", apiConfig.CreateServiceOrderHandler)
	apiRoute.With(apiConfig.AuthMiddleware()).Get("/service_order", apiConfig.GetOrdersByUserHandler)
	apiRoute.With(apiConfig.AuthMiddleware(), apiConfig.RequireRole("admin")).Get("/service_order", apiConfig.GetOrdersByUserHandler)

	router.Mount("/api", apiRoute)
	srv := &http.Server{
		Addr:              ":" + apiConfig.Port,
		Handler:           router,
		ReadHeaderTimeout: time.Minute,
	}

	log.Printf("Serving on port: %s\n", apiConfig.Port)
	log.Fatal(srv.ListenAndServe())
}
