package service

import (
	"mofoto/models"
	"mofoto/repository"
)

type EventService struct {
	eventRepo *repository.EventRepository
}

func NewEventService(eventRepo *repository.EventRepository) *EventService {
	return &EventService{eventRepo: eventRepo}
}

func (s *EventService) CreateEvent(event *models.Event) error {
	return s.eventRepo.Create(event)
}

func (s *EventService) GetEventByID(id uint) (*models.Event, error) {
	return s.eventRepo.FindByID(id)
}

func (s *EventService) GetAllEvents() ([]models.Event, error) {
	return s.eventRepo.FindAll()
}

func (s *EventService) UpdateEvent(event *models.Event) error {
	return s.eventRepo.Update(event)
}

func (s *EventService) DeleteEvent(id uint) error {
	return s.eventRepo.Delete(id)
}

func (s *EventService) SubscribeToEvent(eventID, userID uint) error {
	return s.eventRepo.AddSubscriber(eventID, userID)
}

func (s *EventService) UnsubscribeFromEvent(eventID, userID uint) error {
	return s.eventRepo.RemoveSubscriber(eventID, userID)
}

func (s *EventService) AddPhotoToEvent(photo *models.Photo) error {
	return s.eventRepo.AddPhoto(photo)
}

func (s *EventService) GetEventPhotos(eventID uint) ([]models.Photo, error) {
	return s.eventRepo.GetEventPhotos(eventID)
}

func (s *EventService) GetSubscriberPhotos(eventID, userID uint) ([]models.Photo, error) {
	return s.eventRepo.GetSubscriberPhotos(eventID, userID)
}

