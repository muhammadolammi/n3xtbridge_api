package main

import (
	"log"
	"os"

	"github.com/muhammadolammi/n3xtbridge_api/internal/handlers"
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
	return handlers.Config{
		DBURL:          dburl,
		ClientApiKey:   clientApiKey,
		Paystack:       payment.NewPaystackService(paystackKey),
		PaystackSecret: paystackKey,
	}

}
