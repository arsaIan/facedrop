package service

import (
	"fmt"
	"mofoto/deepface"
	"mofoto/logger"
	"mofoto/models"
	"mofoto/repository"
)

type EventService struct {
	eventRepo *repository.EventRepository
	qrService *QRService
	deepfaceClient *deepface.DeepFaceClient
}

func NewEventService(eventRepo *repository.EventRepository, qrService *QRService, deepfaceClient *deepface.DeepFaceClient) *EventService {
	return &EventService{eventRepo: eventRepo, qrService: qrService, deepfaceClient: deepfaceClient}
}

func (s *EventService) CreateEvent(event *models.Event) error {
	qrCode, err := s.qrService.GenerateEventSubscriptionQR(event.ID)
	if err != nil {
		return err
	}
	event.QRCode = qrCode
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

//todo push event to ready queue
func (s *EventService) PushEventToReadyQueue(eventID uint) error {
	//for now lets put the queue logic here 

	go s.ProcessReadyEvent(eventID)
	
	return s.eventRepo.UpdateEventStatus(eventID, models.EventStatusReady)
}

func (s *EventService) GetEventStatus(eventID uint) (models.EventStatus, error) {
	return s.eventRepo.GetEventStatus(eventID)
}

func (s *EventService) ProcessReadyEvent(eventID uint) (map[string][]string, error) {
	//todo process the event
	//get the subscribers
	subs, err := s.eventRepo.GetSubscribers(eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	//get user faces
	faces, err := s.eventRepo.GetUserFaces(subs)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	//get the photos
	photos, err := s.eventRepo.GetEventPhotos(eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	matches := make(map[string][]string)
	for _, photo := range photos {
		for _, face := range faces {
			//compare the photo with the subscriber's face
			match, err := s.deepfaceClient.CompareFaces(photo.URL, face.FaceURL)
			if err != nil {
				logger.Error(err)
				return nil, err
			}
			if match {
				logger.Info("match found")
				if _, exists := matches[face.User.Email]; !exists {
					matches[face.User.Email] = make([]string, 0)
				}
				matches[face.User.Email] = append(matches[face.User.Email], photo.URL)
			}
		}	
	}
	logger.Info(fmt.Sprintf("Processing event %d finished", eventID))
	return matches, nil
}