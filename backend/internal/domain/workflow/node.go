package workflow

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// NodeType identifies the kind of work a node performs.
type NodeType string

const (
	// Phase 04 — implemented executors
	NodeTypeStart     NodeType = "START"
	NodeTypeEnd       NodeType = "END"
	NodeTypeNoop      NodeType = "NOOP"
	NodeTypeTransform NodeType = "TRANSFORM"
	NodeTypeCondition NodeType = "CONDITION"

	// Future executors — registered but not yet implemented
	NodeTypeVoiceAgent        NodeType = "VOICE_AGENT"
	NodeTypeLanguageDetection NodeType = "LANGUAGE_DETECTION"
	NodeTypeASR               NodeType = "ASR"
	NodeTypeTTS               NodeType = "TTS"
	NodeTypeOCR               NodeType = "OCR"
	NodeTypeDocumentAnalysis  NodeType = "DOCUMENT_ANALYSIS"
	NodeTypeFluencyAnalysis   NodeType = "FLUENCY_ANALYSIS"
	NodeTypeEvidenceValidation NodeType = "EVIDENCE_VALIDATION"
	NodeTypeRule              NodeType = "RULE"
	NodeTypeDecision          NodeType = "DECISION"
	NodeTypeHumanReview       NodeType = "HUMAN_REVIEW"
	NodeTypeHTTP              NodeType = "HTTP"
)

// AllNodeTypes contains every registered node type.
var AllNodeTypes = map[NodeType]bool{
	NodeTypeStart: true, NodeTypeEnd: true, NodeTypeNoop: true,
	NodeTypeTransform: true, NodeTypeCondition: true,
	NodeTypeVoiceAgent: true, NodeTypeLanguageDetection: true,
	NodeTypeASR: true, NodeTypeTTS: true, NodeTypeOCR: true,
	NodeTypeDocumentAnalysis: true, NodeTypeFluencyAnalysis: true,
	NodeTypeEvidenceValidation: true, NodeTypeRule: true,
	NodeTypeDecision: true, NodeTypeHumanReview: true, NodeTypeHTTP: true,
}

// NodePosition holds React Flow-compatible coordinates.
type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// RetryPolicy configures retry behaviour for a node.
type RetryPolicy struct {
	MaxAttempts int `json:"max_attempts"`
	BackoffMs   int `json:"backoff_ms"`
}

// WorkflowNode is a single vertex in the workflow graph.
type WorkflowNode struct {
	ID                uuid.UUID              `json:"id"`
	WorkflowVersionID uuid.UUID              `json:"workflow_version_id"`
	NodeKey           string                 `json:"node_key"` // user-defined stable key, e.g. "node_language"
	Type              NodeType               `json:"type"`
	Name              string                 `json:"name"`
	Config            map[string]interface{} `json:"config"`
	Position          NodePosition           `json:"position"`
	TimeoutMs         int                    `json:"timeout_ms,omitempty"`
	RetryPolicy       *RetryPolicy           `json:"retry_policy,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

func (n *WorkflowNode) ConfigBytes() ([]byte, error) {
	if n.Config == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(n.Config)
}

func (n *WorkflowNode) PositionBytes() ([]byte, error) {
	return json.Marshal(n.Position)
}
