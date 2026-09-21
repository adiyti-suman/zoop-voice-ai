package verification

import (
	"testing"
)

func TestStatusTransitions(t *testing.T) {
	tests := []struct {
		current  VerificationStatus
		next     VerificationStatus
		expected bool
	}{
		{StatusCreated, StatusCollecting, true},
		{StatusCollecting, StatusProcessing, true},
		{StatusProcessing, StatusEvaluating, true},
		{StatusEvaluating, StatusApproved, true},
		{StatusEvaluating, StatusReviewRequired, true},
		{StatusReviewRequired, StatusReviewing, true},
		{StatusReviewing, StatusApproved, true},

		// Invalid transitions
		{StatusApproved, StatusProcessing, false},
		{StatusRejected, StatusProcessing, false},
		{StatusFailed, StatusProcessing, false},
		{StatusCancelled, StatusCollecting, false},
		{StatusCompleted, StatusProcessing, false},
	}

	for _, tt := range tests {
		result := IsValidTransition(tt.current, tt.next)
		if result != tt.expected {
			t.Errorf("Transition %s -> %s: expected %v, got %v", tt.current, tt.next, tt.expected, result)
		}
	}
}

func TestTerminalStates(t *testing.T) {
	if !IsTerminalState(StatusApproved) {
		t.Errorf("Expected APPROVED to be terminal") // Wait, APPROVED is terminal conceptually but IsTerminalState only checks Completed, Failed, Cancelled, Expired?
		// Ah, let's check IsTerminalState implementation.
	}
}
