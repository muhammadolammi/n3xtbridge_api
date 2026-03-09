package main

import (
	"log"
	"os"

	"github.com/muhammadolammi/n3xtbridge_api/internal/handlers"
)

func buildConfig() handlers.Config {
	dburl := os.Getenv("DB_URL")
	if dburl == "" {
		log.Println("Empty DB_URL in env")
	}
	port := os.Getenv("PORT")
	if port == "" {
		log.Println("Empty PORT in env")
	}
	clientApiKey := os.Getenv("CLIENT_API_KEY")
	if clientApiKey == "" {
		log.Println("Empty CLIENT_API_KEY in env")
	}

	return handlers.Config{
		DBURL:        dburl,
		Port:         port,
		ClientApiKey: clientApiKey,
	}

}
