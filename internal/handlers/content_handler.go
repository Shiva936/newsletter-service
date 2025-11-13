package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"newsletter-service/internal/constants"
	"newsletter-service/internal/dtos"
	"newsletter-service/internal/services/content"
)

type ContentHandler struct {
	contentService content.Service
}

func NewContentHandler(contentService content.Service) *ContentHandler {
	return &ContentHandler{
		contentService: contentService,
	}
}

// GetContents retrieves all content
func (h *ContentHandler) GetContents(c *gin.Context) {
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

		contents, total, err := h.contentService.GetAllContentWithPagination(c.Request.Context(), offset, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var response []dtos.ContentResponse
		for _, content := range contents {
			response = append(response, dtos.ContentResponse{
				ID:          content.ID,
				TopicID:     content.TopicID,
				Title:       content.Title,
				Body:        content.Body,
				IsPublished: content.IsPublished,
				PublishedAt: content.PublishedAt,
				CreatedAt:   content.CreatedAt,
				UpdatedAt:   content.UpdatedAt,
			})
		}

		paginationResponse := dtos.CreatePaginationResponse(page, pageSize, total)
		paginatedResponse := dtos.PaginatedResponse[dtos.ContentResponse]{
			Data:       response,
			Pagination: paginationResponse,
		}

		c.JSON(http.StatusOK, paginatedResponse)
	} else {
		// Use non-paginated response for backward compatibility
		contents, err := h.contentService.GetAllContent(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var response []dtos.ContentResponse
		for _, content := range contents {
			response = append(response, dtos.ContentResponse{
				ID:          content.ID,
				TopicID:     content.TopicID,
				Title:       content.Title,
				Body:        content.Body,
				IsPublished: content.IsPublished,
				PublishedAt: content.PublishedAt,
				CreatedAt:   content.CreatedAt,
				UpdatedAt:   content.UpdatedAt,
			})
		}

		c.JSON(http.StatusOK, response)
	}
}

// CreateContent creates new content
func (h *ContentHandler) CreateContent(c *gin.Context) {
	var req dtos.CreateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidRequestBody})
		return
	}

	contentModel := &content.Content{
		TopicID:     req.TopicID,
		Title:       req.Title,
		Body:        req.Body,
		IsPublished: false,
	}

	if err := h.contentService.CreateContent(c.Request.Context(), contentModel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dtos.ContentResponse{
		ID:          contentModel.ID,
		TopicID:     contentModel.TopicID,
		Title:       contentModel.Title,
		Body:        contentModel.Body,
		IsPublished: contentModel.IsPublished,
		PublishedAt: contentModel.PublishedAt,
		CreatedAt:   contentModel.CreatedAt,
		UpdatedAt:   contentModel.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetContentByID retrieves content by ID
func (h *ContentHandler) GetContentByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidContentID})
		return
	}

	contentModel, err := h.contentService.GetContentByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrContentNotFound})
		return
	}

	response := dtos.ContentResponse{
		ID:          contentModel.ID,
		TopicID:     contentModel.TopicID,
		Title:       contentModel.Title,
		Body:        contentModel.Body,
		IsPublished: contentModel.IsPublished,
		PublishedAt: contentModel.PublishedAt,
		CreatedAt:   contentModel.CreatedAt,
		UpdatedAt:   contentModel.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateContent updates content
func (h *ContentHandler) UpdateContent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidContentID})
		return
	}

	var req dtos.UpdateContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidRequestBody})
		return
	}

	updates := make(map[string]interface{})
	if req.TopicID != 0 {
		updates["topic_id"] = req.TopicID
	}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Body != "" {
		updates["body"] = req.Body
	}

	if err := h.contentService.UpdateContent(c.Request.Context(), uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": constants.MsgContentUpdatedSuccessfully})
}

// DeleteContent deletes content
func (h *ContentHandler) DeleteContent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidContentID})
		return
	}

	if err := h.contentService.DeleteContent(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": constants.MsgContentDeletedSuccessfully})
}

// PublishContent publishes content
func (h *ContentHandler) PublishContent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidContentID})
		return
	}

	if err := h.contentService.PublishContent(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": constants.MsgContentPublishedSuccessfully})
}

// GetPendingNotifications gets content that needs notifications sent
func (h *ContentHandler) GetPendingNotifications(c *gin.Context) {
	// Get contents that are published but haven't been sent yet
	pendingContents, err := h.contentService.GetPendingNotifications(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pending_notifications": len(pendingContents),
		"content_ids":           pendingContents,
	})
}
