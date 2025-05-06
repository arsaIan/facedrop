package service

import (
	"facedrop/deepface"
	"facedrop/interfaces"
	"facedrop/logger"
	"facedrop/models"
	"facedrop/repository"
	"fmt"
)

type EventService struct {
	eventRepo *repository.EventRepository
	qrService *QRService
	deepfaceClient *deepface.DeepFaceClient
	processor interfaces.Processor
}

func NewEventService(eventRepo *repository.EventRepository, qrService *QRService, deepfaceClient *deepface.DeepFaceClient, processor interfaces.Processor) *EventService {
	return &EventService{eventRepo: eventRepo, qrService: qrService, deepfaceClient: deepfaceClient, processor: processor}
}

func (s *EventService) CreateEvent(event *models.Event) error {
	qrCode, err := s.qrService.GenerateEventSubscriptionQR(event.ID)
	if err != nil {
		return err
	}
	event.QRCode = qrCode
	errRepo := s.eventRepo.Create(event)
	return errRepo
}

func (s *EventService) GetEventByID(id uint) (*models.Event, error) {
	return s.eventRepo.FindByID(id)
}

func (s *EventService) GetAllEvents(userID uint) ([]models.Event, error) {
	return s.eventRepo.FindAll(userID)
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

//todo push event to ready queue
func (s *EventService) PushEventToReadyQueue(eventID uint) (interface{}, error) {
	//for now lets put the queue logic here 

	result, err := s.processor.Process(eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	logger.Info(fmt.Sprintf("event %d pushed to ready queue %v", eventID, result))
	//s.eventRepo.UpdateEventStatus(eventID, models.EventStatusReady)
	return result, nil
}

func (s *EventService) GetEventStatus(eventID uint) (models.EventStatus, error) {
	return s.eventRepo.GetEventStatus(eventID)
}

func (s *EventService) GetEventSubscribers(eventID uint) ([]models.EventSubscriber, error) {
	return s.eventRepo.GetEventSubscribers(eventID)
}