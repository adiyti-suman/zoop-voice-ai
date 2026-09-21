package language_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	domain "verification-platform/internal/domain/language"
	"verification-platform/internal/service/language"
)

func TestMLClient_DetectLanguage(t *testing.T) {
	// Mock ML server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/detect-language" {
			t.Errorf("Expected path /api/v1/detect-language, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"probabilities": [
				{"language": "hi", "score": 0.95},
				{"language": "en", "score": 0.04}
			],
			"is_speech": true
		}`))
	}))
	defer server.Close()

	client := language.NewMLClient(server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	result, err := client.DetectLanguage(ctx, "session-123", []byte("fake-audio-data"))
	if err != nil {
		t.Fatalf("DetectLanguage failed: %v", err)
	}

	if result.SessionID != "session-123" {
		t.Errorf("Expected SessionID 'session-123', got '%s'", result.SessionID)
	}
	if !result.IsSpeech {
		t.Error("Expected IsSpeech to be true")
	}
	if len(result.Probabilities) != 2 {
		t.Fatalf("Expected 2 probabilities, got %d", len(result.Probabilities))
	}
	if result.Probabilities[0].Language != domain.Hindi || result.Probabilities[0].Score != 0.95 {
		t.Errorf("Expected first probability to be Hindi with score 0.95, got %+v", result.Probabilities[0])
	}
	if top, ok := result.TopLanguage(0.90); !ok || top != domain.Hindi {
		t.Errorf("TopLanguage failed. ok: %v, top: %s", ok, top)
	}
}
