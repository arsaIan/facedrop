package repository

import (
	"facedrop/models"
	"time"

	"gorm.io/gorm"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(event *models.Event) error {
	if err := r.db.Create(event).Error; err != nil {
		return err
	}
	// Preload Creator after creation
	return r.db.Preload("Creator").First(event, event.ID).Error
}

func (r *EventRepository) FindByID(id uint) (*models.Event, error) {
	var event models.Event
	err := r.db.Preload("Creator").Preload("Subscribers").Preload("Photos").First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *EventRepository) FindAll(userID uint) ([]models.Event, error) {
	var events []models.Event
	err := r.db.Preload("Creator").Preload("Subscribers").Preload("Photos").Where("created_by = ?", userID).Find(&events).Error
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *EventRepository) Update(event *models.Event) error {
	return r.db.Save(event).Error
}

func (r *EventRepository) Delete(id uint) error {
	return r.db.Delete(&models.Event{}, id).Error
}

func (r *EventRepository) AddSubscriber(eventID, userID uint) error {
	subscriber := models.EventSubscriber{
		EventID:      eventID,
		UserID:       userID,
		SubscribedAt: time.Now(),
	}
	return r.db.Create(&subscriber).Error
}

func (r *EventRepository) RemoveSubscriber(eventID, userID uint) error {
	return r.db.Where("event_id = ? AND user_id = ?", eventID, userID).Delete(&models.EventSubscriber{}).Error
}

func (r *EventRepository) AddPhoto(photo *models.Photo) error {
	if err := r.db.Create(photo).Error; err != nil {	
		return err
	}	
	// Preload Uploader after creation
	return r.db.Preload("Uploader").Preload("Event").Preload("Event.Creator").First(photo, photo.ID).Error
}

func (r *EventRepository) GetEventPhotos(eventID uint) ([]models.Photo, error) {
	var photos []models.Photo
	err := r.db.Where("event_id = ?", eventID).Preload("Uploader").Find(&photos).Error
	if err != nil {
		return nil, err
	}
	return photos, nil
}

func (r *EventRepository) GetSubscriberPhotos(eventID, userID uint) ([]models.Photo, error) {
	var photos []models.Photo
	err := r.db.Joins("JOIN event_subscribers ON event_subscribers.event_id = photos.event_id").
		Where("event_subscribers.user_id = ? AND photos.event_id = ?", userID, eventID).
		Preload("Uploader").
		Find(&photos).Error
	if err != nil {
		return nil, err
	}
	return photos, nil
} 

func (r *EventRepository) UpdateEventStatus(eventID uint, status models.EventStatus) error {
	return r.db.Model(&models.Event{}).Where("id = ?", eventID).Update("status", status).Error
}

func (r *EventRepository) GetEventStatus(eventID uint) (models.EventStatus, error) {
	var status models.EventStatus
	err := r.db.Model(&models.Event{}).Where("id = ?", eventID).Select("status").Scan(&status).Error
	return status, err
}

func (r *EventRepository) GetSubscribers(eventID uint) ([]models.User, error) {
	var event models.Event
	err := r.db.Preload("Subscribers").First(&event, eventID).Error
	if err != nil {
		return nil, err
	}
	return event.Subscribers, nil
}

func (r *EventRepository) GetUserFaces(subs []models.User) ([]models.UserFace, error) {
	var userIDs []uint
	for _, user := range subs {
		userIDs = append(userIDs, user.ID)
	}

	var userFaces []models.UserFace
	err := r.db.Where("user_id IN ?", userIDs).Find(&userFaces).Error
	return userFaces, err
}


func (r *EventRepository) AddEventMatches(eventID uint, matches []models.EventMatch) error {
	return r.db.Model(&models.EventMatch{}).Create(matches).Error
}

func (r *EventRepository) GetEventMatches(eventID uint) ([]models.EventMatch, error) {
	var eventMatches []models.EventMatch
	err := r.db.Where("event_id = ?", eventID).Find(&eventMatches).Error
	return eventMatches, err
}

func (r *EventRepository) GetEventSubscribers(eventID uint) ([]models.User, error) {
	var event models.Event
	err := r.db.Preload("Subscribers").First(&event, eventID).Error
	if err != nil {
		return nil, err
	}
	return event.Subscribers, nil
}
