package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"newsletter-service/internal/constants"
	"newsletter-service/internal/services/subscriber"
)

type UnsubscribeHandler struct {
	subscriberService subscriber.Service
}

func NewUnsubscribeHandler(subscriberService subscriber.Service) *UnsubscribeHandler {
	return &UnsubscribeHandler{
		subscriberService: subscriberService,
	}
}

// UnsubscribeGet handles GET requests to the unsubscribe page
func (h *UnsubscribeHandler) UnsubscribeGet(c *gin.Context) {
	subscriberIDStr := c.Query("subscriber")
	contentIDStr := c.Query("content")

	if subscriberIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscriber ID is required"})
		return
	}

	subscriberID, err := strconv.ParseUint(subscriberIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	// Get subscriber details
	subscriber, topicNames, err := h.subscriberService.GetSubscriberByIDWithTopics(c.Request.Context(), uint(subscriberID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscriber not found"})
		return
	}

	// Render unsubscribe page
	html := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Unsubscribe - Newsletter Service</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 50px auto;
            padding: 20px;
            background-color: #f4f4f4;
        }
        .container {
            background-color: white;
            padding: 40px;
            border-radius: 10px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
            text-align: center;
        }
        h1 {
            color: #007bff;
            margin-bottom: 20px;
        }
        .email-info {
            background-color: #f8f9fa;
            padding: 15px;
            border-radius: 5px;
            margin: 20px 0;
        }
        .topic-list {
            text-align: left;
            margin: 20px 0;
        }
        .topic-item {
            padding: 5px 0;
        }
        .btn {
            display: inline-block;
            padding: 10px 20px;
            margin: 10px;
            border: none;
            border-radius: 5px;
            text-decoration: none;
            cursor: pointer;
            font-size: 16px;
        }
        .btn-danger {
            background-color: #dc3545;
            color: white;
        }
        .btn-secondary {
            background-color: #6c757d;
            color: white;
        }
        .btn:hover {
            opacity: 0.8;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Unsubscribe from Newsletter</h1>
        
        <div class="email-info">
            <strong>Email:</strong> ` + subscriber.Email + `<br>
            <strong>Name:</strong> ` + subscriber.Name + `
        </div>

        <p>You are currently subscribed to the following topics:</p>
        <div class="topic-list">`

	for _, topic := range topicNames {
		html += `<div class="topic-item">• ` + topic + `</div>`
	}

	html += `</div>
        
        <p>Are you sure you want to unsubscribe from all newsletters?</p>
        
        <form method="POST" action="/unsubscribe" style="display: inline;">
            <input type="hidden" name="subscriber" value="` + subscriberIDStr + `">
            <input type="hidden" name="content" value="` + contentIDStr + `">
            <button type="submit" class="btn btn-danger">Yes, Unsubscribe</button>
        </form>
        
        <a href="#" onclick="history.back()" class="btn btn-secondary">Cancel</a>
        
        <p style="margin-top: 30px; font-size: 12px; color: #666;">
            If you clicked this link by mistake, you can simply close this page.
        </p>
    </div>
</body>
</html>`

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// UnsubscribePost handles POST requests to unsubscribe a user
func (h *UnsubscribeHandler) UnsubscribePost(c *gin.Context) {
	subscriberIDStr := c.PostForm("subscriber")
	if subscriberIDStr == "" {
		subscriberIDStr = c.Query("subscriber")
	}

	if subscriberIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscriber ID is required"})
		return
	}

	subscriberID, err := strconv.ParseUint(subscriberIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscriber ID"})
		return
	}

	// Deactivate subscriber instead of deleting
	updates := map[string]interface{}{
		"is_active": false,
	}

	if err := h.subscriberService.UpdateSubscriber(c.Request.Context(), uint(subscriberID), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unsubscribe"})
		return
	}

	// Render success page
	html := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Unsubscribed - Newsletter Service</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 50px auto;
            padding: 20px;
            background-color: #f4f4f4;
        }
        .container {
            background-color: white;
            padding: 40px;
            border-radius: 10px;
            box-shadow: 0 2px 5px rgba(0,0,0,0.1);
            text-align: center;
        }
        .success-icon {
            font-size: 48px;
            color: #28a745;
            margin-bottom: 20px;
        }
        h1 {
            color: #28a745;
            margin-bottom: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success-icon">✓</div>
        <h1>Successfully Unsubscribed</h1>
        <p>You have been successfully unsubscribed from our newsletter.</p>
        <p>We're sorry to see you go! If you change your mind, you can always subscribe again.</p>
        <p style="margin-top: 30px; font-size: 12px; color: #666;">
            This action has been completed. You can safely close this page.
        </p>
    </div>
</body>
</html>`

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

// ResubscribeHandler allows users to reactivate their subscription
func (h *UnsubscribeHandler) Resubscribe(c *gin.Context) {
	subscriberIDStr := c.Param("id")

	subscriberID, err := strconv.ParseUint(subscriberIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": constants.ErrInvalidSubscriberID})
		return
	}

	// Reactivate subscriber
	updates := map[string]interface{}{
		"is_active": true,
	}

	if err := h.subscriberService.UpdateSubscriber(c.Request.Context(), uint(subscriberID), updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully resubscribed to newsletter"})
}
