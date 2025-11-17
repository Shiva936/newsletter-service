package main

import (
	"context"
	"log"
	"time"

	"newsletter-service/internal/config"
	"newsletter-service/internal/connections"
	"newsletter-service/internal/schedulers"
	"newsletter-service/internal/services/content"
	"newsletter-service/internal/services/notification"
	"newsletter-service/internal/services/subscriber"
	"newsletter-service/internal/services/topic"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := connections.NewPostgresDB(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Get underlying sql.DB for connection management
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	defer sqlDB.Close()

	// Connect to Redis
	redisClient, err := connections.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize repositories
	contentRepo := content.NewRepository(db)
	subscriberRepo := subscriber.NewRepository(db)
	topicRepo := topic.NewRepository(db)

	// Initialize services
	topicService := topic.NewService(topicRepo)
	contentService := content.NewService(contentRepo)
	subscriberService := subscriber.NewServiceWithTopic(subscriberRepo, topicService)

	// Initialize notification service with multi-provider support
	notificationService, err := notification.NewServiceWithProviders(db, contentService, subscriberService, cfg)
	if err != nil {
		log.Fatalf("Failed to create notification service with providers: %v", err)
	}
	log.Printf("Initialized notification service with multi-provider support")

	// Initialize scheduler
	scheduler := schedulers.NewNotificationScheduler(contentService, notificationService)

	// Start worker
	log.Println("Worker started, checking for pending notifications every minute...")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := scheduler.ProcessPendingNotifications(context.Background()); err != nil {
				log.Printf("Error processing notifications: %v", err)
			}
		}
	}
}
