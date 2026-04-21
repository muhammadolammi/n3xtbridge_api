package handlers

import (
	"database/sql"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/muhammadolammi/goauth"
	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
	"github.com/muhammadolammi/n3xtbridge_api/internal/mailer"
	payment "github.com/muhammadolammi/n3xtbridge_api/internal/payments"
	"github.com/muhammadolammi/n3xtbridge_api/shared"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	DBURL        string
	DBQueries    *database.Queries
	DBConn       *sql.DB
	ClientApiKey string
	JwtSecret    string
	// RateLimit      int
	Paystack       *payment.PaystackService
	PaystackSecret string
	IsProd         bool
	AwsConfig      *aws.Config
	PresignClient  *s3.PresignClient

	// Email configuration (Zoho SMTP)
	EmailSender *mailer.Mailer
	AuthService *goauth.AuthService
	// R2 CONFIG
	R2 *R2Config
	// Redis
	RedisClient *redis.Client
	RedisURL    string
}
type R2Config struct {
	AccountID     string
	PublicBucket  string
	PrivateBucket string

	AccessKey string
	SecretKey string
}
type PresignResponse struct {
	UploadURL  string `json:"upload_url"`
	ObjectKey  string `json:"object_key"`
	Expiration int64  `json:"expiration"`
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
	Image       string    `json:"image"`
	Tags        []string  `json:"tags"`
	PromoIDs    []string  `json:"promo_ids"`
	CreatedAt   time.Time `json:"created_at"`
	MinPrice    string    `json:"min_price"`
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
	ID             uuid.UUID         `json:"id"`
	UserID         uuid.UUID         `json:"user_id"`
	QuoteRequestID uuid.UUID         `json:"quote_request_id"`
	Amount         string            `json:"amount"`
	Breakdown      []shared.Item     `json:"breakdown"`
	Discounts      []shared.Discount `json:"discounts"`
	PromoIDs       []string          `json:"promo_ids"`
	Notes          string            `json:"notes"`
	Status         QuoteStatus       `json:"status"`
	ExpiresAt      time.Time         `json:"expires_at"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type QuoteRequest struct {
	ID          uuid.UUID          `json:"id"`
	UserID      uuid.UUID          `json:"user_id"`
	ServiceID   uuid.UUID          `json:"service_id"`
	Description string             `json:"description"`
	Attachments []string           `json:"attachments"`
	PromoIDs    []string           `json:"promo_ids"`
	Status      QuoteRequestStatus `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	VnR2Key     string             `json:"vn_key"`
	VideoKey    string             `json:"video_key"`
}

type GetQuoteRequestsRow struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	ServiceID   uuid.UUID `json:"service_id"`
	Description string    `json:"description"`
	Attachments []string  `json:"attachments"`
	PromoIDs    []string  `json:"promo_ids"`

	Status      QuoteRequestStatus `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	UserEmail   string             `json:"user_email"`
	UserName    string             `json:"user_name"`
	ServiceName string             `json:"service_name"`
	VnR2Key     string             `json:"vn_key"`
	VideoKey    string             `json:"video_key"`
}

type GetUserQuoteRequestsRow struct {
	ID       uuid.UUID `json:"id"`
	UserID   uuid.UUID `json:"user_id"`
	QuoteID  uuid.UUID `json:"quote_id"`
	PromoIDs []string  `json:"promo_ids"`

	ServiceID   uuid.UUID          `json:"service_id"`
	Description string             `json:"description"`
	Attachments []string           `json:"attachments"`
	Status      QuoteRequestStatus `json:"status"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`

	ServiceName string `json:"service_name"`
	VnR2Key     string `json:"vn_key"`
	VideoKey    string `json:"video_key"`
}

type GetUserQuotesWithServiceRow struct {
	ID             uuid.UUID         `json:"id"`
	UserID         uuid.UUID         `json:"user_id"`
	QuoteRequestID uuid.UUID         `json:"quote_request_id"`
	Amount         string            `json:"amount"`
	Breakdown      []shared.Item     `json:"breakdown"`
	Discounts      []shared.Discount `json:"discounts"`
	PromoIDs       []string          `json:"promo_ids"`

	Notes       string      `json:"notes"`
	Status      QuoteStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	ExpiresAt   time.Time   `json:"expires_at"`
	ServiceName string      `json:"service_name"`
	ServiceID   uuid.UUID   `json:"service_id"`
}

// invoice

type DBItem struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Price       string  `json:"price"`
}

type DBDiscount struct {
	Name        string `json:"name"`
	Amount      string `json:"amount"`
	Type        string `json:"type"`
	Description string `json:"description"`
	ItemName    string `json:"item_name"`
}

type Invoice struct {
	ID             uuid.UUID         `json:"id"`
	UserId         uuid.UUID         `json:"user_id"`
	QuoteID        uuid.UUID         `json:"quote_id"`
	CustomerName   string            `json:"customer_name"`
	InvoiceNumber  string            `json:"invoice_number"`
	CustomerEmail  string            `json:"customer_email"`
	CustomerPhone  string            `json:"customer_phone"`
	Items          []shared.Item     `json:"items"`
	Discounts      []shared.Discount `json:"discounts"`
	Total          float64           `json:"total"`
	Notes          string            `json:"notes"`
	Status         string            `json:"status"`
	CreatedAt      time.Time         `json:"created_at"`
	DeletedAt      sql.NullTime      `json:"deleted_at"`
	PaymentToken   string            `json:"payment_token"`
	ReminderSentAt time.Time         `json:"reminder_sent_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type Promotion struct {
	ID          uuid.UUID         `json:"id"`
	Code        string            `json:"code"`
	Name        string            `json:"name"`
	Description sql.NullString    `json:"description"`
	Breakdown   []shared.Discount `json:"breakdown"`
	IsActive    bool              `json:"is_active"`
	StartsAt    time.Time         `json:"starts_at"`
	ExpiresAt   time.Time         `json:"expires_at"`
	CreatedAt   time.Time         `json:"created_at"`
	ServiceID   string            `json:"service_id"`
	Attachments []string          `json:"attachments"`
}
