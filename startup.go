package main

import (
	"log"
	"os"

	"github.com/muhammadolammi/n3xtbridge_api/internal/handlers"
	"github.com/muhammadolammi/n3xtbridge_api/internal/mailer"
	payment "github.com/muhammadolammi/n3xtbridge_api/internal/payments"
)

func buildConfig() handlers.Config {
	envMode := os.Getenv("ENV")
	if envMode == "" {
		log.Fatal("cant start up with empty ENV, security risk")
	}
	dburl := os.Getenv("DB_URL")
	if dburl == "" {
		log.Println("Empty DB_URL in env")
	}

	clientApiKey := os.Getenv("CLIENT_API_KEY")
	if clientApiKey == "" {
		log.Println("Empty CLIENT_API_KEY in env")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Println("Empty JWT_SECRET in env")
	}
	paystackKey := os.Getenv("PAYSTACK_SECRET_KEY")
	if paystackKey == "" {
		log.Panic("Empty PAYSTACK_SECRET_KEY in env, server wont be able to make payment")

	}

	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	if smtpServer == "" || smtpPort == "" || smtpUsername == "" || smtpPassword == "" {
		log.Panic("Incomplete SMTP configuration in environment variables. Please set SMTP_SERVER, SMTP_PORT, SMTP_USERNAME, and SMTP_PASSWORD.")
	}

	return handlers.Config{
		DBURL:        dburl,
		ClientApiKey: clientApiKey,
		Paystack:     payment.NewPaystackService(paystackKey),
		EmailSender: mailer.NewMailer(mailer.NewMailerParams{
			Server:   smtpServer,
			Port:     smtpPort,
			Username: smtpUsername,
			Password: smtpPassword,
		}),
		PaystackSecret: paystackKey,
	}

}
