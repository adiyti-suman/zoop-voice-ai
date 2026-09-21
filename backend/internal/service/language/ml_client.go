package language

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	domain "verification-platform/internal/domain/language"
)

// MLClient handles communication with the Python FastAPI ML Sidecar
type MLClient interface {
	DetectLanguage(ctx context.Context, sessionID string, pcmData []byte) (*domain.DetectionResult, error)
}

type mlClientImpl struct {
	baseURL    string
	httpClient *http.Client
}

func NewMLClient(baseURL string) MLClient {
	return &mlClientImpl{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *mlClientImpl) DetectLanguage(ctx context.Context, sessionID string, pcmData []byte) (*domain.DetectionResult, error) {
	// Create a multipart form
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add audio file part
	fw, err := w.CreateFormFile("audio", fmt.Sprintf("%s.pcm", sessionID))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(fw, bytes.NewReader(pcmData)); err != nil {
		return nil, err
	}
	w.Close()

	// Build request
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v1/detect-language", c.baseURL), &b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML service error: %s", string(body))
	}

	// Parse response
	var result domain.DetectionResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	result.SessionID = sessionID
	result.Timestamp = time.Now().UTC()

	return &result, nil
}
