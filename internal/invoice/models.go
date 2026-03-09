package invoice

import "time"

type Item struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type InvoiceInput struct {
	CustomerName  string  `json:"customer_name"`
	CustomerEmail string  `json:"customer_email"`
	CustomerPhone string  `json:"customer_phone"`
	Items         []Item  `json:"items"`
	Discount      float64 `json:"discount"`
	Notes         string  `json:"notes"`
}

type Invoice struct {
	ID            string `json:"id"`
	CustomerName  string `json:"customer_name"`
	InvoiceNumber string `json:"invoice_number"`

	CustomerEmail string    `json:"customer_email"`
	CustomerPhone string    `json:"customer_phone"`
	Items         []Item    `json:"items"`
	Discount      float64   `json:"discount"`
	Total         float64   `json:"total"`
	CreatedAt     time.Time `json:"created_at"`
}
