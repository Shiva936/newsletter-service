package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // Postgres driver

	"github.com/Shiva936/newsletter-service/internal/config"
)

func NewPostgresDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Optional: test the connection
	if err := db.Ping(); err != nil {
		log.Printf("Failed to connect to Postgres: %v\n", err)
		return nil, err
	}

	log.Println("Connected to Postgres successfully")
	return db, nil
}
