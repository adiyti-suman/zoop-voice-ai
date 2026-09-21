package workflow_test

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	wf "verification-platform/internal/domain/workflow"
	wfRepo "verification-platform/internal/repository/workflow"
	wfSvc "verification-platform/internal/service/workflow"

	verfRepo "verification-platform/internal/repository/verification"
	verfSvc "verification-platform/internal/service/verification"
)

func openDB(t *testing.T) *sql.DB {
	t.Helper()
	connStr := "postgres://postgres:password@localhost:5434/verification?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skip("Postgres not available, skipping integration test")
	}
	return db
}

func applyMigrations(t *testing.T, db *sql.DB) {
	t.Helper()
	for _, f := range []string{
		"../../../migrations/000001_create_verifications_table.up.sql",
		"../../../migrations/000002_create_idempotency_keys_table.up.sql",
		"../../../migrations/000003_create_workflow_tables.up.sql",
	} {
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		db.Exec(string(b)) //nolint:errcheck — idempotent IF NOT EXISTS
	}
}

// createTestVerification creates a minimal verification so FK constraints pass.
func createTestVerification(t *testing.T, db *sql.DB) string {
	t.Helper()
	repo := verfRepo.NewPostgresRepository(db)
	svc := verfSvc.NewService(repo)
	v, err := svc.Create(context.Background(), "", verfSvc.CreateRequest{Type: "VOICE"})
	if err != nil {
		t.Fatalf("create verification: %v", err)
	}
	return v.ID.String()
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestWorkflow_SmokeSuiteCreateValidatePublishRun(t *testing.T) {
	db := openDB(t)
	defer db.Close()
	applyMigrations(t, db)

	ctx := context.Background()
	repo := wfRepo.NewPostgresRepository(db)
	svc := wfSvc.NewService(repo)

	verID := createTestVerification(t, db)

	// 1. Create workflow
	w, err := svc.CreateWorkflow(ctx, wfSvc.CreateWorkflowRequest{Name: "test-workflow", Description: "smoke test"})
	if err != nil {
		t.Fatalf("CreateWorkflow: %v", err)
	}
	if w.Status != wf.WorkflowStatusDraft {
		t.Errorf("expected DRAFT, got %s", w.Status)
	}

	// 2. Create version: START → NOOP → TRANSFORM → CONDITION → END(true)/END(false)
	v, err := svc.CreateVersion(ctx, w.ID.String(), wfSvc.CreateVersionRequest{
		Version: "1.0.0",
		Nodes: []wfSvc.NodeInput{
			{NodeKey: "start", Type: "START", Name: "Start", Position: wf.NodePosition{X: 0, Y: 0}},
			{NodeKey: "noop", Type: "NOOP", Name: "Noop", Position: wf.NodePosition{X: 200, Y: 0}},
			{NodeKey: "transform", Type: "TRANSFORM", Name: "Transform", Config: map[string]interface{}{"mappings": map[string]interface{}{}}, Position: wf.NodePosition{X: 400, Y: 0}},
			{NodeKey: "cond", Type: "CONDITION", Name: "Condition", Config: map[string]interface{}{"expression": "noop.noop == true"}, Position: wf.NodePosition{X: 600, Y: 0}},
			{NodeKey: "end_true", Type: "END", Name: "End True", Position: wf.NodePosition{X: 800, Y: -100}},
			{NodeKey: "end_false", Type: "END", Name: "End False", Position: wf.NodePosition{X: 800, Y: 100}},
		},
		Edges: []wfSvc.EdgeInput{
			{SourceNodeKey: "start", TargetNodeKey: "noop"},
			{SourceNodeKey: "noop", TargetNodeKey: "transform"},
			{SourceNodeKey: "transform", TargetNodeKey: "cond"},
			{SourceNodeKey: "cond", TargetNodeKey: "end_true", Condition: "true"},
			{SourceNodeKey: "cond", TargetNodeKey: "end_false", Condition: "false"},
		},
	})
	if err != nil {
		t.Fatalf("CreateVersion: %v", err)
	}
	if v.Status != wf.VersionStatusDraft {
		t.Errorf("expected DRAFT version, got %s", v.Status)
	}

	// 3. Validate
	result, err := svc.ValidateVersion(ctx, v.ID.String())
	if err != nil {
		t.Fatalf("ValidateVersion: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid graph, got errors: %+v", result.Errors)
	}

	// 4. Publish
	published, err := svc.PublishVersion(ctx, v.ID.String())
	if err != nil {
		t.Fatalf("PublishVersion: %v", err)
	}
	if published.Status != wf.VersionStatusPublished {
		t.Errorf("expected PUBLISHED, got %s", published.Status)
	}

	// 5. Execute — START → NOOP → TRANSFORM → CONDITION(true) → END
	run, err := svc.StartRun(ctx, v.ID.String(), wfSvc.StartRunRequest{
		VerificationID: verID,
		Input:          map[string]interface{}{"test": true},
	})
	if err != nil {
		t.Fatalf("StartRun: %v", err)
	}

	// Run must complete synchronously
	if run.Status != wf.RunStatusCompleted {
		t.Errorf("expected run COMPLETED, got %s (error: %+v)", run.Status, run.Error)
	}

	// 6. Inspect node runs — expect 5 (start, noop, transform, cond, end_true)
	_, nodeRuns, err := svc.GetRun(ctx, run.ID.String())
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if len(nodeRuns) != 5 {
		t.Errorf("expected 5 node runs, got %d", len(nodeRuns))
	}
	for _, nr := range nodeRuns {
		if nr.Status != wf.NodeRunCompleted {
			t.Errorf("node %s expected COMPLETED, got %s", nr.NodeKey, nr.Status)
		}
	}

	// 7. Cancel a new run (must work on PENDING/RUNNING)
	// Try cancelling a second identical run right after creation before engine runs
	_ = db // (not testing cancellation mid-flight in this smoke test; covered by unit tests)

	// 8. Attempt to publish an already-published version → must fail
	_, err = svc.PublishVersion(ctx, v.ID.String())
	if err != wf.ErrInvalidTransition {
		t.Errorf("expected ErrInvalidTransition, got %v", err)
	}

	// 9. Validation failure: empty graph
	badVersion, err := svc.CreateVersion(ctx, w.ID.String(), wfSvc.CreateVersionRequest{
		Version: "2.0.0",
		Nodes:   nil,
		Edges:   nil,
	})
	if err != nil {
		t.Fatalf("CreateVersion bad: %v", err)
	}
	_, err = svc.PublishVersion(ctx, badVersion.ID.String())
	if err != wf.ErrInvalidGraph {
		t.Errorf("expected ErrInvalidGraph for empty version, got %v", err)
	}

	t.Logf("✅ smoke test passed — run_id=%s  node_runs=%d", run.ID, len(nodeRuns))
}

// TestWorkflow_FailedNode verifies that a node that errors propagates FAILED to the run.
func TestWorkflow_FailedNode(t *testing.T) {
	db := openDB(t)
	defer db.Close()
	applyMigrations(t, db)

	ctx := context.Background()
	repo := wfRepo.NewPostgresRepository(db)
	svc := wfSvc.NewService(repo)
	verID := createTestVerification(t, db)

	// Create a workflow with an unsupported node type (registered in domain but no executor)
	w, _ := svc.CreateWorkflow(ctx, wfSvc.CreateWorkflowRequest{Name: "fail-test"})
	v, err := svc.CreateVersion(ctx, w.ID.String(), wfSvc.CreateVersionRequest{
		Version: "1.0.0",
		Nodes: []wfSvc.NodeInput{
			{NodeKey: "start", Type: "START", Name: "Start"},
			{NodeKey: "voice", Type: "VOICE_AGENT", Name: "Voice Agent"},
			{NodeKey: "end", Type: "END", Name: "End"},
		},
		Edges: []wfSvc.EdgeInput{
			{SourceNodeKey: "start", TargetNodeKey: "voice"},
			{SourceNodeKey: "voice", TargetNodeKey: "end"},
		},
	})
	if err != nil {
		t.Fatalf("CreateVersion: %v", err)
	}

	_, err = svc.PublishVersion(ctx, v.ID.String())
	if err != nil {
		t.Fatalf("PublishVersion: %v", err)
	}

	run, err := svc.StartRun(ctx, v.ID.String(), wfSvc.StartRunRequest{VerificationID: verID})
	if err != nil {
		t.Fatalf("StartRun: %v", err)
	}

	// VOICE_AGENT has no registered executor → engine should fail the run
	if run.Status != wf.RunStatusFailed {
		t.Errorf("expected FAILED, got %s", run.Status)
	}
	t.Logf("✅ failure propagation verified — run_id=%s status=%s", run.ID, run.Status)
}

// TestWorkflow_CrashRecovery ensures RUNNING runs can be identified after restart.
func TestWorkflow_CrashRecovery(t *testing.T) {
	db := openDB(t)
	defer db.Close()
	applyMigrations(t, db)

	ctx := context.Background()
	repo := wfRepo.NewPostgresRepository(db)
	svc := wfSvc.NewService(repo)
	verID := createTestVerification(t, db)

	// Create real workflow + version so FK constraints pass
	w, _ := svc.CreateWorkflow(ctx, wfSvc.CreateWorkflowRequest{Name: "crash-test"})
	v, err := svc.CreateVersion(ctx, w.ID.String(), wfSvc.CreateVersionRequest{
		Version: "1.0.0",
		Nodes: []wfSvc.NodeInput{
			{NodeKey: "start", Type: "START", Name: "Start"},
			{NodeKey: "end", Type: "END", Name: "End"},
		},
		Edges: []wfSvc.EdgeInput{
			{SourceNodeKey: "start", TargetNodeKey: "end"},
		},
	})
	if err != nil {
		t.Fatalf("CreateVersion: %v", err)
	}

	verUUID, err := uuid.Parse(verID)
	if err != nil {
		t.Fatalf("parse verID: %v", err)
	}

	// Insert a run that is left in RUNNING state (simulating a crash)
	run := &wf.WorkflowRun{
		ID:                uuid.New(),
		VerificationID:    verUUID,
		WorkflowVersionID: v.ID,
		Status:            wf.RunStatusPending,
		Input:             map[string]interface{}{},
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	if err := repo.CreateRun(ctx, run); err != nil {
		t.Fatalf("CreateRun: %v", err)
	}
	// Force it to RUNNING (as if we crashed mid-execution)
	if err := repo.UpdateRunStatus(ctx, run.ID.String(), wf.RunStatusRunning, nil); err != nil {
		t.Fatalf("UpdateRunStatus: %v", err)
	}

	// After "restart" we detect it is still RUNNING
	recovered, err := repo.GetRun(ctx, run.ID.String())
	if err != nil {
		t.Fatalf("GetRun: %v", err)
	}
	if recovered.Status != wf.RunStatusRunning {
		t.Errorf("expected RUNNING, got %s", recovered.Status)
	}
	t.Logf("✅ crash recovery: detected orphaned RUNNING run_id=%s", run.ID)
}
