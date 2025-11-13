package main

import (
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"

	"newsletter-service/internal/config"
	"newsletter-service/internal/connections"
	"newsletter-service/internal/handlers"
	"newsletter-service/internal/router"
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

	// Connect to PostgreSQL
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
	redisAddr := cfg.Redis.Host + ":" + strconv.Itoa(cfg.Redis.Port)
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test Redis connection
	_, err = redisClient.Ping(redisClient.Context()).Result()
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Rate limiting will fall back to memory storage")
		redisClient = nil
	} else {
		log.Println("Connected to Redis successfully")
	}

	// Initialize repositories
	topicRepo := topic.NewRepository(db)
	subscriberRepo := subscriber.NewRepository(db)
	contentRepo := content.NewRepository(db)

	// Initialize services
	topicService := topic.NewService(topicRepo)
	subscriberService := subscriber.NewServiceWithTopic(subscriberRepo, topicService)
	contentService := content.NewService(contentRepo)

	// Initialize notification service (without email provider - web API doesn't send emails directly)
	// Email sending is handled by the worker process
	notificationService := notification.NewService(db, contentService, subscriberService)

	// Initialize handlers
	handler := handlers.NewHandler(topicService, subscriberService, contentService, notificationService)

	// Setup routes
	router := router.SetupRoutes(handler, cfg, redisClient)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Starting newsletter service on port %s...", port)
	log.Printf("Env: %s", cfg.Env)
	log.Printf("Rate limiting: %v (storage: %s)", cfg.RateLimit.Enabled, cfg.RateLimit.Storage)
	log.Printf("Auto-migration: %v", cfg.Database.AutoMigrate)
	log.Printf("Scheduler auth: %v", cfg.Scheduler.Enabled)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
