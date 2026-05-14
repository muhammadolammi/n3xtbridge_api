package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/muhammadolammi/n3xtbridge_api/internal/handlers"
	"github.com/muhammadolammi/n3xtbridge_api/internal/mailer"
	payment "github.com/muhammadolammi/n3xtbridge_api/internal/payments"
	"github.com/redis/go-redis/v9"
)

func buildConfig() handlers.Config {
	env := os.Getenv("ENV")
	isCloudRun := os.Getenv("K_SERVICE") // This is always present on Google Cloud Run
	isProd := (env == "production" || env == "deployment" || isCloudRun != "")

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

	APIUrl := os.Getenv("API_URL")
	if APIUrl == "" {
		log.Panicln("API_URL is required to determine the base URL for the application")
	}

	// SMTP CONFIG
	smtpServer := os.Getenv("SMTP_SERVER")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")

	if smtpServer == "" || smtpPort == "" || smtpUsername == "" || smtpPassword == "" {
		log.Panic("Incomplete SMTP configuration in environment variables. Please set SMTP_SERVER, SMTP_PORT, SMTP_USERNAME, and SMTP_PASSWORD.")
	}

	// R2 CONFIG
	r2AccountId := os.Getenv("R2_ACCOUNT_ID")
	r2Bucket := os.Getenv("R2_BUCKET")
	r2PubBucket := os.Getenv("R2_BUCKET_PUB")
	r2SecretKey := os.Getenv("R2_SECRET_KEY")
	r2AccessKey := os.Getenv("R2_ACCESS_KEY")
	if r2AccountId == "" || r2Bucket == "" || r2PubBucket == "" || r2SecretKey == "" || r2AccessKey == "" {
		log.Panicln("Incomplete R2 configuration in environment variables. Please set R2_ACCOUNT_ID, R2_BUCKET,R2_BUCKET_PUB, R2_SECRET_KEY, and R2_ACCESS_KEY.")
	}

	r2Config := handlers.R2Config{
		AccountID:     r2AccountId,
		AccessKey:     r2AccessKey,
		SecretKey:     r2SecretKey,
		PrivateBucket: r2Bucket,
		PublicBucket:  r2PubBucket,
	}
	// REDIS
	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		log.Panicln("Incomplete Redis configuration in environment variables. Please set REDIS_URL")
	}

	return handlers.Config{
		DBURL:        dburl,
		IsProd:       isProd,
		ApiUrl:       APIUrl,
		ClientApiKey: clientApiKey,
		Paystack:     payment.NewPaystackService(paystackKey),
		EmailSender: mailer.NewMailer(mailer.NewMailerParams{
			Server:   smtpServer,
			Port:     smtpPort,
			Username: smtpUsername,
			Password: smtpPassword,
		}),
		PaystackSecret: paystackKey,
		R2:             &r2Config,
		RedisURL:       redisUrl,
	}

}

func loadAWSConfig(cfg *handlers.Config, r2Config *handlers.R2Config) {
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(r2Config.AccessKey, r2Config.SecretKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		log.Println("error creating aws config", err)
		return
	}
	cfg.AwsConfig = &awsConfig
	client := s3.NewFromConfig(*cfg.AwsConfig, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2.AccountID))
		o.UsePathStyle = true
	})

	presignClient := s3.NewPresignClient(client)
	cfg.PresignClient = presignClient
}

func loadRedisClient(cfg *handlers.Config) {
	option, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Panicln("invalid REDIS_URL:", err)
	}

	client := redis.NewClient(option)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		// log.Panicln("redis connection failed:", err)
		// no panic because we want the server to start even if redis is not available, but we log the error
		log.Println("redis connection failed:", err)
		return

	}
	log.Println("✅ Redis connected")

	// client.FlushDB(context.Background())

	cfg.RedisClient = client
}

func configureGoth(cfg *handlers.Config) {
	key := os.Getenv("GOTHIC_KEY")
	if key == "" {
		log.Panicln("GOTHIC_KEY is required for authentication to work")
	}

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if googleClientID == "" || googleClientSecret == "" {
		log.Panicln("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET are required for Google authentication")
	}
	baseUrl := "http://localhost:8080"
	if cfg.IsProd {
		baseUrl = cfg.ApiUrl
	}
	goth.UseProviders(
		google.New(googleClientID, googleClientSecret, baseUrl+"/auth/google/callback"),
	)
	store := sessions.NewCookieStore([]byte(key))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		Secure:   cfg.IsProd,
		// Add this line - it helps with cross-site redirects
		SameSite: http.SameSiteLaxMode,
	}
	gothic.Store = store

}
