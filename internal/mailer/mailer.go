package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/smtp"
	"time"
)

// ApplianceDetails represents an appliance item in the order
type ApplianceDetails struct {
	Name           string  `json:"name"`
	Quantity       int     `json:"quantity"`
	EstimatedCost  float64 `json:"estimated_cost"`
	PartnerVendor  string  `json:"partner_vendor,omitempty"`
}

// Mailer struct holds SMTP configuration
type Mailer struct {
	Server    string
	Port      int
	Username  string
	Password  string
	FromEmail string
	FromName  string
}

// NewMailer creates a new Mailer instance
func NewMailer(server string, port int, username, password, fromEmail, fromName string) *Mailer {
	return &Mailer{
		Server:    server,
		Port:      port,
		Username:  username,
		Password:  password,
		FromEmail: fromEmail,
		FromName:  fromName,
	}
}

// EmailData holds data for email templates
type EmailData struct {
	OrderNumber      string
	CustomerName     string
	BusinessName     string
	ServiceType      string
	TotalAmount      float64
	TransportFee     float64
	DeliveryAddress  string
	Appliances       []ApplianceDetails
	CreatedAt        time.Time
	CompanySize      string
	ReferralSource   string
	WhatsappPhone    string
	ContactEmail     string
}

// SendConfirmationEmail sends order confirmation to customer
func (m *Mailer) SendConfirmationEmail(toEmail, toName string, data EmailData) error {
	subject := fmt.Sprintf("Service Order Confirmation - %s", data.OrderNumber)

	// Parse HTML template
	tmpl, err := template.New("order_confirmation").Funcs(template.FuncMap{
		"printf": fmt.Sprintf,
	}).Parse(orderConfirmationHTML)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, struct {
		EmailData
		ToName    string
		FromName  string
		FromEmail string
	}{
		EmailData: data,
		ToName:    toName,
		FromName:  m.FromName,
		FromEmail: m.FromEmail,
	}); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	// Send email via SMTP
	return m.sendMail(toEmail, subject, body.String())
}

// SendAdminNotification sends a notification to admin about new order
func (m *Mailer) SendAdminNotification(adminEmail string, data EmailData) error {
	subject := fmt.Sprintf("New Service Order Received - %s", data.OrderNumber)

	tmpl, err := template.New("admin_notification").Funcs(template.FuncMap{
		"printf": fmt.Sprintf,
	}).Parse(adminNotificationHTML)
	if err != nil {
		return fmt.Errorf("failed to parse admin template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, struct {
		EmailData
	}{
		EmailData: data,
	}); err != nil {
		return fmt.Errorf("failed to execute admin template: %w", err)
	}

	return m.sendMail(adminEmail, subject, body.String())
}

// sendMail handles the actual SMTP sending
func (m *Mailer) sendMail(to, subject, body string) error {
	// Set up authentication
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Server)

	// Build headers
	headers := map[string]string{
		"From":         fmt.Sprintf("%s <%s>", m.FromName, m.FromEmail),
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=UTF-8",
	}

	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", m.Server, m.Port)
	if err := smtp.SendMail(addr, auth, m.FromEmail, []string{to}, msg.Bytes()); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// ParseAppliances converts JSON bytes to ApplianceDetails slice
func ParseAppliances(rawMsg []byte) ([]ApplianceDetails, error) {
	if len(rawMsg) == 0 || string(rawMsg) == "null" {
		return []ApplianceDetails{}, nil
	}

	var appliances []ApplianceDetails
	if err := json.Unmarshal(rawMsg, &appliances); err != nil {
		return nil, fmt.Errorf("failed to parse appliances: %w", err)
	}

	return appliances, nil
}

// HTML Email Templates

