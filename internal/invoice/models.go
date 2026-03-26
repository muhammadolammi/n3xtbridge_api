package invoice

import (
	"time"

	"github.com/google/uuid"
)

type Item struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type Discount struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
}

type InvoiceInput struct {
	CustomerName  string     `json:"customer_name"`
	CustomerEmail string     `json:"customer_email"`
	CustomerPhone string     `json:"customer_phone"`
	Items         []Item     `json:"items"`
	Discounts     []Discount `json:"discounts"`
	Notes         string     `json:"notes"`
}

type Invoice struct {
	ID            uuid.UUID  `json:"id"`
	UserId        uuid.UUID  `json:"user_id"`
	CustomerName  string     `json:"customer_name"`
	InvoiceNumber string     `json:"invoice_number"`
	CustomerEmail string     `json:"customer_email"`
	CustomerPhone string     `json:"customer_phone"`
	Items         []Item     `json:"items"`
	Discounts     []Discount `json:"discounts"`
	Total         float64    `json:"total"`
	Notes         string     `json:"notes"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
