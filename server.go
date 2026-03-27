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
	corsOptions := cors.Options{
		AllowedOrigins: []string{"http://localhost:5173", "http://localhost:8081", "https://n3xtbridge.com", "https://n3xtbridge-backend-755404739186.us-east1.run.app"},

		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
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

	// 1. GLOBAL MIDDLEWARE (Must be in this order)
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(cors.Handler(corsOptions)) // CORS MUST BE BEFORE AUTH
	router.Use(middleware.Recoverer)

	// 2. DEFINE THE API ROUTE
	apiRoute := chi.NewRouter()

	// 3. APPLY CLIENT AUTH ONLY TO THE API GROUP (Not globally)
	apiRoute.Use(apiConfig.ClientAuth())

	// Public Health Checks
	apiRoute.Get("/hello", handlers.HelloReady)

	// Auth routes
	apiRoute.Post("/auth/signup", apiConfig.SignupHandler)
	apiRoute.Post("/auth/signin", apiConfig.AuthService.LoginHandler)
	apiRoute.Post("/auth/refresh", apiConfig.AuthService.RefreshHandler)

	// unprotected routes
	apiRoute.Get("/services", apiConfig.GetActiveServicesHandler)

	// Authenticated Auth routes
	apiRoute.Group(func(r chi.Router) {
		r.Use(apiConfig.AuthService.RequireAuth)
		r.Post("/auth/signout", apiConfig.AuthService.LogoutHandler)
		r.Get("/auth/user", apiConfig.GetUserHandler)
		r.Post("/customer/quotes/requests", apiConfig.CreateQuoteRequestHandler)
		r.Get("/customer/quotes/my-requests", apiConfig.GetUserQuoteRequestsHandler)
		r.Get("/customer/quotes", apiConfig.GetUserQuotesWithServiceHandler)
		r.Get("/customer/quotes/{id}", apiConfig.GetUserQuoteWithServiceHandler)

		r.Get("/customer/invoices", apiConfig.GetCustomerInvoicesHandler)
		r.Get("/invoices/{id}", apiConfig.GetInvoiceHandler)
		r.Patch("/customer/quotes/requests/{id}/description", apiConfig.UpdateUserQuoteRequestDescriptionHandler)

	})

	// Invoice routes
	apiRoute.Group(func(r chi.Router) {
		r.Use(apiConfig.AuthService.RequireAuth)
		r.Use(apiConfig.RequireRole("admin", "staff"))
		r.Post("/worker/invoices", apiConfig.CreateInvoiceHandler)
		r.Get("/worker/invoices", apiConfig.GetWorkersCreatedInvoicesHandler)
	})

	// Admin only
	apiRoute.Group(func(r chi.Router) {
		r.Use(apiConfig.AuthService.RequireAuth)
		r.Use(apiConfig.RequireRole("admin"))
		r.Get("/admin/invoices", apiConfig.AdminListAllInvoicesHandler)
		r.Post("/admin/services", apiConfig.CreateServiceeHandler)
		r.Get("/admin/services", apiConfig.AdminListAllServicesHandler)
		r.Patch("/admin/services/{id}/status", apiConfig.AdminUpdateServiceStatusHandler)
		r.Get("/admin/quote-requests", apiConfig.AdminGetQuoteRequestsHandler)
		r.Post("/admin/quotes", apiConfig.AdminCreateQuoteHandler)
		r.Get("/admin/quotes", apiConfig.AdminGetQuotesHandler)

		r.Patch("/admin/quotes/{id}/status", apiConfig.AdminUpdateQuoteStatusHandler)

	})

	// Mount everything under /api
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
