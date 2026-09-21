package workflow

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// WorkflowEdge is a directed connection between two nodes in the graph.
type WorkflowEdge struct {
	ID                uuid.UUID              `json:"id"`
	WorkflowVersionID uuid.UUID              `json:"workflow_version_id"`
	SourceNodeKey     string                 `json:"source_node_key"`
	TargetNodeKey     string                 `json:"target_node_key"`
	Condition         string                 `json:"condition,omitempty"` // safe expression, e.g. "result.status == \"ok\""
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
}

func (e *WorkflowEdge) MetadataBytes() ([]byte, error) {
	if e.Metadata == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(e.Metadata)
}
