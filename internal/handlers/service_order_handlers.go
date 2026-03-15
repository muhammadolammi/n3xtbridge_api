package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/muhammadolammi/n3xtbridge_api/internal/auth"
	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
	"github.com/muhammadolammi/n3xtbridge_api/internal/helpers"
	"github.com/muhammadolammi/n3xtbridge_api/internal/mailer"
	"github.com/sqlc-dev/pqtype"
)

// GenerateOrderNumber generates a unique order number
func GenerateOrderNumber() string {
	year := time.Now().Year()
	counter := time.Now().UnixNano() % 1000000
	return fmt.Sprintf("ORD-%d-%06d", year, counter)
}

var transportData = map[string]string{
	"lagos": "15000.00",

	// South West
	"ogun":  "20000.00",
	"oyo":   "35000.00",
	"osun":  "45000.00",
	"ondo":  "45000.00",
	"ekiti": "50000.00",

	// South South
	"edo":         "55000.00",
	"delta":       "60000.00",
	"bayelsa":     "70000.00",
	"rivers":      "70000.00",
	"akwa ibom":   "75000.00",
	"cross river": "80000.00",

	// South East
	"abia":    "70000.00",
	"anambra": "65000.00",
	"ebonyi":  "70000.00",
	"enugu":   "65000.00",
	"imo":     "70000.00",

	// North Central
	"kwara":    "45000.00",
	"kogi":     "50000.00",
	"benue":    "65000.00",
	"niger":    "60000.00",
	"plateau":  "70000.00",
	"nasarawa": "65000.00",
	"fct":      "65000.00",

	// North West
	"kaduna":  "75000.00",
	"kano":    "85000.00",
	"katsina": "90000.00",
	"kebbi":   "95000.00",
	"sokoto":  "100000.00",
	"zamfara": "95000.00",
	"jigawa":  "90000.00",

	// North East
	"adamawa": "100000.00",
	"bauchi":  "85000.00",
	"borno":   "110000.00",
	"gombe":   "85000.00",
	"taraba":  "90000.00",
	"yobe":    "95000.00",
}

