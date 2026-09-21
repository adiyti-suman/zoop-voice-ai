package voice

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	domain "verification-platform/internal/domain/voice"
	svc "verification-platform/internal/service/voice"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all for MVP
	},
}

// WSControlMessage is the JSON payload for text frames
type WSControlMessage struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}

type StreamHandler struct {
	svc    svc.Service
	buffer *svc.AudioBuffer
}

func NewStreamHandler(s svc.Service, buf *svc.AudioBuffer) *StreamHandler {
	return &StreamHandler{svc: s, buffer: buf}
}

// GET /api/v1/voice/sessions/{id}/stream
func (h *StreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sessionID := extractID(r.URL.Path, "/api/v1/voice/sessions/")
	sessionID = strings.TrimSuffix(sessionID, "/stream")

	// Verify session exists
	_, err := h.svc.GetSession(r.Context(), sessionID)
	if err != nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Move to CONNECTING, then CONNECTED
	_ = h.svc.TransitionStatus(r.Context(), sessionID, domain.SessionStatusConnecting)
	_ = h.svc.TransitionStatus(r.Context(), sessionID, domain.SessionStatusConnected)

	h.sendControl(conn, "CONNECTED", "")

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WS read err: %v", err)
			break
		}

		switch messageType {
		case websocket.TextMessage:
			h.handleTextFrame(conn, p, sessionID)
		case websocket.BinaryMessage:
			h.handleBinaryFrame(conn, p)
		}
	}

	// On disconnect, mark completed if not failed
	_ = h.svc.TransitionStatus(context.Background(), sessionID, domain.SessionStatusCompleted)
}

func (h *StreamHandler) handleTextFrame(conn *websocket.Conn, p []byte, sessionID string) {
	var msg WSControlMessage
	if err := json.Unmarshal(p, &msg); err != nil {
		h.sendControl(conn, "ERROR", "Invalid JSON")
		return
	}

	switch msg.Type {
	case "START":
		h.svc.TransitionStatus(context.Background(), sessionID, domain.SessionStatusListening)
		h.sendControl(conn, "READY", "")
	case "PAUSE":
		h.svc.TransitionStatus(context.Background(), sessionID, domain.SessionStatusPaused)
		h.sendControl(conn, "ACK", "Paused")
	case "RESUME":
		h.svc.TransitionStatus(context.Background(), sessionID, domain.SessionStatusListening)
		h.sendControl(conn, "ACK", "Resumed")
	case "STOP":
		h.svc.TransitionStatus(context.Background(), sessionID, domain.SessionStatusStopping)
		h.sendControl(conn, "STOPPED", "")
	case "PING":
		h.sendControl(conn, "PONG", "")
	default:
		h.sendControl(conn, "WARNING", "Unknown command")
	}
}

func (h *StreamHandler) handleBinaryFrame(conn *websocket.Conn, p []byte) {
	// Protocol: Sequence (4 bytes) + TimestampMs (8 bytes) + PCM data
	if len(p) < 12 {
		h.sendControl(conn, "ERROR", "Binary frame too small")
		return
	}

	seq := binary.LittleEndian.Uint32(p[0:4])
	ts := binary.LittleEndian.Uint64(p[4:12])
	data := p[12:]

	chunk := svc.AudioChunk{
		Sequence:    seq,
		TimestampMs: ts,
		Data:        data,
	}

	if err := h.buffer.Push(chunk); err != nil {
		if err == domain.ErrSequenceInvalid {
			h.sendControl(conn, "WARNING", "Invalid sequence number")
		} else if err == domain.ErrBufferOverflow {
			h.sendControl(conn, "WARNING", "Buffer overflow, dropping frame")
		}
	}
}

func (h *StreamHandler) sendControl(conn *websocket.Conn, typ, msg string) {
	b, _ := json.Marshal(WSControlMessage{Type: typ, Message: msg})
	_ = conn.WriteMessage(websocket.TextMessage, b)
}
