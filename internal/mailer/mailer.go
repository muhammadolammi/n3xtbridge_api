package mailer

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/smtp"
	"net/textproto"

	"github.com/muhammadolammi/email"
)

//go:embed templates/*
var templateFS embed.FS

func NewMailer(params NewMailerParams) *Mailer {
	return &Mailer{
		Server:   params.Server,
		Port:     params.Port,
		Username: params.Username,
		Password: params.Password,
	}
}

func (m *Mailer) SendInvoice(data InvoiceData) error {
	subject := fmt.Sprintf("Invoice #%s from N3xtBridge", data.InvoiceNumber)

	// Parse HTML template
	tmpl, err := template.ParseFS(templateFS, "templates/invoice.html")

	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}
	// log.Println(body.String())
	fromEmail := "sales@n3xtbridge.com"
	fromName := "N3xtBridge Sales"

	return m.SendMail(data.CustomerEmail, fromName, fromEmail, subject, body.String())
}

// sendMail handles the actual SMTP sending
func (m *Mailer) SendMail(to, fromName, fromEmail, subject, body string) error {
	fromFormatted := fmt.Sprintf(`"%s" <%s>`, fromName, fromEmail)
	email := email.Email{
		From:    fromFormatted,
		To:      []string{to},
		Subject: subject,
		HTML:    []byte(body),
		Headers: textproto.MIMEHeader{},
	}
	email.Headers.Set("From", fromFormatted)
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Server)
	err := email.Send(m.Server+":"+m.Port, auth)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil

}
