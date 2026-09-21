package verification_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	domain "verification-platform/internal/domain/verification"
	verfRepo "verification-platform/internal/repository/verification"
	verfSvc "verification-platform/internal/service/verification"
)

func TestIntegration_LifecycleAndIdempotency(t *testing.T) {
	// Skip if running in an environment without Postgres
	connStr := "postgres://postgres:password@localhost:5434/verification?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer db.Close()
	
	if err := db.Ping(); err != nil {
		t.Skip("Postgres not available, skipping integration test")
	}

	// Apply migrations
	mig1, err := os.ReadFile("../../../migrations/000001_create_verifications_table.up.sql")
	if err == nil {
		db.Exec(string(mig1))
	}
	mig2, err := os.ReadFile("../../../migrations/000002_create_idempotency_keys_table.up.sql")
	if err == nil {
		db.Exec(string(mig2))
	}

	repo := verfRepo.NewPostgresRepository(db)
	svc := verfSvc.NewService(repo)

	ctx := context.Background()
	ik := "test-idem-" + uuid.New().String()

	lang := "en-US"
	req := verfSvc.CreateRequest{
		Type:     "VOICE",
		Language: &lang,
	}

	// 1. Create Verification
	v1, err := svc.Create(ctx, ik, req)
	if err != nil {
		t.Fatalf("Failed to create verification: %v", err)
	}

	if v1.Status != domain.StatusCreated {
		t.Errorf("Expected status CREATED, got %s", v1.Status)
	}

	// 2. Idempotency Check
	v2, err := svc.Create(ctx, ik, req)
	if err != nil {
		t.Fatalf("Failed idempotent create: %v", err)
	}

	if v1.ID != v2.ID {
		t.Errorf("Idempotency failed. Expected ID %s, got %s", v1.ID, v2.ID)
	}

	// 3. Lifecycle Transitions
	transitions := []struct {
		expected domain.VerificationStatus
		next     domain.VerificationStatus
	}{
		{domain.StatusCreated, domain.StatusCollecting},
		{domain.StatusCollecting, domain.StatusProcessing},
		{domain.StatusProcessing, domain.StatusEvaluating},
		{domain.StatusEvaluating, domain.StatusReviewRequired},
		{domain.StatusReviewRequired, domain.StatusReviewing},
		{domain.StatusReviewing, domain.StatusApproved},
	}

	for _, tr := range transitions {
		err := svc.Transition(ctx, v1.ID.String(), tr.expected, tr.next, "SYSTEM", "")
		if err != nil {
			t.Fatalf("Failed transition from %s to %s: %v", tr.expected, tr.next, err)
		}

		vCheck, err := svc.Get(ctx, v1.ID.String())
		if err != nil {
			t.Fatalf("Failed to fetch verification: %v", err)
		}
		if vCheck.Status != tr.next {
			t.Errorf("Expected status %s, got %s", tr.next, vCheck.Status)
		}
	}

	// 4. Failure scenario: Invalid transition
	err = svc.Transition(ctx, v1.ID.String(), domain.StatusApproved, domain.StatusProcessing, "SYSTEM", "")
	if err == nil {
		t.Errorf("Expected error transitioning from APPROVED to PROCESSING")
	} else if err != domain.ErrInvalidStateTransition {
		t.Errorf("Expected ErrInvalidStateTransition, got %v", err)
	}

	// 5. Expiration (manual DB update for test)
	past := time.Now().Add(-24 * time.Hour)
	_, err = db.Exec("UPDATE verifications SET expires_at = $1 WHERE id = $2", past, v1.ID)
	if err != nil {
		t.Fatalf("Failed to expire verification manually: %v", err)
	}
}