// CreateServiceOrderHandler handles creation of a new service order (public endpoint)
func (cfg *Config) CreateServiceOrderHandler(w http.ResponseWriter, r *http.Request) {
	var input ServiceOrderRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	// Validate required fields
	if input.Email == "" || input.FullName == "" || input.BusinessName == "" || input.Phone == "" || input.ServiceType == "" || input.DeliveryAddress == "" || input.DeliveryCity == "" || input.DeliveryState == "" || input.DeliveryCountry == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "input required data")
		return
	}

	// if input.ServiceType != "solar" && input.ServiceType != "security" && input.ServiceType != "security" {
	// 	helpers.RespondWithError(w, http.StatusBadRequest, "service_type must be 'solar', 'security', or 'both'")
	// 	return
	// }

	if len(input.Appliances) == 0 {
		helpers.RespondWithError(w, http.StatusBadRequest, "at least one appliance is required")
		return
	}

	// Validate company size
	validCompanySizes := map[string]bool{
		"sole_proprietor": true,
		"1-10":            true,
		"11-50":           true,
		"51-200":          true,
		"201-1000":        true,
		"1000+":           true,
	}
	if !validCompanySizes[input.CompanySize] {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid company_size")
		return
	}

	// Generate order number
	orderNumber := GenerateOrderNumber()
	// Convert appliances to JSON for storage
	appliancesJSON, err := json.Marshal(input.Appliances)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to process appliances")
		return
	}

	// Determine user ID if user is authenticated
	//  if no user id then we know the order isnt from a signed up user
	var userID uuid.NullUUID
	if claims, ok := r.Context().Value("user").(*auth.Claims); ok && claims.UserID != "" {
		if parsedUUID, err := uuid.Parse(claims.UserID); err == nil {
			userID = uuid.NullUUID{UUID: parsedUUID, Valid: true}
		}
	} else {
		helpers.RespondWithError(w, http.StatusUnauthorized, "please log in")
		return
	}

	tranportfee, ok := transportData[strings.ToLower(input.DeliveryState)]
	if !ok {
		helpers.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("service not available in %s state currently.", input.DeliveryState))
		return
	}

	// Build DB params
	dbParams := database.CreateServiceOrderParams{
		OrderNumber:      orderNumber,
		Email:            input.Email,
		FullName:         input.FullName,
		BusinessName:     input.BusinessName,
		Phone:            input.Phone,
		WhatsappPhone:    sql.NullString{String: input.WhatsappPhone, Valid: input.WhatsappPhone != ""},
		CompanySize:      input.CompanySize,
		ReferralSource:   input.ReferralSource,
		ServiceType:      input.ServiceType,
		ApplianceDetails: pqtype.NullRawMessage{RawMessage: appliancesJSON, Valid: true},
		DeliveryAddress:  input.DeliveryAddress,
		TransportFee:     tranportfee, // Promo: always 1 USD
		PromoApplied:     sql.NullBool{Bool: true, Valid: true},
		Status:           sql.NullString{String: "pending", Valid: true},
		UserID:           userID,
		Notes:            sql.NullString{String: input.Notes, Valid: input.Notes != ""},
	}

	// Save to database
	ctx := r.Context()
	order, err := cfg.DB.CreateServiceOrder(ctx, dbParams)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to create order: "+err.Error())
		return
	}

	// Parse appliances from JSON for response
	mailerAppliances, err := mailer.ParseAppliances(order.ApplianceDetails.RawMessage)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to process order details")
		return
	}

	// Convert to handler Appliance type
	appliances := make([]Appliance, len(mailerAppliances))
	for i, a := range mailerAppliances {
		appliances[i] = Appliance{
			Name:          a.Name,
			Quantity:      a.Quantity,
			Price:         a.EstimatedCost, // Map EstimatedCost to Price in handler type
			PartnerVendor: a.PartnerVendor,
		}
	}

	// Prepare response
	response := ServiceOrderResponse{
		OrderNumber:     order.OrderNumber,
		Email:           order.Email,
		FullName:        order.FullName,
		BusinessName:    order.BusinessName,
		Phone:           order.Phone,
		WhatsappPhone:   order.WhatsappPhone.String,
		CompanySize:     order.CompanySize,
		ReferralSource:  order.ReferralSource,
		ServiceType:     order.ServiceType,
		Appliances:      appliances,
		DeliveryAddress: order.DeliveryAddress,
		TransportFee:    1.00,
		PromoApplied:    true,
		Status:          "pending",
		Notes:           order.Notes.String,
		CreatedAt:       order.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		Message:         "Order created successfully! Check your email for confirmation.",
	}

	// Send confirmation email (async, don't block)
	go func() {
		if cfg.SMTPServer != "" && cfg.SMTPUsername != "" && cfg.SMTPPassword != "" {
			// Convert appliances for email
			mailerAppliances := make([]mailer.ApplianceDetails, len(appliances))
			for i, a := range appliances {
				mailerAppliances[i] = mailer.ApplianceDetails{
					Name:          a.Name,
					Quantity:      a.Quantity,
					EstimatedCost: a.Price,
					PartnerVendor: a.PartnerVendor,
				}
			}

			emailData := mailer.EmailData{
				OrderNumber:     order.OrderNumber,
				CustomerName:    order.FullName,
				BusinessName:    order.BusinessName,
				ServiceType:     order.ServiceType,
				TransportFee:    1.00,
				DeliveryAddress: order.DeliveryAddress,
				Appliances:      mailerAppliances,
				CreatedAt:       order.CreatedAt.Time,
				CompanySize:     order.CompanySize,
				ReferralSource:  order.ReferralSource,
				WhatsappPhone:   order.WhatsappPhone.String,
				ContactEmail:    order.Email,
			}

			mailer := mailer.NewMailer(
				cfg.SMTPServer,
				cfg.SMTPPort,
				cfg.SMTPUsername,
				cfg.SMTPPassword,
				cfg.FromEmail,
				cfg.FromName,
			)

			if err := mailer.SendConfirmationEmail(order.Email, order.FullName, emailData); err != nil {
				fmt.Printf("Failed to send confirmation email: %v\n", err)
			} else {
				fmt.Printf("Confirmation email sent to %s for order %s\n", order.Email, order.OrderNumber)
			}

			// Send admin notification (use FromEmail as admin notification address)
			_ = mailer.SendAdminNotification(cfg.FromEmail, emailData)
		} else {
			fmt.Println("Email not configured - skipping email sends")
		}
	}()

	helpers.RespondWithJson(w, http.StatusCreated, response)
}

