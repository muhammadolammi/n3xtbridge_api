package handlers

import (
	"database/sql"

	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
)

type Config struct {
	DBURL  string
	DB     *database.Queries
	DBConn *sql.DB
	// JwtKey       string
	ClientApiKey string
	Port         string
	// RABBITMQUrl                string
	// RabbitConn                 *amqp.Connection
	// RabbitChan                 *amqp.Channel
	// RefreshTokenEXpirationTime int //in minute
	// AcessTokenEXpirationTime   int //in minute
	RateLimit int
	// HttpClient *http.Client // this should be used for all internal and external http communication
}
