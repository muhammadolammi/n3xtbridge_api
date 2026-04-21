package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/muhammadolammi/n3xtbridge_api/internal/infra"
)

func main() {

	_ = godotenv.Load()

	cfg := buildConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// // Connect services in goroutines
	// go infra.ConnectRabbit(ctx, &cfg)
	go loadAWSConfig(&cfg, cfg.R2)
	// go infra.ConnectPubSub(ctx, &cfg)
	go loadRedisClient(&cfg)

	infra.ConnectDB(ctx, &cfg)

	// Start your server in goroutine
	go func() {
		server(&cfg)
	}()

	// Wait for Cloud Run shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)

	<-stop
	log.Println("SIGTERM received: shutting down gracefully...")

	_, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Close all connections safely
	if cfg.DBConn != nil {
		cfg.DBConn.Close()
		log.Println("Postgres connection closed")
	}

	log.Println("All resources cleaned up. Exiting...")
}
