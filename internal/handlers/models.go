package handlers

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	goauth "github.com/muhammadolammi/goauth/pkg/auth"
	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
)

type Config struct {
	DBURL                      string
	DB                         *database.Queries
	DBConn                     *sql.DB
	ClientApiKey               string
	JwtSecret                  string
	RateLimit                  int
	RefreshTokenEXpirationTime int //in minute
	AcessTokenEXpirationTime   int //in minute
	// Email configuration (Zoho SMTP)
	SMTPServer   string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string
	AuthService  *goauth.AuthService
}

type User struct {
	ID          uuid.UUID      `json:"id"`
	Email       string         `json:"email"`
	FirstName   string         `json:"first_name"`
	LastName    string         `json:"last_name"`
	PhoneNumber sql.NullString `json:"phone_number"`
	Address     string         `json:"address"`
	Country     string         `json:"country"`
	State       string         `json:"state"`
	Role        string         `json:"role"`
	CreatedAt   sql.NullTime   `json:"created_at"`
}

// Appliance represents an item to be purchased from a partner
type Appliance struct {
	Name          string  `json:"name"`
	Quantity      int     `json:"quantity"`
	Price         float64 `json:"price"` // Estimated cost from partner
	PartnerVendor string  `json:"partner_vendor,omitempty"`
}

// ServiceOrderRequest represents the request to create a service order
type ServiceOrderRequest struct {
	Email           string      `json:"email"`
	FullName        string      `json:"full_name"`
	BusinessName    string      `json:"business_name"`
	Phone           string      `json:"phone"`
	WhatsappPhone   string      `json:"whatsapp_phone,omitempty"`
	CompanySize     string      `json:"company_size"` // sole_proprietor, 1-10, 11-50, 51-200, 201-1000, 1000+
	ReferralSource  string      `json:"referral_source"`
	ServiceType     string      `json:"service_type"` // solar, security, both
	Appliances      []Appliance `json:"appliances"`
	DeliveryAddress string      `json:"delivery_address"`
	DeliveryCity    string      `json:"delivery_city"`
	DeliveryState   string      `json:"delivery_state"`
	DeliveryCountry string      `json:"delivery_country"`

	Notes   string `json:"notes,omitempty"`
	IsPromo bool   `json:"is_promo"`
}

// ServiceOrderResponse represents the response after creating an order
type ServiceOrderResponse struct {
	OrderNumber     string      `json:"order_number"`
	Email           string      `json:"email"`
	FullName        string      `json:"full_name"`
	BusinessName    string      `json:"business_name"`
	Phone           string      `json:"phone"`
	WhatsappPhone   string      `json:"whatsapp_phone,omitempty"`
	CompanySize     string      `json:"company_size"`
	ReferralSource  string      `json:"referral_source"`
	ServiceType     string      `json:"service_type"`
	Appliances      []Appliance `json:"appliances"`
	DeliveryAddress string      `json:"delivery_address"`
	TransportFee    float64     `json:"transport_fee"`
	PromoApplied    bool        `json:"promo_applied"`
	Status          string      `json:"status"`
	Notes           string      `json:"notes,omitempty"`
	CreatedAt       string      `json:"created_at"`
	Message         string      `json:"message,omitempty"`
}

type Service struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	IsActive    bool      `json:"is_active"`
	IsFeatured  bool      `json:"is_featured"`
	Icon        string    `json:"icon"`
	Image       string    `json:"image"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
}

type QuoteRequestStatus string

const (
	QuoteRequestStatusPending   QuoteRequestStatus = "pending"
	QuoteRequestStatusReviewing QuoteRequestStatus = "reviewing"
	QuoteRequestStatusQuoted    QuoteRequestStatus = "quoted"
	QuoteRequestStatusRejected  QuoteRequestStatus = "rejected"
)

type QuoteStatus string

const (
	QuoteStatusDraft    QuoteStatus = "draft"
	QuoteStatusSent     QuoteStatus = "sent"
	QuoteStatusAccepted QuoteStatus = "accepted"
	QuoteStatusDeclined QuoteStatus = "declined"
	QuoteStatusExpired  QuoteStatus = "expired"
)

type QuoteBreakdown struct {
	Name        string `json:"name"`
	Cost        string `json:"cost"`
	Description string `json:"description"`
}
type Quote struct {
	ID             uuid.UUID        `json:"id"`
	QuoteRequestID uuid.UUID        `json:"quote_request_id"`
	Amount         string           `json:"amount"`
	Breakdown      []QuoteBreakdown `json:"breakdown"`
	Notes          string           `json:"notes"`
	Status         QuoteStatus      `json:"status"`
	ExpiresAt      time.Time        `json:"exire_at"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

type QuoteRequest struct {
	ID          uuid.UUID          `json:"id"`
	UserID      uuid.UUID          `json:"user_id"`
	ServiceID   uuid.UUID          `json:"service_id"`
	Description string             `json:"description"`
	Attachments []string           `json:"attachments"`
	Status      QuoteRequestStatus `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type GetQuoteRequestsRow struct {
	ID          uuid.UUID          `json:"id"`
	UserID      uuid.UUID          `json:"user_id"`
	ServiceID   uuid.UUID          `json:"service_id"`
	Description string             `json:"description"`
	Attachments []string           `json:"attachments"`
	Status      QuoteRequestStatus `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	UserEmail   string             `json:"user_email"`
	UserName    string             `json:"user_name"`
	ServiceName string             `json:"service_name"`
}
