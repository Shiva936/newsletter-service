package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"newsletter-service/internal/constants"
	"newsletter-service/internal/dtos"
	"newsletter-service/internal/services/topic"
)

type TopicHandler struct {
	topicService topic.Service
}

func NewTopicHandler(topicService topic.Service) *TopicHandler {
	return &TopicHandler{
		topicService: topicService,
	}
}

// GetTopics retrieves all topics
func (h *TopicHandler) GetTopics(c *gin.Context) {
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

		topics, total, err := h.topicService.GetAllTopicsWithPagination(c.Request.Context(), offset, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var response []dtos.TopicResponse
		for _, topic := range topics {
			response = append(response, dtos.TopicResponse{
				ID:          topic.ID,
				Name:        topic.Name,
				Description: topic.Description,
				CreatedAt:   topic.CreatedAt,
				UpdatedAt:   topic.UpdatedAt,
			})
		}

		paginationResponse := dtos.CreatePaginationResponse(page, pageSize, total)
		paginatedResponse := dtos.PaginatedResponse[dtos.TopicResponse]{
			Data:       response,
			Pagination: paginationResponse,
		}

		c.JSON(http.StatusOK, paginatedResponse)
	} else {
		// Use non-paginated response for backward compatibility
		topics, err := h.topicService.GetAllTopics(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var response []dtos.TopicResponse
		for _, topic := range topics {
			response = append(response, dtos.TopicResponse{
				ID:          topic.ID,
				Name:        topic.Name,
				Description: topic.Description,
				CreatedAt:   topic.CreatedAt,
				UpdatedAt:   topic.UpdatedAt,
			})
		}

		c.JSON(http.StatusOK, response)
	}
}

// CreateTopic creates a new topic
func (h *TopicHandler) CreateTopic(c *gin.Context) {
	var req dtos.CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidRequestBody})
		return
	}

	topicModel := &topic.Topic{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.topicService.CreateTopic(c.Request.Context(), topicModel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dtos.TopicResponse{
		ID:          topicModel.ID,
		Name:        topicModel.Name,
		Description: topicModel.Description,
		CreatedAt:   topicModel.CreatedAt,
		UpdatedAt:   topicModel.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetTopicByID retrieves a topic by ID
func (h *TopicHandler) GetTopicByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidTopicID})
		return
	}

	topicModel, err := h.topicService.GetTopicByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrTopicNotFound})
		return
	}

	response := dtos.TopicResponse{
		ID:          topicModel.ID,
		Name:        topicModel.Name,
		Description: topicModel.Description,
		CreatedAt:   topicModel.CreatedAt,
		UpdatedAt:   topicModel.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateTopic updates a topic
func (h *TopicHandler) UpdateTopic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidTopicID})
		return
	}

	var req dtos.UpdateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidRequestBody})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	if err := h.topicService.UpdateTopic(c.Request.Context(), uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": constants.MsgTopicUpdatedSuccessfully})
}

// DeleteTopic deletes a topic
func (h *TopicHandler) DeleteTopic(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidTopicID})
		return
	}

	if err := h.topicService.DeleteTopic(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": constants.MsgTopicDeletedSuccessfully})
}
