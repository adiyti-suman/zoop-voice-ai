package workflow

import "testing"

func TestVersionTransitions(t *testing.T) {
	tests := []struct {
		from WorkflowVersionStatus
		to   WorkflowVersionStatus
		ok   bool
	}{
		{VersionStatusDraft, VersionStatusValidating, true},
		{VersionStatusDraft, VersionStatusArchived, true},
		{VersionStatusValidating, VersionStatusPublished, true},
		{VersionStatusValidating, VersionStatusDraft, true},
		{VersionStatusPublished, VersionStatusArchived, true},

		// Invalid
		{VersionStatusPublished, VersionStatusDraft, false},
		{VersionStatusPublished, VersionStatusValidating, false},
		{VersionStatusArchived, VersionStatusDraft, false},
	}
	for _, tt := range tests {
		got := IsVersionTransitionValid(tt.from, tt.to)
		if got != tt.ok {
			t.Errorf("IsVersionTransitionValid(%s→%s) = %v, want %v", tt.from, tt.to, got, tt.ok)
		}
	}
}

func TestRunTransitions(t *testing.T) {
	tests := []struct {
		from WorkflowRunStatus
		to   WorkflowRunStatus
		ok   bool
	}{
		{RunStatusPending, RunStatusRunning, true},
		{RunStatusRunning, RunStatusCompleted, true},
		{RunStatusRunning, RunStatusFailed, true},
		{RunStatusRunning, RunStatusCancelled, true},
		{RunStatusRunning, RunStatusWaiting, true},
		{RunStatusWaiting, RunStatusRunning, true},

		// Invalid
		{RunStatusCompleted, RunStatusRunning, false},
		{RunStatusFailed, RunStatusRunning, false},
		{RunStatusCancelled, RunStatusCompleted, false},
	}
	for _, tt := range tests {
		got := IsRunTransitionValid(tt.from, tt.to)
		if got != tt.ok {
			t.Errorf("IsRunTransitionValid(%s→%s) = %v, want %v", tt.from, tt.to, got, tt.ok)
		}
	}
}

func TestIsRunTerminal(t *testing.T) {
	for _, s := range []WorkflowRunStatus{RunStatusCompleted, RunStatusFailed, RunStatusCancelled, RunStatusTimedOut} {
		if !IsRunTerminal(s) {
			t.Errorf("expected %s to be terminal", s)
		}
	}
	for _, s := range []WorkflowRunStatus{RunStatusPending, RunStatusRunning, RunStatusWaiting, RunStatusPaused} {
		if IsRunTerminal(s) {
			t.Errorf("expected %s to NOT be terminal", s)
		}
	}
}
