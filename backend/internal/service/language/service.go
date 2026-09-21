package language

import (
	"context"
	"log"

	domain "verification-platform/internal/domain/language"
	voiceSvc "verification-platform/internal/service/voice"
)

// Service coordinates Language Detection.
type Service interface {
	DetectFromBuffer(ctx context.Context, sessionID string, buffer []byte) (*domain.DetectionResult, error)
}

type serviceImpl struct {
	mlClient     MLClient
	voiceService voiceSvc.Service // to push detection events back to the session if needed
}

func NewService(mlClient MLClient, voiceService voiceSvc.Service) Service {
	return &serviceImpl{
		mlClient:     mlClient,
		voiceService: voiceService,
	}
}

func (s *serviceImpl) DetectFromBuffer(ctx context.Context, sessionID string, buffer []byte) (*domain.DetectionResult, error) {
	// Call Python ML service
	result, err := s.mlClient.DetectLanguage(ctx, sessionID, buffer)
	if err != nil {
		log.Printf("[LanguageService] ML Client error: %v", err)
		return nil, err
	}

	// Update VoiceSession if a confident language is detected
	if topLang, ok := result.TopLanguage(0.70); ok {
		// Log the detection. 
		// Note: We might update the DB session here (e.g., s.voiceService.SetLanguage(sessionID, topLang))
		log.Printf("[LanguageService] Confident language detected for %s: %s (score=%.2f)", sessionID, topLang, result.Probabilities[0].Score)
	}

	return result, nil
}
