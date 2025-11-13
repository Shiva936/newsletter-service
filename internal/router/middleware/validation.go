package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"newsletter-service/internal/constants"
)

// ValidationMiddleware adds input validation to gin context
func ValidationMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Add validator instance to context
		c.Set("validator", validator.New())
		c.Next()
	})
}

// ValidateJSON validates JSON request body against struct validation tags
func ValidateJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   constants.ErrInvalidRequestBody,
			"details": err.Error(),
		})
		return false
	}

	// Get validator from context
	validate, exists := c.Get("validator")
	if !exists {
		validate = validator.New()
	}

	if err := validate.(*validator.Validate).Struct(obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return false
	}

	return true
}

// ValidateQuery validates query parameters against struct validation tags
func ValidateQuery(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   constants.ErrInvalidPaginationParams,
			"details": err.Error(),
		})
		return false
	}

	// Get validator from context
	validate, exists := c.Get("validator")
	if !exists {
		validate = validator.New()
	}

	if err := validate.(*validator.Validate).Struct(obj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
		})
		return false
	}

	return true
}
