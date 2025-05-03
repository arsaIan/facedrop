package events

import (
	"facedrop/deepface"
	"facedrop/logger"
	"facedrop/models"
	"facedrop/repository"
	"fmt"
)

type QueueProcessor struct {
	eventRepo *repository.EventRepository
	deepfaceClient *deepface.DeepFaceClient
}
func NewQueueProcessor(eventRepo *repository.EventRepository, deepfaceClient *deepface.DeepFaceClient) *QueueProcessor {
	return &QueueProcessor{eventRepo: eventRepo, deepfaceClient: deepfaceClient}
}

func (qp *QueueProcessor) ProcessReadyEvent(eventID uint) (map[uint][]string, error) {
	//todo process the event
	logger.Info("start processing")
	//get the subscribers
	subs, err := qp.eventRepo.GetSubscribers(eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	logger.Info(fmt.Sprintf("fetch %d subs for event %d", len(subs), eventID))
	//get user faces
	faces, err := qp.eventRepo.GetUserFaces(subs)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	logger.Info(fmt.Sprintf("fetch %d faces for event %d", len(faces), eventID))

	//get the photos
	photos, err := qp.eventRepo.GetEventPhotos(eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	logger.Info(fmt.Sprintf("fetch %d photos for event %d", len(photos), eventID))
	matches := make(map[uint][]string)
	for i, photo := range photos {
		logger.Info(fmt.Sprintf("comparing photo %d", i))
		for _, face := range faces {
			//compare the photo with the subscriber's face
			match, err := qp.deepfaceClient.CompareFaces(photo.URL, face.FaceURL)
			if err != nil {
				logger.Error(err)
				return nil, err
			}
			if match {
				logger.Info("match found")
				if _, exists := matches[face.User.ID]; !exists {
					matches[face.User.ID] = make([]string, 0)
				}
				matches[face.User.ID] = append(matches[face.User.ID], photo.URL)
			}
		}	
	}
	logger.Info(fmt.Sprintf("Processing event %d finished", eventID))
	//save the matches
	var eventMatches []models.EventMatch;
	for userId, photos := range matches {	
		for _, photo := range photos {
			eventMatches = append(eventMatches, models.EventMatch{UserID: userId, PhotoURL: photo})
		}
	}
	qp.eventRepo.AddEventMatches(eventID, eventMatches)
	return matches, nil
}