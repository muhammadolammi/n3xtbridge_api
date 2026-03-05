package infra

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/muhammadolammi/n3xtbridge_api/internal/database"
	"github.com/muhammadolammi/n3xtbridge_api/internal/handlers"
)

func ConnectDB(ctx context.Context, cfg *handlers.Config) {
	backoff := time.Second
	const maxBackoff = 30 * time.Second

	for {
		select {
		case <-ctx.Done():
			log.Println("DB connect cancelled")
			return
		default:
		}

		db, err := sql.Open("postgres", cfg.DBURL)
		if err != nil {
			log.Println("DB open error:", err)
			sleepBackoff(&backoff, maxBackoff)
			continue
		}

		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = db.PingContext(pingCtx)
		cancel()

		if err != nil {
			log.Println("DB not ready:", err)
			db.Close()
			sleepBackoff(&backoff, maxBackoff)
			continue
		}

		// Pool tuning (Cloud Run friendly)
		db.SetMaxOpenConns(5)
		db.SetMaxIdleConns(2)
		db.SetConnMaxLifetime(5 * time.Minute)

		cfg.DB = database.New(db)
		log.Println("✅ Postgres connected")
		return
	}
}

func sleepBackoff(b *time.Duration, max time.Duration) {
	time.Sleep(*b)
	*b *= 2
	if *b > max {
		*b = max
	}
}
