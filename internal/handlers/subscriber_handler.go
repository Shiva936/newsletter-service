package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"newsletter-service/internal/constants"
	"newsletter-service/internal/dtos"
	"newsletter-service/internal/router/middleware"
	"newsletter-service/internal/services/subscriber"
)

type SubscriberHandler struct {
	subscriberService subscriber.Service
}

func NewSubscriberHandler(subscriberService subscriber.Service) *SubscriberHandler {
	return &SubscriberHandler{
		subscriberService: subscriberService,
	}
}

// GetSubscribers retrieves all subscribers
func (h *SubscriberHandler) GetSubscribers(c *gin.Context) {
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

		subscribers, total, err := h.subscriberService.GetAllSubscribersWithPagination(c.Request.Context(), offset, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var response []dtos.SubscriberResponse
		for _, sub := range subscribers {
			response = append(response, dtos.SubscriberResponse{
				ID:        sub.ID,
				Email:     sub.Email,
				Name:      sub.Name,
				IsActive:  sub.IsActive,
				CreatedAt: sub.CreatedAt,
				UpdatedAt: sub.UpdatedAt,
			})
		}

		paginationResponse := dtos.CreatePaginationResponse(page, pageSize, total)
		paginatedResponse := dtos.PaginatedResponse[dtos.SubscriberResponse]{
			Data:       response,
			Pagination: paginationResponse,
		}

		c.JSON(http.StatusOK, paginatedResponse)
	} else {
		// Use non-paginated response for backward compatibility
		subscribers, err := h.subscriberService.GetAllSubscribers(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var response []dtos.SubscriberResponse
		for _, sub := range subscribers {
			response = append(response, dtos.SubscriberResponse{
				ID:        sub.ID,
				Email:     sub.Email,
				Name:      sub.Name,
				IsActive:  sub.IsActive,
				CreatedAt: sub.CreatedAt,
				UpdatedAt: sub.UpdatedAt,
			})
		}

		c.JSON(http.StatusOK, response)
	}
}

// CreateSubscriber creates a new subscriber
func (h *SubscriberHandler) CreateSubscriber(c *gin.Context) {
	var req dtos.CreateSubscriberRequest
	if !middleware.ValidateJSON(c, &req) {
		return
	}

	subscriberModel := &subscriber.Subscriber{
		Email:    req.Email,
		Name:     req.Name,
		IsActive: true,
	}

	var err error
	if len(req.SubscribedTopics) > 0 {
		// Create subscriber with topics
		err = h.subscriberService.CreateSubscriberWithTopics(c.Request.Context(), subscriberModel, req.SubscribedTopics)
	} else {
		// Create subscriber without topics
		err = h.subscriberService.CreateSubscriber(c.Request.Context(), subscriberModel)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get subscriber with subscribed topics for response
	subscriberWithTopics, topicNames, err := h.subscriberService.GetSubscriberByIDWithTopics(c.Request.Context(), subscriberModel.ID)
	if err != nil {
		// Fallback to basic response if getting topics fails
		response := dtos.SubscriberResponse{
			ID:               subscriberModel.ID,
			Email:            subscriberModel.Email,
			Name:             subscriberModel.Name,
			IsActive:         subscriberModel.IsActive,
			SubscribedTopics: req.SubscribedTopics,
			CreatedAt:        subscriberModel.CreatedAt,
			UpdatedAt:        subscriberModel.UpdatedAt,
		}
		c.JSON(http.StatusCreated, response)
		return
	}

	response := dtos.SubscriberResponse{
		ID:               subscriberWithTopics.ID,
		Email:            subscriberWithTopics.Email,
		Name:             subscriberWithTopics.Name,
		IsActive:         subscriberWithTopics.IsActive,
		SubscribedTopics: topicNames,
		CreatedAt:        subscriberWithTopics.CreatedAt,
		UpdatedAt:        subscriberWithTopics.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetSubscriberByID retrieves a subscriber by ID
func (h *SubscriberHandler) GetSubscriberByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidSubscriberID})
		return
	}

	subscriberModel, topicNames, err := h.subscriberService.GetSubscriberByIDWithTopics(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrSubscriberNotFound})
		return
	}

	response := dtos.SubscriberResponse{
		ID:               subscriberModel.ID,
		Email:            subscriberModel.Email,
		Name:             subscriberModel.Name,
		IsActive:         subscriberModel.IsActive,
		SubscribedTopics: topicNames,
		CreatedAt:        subscriberModel.CreatedAt,
		UpdatedAt:        subscriberModel.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateSubscriber updates a subscriber
func (h *SubscriberHandler) UpdateSubscriber(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidSubscriberID})
		return
	}

	var req dtos.UpdateSubscriberRequest
	if !middleware.ValidateJSON(c, &req) {
		return
	}

	updates := make(map[string]interface{})
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := h.subscriberService.UpdateSubscriberWithTopics(c.Request.Context(), uint(id), updates, req.SubscribedTopics); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": constants.MsgSubscriberUpdatedSuccessfully})
}

