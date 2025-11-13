package connections

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"newsletter-service/internal/config"
	"newsletter-service/internal/services/content"
	"newsletter-service/internal/services/notification"
	"newsletter-service/internal/services/subscriber"
	"newsletter-service/internal/services/topic"
)

func NewPostgresDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to PostgreSQL successfully")

	// Only run auto-migration if explicitly enabled in config
	// This is now disabled by default in favor of Goose migrations
	if cfg.AutoMigrate {
		log.Println("Auto-migration is enabled, running GORM auto-migrate...")
		if err := autoMigrate(db); err != nil {
			return nil, fmt.Errorf("auto-migration failed: %w", err)
		}
		log.Println("Auto-migration completed successfully")
	} else {
		log.Println("Auto-migration is disabled. Use Goose for schema management.")
	}

	return db, nil
}

func autoMigrate(db *gorm.DB) error {
	log.Println("Running auto-migrations...")

	err := db.AutoMigrate(
		&topic.Topic{},
		&subscriber.Subscriber{},
		&subscriber.Subscription{},
		&content.Content{},
		&notification.EmailLog{},
	)
	if err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	log.Println("Auto-migrations completed successfully")
	return nil
}
