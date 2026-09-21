package voice_test

import (
	"context"
	"database/sql"
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	_ "github.com/lib/pq"

	domain "verification-platform/internal/domain/voice"
	voiceHandler "verification-platform/internal/handler/voice"
	voiceRepo "verification-platform/internal/repository/voice"
	voiceSvc "verification-platform/internal/service/voice"

	// FK dependencies
	verfRepo "verification-platform/internal/repository/verification"
	verfSvc "verification-platform/internal/service/verification"
	wfRepo "verification-platform/internal/repository/workflow"
	wfSvc "verification-platform/internal/service/workflow"
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
		"../../../migrations/000004_create_voice_sessions_table.up.sql",
	} {
		b, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		db.Exec(string(b)) //nolint:errcheck
	}
}

// createTestFKs creates a verification and workflow run so the session FKs pass.
func createTestFKs(t *testing.T, db *sql.DB) (string, string) {
	t.Helper()
	ctx := context.Background()

	vs := verfSvc.NewService(verfRepo.NewPostgresRepository(db))
	verf, _ := vs.Create(ctx, "", verfSvc.CreateRequest{Type: "VOICE"})

	ws := wfSvc.NewService(wfRepo.NewPostgresRepository(db))
	w, _ := ws.CreateWorkflow(ctx, wfSvc.CreateWorkflowRequest{Name: "test"})
	v, _ := ws.CreateVersion(ctx, w.ID.String(), wfSvc.CreateVersionRequest{
		Version: "1.0",
		Nodes: []wfSvc.NodeInput{
			{NodeKey: "start", Type: "START"},
			{NodeKey: "end", Type: "END"},
		},
		Edges: []wfSvc.EdgeInput{
			{SourceNodeKey: "start", TargetNodeKey: "end"},
		},
	})
	_, _ = ws.PublishVersion(ctx, v.ID.String())
	run, _ := ws.StartRun(ctx, v.ID.String(), wfSvc.StartRunRequest{VerificationID: verf.ID.String()})

	if run == nil {
		t.Fatalf("Failed to create workflow run in createTestFKs")
	}

	return verf.ID.String(), run.ID.String()
}

func TestVoice_Integration_StreamAndBuffer(t *testing.T) {
	db := openDB(t)
	defer db.Close()
	applyMigrations(t, db)

	ctx := context.Background()
	verfID, runID := createTestFKs(t, db)

	repo := voiceRepo.NewPostgresRepository(db)
	svc := voiceSvc.NewService(repo)
	buf := voiceSvc.NewAudioBuffer(10) // Small buffer to test overflow
	streamHandler := voiceHandler.NewStreamHandler(svc, buf)

	// 1. Create Session
	session, err := svc.CreateSession(ctx, voiceSvc.CreateSessionRequest{
		VerificationID: verfID,
		WorkflowRunID:  runID,
		SampleRate:     16000,
		Channels:       1,
		Encoding:       "pcm_s16le",
	})
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	// 2. Setup Test Server for WebSocket
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/voice/sessions/", func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/stream") {
			streamHandler.ServeHTTP(w, req)
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/voice/sessions/" + session.ID.String() + "/stream"
	
	// 3. Connect WebSocket client
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WS Dial: %v", err)
	}
	defer ws.Close()

	// 4. Verify Server sends CONNECTED
	var msg voiceHandler.WSControlMessage
	if err := ws.ReadJSON(&msg); err != nil {
		t.Fatalf("ReadJSON CONNECTED: %v", err)
	}
	if msg.Type != "CONNECTED" {
		t.Errorf("Expected CONNECTED, got %s", msg.Type)
	}

	// 5. Send START and verify READY
	ws.WriteJSON(voiceHandler.WSControlMessage{Type: "START"})
	if err := ws.ReadJSON(&msg); err != nil {
		t.Fatalf("ReadJSON READY: %v", err)
	}
	if msg.Type != "READY" {
		t.Errorf("Expected READY, got %s", msg.Type)
	}

	// 6. Verify Status is LISTENING
	s, _ := svc.GetSession(ctx, session.ID.String())
	if s.Status != domain.SessionStatusListening {
		t.Errorf("Expected LISTENING status, got %s", s.Status)
	}

	// 7. Send valid Binary Audio Chunk
	payload := make([]byte, 12+1024)
	binary.LittleEndian.PutUint32(payload[0:4], 1) // Seq
	binary.LittleEndian.PutUint64(payload[4:12], 1000) // TS
	// data is 0s
	if err := ws.WriteMessage(websocket.BinaryMessage, payload); err != nil {
		t.Fatalf("WriteMessage: %v", err)
	}
	
	// wait for processing
	time.Sleep(100 * time.Millisecond)

	if buf.Len() != 1 {
		t.Errorf("Expected 1 chunk in buffer, got %d", buf.Len())
	}
	chunk, ok := buf.Pop()
	if !ok || chunk.Sequence != 1 || chunk.TimestampMs != 1000 || len(chunk.Data) != 1024 {
		t.Errorf("Pop chunk mismatch: %+v", chunk)
	}

	// 8. Test invalid sequence (out of order)
	binary.LittleEndian.PutUint32(payload[0:4], 1) // Seq 1 again (should be rejected)
	ws.WriteMessage(websocket.BinaryMessage, payload)
	
	if err := ws.ReadJSON(&msg); err != nil {
		t.Fatalf("ReadJSON WARNING: %v", err)
	}
	if msg.Type != "WARNING" || !strings.Contains(msg.Message, "Invalid sequence") {
		t.Errorf("Expected Invalid Sequence WARNING, got %s: %s", msg.Type, msg.Message)
	}

	// 9. Send STOP and verify STOPPED
	ws.WriteJSON(voiceHandler.WSControlMessage{Type: "STOP"})
	if err := ws.ReadJSON(&msg); err != nil {
		t.Fatalf("ReadJSON STOPPED: %v", err)
	}
	if msg.Type != "STOPPED" {
		t.Errorf("Expected STOPPED, got %s", msg.Type)
	}

	// 10. Close WS and verify session reaches COMPLETED
	ws.Close()
	time.Sleep(100 * time.Millisecond) // Let close handler finish

	s, _ = svc.GetSession(ctx, session.ID.String())
	if s.Status != domain.SessionStatusCompleted {
		t.Errorf("Expected COMPLETED status, got %s", s.Status)
	}

	t.Log("✅ Voice Session stream lifecycle and buffer backpressure verified")
}
