package controller

import (
	"facedrop/config"
	"facedrop/logger"
	"facedrop/models"
	"facedrop/service"
	"facedrop/utils"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type EventController struct {
	eventService *service.EventService
	storageClient *service.StorageClient
	cfg *config.Config
}

func NewEventController(eventService *service.EventService, storageClient *service.StorageClient, cfg *config.Config) *EventController {
	return &EventController{eventService: eventService, storageClient: storageClient, cfg: cfg}
}

func (c *EventController) CreateEvent(ctx *gin.Context) {
	var event models.Event
	if err := ctx.ShouldBindJSON(&event); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	
	event.CreatedBy = ctx.GetUint("user_id")
	event.CreatedAt = time.Now()
	if err := c.eventService.CreateEvent(&event); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	ctx.JSON(http.StatusCreated, event)
}

func (c *EventController) GetEvent(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	event, err := c.eventService.GetEventByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	ctx.JSON(http.StatusOK, event)
}

func (c *EventController) GetAllEvents(ctx *gin.Context) {
	userID := ctx.GetUint("user_id")
	events, err := c.eventService.GetAllEvents(userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	ctx.JSON(http.StatusOK, events)
}



func (c *EventController) UpdateEvent(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var event models.Event
	if err := ctx.ShouldBindJSON(&event); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event.ID = uint(id)
	if err := c.eventService.UpdateEvent(&event); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update event"})
		return
	}

	ctx.JSON(http.StatusOK, event)
}

func (c *EventController) DeleteEvent(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	if err := c.eventService.DeleteEvent(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *EventController) SubscribeToEvent(ctx *gin.Context) {
	eventID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := c.eventService.SubscribeToEvent(uint(eventID), userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to subscribe to event"})
		return
	}

	ctx.Status(http.StatusOK)
}

func (c *EventController) UnsubscribeFromEvent(ctx *gin.Context) {
	eventID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := c.eventService.UnsubscribeFromEvent(uint(eventID), userID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unsubscribe from event"})
		return
	}

	ctx.Status(http.StatusOK)
}

func (c *EventController) AddPhoto(ctx *gin.Context) {
	eventID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}
	event, err := c.eventService.GetEventByID(uint(eventID))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}
	userID := ctx.GetUint("user_id")
	if event.CreatedBy != userID{
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not allowed to add photos to this event"})
		return
	}
	

	file, err := ctx.FormFile("photo")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Photo file is required"})
		return
	}

	// Open the file
	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	// Read the file content
	fileContent, err := io.ReadAll(src)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Upload file to S3
	// Remove spaces from filename
	file.Filename = utils.CleanFilename(file.Filename)
	fileKey := fmt.Sprintf("events/%d/%s", eventID, file.Filename)
	photoURL, err := c.storageClient.UploadFile(ctx, fileKey, fileContent, c.cfg.StorageConfig.EventBucket)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload photo"})
		return
	}
	logger.Info("Photo uploaded to S3", logger.String("photoURL", photoURL))

	photo := &models.Photo{
		EventID:     uint(eventID),
		UploadedBy:  userID,
		URL:         photoURL,
		UploadedAt:  time.Now(),
		IsPublic:    true,
		Description: ctx.PostForm("description"),
	}

	if err := c.eventService.AddPhotoToEvent(photo); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add photo"})
		return
	}

	ctx.JSON(http.StatusCreated, photo)
}



func (c *EventController) GetEventPhotos(ctx *gin.Context) {
	// Get pagination parameters from query string
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("limit", "1"))

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Calculate offset
	offset := (page - 1) * pageSize
	// Get event ID from URL parameter
	eventID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	// Get total count of photos
	photos, err := c.eventService.GetEventPhotos(uint(eventID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch photos"})
		return
	}

	totalPhotos := len(photos)
	totalPages := (totalPhotos + pageSize - 1) / pageSize

	// Apply pagination
	start := offset
	end := offset + pageSize
	if start >= totalPhotos {
		ctx.JSON(http.StatusOK, gin.H{
			"photos":      []models.Photo{},
			"pagination": gin.H{
				"current_page": page,
				"total_pages": totalPages,
				"page_size":   pageSize,
				"total_items": totalPhotos,
			},
		})
		return
	}
	if end > totalPhotos {
		end = totalPhotos
	}

	paginatedPhotos := photos[start:end]

	ctx.JSON(http.StatusOK, gin.H{
		"photos":      paginatedPhotos,
		"pagination": gin.H{
			"current_page": page,
			"total_pages": totalPages,
			"page_size":   pageSize,
			"total_items": totalPhotos,
		},
	})
	return
}

func (c *EventController) GetSubscriberPhotos(ctx *gin.Context) {
	eventID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	userID := ctx.GetUint("user_id")
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	photos, err := c.eventService.GetSubscriberPhotos(uint(eventID), userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch photos"})
		return
	}

	ctx.JSON(http.StatusOK, photos)
} 

func (c *EventController) PushEventToReadyQueue(ctx *gin.Context) {
	eventID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	_, err = c.eventService.PushEventToReadyQueue(uint(eventID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to push event to ready queue"})
		return 
	}
	//utils.AddZipDownloadHeader(ctx, fmt.Sprintf("attachment; filename=event_%d_photos.zip", eventID), zipData.([]byte))
	// Send the file
	ctx.JSON(http.StatusOK, gin.H{"message": "Files sent to subscribers"})
}

func (c *EventController) GetEventSubscribers(ctx *gin.Context) {
	eventID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	subscribers, err := c.eventService.GetEventSubscribers(uint(eventID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch subscribers"})
		return
	}
	ctx.JSON(http.StatusOK, subscribers)
}