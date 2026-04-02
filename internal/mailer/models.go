package mailer

import "github.com/muhammadolammi/n3xtbridge_api/shared"

type InvoiceData struct {
	InvoiceNumber string
	CustomerName  string
	CustomerEmail string
	Date          string
	Items         []shared.Item
	Discounts     []shared.Discount
	Total         float64
	PaymentLink   string
	Notes         string
}

type Mailer struct {
	Server string
	Port   string

	Username string
	Password string
}

type NewMailerParams struct {
	Server    string
	Port      string
	Username  string
	Password  string
	FromEmail string
	FromName  string
}
