package errors

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"newsletter-service/internal/logger"
)

// AppError represents a standardized application error
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	StatusCode int    `json:"-"`
	Err        error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// Error constructors
func NewValidationError(message string, err error) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    message,
		StatusCode: http.StatusBadRequest,
		Err:        err,
	}
}

func NewNotFoundError(resource string, id interface{}) *AppError {
	return &AppError{
		Code:       "RESOURCE_NOT_FOUND",
		Message:    fmt.Sprintf("%s with ID %v not found", resource, id),
		StatusCode: http.StatusNotFound,
	}
}

func NewConflictError(message string, err error) *AppError {
	return &AppError{
		Code:       "RESOURCE_CONFLICT",
		Message:    message,
		StatusCode: http.StatusConflict,
		Err:        err,
	}
}

func NewInternalError(message string, err error) *AppError {
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Err:        err,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:       "UNAUTHORIZED",
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:       "FORBIDDEN",
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

func NewTooManyRequestsError(message string) *AppError {
	return &AppError{
		Code:       "TOO_MANY_REQUESTS",
		Message:    message,
		StatusCode: http.StatusTooManyRequests,
	}
}

// ErrorHandler middleware for centralized error handling
func ErrorHandler() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			ctx := c.Request.Context()

			var appErr *AppError
			var ok bool

			// Check if it's already an AppError
			if appErr, ok = err.(*AppError); !ok {
				// Convert generic error to AppError
				appErr = NewInternalError("An internal error occurred", err)
			}

			// Log the error
			if appErr.StatusCode >= 500 {
				logger.Error(ctx, "Internal server error: %v", appErr.Error())
			} else {
				logger.Warn(ctx, "Client error: %v", appErr.Error())
			}

			// Return structured error response
			response := gin.H{
				"error": gin.H{
					"code":    appErr.Code,
					"message": appErr.Message,
				},
			}

			if appErr.Details != "" {
				response["error"].(gin.H)["details"] = appErr.Details
			}

			c.JSON(appErr.StatusCode, response)
		}
	})
}

// HandleError is a helper function to handle errors in handlers
func HandleError(c *gin.Context, err error) {
	if err != nil {
		c.Error(err)
		c.Abort()
	}
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// WrapWithCode wraps an error with a specific error code
func WrapWithCode(err error, code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
	}
}
