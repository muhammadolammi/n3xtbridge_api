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
			"Content-Type", "Authorization", "X-Requested-With",
			"client-api-key", "X-CSRF-Token", "x-paystack-signature", "Accept",
		},
		AllowCredentials: true,
		MaxAge:           300,
	}

	router := chi.NewRouter()

	// 1. GLOBAL MIDDLEWARE
	router.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, cors.Handler(corsOptions), middleware.Recoverer)

	// --- PUBLIC WEBHOOKS (Strict Rate Limit) ---
	router.With(apiConfig.RateLimiter(5, 10*time.Minute)).Post("/api/webhooks/paystack", apiConfig.PaystackWebhookHandler)

	// 2. DEFINE THE API ROUTE
	protectedRoute := chi.NewRouter()
	protectedRoute.Use(apiConfig.ClientAuth())

	// Public Health Checks (Standard Limit)
	protectedRoute.With(apiConfig.RateLimiter(60, time.Minute)).Get("/hello", handlers.HelloReady)

	protectedRoute.Group(func(r chi.Router) {
		r.Use(apiConfig.RateLimiter(5, 10*time.Minute))
		r.Post("/auth/signup", apiConfig.SignupHandler)
		r.Post("/auth/signin", apiConfig.AuthService.LoginHandler)
	})

	// --- TIER: HIGH FREQUENCY AUTH (Refresh/Check) ---
	protectedRoute.Group(func(r chi.Router) {
		r.Use(apiConfig.RateLimiter(30, 10*time.Minute)) // Much more breathing room
		r.Post("/auth/refresh", apiConfig.AuthService.RefreshHandler)
		r.Post("/auth/check-lead", apiConfig.CheckLeadHandler)
	})

	// --- TIER: PUBLIC DATA (Standard) ---
	protectedRoute.Group(func(r chi.Router) {
		r.Use(apiConfig.RateLimiter(60, time.Minute))
		r.Get("/services", apiConfig.GetActiveServicesHandler)
		r.Get("/services/{id}", apiConfig.GetServiceHandler)
		r.Get("/promotions", apiConfig.GetActivePromosHandler)
		r.Get("/promotions/{id}", apiConfig.GetPromoHandler)
		r.Get("/p/invoices/{id}", apiConfig.PublicGetInvoiceHandler)
	})

	// --- TIER: SENSITIVE ACTIONS (Moderate) ---
	protectedRoute.Group(func(r chi.Router) {
		r.Use(apiConfig.RateLimiter(15, time.Minute))
		r.Get("/promotions/verify/{code}", apiConfig.VerifyPromoHandler)
		r.Post("/payments/{id}", apiConfig.InitializePaymentHandler)
		r.Get("/payments/verify/{ref}", apiConfig.VerifyPaymentStatusHandler)
	})

	// --- TIER: AUTHENTICATED CUSTOMER ROUTES ---
	protectedRoute.Group(func(r chi.Router) {
		r.Use(apiConfig.AuthService.RequireAuth)
		r.Use(apiConfig.RateLimiter(100, time.Minute))

		r.Post("/auth/signout", apiConfig.AuthService.LogoutHandler)
		r.Get("/auth/user", apiConfig.GetUserHandler)
		// Customer Operations
		r.Post("/customer/quotes/requests", apiConfig.CreateQuoteRequestHandler)
		r.Get("/customer/quotes/my-requests", apiConfig.GetUserQuoteRequestsHandler)
		r.Get("/customer/quotes", apiConfig.GetUserQuotesWithServiceHandler)
		r.Get("/customer/quotes/{id}", apiConfig.GetUserQuoteWithServiceHandler)
		r.Get("/customer/invoices", apiConfig.GetCustomerInvoicesHandler)
		r.Get("/invoices/{id}", apiConfig.GetInvoiceHandler)
		r.Get("/quotes/invoices/{id}", apiConfig.GetQuoteInvoiceHandler)
		r.Patch("/customer/quotes/requests/{id}/description", apiConfig.UpdateUserQuoteRequestDescriptionHandler)
		r.Patch("/customer/quotes/{id}/status", apiConfig.CustomerUpdateQuoteStatusHandler)
		r.Post("/storage/presign", apiConfig.PresignUploadHandler)
		r.Get("/storage/presign/*", apiConfig.PresignGetHandler)
	})

	// --- TIER: WORKER/STAFF ROUTES ---
	protectedRoute.Group(func(r chi.Router) {
		r.Use(apiConfig.AuthService.RequireAuth)
		r.Use(apiConfig.RequireRole("admin", "staff"))
		r.Use(apiConfig.RateLimiter(30, time.Minute))

		r.Post("/worker/invoices", apiConfig.CreateInvoiceHandler)
		r.Get("/worker/invoices", apiConfig.GetWorkersCreatedInvoicesHandler)
	})

	// --- TIER: ADMIN ONLY ROUTES ---
	protectedRoute.Group(func(r chi.Router) {
		r.Use(apiConfig.AuthService.RequireAuth)
		r.Use(apiConfig.RequireRole("admin"))
		r.Use(apiConfig.RateLimiter(30, time.Minute))

		r.Get("/admin/invoices", apiConfig.AdminListAllInvoicesHandler)
		r.Post("/admin/services", apiConfig.CreateServiceHandler)
		r.Get("/admin/services", apiConfig.AdminListAllServicesHandler)
		r.Patch("/admin/services/{id}/status", apiConfig.AdminUpdateServiceStatusHandler)
		r.Get("/admin/quote-requests", apiConfig.AdminGetQuoteRequestsHandler)
		r.Post("/admin/quotes", apiConfig.AdminCreateQuoteHandler)
		r.Get("/admin/quotes", apiConfig.AdminGetQuotesHandler)
		r.Patch("/admin/quotes/{id}/status", apiConfig.AdminUpdateQuoteStatusHandler)
		r.Post("/admin/promotions", apiConfig.AdminCreatePromotionHandler)
		r.Get("/admin/promotions", apiConfig.AdminListPromotionsHandler)
		r.Post("/admin/mail/invoices/{id}", apiConfig.AdminSendInvoiceEmailHandler)
	})

	// Mount everything under /api
	router.Mount("/api", protectedRoute)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
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