// DeleteSubscriber deletes a subscriber
func (h *SubscriberHandler) DeleteSubscriber(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidSubscriberID})
		return
	}

	if err := h.subscriberService.DeleteSubscriber(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": constants.MsgSubscriberDeletedSuccessfully})
}

// CreateSubscription creates a new subscription
func (h *SubscriberHandler) CreateSubscription(c *gin.Context) {
	var req dtos.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidRequestBody})
		return
	}

	if err := h.subscriberService.Subscribe(c.Request.Context(), req.SubscriberID, req.TopicID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": constants.MsgSubscriptionCreatedSuccessfully})
}

// GetSubscriptions retrieves all subscriptions
func (h *SubscriberHandler) GetSubscriptions(c *gin.Context) {
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

		subscriptions, total, err := h.subscriberService.GetAllSubscriptionsWithPagination(c.Request.Context(), offset, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var response []dtos.SubscriptionResponse
		for _, sub := range subscriptions {
			response = append(response, dtos.SubscriptionResponse{
				ID:           sub.ID,
				SubscriberID: sub.SubscriberID,
				TopicID:      sub.TopicID,
				CreatedAt:    sub.CreatedAt,
			})
		}

		paginationResponse := dtos.CreatePaginationResponse(page, pageSize, total)
		paginatedResponse := dtos.PaginatedResponse[dtos.SubscriptionResponse]{
			Data:       response,
			Pagination: paginationResponse,
		}

		c.JSON(http.StatusOK, paginatedResponse)
	} else {
		// Use non-paginated response for backward compatibility
		subscriptions, err := h.subscriberService.GetAllSubscriptions(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var response []dtos.SubscriptionResponse
		for _, sub := range subscriptions {
			response = append(response, dtos.SubscriptionResponse{
				ID:           sub.ID,
				SubscriberID: sub.SubscriberID,
				TopicID:      sub.TopicID,
				CreatedAt:    sub.CreatedAt,
			})
		}

		c.JSON(http.StatusOK, response)
	}
}

// GetSubscriptionsBySubscriber retrieves subscriptions by subscriber ID
func (h *SubscriberHandler) GetSubscriptionsBySubscriber(c *gin.Context) {
	subscriberID, err := strconv.ParseUint(c.Param("subscriber_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidSubscriberID})
		return
	}

	subscriptions, err := h.subscriberService.GetSubscriptionsBySubscriberID(c.Request.Context(), uint(subscriberID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []dtos.SubscriptionResponse
	for _, sub := range subscriptions {
		response = append(response, dtos.SubscriptionResponse{
			ID:           sub.ID,
			SubscriberID: sub.SubscriberID,
			TopicID:      sub.TopicID,
			CreatedAt:    sub.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetSubscriptionsByTopic retrieves subscriptions by topic ID
func (h *SubscriberHandler) GetSubscriptionsByTopic(c *gin.Context) {
	topicID, err := strconv.ParseUint(c.Param("topic_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidTopicID})
		return
	}

	subscriptions, err := h.subscriberService.GetSubscriptionsByTopicID(c.Request.Context(), uint(topicID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []dtos.SubscriptionResponse
	for _, sub := range subscriptions {
		response = append(response, dtos.SubscriptionResponse{
			ID:           sub.ID,
			SubscriberID: sub.SubscriberID,
			TopicID:      sub.TopicID,
			CreatedAt:    sub.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

// DeleteSubscription deletes a subscription
func (h *SubscriberHandler) DeleteSubscription(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidSubscriptionID})
		return
	}

	if err := h.subscriberService.Unsubscribe(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": constants.MsgSubscriptionDeletedSuccessfully})
}

// BulkCreateSubscribers creates multiple subscribers at once
func (h *SubscriberHandler) BulkCreateSubscribers(c *gin.Context) {
	var req dtos.BulkCreateSubscribersRequest
	if !middleware.ValidateJSON(c, &req) {
		return
	}

	startTime := time.Now()
	var subscribers []*subscriber.Subscriber
	var topicNamesList [][]string
	var errors []dtos.BulkError

	// Prepare subscriber models
	for _, createReq := range req.Subscribers {
		subscriberModel := &subscriber.Subscriber{
			Email:    createReq.Email,
			Name:     createReq.Name,
			IsActive: true,
		}
		subscribers = append(subscribers, subscriberModel)
		topicNamesList = append(topicNamesList, createReq.SubscribedTopics)
	}

	// Perform bulk create
	successIDs, bulkErrors := h.subscriberService.BulkCreateSubscribers(c.Request.Context(), subscribers, topicNamesList)

	// Convert errors to response format
	for i, err := range bulkErrors {
		if err != nil {
			var email string
			if i < len(req.Subscribers) {
				email = req.Subscribers[i].Email
			}
			errors = append(errors, dtos.BulkError{
				Index: i,
				Email: email,
				Error: err.Error(),
			})
		}
	}

	// Build success response
	var successResponses []dtos.SubscriberResponse
	for i, id := range successIDs {
		if i < len(subscribers) {
			sub := subscribers[i]
			if sub.ID == id {
				var topics []string
				if i < len(topicNamesList) {
					topics = topicNamesList[i]
				}
				successResponses = append(successResponses, dtos.SubscriberResponse{
					ID:               sub.ID,
					Email:            sub.Email,
					Name:             sub.Name,
					IsActive:         sub.IsActive,
					SubscribedTopics: topics,
					CreatedAt:        sub.CreatedAt,
					UpdatedAt:        sub.UpdatedAt,
				})
			}
		}
	}

	endTime := time.Now()
	summary := dtos.BulkOperationSummary{
		Total:       len(req.Subscribers),
		Success:     len(successIDs),
		Errors:      len(errors),
		StartedAt:   startTime,
		CompletedAt: endTime,
		Duration:    endTime.Sub(startTime).String(),
	}

	response := dtos.BulkCreateSubscribersResponse{
		Success: successResponses,
		Errors:  errors,
		Summary: summary,
	}

	statusCode := http.StatusCreated
	if len(errors) > 0 && len(successIDs) == 0 {
		statusCode = http.StatusBadRequest
	} else if len(errors) > 0 {
		statusCode = http.StatusMultiStatus
	}

	c.JSON(statusCode, response)
}

// BulkUpdateSubscribers updates multiple subscribers at once
func (h *SubscriberHandler) BulkUpdateSubscribers(c *gin.Context) {
	var req dtos.BulkUpdateSubscribersRequest
	if !middleware.ValidateJSON(c, &req) {
		return
	}

	startTime := time.Now()
	var bulkUpdates []subscriber.BulkSubscriberUpdate
	var errors []dtos.BulkError

	// Prepare updates
	for _, updateReq := range req.Updates {
		updates := make(map[string]interface{})
		if updateReq.Email != "" {
			updates["email"] = updateReq.Email
		}
		if updateReq.Name != "" {
			updates["name"] = updateReq.Name
		}
		if updateReq.IsActive != nil {
			updates["is_active"] = *updateReq.IsActive
		}

		bulkUpdate := subscriber.BulkSubscriberUpdate{
			ID:         updateReq.ID,
			Updates:    updates,
			TopicNames: updateReq.SubscribedTopics,
		}
		bulkUpdates = append(bulkUpdates, bulkUpdate)
	}

	// Perform bulk update
	bulkErrors := h.subscriberService.BulkUpdateSubscribers(c.Request.Context(), bulkUpdates)

	// Convert errors to response format
	for i, err := range bulkErrors {
		if err != nil {
			var id uint
			if i < len(req.Updates) {
				id = req.Updates[i].ID
			}
			errors = append(errors, dtos.BulkError{
				Index: i,
				ID:    id,
				Error: err.Error(),
			})
		}
	}

	endTime := time.Now()
	summary := dtos.BulkOperationSummary{
		Total:       len(req.Updates),
		Success:     len(req.Updates) - len(errors),
		Errors:      len(errors),
		StartedAt:   startTime,
		CompletedAt: endTime,
		Duration:    endTime.Sub(startTime).String(),
	}

	response := dtos.BulkResponse{
		Success: gin.H{"message": "Bulk update completed"},
		Errors:  errors,
		Summary: summary,
	}

	statusCode := http.StatusOK
	if len(errors) > 0 && summary.Success == 0 {
		statusCode = http.StatusBadRequest
	} else if len(errors) > 0 {
		statusCode = http.StatusMultiStatus
	}

	c.JSON(statusCode, response)
}

// BulkDeleteSubscribers deletes multiple subscribers at once
func (h *SubscriberHandler) BulkDeleteSubscribers(c *gin.Context) {
	var req dtos.BulkDeleteSubscribersRequest
	if !middleware.ValidateJSON(c, &req) {
		return
	}

	startTime := time.Now()
	var errors []dtos.BulkError

	// Perform bulk delete
	bulkErrors := h.subscriberService.BulkDeleteSubscribers(c.Request.Context(), req.IDs)

	// Convert errors to response format
	for i, err := range bulkErrors {
		if err != nil {
			var id uint
			if i < len(req.IDs) {
				id = req.IDs[i]
			}
			errors = append(errors, dtos.BulkError{
				Index: i,
				ID:    id,
				Error: err.Error(),
			})
		}
	}

	endTime := time.Now()
	summary := dtos.BulkOperationSummary{
		Total:       len(req.IDs),
		Success:     len(req.IDs) - len(errors),
		Errors:      len(errors),
		StartedAt:   startTime,
		CompletedAt: endTime,
		Duration:    endTime.Sub(startTime).String(),
	}

	response := dtos.BulkResponse{
		Success: gin.H{"message": "Bulk delete completed"},
		Errors:  errors,
		Summary: summary,
	}

	statusCode := http.StatusOK
	if len(errors) > 0 && summary.Success == 0 {
		statusCode = http.StatusBadRequest
	} else if len(errors) > 0 {
		statusCode = http.StatusMultiStatus
	}

	c.JSON(statusCode, response)
}
