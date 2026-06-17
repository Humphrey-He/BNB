package app

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

type EnvConfig struct {
	PostgresHost string
	PostgresPort string
	PostgresDB   string
	PostgresUser string
	PostgresPass string
	NATSURL      string
}

func LoadEnvConfig() EnvConfig {
	return EnvConfig{
		PostgresHost: getenv("POSTGRES_HOST", "127.0.0.1"),
		PostgresPort: getenv("POSTGRES_PORT", "5432"),
		PostgresDB:   getenv("POSTGRES_DB", "asset_platform"),
		PostgresUser: getenv("POSTGRES_USER", "platform"),
		PostgresPass: os.Getenv("POSTGRES_PASSWORD"),
		NATSURL:      getenv("NATS_URL", "nats://127.0.0.1:4222"),
	}
}

func OpenPostgres(cfg EnvConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDB,
		cfg.PostgresUser,
		cfg.PostgresPass,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func ConnectNATS(cfg EnvConfig) (*nats.Conn, error) {
	return nats.Connect(cfg.NATSURL)
}

func NewLogger(service string) *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})).With("service", service)
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
