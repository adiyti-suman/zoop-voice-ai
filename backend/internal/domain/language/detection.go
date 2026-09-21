package language

import (
	"time"
)

// LanguageProbability represents the ML model's confidence in a detected language.
type LanguageProbability struct {
	Language Code    `json:"language"`
	Score    float32 `json:"score"` // 0.0 to 1.0
}

// DetectionResult is the output from the ML service.
type DetectionResult struct {
	SessionID   string                `json:"session_id"`
	Probabilities []LanguageProbability `json:"probabilities"`
	IsSpeech    bool                  `json:"is_speech"` // Did it detect any human speech?
	Timestamp   time.Time             `json:"timestamp"`
}

// TopLanguage returns the language with the highest probability, provided it meets the threshold.
func (r *DetectionResult) TopLanguage(threshold float32) (Code, bool) {
	if len(r.Probabilities) == 0 || !r.IsSpeech {
		return "", false
	}
	top := r.Probabilities[0]
	if top.Score >= threshold {
		return top.Language, true
	}
	return "", false
}
