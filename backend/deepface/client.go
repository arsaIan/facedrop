package deepface

import (
	"bytes"
	"encoding/json"
	"facedrop/logger"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

type DeepFaceClient struct {
	Url string
}

func NewDeepFaceClient(url string) *DeepFaceClient {
	return &DeepFaceClient{Url: url}
}

type FaceEmbeddingRequest struct {
	ModelName string `json:"model_name"`
	Img       string `json:"img"`
}

type FaceVerificationRequest struct {
	ModelName       string `json:"model_name"`
	Img1           string `json:"img1"`
	Img2           string `json:"img2"`
	DetectorBackend string `json:"detector_backend,omitempty"`
	DistanceMetric  string `json:"distance_metric,omitempty"`
	EnforceDetection bool	`json:"enforce_detection"`
}

func (c *DeepFaceClient) GetFaceEmbedding(imageUrl string) ([]float32, error) {
	// Create request body with default model
	requestBody := FaceEmbeddingRequest{
		ModelName: "Facenet",
		Img:       imageUrl,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Make POST request to /represent endpoint
	resp, err := http.Post(c.Url+"/represent", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse response
	var result struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return result.Embedding, nil
}

func (c *DeepFaceClient) CompareFaces(imageUrl1 string, imageUrl2 string) (bool, error) {
	// Create request body with default parameters
	requestBody := FaceVerificationRequest{
		ModelName:       "Facenet",
		Img1:           imageUrl1,
		Img2:           imageUrl2,
		DetectorBackend: "opencv",
		DistanceMetric:  "euclidean",
		EnforceDetection: false,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error(err)
		return false, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Make POST request to /verify endpoint
	resp, err := http.Post(c.Url+"/verify", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.Error(err)
		return false, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return false, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse response
	var result struct {
		Verified bool `json:"verified"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Error(err)
		return false, fmt.Errorf("failed to parse response: %v", err)
	}
	return result.Verified, nil
}

// Helper function to create multipart form data request
func createMultipartFormData(fields map[string]string, files map[string]string) (*bytes.Buffer, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, "", err
		}
	}

	// Add files
	for key, filePath := range files {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, "", err
		}
		defer file.Close()

		part, err := writer.CreateFormFile(key, filePath)
		if err != nil {
			return nil, "", err
		}

		if _, err := io.Copy(part, file); err != nil {
			return nil, "", err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", err
	}

	return body, writer.FormDataContentType(), nil
}

