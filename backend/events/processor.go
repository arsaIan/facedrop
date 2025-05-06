package events

import (
	"context"
	"facedrop/config"
	"facedrop/logger"
	"facedrop/models"
	"facedrop/repository"
	"facedrop/sender"
	"facedrop/storage"
	"fmt"
)

type EventProcessor struct {
	eventRepo *repository.EventRepository
	storageClient *storage.S3Client
	cfg *config.Config
	sender *sender.EmailSender
}
func NewEventProcessor(eventRepo *repository.EventRepository, storageClient *storage.S3Client, cfg *config.Config, sender *sender.EmailSender) *EventProcessor {
	return &EventProcessor{eventRepo: eventRepo, storageClient: storageClient, cfg: cfg, sender: sender}
}
func (ep *EventProcessor) Process(eventID uint) (interface{}, error) {
	logger.Info(fmt.Sprintf("processing event with dummy processor %d", eventID))
	event, err := ep.eventRepo.FindByID(eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	subs, err := ep.eventRepo.GetSubscribers(eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	logger.Info(fmt.Sprintf("fetch %d subs for event %d", len(subs), eventID))
	//get user faces
	faces, err := ep.eventRepo.GetUserFaces(subs)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	logger.Info(fmt.Sprintf("fetch %d faces for event %d", len(faces), eventID))

	//get the photos
	photos, err := ep.eventRepo.GetEventPhotos(eventID)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	logger.Info(fmt.Sprintf("fetch %d photos for event %d", len(photos), eventID))
	photoURLs := make([]string, 0)
	for _, p := range photos {
		photoURLs = append(photoURLs, p.URL)
	}

	zipFile, err := ep.storageClient.GetZippedFiles(context.Background(), photoURLs, ep.cfg.StorageConfig.EventBucket) 
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	downloadLink, err := ep.storageClient.UploadZipFile(context.Background(), zipFile, fmt.Sprintf("event-%d", eventID), ep.cfg.StorageConfig.EventBucket)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	for _, sub := range subs {
		emailData := sender.EmailData{
			Subject: fmt.Sprintf("Event Photos #%s", event.Title),
			Body: fmt.Sprintf("Hey %s, your photos from %s are ready to download. Click the button below to download them:", sub.FirstName, event.Title),
			DownloadLink: downloadLink,
			Username: sub.FirstName,
			EventName: event.Title,
		}
		err := ep.sender.SendZipFile([]string{sub.Email}, emailData)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		logger.Info(fmt.Sprintf("sent event %d to %s", eventID, sub.Email))
		ep.eventRepo.UpdateSubscriberStatus(sub.ID, models.Delivered)
	}
	return zipFile, nil
}
