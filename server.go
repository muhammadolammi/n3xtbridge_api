package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/muhammadolammi/n3xtbridge_api/internal/handlers"
)

func server(apiConfig *handlers.Config) {

	// Define CORS options
	corsOptions := cors.Options{
		AllowedOrigins: []string{
			"http://localhost:3000", "http://localhost:5173", "https://n3xtbridge.com", // Your Firebase URL
		}, // You can customize this based on your needs

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
	apiRoute.Post("/auth/signin", apiConfig.AuthService.LoginHandler)
	apiRoute.With(apiConfig.AuthService.RequireAuth).Post("/auth/signout", apiConfig.AuthService.LogoutHandler)
	apiRoute.Post("/auth/refresh", apiConfig.AuthService.RefreshHandler)
	apiRoute.With(apiConfig.AuthService.RequireAuth).Get("/auth/user", apiConfig.GetUserHandler)

	// Invoice routes (requires JWT auth + staff or admin role)
	apiRoute.With(apiConfig.AuthService.RequireAuth, apiConfig.RequireRole("admin", "staff")).Post("/invoices", apiConfig.CreateInvoiceHandler)
	apiRoute.With(apiConfig.AuthService.RequireAuth, apiConfig.RequireRole("admin", "staff")).Get("/invoices", apiConfig.GetInvoicesHandler)
	apiRoute.With(apiConfig.AuthService.RequireAuth, apiConfig.RequireRole("admin", "staff")).Get("/invoices/{id}", apiConfig.GetInvoiceHandler)
	apiRoute.With(apiConfig.AuthService.RequireAuth, apiConfig.RequireRole("admin")).Get("/admin/invoices", apiConfig.AdminListAllInvoicesHandler)

	router.Mount("/api", apiRoute)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Fallback for local development
	}

	addr := ":" + port
	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: time.Minute,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(srv.ListenAndServe())
}