// GetOrdersByEmailHandler retrieves all orders for a given email (for admin)
func (cfg *Config) GetOrdersByEmailHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "email query parameter is required")
		return
	}

	orders, err := cfg.DB.GetServiceOrdersByEmail(r.Context(), email)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to fetch orders")
		return
	}

	// Convert to response format
	responses := make([]ServiceOrderResponse, 0, len(orders))
	for _, order := range orders {
		mailerAppliances, _ := mailer.ParseAppliances(order.ApplianceDetails.RawMessage)
		// Convert to handler Appliance typek
		appliances := make([]Appliance, len(mailerAppliances))
		for i, a := range mailerAppliances {
			appliances[i] = Appliance{
				Name:          a.Name,
				Quantity:      a.Quantity,
				Price:         a.EstimatedCost,
				PartnerVendor: a.PartnerVendor,
			}
		}

		transportFee := 1.00
		if order.TransportFee != "" {
			fmt.Sscanf(order.TransportFee, "%f", &transportFee)
		}

		status := "pending"
		if order.Status.Valid {
			status = order.Status.String
		}

		responses = append(responses, ServiceOrderResponse{
			OrderNumber:     order.OrderNumber,
			Email:           order.Email,
			FullName:        order.FullName,
			BusinessName:    order.BusinessName,
			Phone:           order.Phone,
			WhatsappPhone:   order.WhatsappPhone.String,
			CompanySize:     order.CompanySize,
			ReferralSource:  order.ReferralSource,
			ServiceType:     order.ServiceType,
			Appliances:      appliances,
			DeliveryAddress: order.DeliveryAddress,
			TransportFee:    transportFee,
			PromoApplied:    order.PromoApplied.Bool,
			Status:          status,
			Notes:           order.Notes.String,
			CreatedAt:       order.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		})
	}

	helpers.RespondWithJson(w, http.StatusOK, responses)
}

// GetOrdersByUserHandler retrieves all orders for the currently authenticated user (requires JWT)
func (cfg *Config) GetOrdersByUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by AuthMiddleware)
	claimsVal := r.Context().Value("user")
	if claimsVal == nil {
		helpers.RespondWithError(w, http.StatusUnauthorized, "user not authenticated")
		return
	}

	// Assert as *auth.Claims
	claims, ok := claimsVal.(*auth.Claims)
	if !ok {
		helpers.RespondWithError(w, http.StatusInternalServerError, "invalid user context")
		return
	}

	// Parse user ID from claims
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "invalid user ID in token")
		return
	}

	// Fetch orders by user ID
	orders, err := cfg.DB.GetServiceOrdersByUserID(r.Context(), uuid.NullUUID{UUID: userID, Valid: true})
	if err != nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to fetch orders")
		return
	}

	// Convert to response format
	responses := make([]ServiceOrderResponse, 0, len(orders))
	for _, order := range orders {
		mailerAppliances, _ := mailer.ParseAppliances(order.ApplianceDetails.RawMessage)
		// Convert to handler Appliance type
		appliances := make([]Appliance, len(mailerAppliances))
		for i, a := range mailerAppliances {
			appliances[i] = Appliance{
				Name:          a.Name,
				Quantity:      a.Quantity,
				Price:         a.EstimatedCost,
				PartnerVendor: a.PartnerVendor,
			}
		}

		transportFee := 1.00
		if order.TransportFee != "" {
			fmt.Sscanf(order.TransportFee, "%f", &transportFee)
		}

		status := "pending"
		if order.Status.Valid {
			status = order.Status.String
		}

		responses = append(responses, ServiceOrderResponse{
			OrderNumber:     order.OrderNumber,
			Email:           order.Email,
			FullName:        order.FullName,
			BusinessName:    order.BusinessName,
			Phone:           order.Phone,
			WhatsappPhone:   order.WhatsappPhone.String,
			CompanySize:     order.CompanySize,
			ReferralSource:  order.ReferralSource,
			ServiceType:     order.ServiceType,
			Appliances:      appliances,
			DeliveryAddress: order.DeliveryAddress,
			TransportFee:    transportFee,
			PromoApplied:    order.PromoApplied.Bool,
			Status:          status,
			Notes:           order.Notes.String,
			CreatedAt:       order.CreatedAt.Time.Format("2006-01-02 15:04:05"),
		})
	}

	helpers.RespondWithJson(w, http.StatusOK, responses)
}
