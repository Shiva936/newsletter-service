package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"newsletter-service/internal/config"
)

// TokenBucket represents a leaky bucket for rate limiting
type TokenBucket struct {
	Capacity   int           `json:"capacity"`    // Maximum tokens
	Tokens     int           `json:"tokens"`      // Current tokens
	RefillSize int           `json:"refill_size"` // Tokens added per refill
	RefillRate time.Duration `json:"refill_rate"` // How often to refill
	LastRefill time.Time     `json:"last_refill"` // Last refill time
}

// RateLimiter interface for different storage backends
type RateLimiter interface {
	Allow(key string, rule config.RateLimitRule) (bool, error)
	CleanupExpired() error
}

// RedisRateLimiter implements RateLimiter using Redis
type RedisRateLimiter struct {
	client *redis.Client
}

// MemoryRateLimiter implements RateLimiter using in-memory storage
type MemoryRateLimiter struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
	}
}

// NewMemoryRateLimiter creates a new memory-based rate limiter
func NewMemoryRateLimiter() *MemoryRateLimiter {
	return &MemoryRateLimiter{
		buckets: make(map[string]*TokenBucket),
	}
}

// Allow checks if a request should be allowed based on rate limiting rules
func (r *RedisRateLimiter) Allow(key string, rule config.RateLimitRule) (bool, error) {
	now := time.Now()
	bucketKey := fmt.Sprintf("rate_limit:%s", key)

	// Get existing bucket from Redis
	data, err := r.client.Get(r.client.Context(), bucketKey).Result()
	var bucket *TokenBucket

	if err == redis.Nil {
		// Create new bucket
		bucket = &TokenBucket{
			Capacity:   rule.BucketSize,
			Tokens:     rule.BucketSize - 1, // Consume one token immediately
			RefillSize: rule.RefillSize,
			RefillRate: rule.RefillDuration,
			LastRefill: now,
		}
	} else if err != nil {
		return false, err
	} else {
		// Parse existing bucket
		bucket = &TokenBucket{}
		if err := json.Unmarshal([]byte(data), bucket); err != nil {
			return false, err
		}

		// Refill tokens if enough time has passed
		r.refillTokens(bucket, now)

		// Check if we have tokens available
		if bucket.Tokens <= 0 {
			// Save updated bucket back to Redis
			r.saveBucket(bucketKey, bucket)
			return false, nil
		}

		// Consume a token
		bucket.Tokens--
	}

	// Save updated bucket back to Redis with expiration
	if err := r.saveBucket(bucketKey, bucket); err != nil {
		return false, err
	}

	return true, nil
}

// CleanupExpired removes expired buckets (handled automatically by Redis TTL)
func (r *RedisRateLimiter) CleanupExpired() error {
	// Redis handles expiration automatically, nothing to do
	return nil
}

// Allow checks if a request should be allowed based on rate limiting rules
func (m *MemoryRateLimiter) Allow(key string, rule config.RateLimitRule) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	bucket, exists := m.buckets[key]

	if !exists {
		// Create new bucket
		bucket = &TokenBucket{
			Capacity:   rule.BucketSize,
			Tokens:     rule.BucketSize - 1, // Consume one token immediately
			RefillSize: rule.RefillSize,
			RefillRate: rule.RefillDuration,
			LastRefill: now,
		}
		m.buckets[key] = bucket
		return true, nil
	}

	// Refill tokens if enough time has passed
	m.refillTokens(bucket, now)

	// Check if we have tokens available
	if bucket.Tokens <= 0 {
		return false, nil
	}

	// Consume a token
	bucket.Tokens--
	return true, nil
}

// CleanupExpired removes expired buckets from memory
func (m *MemoryRateLimiter) CleanupExpired() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	expiry := 24 * time.Hour // Remove buckets older than 24 hours

	for key, bucket := range m.buckets {
		if now.Sub(bucket.LastRefill) > expiry {
			delete(m.buckets, key)
		}
	}

	return nil
}

// Helper methods

func (r *RedisRateLimiter) refillTokens(bucket *TokenBucket, now time.Time) {
	// Calculate how many refill periods have passed
	elapsed := now.Sub(bucket.LastRefill)
	refillCount := int(elapsed / bucket.RefillRate)

	if refillCount > 0 {
		// Add tokens up to capacity
		newTokens := bucket.Tokens + (refillCount * bucket.RefillSize)
		if newTokens > bucket.Capacity {
			newTokens = bucket.Capacity
		}
		bucket.Tokens = newTokens
		bucket.LastRefill = bucket.LastRefill.Add(time.Duration(refillCount) * bucket.RefillRate)
	}
}

func (m *MemoryRateLimiter) refillTokens(bucket *TokenBucket, now time.Time) {
	// Calculate how many refill periods have passed
	elapsed := now.Sub(bucket.LastRefill)
	refillCount := int(elapsed / bucket.RefillRate)

	if refillCount > 0 {
		// Add tokens up to capacity
		newTokens := bucket.Tokens + (refillCount * bucket.RefillSize)
		if newTokens > bucket.Capacity {
			newTokens = bucket.Capacity
		}
		bucket.Tokens = newTokens
		bucket.LastRefill = bucket.LastRefill.Add(time.Duration(refillCount) * bucket.RefillRate)
	}
}

func (r *RedisRateLimiter) saveBucket(key string, bucket *TokenBucket) error {
	data, err := json.Marshal(bucket)
	if err != nil {
		return err
	}

	// Set with 1 hour TTL to prevent memory leaks
	return r.client.Set(r.client.Context(), key, data, time.Hour).Err()
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(cfg *config.Config, limiter RateLimiter) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Skip if rate limiting is disabled
		if !cfg.RateLimit.Enabled {
			c.Next()
			return
		}

		// Determine which rule to apply
		rule := cfg.RateLimit.DefaultRule

		// Check for route-specific rules
		path := c.Request.URL.Path
		method := c.Request.Method
		routeKey := fmt.Sprintf("%s:%s", method, path)

		if routeRule, exists := cfg.RateLimit.Routes[routeKey]; exists && routeRule.Enabled {
			rule = routeRule
		}

		// Skip if rule is disabled
		if !rule.Enabled {
			c.Next()
			return
		}

		// Generate identifier based on rule configuration
		var identifier string
		switch rule.IdentifyBy {
		case "api_key":
			apiKey := c.GetHeader("X-API-Key")
			if apiKey == "" {
				apiKey = c.GetHeader("Authorization")
				if strings.HasPrefix(apiKey, "Bearer ") {
					apiKey = strings.TrimPrefix(apiKey, "Bearer ")
				}
			}
			if apiKey == "" {
				apiKey = "anonymous"
			}
			identifier = fmt.Sprintf("api_key:%s", apiKey)
		case "ip":
			fallthrough
		default:
			// Default to IP-based rate limiting
			clientIP := c.ClientIP()
			identifier = fmt.Sprintf("ip:%s", clientIP)
		}

		// Check if request is allowed
		allowed, err := limiter.Allow(identifier, rule)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal server error",
				"message": "Rate limiting service unavailable",
			})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     "Too many requests. Please try again later.",
				"retry_after": rule.RefillDuration.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	})
}
