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
	DBQueries                  *database.Queries
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

type Quote struct {
	ID             uuid.UUID   `json:"id"`
	UserID         uuid.UUID   `json:"user_id"`
	QuoteRequestID uuid.UUID   `json:"quote_request_id"`
	Amount         string      `json:"amount"`
	Breakdown      []Item      `json:"breakdown"`
	Discounts      []Discount  `json:"discounts"`
	Notes          string      `json:"notes"`
	Status         QuoteStatus `json:"status"`
	ExpiresAt      time.Time   `json:"expires_at"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
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

type GetUserQuoteRequestsRow struct {
	ID      uuid.UUID `json:"id"`
	UserID  uuid.UUID `json:"user_id"`
	QuoteID uuid.UUID `json:"quote_id"`

	ServiceID   uuid.UUID          `json:"service_id"`
	Description string             `json:"description"`
	Attachments []string           `json:"attachments"`
	Status      QuoteRequestStatus `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`

	ServiceName string `json:"service_name"`
}

type GetUserQuotesWithServiceRow struct {
	ID             uuid.UUID   `json:"id"`
	UserID         uuid.UUID   `json:"user_id"`
	QuoteRequestID uuid.UUID   `json:"quote_request_id"`
	Amount         string      `json:"amount"`
	Breakdown      []Item      `json:"breakdown"`
	Discounts      []Discount  `json:"discounts"`
	Notes          string      `json:"notes"`
	Status         QuoteStatus `json:"status"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	ExpiresAt      time.Time   `json:"expires_at"`
	ServiceIcon    string      `json:"service_icon"`
	ServiceName    string      `json:"service_name"`
	ServiceID      uuid.UUID   `json:"service_id"`
}

// invoice
type Item struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}

type Discount struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type DBItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	Price       string `json:"price"`
}

type DBDiscount struct {
	Name   string `json:"name"`
	Amount string `json:"amount"`
}

type Invoice struct {
	ID            uuid.UUID    `json:"id"`
	UserId        uuid.UUID    `json:"user_id"`
	QuoteID       uuid.UUID    `json:"quote_id"`
	CustomerName  string       `json:"customer_name"`
	InvoiceNumber string       `json:"invoice_number"`
	CustomerEmail string       `json:"customer_email"`
	CustomerPhone string       `json:"customer_phone"`
	Items         []Item       `json:"items"`
	Discounts     []Discount   `json:"discounts"`
	Total         float64      `json:"total"`
	Notes         string       `json:"notes"`
	Status        string       `json:"status"`
	CreatedAt     time.Time    `json:"created_at"`
	DeletedAt     sql.NullTime `json:"deleted_at"`

	UpdatedAt time.Time `json:"updated_at"`
}
