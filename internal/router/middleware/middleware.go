package middleware

import (
	"net/http"
	"strings"

	"newsletter-service/internal/config"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware provides basic authentication
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		cfg.Auth.Username: cfg.Auth.Password,
	})
}

// SchedulerAuthMiddleware provides separate authentication for scheduler APIs
func SchedulerAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip if scheduler auth is disabled
		if !cfg.Scheduler.Enabled {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Service unavailable",
				"message": "Scheduler service is disabled",
			})
			c.Abort()
			return
		}

		// Check for basic auth header
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Basic ") {
			c.Header("WWW-Authenticate", "Basic realm=\"Scheduler API\"")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Scheduler API requires basic authentication",
			})
			c.Abort()
			return
		}

		// Use Gin's built-in basic auth with scheduler credentials
		accounts := gin.Accounts{
			cfg.Scheduler.Username: cfg.Scheduler.Password,
		}

		basicAuth := gin.BasicAuth(accounts)
		basicAuth(c)
	})
}

// CORSMiddleware adds CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-API-Key")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
