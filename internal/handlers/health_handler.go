package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"newsletter-service/internal/constants"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health endpoint for main service health check
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  constants.HealthStatusHealthy,
		"service": constants.ServiceNameMain,
	})
}

// SchedulerHealth endpoint for scheduler service health check
func (h *HealthHandler) SchedulerHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  constants.HealthStatusHealthy,
		"service": constants.ServiceNameScheduler,
	})
}
