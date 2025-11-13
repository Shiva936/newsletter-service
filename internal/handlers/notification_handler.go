package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"newsletter-service/internal/constants"
	"newsletter-service/internal/dtos"
	"newsletter-service/internal/services/notification"
)

type NotificationHandler struct {
	notificationService notification.Service
}

func NewNotificationHandler(notificationService notification.Service) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// GetEmailLogs retrieves all email logs
func (h *NotificationHandler) GetEmailLogs(c *gin.Context) {
	var pagination dtos.PaginationRequest
	if err := c.ShouldBindQuery(&pagination); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidPaginationParams})
		return
	}

	// Check if pagination parameters were provided
	if pagination.Page > 0 || pagination.PageSize > 0 {
		// Use paginated response
		page, pageSize := pagination.GetDefaults()
		offset := pagination.CalculateOffset()

		logs, total, err := h.notificationService.GetEmailLogsWithPagination(c.Request.Context(), offset, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		paginationResponse := dtos.CreatePaginationResponse(page, pageSize, total)
		paginatedResponse := dtos.PaginatedResponse[*notification.EmailLog]{
			Data:       logs,
			Pagination: paginationResponse,
		}

		c.JSON(http.StatusOK, paginatedResponse)
	} else {
		// Use non-paginated response for backward compatibility
		logs, err := h.notificationService.GetEmailLogs(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, logs)
	}
}

// GetEmailLogByID retrieves an email log by ID
func (h *NotificationHandler) GetEmailLogByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidEmailLogID})
		return
	}

	log, err := h.notificationService.GetEmailLogByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrEmailLogNotFound})
		return
	}

	c.JSON(http.StatusOK, log)
}

// SendNotifications sends notifications for specific content (Scheduler endpoint)
func (h *NotificationHandler) SendNotifications(c *gin.Context) {
	var req struct {
		ContentID uint `json:"content_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidRequestBody})
		return
	}

	if err := h.notificationService.SendNotificationsByContentID(c.Request.Context(), req.ContentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": constants.MsgNotificationsSentSuccessfully})
}

// RetryFailedNotifications retries failed email deliveries (Scheduler endpoint)
func (h *NotificationHandler) RetryFailedNotifications(c *gin.Context) {
	if err := h.notificationService.RetryFailedEmails(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": constants.MsgFailedNotificationsRetryInitiated})
}
