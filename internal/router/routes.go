package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"newsletter-service/internal/config"
	"newsletter-service/internal/errors"
	"newsletter-service/internal/handlers"
	"newsletter-service/internal/logger"
	"newsletter-service/internal/router/middleware"
)

func SetupRoutes(h *handlers.Handler, cfg *config.Config, redisClient *redis.Client) *gin.Engine {
	r := gin.Default()

	// Apply global middleware
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.ValidationMiddleware())
	r.Use(logger.LoggerMiddleware())
	r.Use(errors.ErrorHandler())

	// Initialize rate limiter based on configuration
	var rateLimiter middleware.RateLimiter
	if cfg.RateLimit.Storage == "redis" && redisClient != nil {
		rateLimiter = middleware.NewRedisRateLimiter(redisClient)
	} else {
		rateLimiter = middleware.NewMemoryRateLimiter()
	}

	// Apply rate limiting middleware globally
	r.Use(middleware.RateLimitMiddleware(cfg, rateLimiter))

	// Public API routes (with basic auth)
	v1 := r.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware(cfg))
	{
		// Topic routes
		v1.GET("/topics", h.Topic.GetTopics)
		v1.POST("/topics", h.Topic.CreateTopic)
		v1.GET("/topics/:id", h.Topic.GetTopicByID)
		v1.PUT("/topics/:id", h.Topic.UpdateTopic)
		v1.DELETE("/topics/:id", h.Topic.DeleteTopic)

		// Subscriber routes
		v1.GET("/subscribers", h.Subscriber.GetSubscribers)
		v1.POST("/subscribers", h.Subscriber.CreateSubscriber)
		v1.POST("/subscribers/bulk", h.Subscriber.BulkCreateSubscribers)
		v1.PUT("/subscribers/bulk", h.Subscriber.BulkUpdateSubscribers)
		v1.DELETE("/subscribers/bulk", h.Subscriber.BulkDeleteSubscribers)
		v1.GET("/subscribers/:id", h.Subscriber.GetSubscriberByID)
		v1.PUT("/subscribers/:id", h.Subscriber.UpdateSubscriber)
		v1.DELETE("/subscribers/:id", h.Subscriber.DeleteSubscriber)

		// Subscription routes
		v1.POST("/subscriptions", h.Subscriber.CreateSubscription)
		v1.GET("/subscriptions", h.Subscriber.GetSubscriptions)
		v1.GET("/subscriptions/subscriber/:subscriber_id", h.Subscriber.GetSubscriptionsBySubscriber)
		v1.GET("/subscriptions/topic/:topic_id", h.Subscriber.GetSubscriptionsByTopic)
		v1.DELETE("/subscriptions/:id", h.Subscriber.DeleteSubscription)

		// Content routes
		v1.GET("/contents", h.Content.GetContents)
		v1.POST("/contents", h.Content.CreateContent)
		v1.GET("/contents/:id", h.Content.GetContentByID)
		v1.PUT("/contents/:id", h.Content.UpdateContent)
		v1.DELETE("/contents/:id", h.Content.DeleteContent)
		v1.POST("/contents/:id/publish", h.Content.PublishContent)

		// Email log routes
		v1.GET("/email-logs", h.Notification.GetEmailLogs)
		v1.GET("/email-logs/:id", h.Notification.GetEmailLogByID)
	}

	// Scheduler API routes (with separate authentication)
	scheduler := r.Group("/scheduler/v1")
	scheduler.Use(middleware.SchedulerAuthMiddleware(cfg))
	{
		// Notification endpoints for scheduled tasks
		scheduler.POST("/notifications/send", h.Notification.SendNotifications)
		scheduler.GET("/notifications/pending", h.Content.GetPendingNotifications)
		scheduler.POST("/notifications/retry-failed", h.Notification.RetryFailedNotifications)

		// Health check for scheduler
		scheduler.GET("/health", h.Health.SchedulerHealth)
	}

	// Health check endpoint (no auth required)
	r.GET("/health", h.Health.Health)

	// Unsubscribe endpoints (no auth required for user convenience)
	r.GET("/unsubscribe", h.Unsubscribe.UnsubscribeGet)
	r.POST("/unsubscribe", h.Unsubscribe.UnsubscribePost)
	r.POST("/subscribers/:id/resubscribe", h.Unsubscribe.Resubscribe)

	return r
}