const orderConfirmationHTML = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .order-details { background-color: #f9f9f9; padding: 15px; margin: 15px 0; border-left: 4px solid #4CAF50; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #ddd; font-size: 12px; color: #777; }
        table { width: 100%; border-collapse: collapse; margin: 10px 0; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Order Confirmed!</h1>
            <p>Thank you for choosing n3xtbridge</p>
        </div>

        <div class="content">
            <p>Dear {{.ToName}},</p>
            <p>Your service order has been successfully registered. Below are your order details:</p>

            <div class="order-details">
                <h3>Order Information</h3>
                <p><strong>Order Number:</strong> {{.OrderNumber}}</p>
                <p><strong>Service Type:</strong> {{.ServiceType}}</p>
                <p><strong>Order Date:</strong> {{printf "%s" (.CreatedAt.Format "January 2, 2006 15:04")}}</p>
                <p><strong>Status:</strong> <span style="color: #4CAF50; font-weight: bold;">Pending</span></p>
            </div>

            <div class="order-details">
                <h3>Business Information</h3>
                <p><strong>Business Name:</strong> {{.BusinessName}}</p>
                <p><strong>Company Size:</strong> {{.CompanySize}}</p>
                <p><strong>Delivery Address:</strong> {{.DeliveryAddress}}</p>
            </div>

            <h3>Promotion Details</h3>
            <ul>
                <li><strong>Installation Fee:</strong> $1.00 (Promotional Offer)</li>
                <li><strong>Transport Fee:</strong> ${{printf "%.2f" .TransportFee}}</li>
                <li>You only pay for appliances from our certified partners</li>
            </ul>

            <h3>Appliance List</h3>
            <table>
                <thead>
                    <tr>
                        <th>Item</th>
                        <th>Quantity</th>
                        <th>Est. Cost</th>
                        {{if .PartnerVendor}}<th>Partner</th>{{end}}
                    </tr>
                </thead>
                <tbody>
                {{range .Appliances}}
                    <tr>
                        <td>{{.Name}}</td>
                        <td>{{.Quantity}}</td>
                        <td>${{printf "%.2f" .EstimatedCost}}</td>
                        {{if .PartnerVendor}}<td>{{.PartnerVendor}}</td>{{end}}
                    </tr>
                {{end}}
                </tbody>
            </table>

            <p><strong>Next Steps:</strong></p>
            <ol>
                <li>Our team will contact you within 24-48 hours to confirm your order.</li>
                <li>We'll help you purchase the required appliances from our partners.</li>
                <li>Installation will be scheduled at your convenience.</li>
                <li>Bring your WhatsApp number for easy communication: {{.WhatsappPhone}}</li>
            </ol>

            <p>For any questions, reply to this email or contact us directly.</p>

            <p>Best regards,<br>
            <strong>{{.FromName}}</strong><br>
            <a href="https://n3xtbridge.com">n3xtbridge.com</a></p>
        </div>

        <div class="footer">
            <p>© 2025 n3xtbridge. All rights reserved.</p>
            <p>You received this email because you placed an order with us.</p>
        </div>
    </div>
</body>
</html>`

const adminNotificationHTML = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 800px; margin: 0 auto; padding: 20px; }
        .header { background-color: #2196F3; color: white; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .order-details { background-color: #f9f9f9; padding: 15px; margin: 15px 0; border-left: 4px solid #2196F3; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>New Service Order!</h1>
            <p>Action Required: Contact Customer</p>
        </div>

        <div class="content">
            <p>A new service order has been placed on n3xtbridge.com</p>

            <div class="order-details">
                <h3>Order Summary</h3>
                <p><strong>Order Number:</strong> {{.OrderNumber}}</p>
                <p><strong>Customer:</strong> {{.CustomerName}}</p>
                <p><strong>Business:</strong> {{.BusinessName}}</p>
                <p><strong>Service Type:</strong> {{.ServiceType}}</p>
                <p><strong>Email:</strong> {{.ContactEmail}}</p>
                <p><strong>Phone:</strong> {{.ContactEmail}}</p>
                {{if .WhatsappPhone}}<p><strong>WhatsApp:</strong> {{.WhatsappPhone}}</p>{{end}}
                <p><strong>Company Size:</strong> {{.CompanySize}}</p>
                <p><strong>Referral Source:</strong> {{.ReferralSource}}</p>
            </div>

            <div class="order-details">
                <h3>Delivery Information</h3>
                <p><strong>Address:</strong> {{.DeliveryAddress}}</p>
                <p><strong>Transport Fee:</strong> ${{printf "%.2f" .TransportFee}}</p>
            </div>

            <h3>Appliance Requirements</h3>
            <table border="1" cellpadding="5" cellspacing="0">
                <tr>
                    <th>Item</th>
                    <th>Quantity</th>
                    <th>Est. Cost</th>
                    {{if .PartnerVendor}}<th>Partner</th>{{end}}
                </tr>
                {{range .Appliances}}
                <tr>
                    <td>{{.Name}}</td>
                    <td>{{.Quantity}}</td>
                    <td>${{printf "%.2f" .EstimatedCost}}</td>
                    {{if .PartnerVendor}}<td>{{.PartnerVendor}}</td>{{end}}
                </tr>
                {{end}}
            </table>

            <p><strong>Total Appliances Cost (Est.):</strong> ${{printf "%.2f" .TotalAmount}}</p>
            <p><strong>Promo Applied:</strong> Yes (Installation fee = $1.00)</p>

            <hr>
            <p><strong>Customer Notes:</strong> {{.Notes}}</p>
        </div>

        <div class="footer">
            <p>Please contact the customer within 24 hours to proceed.</p>
            <p>n3xtbridge Admin System</p>
        </div>
    </div>
</body>
</html>`
