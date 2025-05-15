package service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/skip2/go-qrcode"
)

type QRService struct {
	baseURL string
}

func NewQRService(baseURL string) *QRService {
	return &QRService{
		baseURL: baseURL,
	}
}

// GenerateEventSubscriptionQR generates a QR code for event subscription
// Returns the QR code as a base64 encoded PNG image
func (s *QRService) GenerateEventSubscriptionQR(eventID uint) (string, error) {
	// Create the subscription URL
	subscriptionURL := fmt.Sprintf("%s/events/subscribe?event=%d", s.baseURL, eventID)

	// Generate QR code
	qr, err := qrcode.New(subscriptionURL, qrcode.Medium)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %v", err)
	}

	// Convert QR code to PNG
	var buf bytes.Buffer
	if err := qr.Write(256, &buf); err != nil {
		return "", fmt.Errorf("failed to write QR code: %v", err)
	}

	// Encode PNG to base64
	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	return base64Str, nil
}

// SaveQRCodeToFile saves the QR code to a file
func (s *QRService) SaveQRCodeToFile(eventID string, outputDir string) (string, error) {
	// Create the subscription URL
	subscriptionURL := fmt.Sprintf("%s/events/%s/subscribe", s.baseURL, eventID)

	// Generate QR code
	qr, err := qrcode.New(subscriptionURL, qrcode.Medium)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %v", err)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %v", err)
	}

	// Create file path
	filePath := filepath.Join(outputDir, fmt.Sprintf("event_%s_qr.png", eventID))

	// Save QR code to file
	if err := qr.WriteFile(256, filePath); err != nil {
		return "", fmt.Errorf("failed to save QR code to file: %v", err)
	}

	return filePath, nil
} 