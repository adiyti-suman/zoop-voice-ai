// ── Status enums ─────────────────────────────────────────────────────────────

export type WorkflowStatus = "DRAFT" | "ACTIVE" | "ARCHIVED";

export type WorkflowVersionStatus =
  | "DRAFT"
  | "VALIDATING"
  | "PUBLISHED"
  | "ARCHIVED";

export type WorkflowRunStatus =
  | "PENDING"
  | "RUNNING"
  | "PAUSED"
  | "WAITING"
  | "COMPLETED"
  | "FAILED"
  | "CANCELLED"
  | "TIMED_OUT";

export type NodeRunStatus =
  | "PENDING"
  | "RUNNING"
  | "WAITING"
  | "COMPLETED"
  | "FAILED"
  | "SKIPPED"
  | "CANCELLED"
  | "TIMED_OUT";

export type NodeType =
  | "START"
  | "END"
  | "NOOP"
  | "CONDITION"
  | "TRANSFORM"
  | "VOICE_AGENT"
  | "LANGUAGE_DETECTION"
  | "ASR"
  | "TTS"
  | "OCR"
  | "DOCUMENT_ANALYSIS"
  | "FLUENCY_ANALYSIS"
  | "EVIDENCE_VALIDATION"
  | "RULE"
  | "DECISION"
  | "HUMAN_REVIEW"
  | "HTTP";

// ── Core entities ─────────────────────────────────────────────────────────────

export interface Workflow {
  id: string;
  name: string;
  description: string;
  status: WorkflowStatus;
  created_at: string;
  updated_at: string;
}

export interface WorkflowVersion {
  id: string;
  workflow_id: string;
  version: string;
  status: WorkflowVersionStatus;
  created_at: string;
  published_at?: string;
  archived_at?: string;
}

// ── React Flow compatible graph types ─────────────────────────────────────────

/** Position compatible with React Flow's NodePositionChange */
export interface NodePosition {
  x: number;
  y: number;
}

export interface RetryPolicy {
  max_attempts: number;
  backoff_ms: number;
}

/** Backend WorkflowNode mapped for React Flow usage */
export interface WorkflowNode {
  id: string;
  workflow_version_id: string;
  node_key: string;
  type: NodeType;
  name: string;
  config: Record<string, unknown>;
  position: NodePosition;
  timeout_ms?: number;
  retry_policy?: RetryPolicy;
  created_at: string;
}

/** Backend WorkflowEdge mapped for React Flow usage */
export interface WorkflowEdge {
  id: string;
  workflow_version_id: string;
  source_node_key: string; // maps to React Flow `source`
  target_node_key: string; // maps to React Flow `target`
  condition?: string;
  metadata?: Record<string, unknown>;
  created_at: string;
}

/** Complete graph for a workflow version */
export interface WorkflowGraph {
  version: WorkflowVersion;
  nodes: WorkflowNode[];
  edges: WorkflowEdge[];
}

// ── Runs ──────────────────────────────────────────────────────────────────────

export interface RunError {
  code: string;
  message: string;
  node_key?: string;
}

export interface WorkflowRun {
  id: string;
  verification_id: string;
  workflow_version_id: string;
  status: WorkflowRunStatus;
  current_node_key?: string;
  input: Record<string, unknown>;
  output?: Record<string, unknown>;
  error?: RunError;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface NodeRun {
  id: string;
  workflow_run_id: string;
  node_key: string;
  attempt: number;
  status: NodeRunStatus;
  input?: Record<string, unknown>;
  output?: Record<string, unknown>;
  error?: RunError;
  started_at?: string;
  completed_at?: string;
  duration_ms?: number;
  created_at: string;
  updated_at: string;
}

export interface WorkflowRunDetail {
  run: WorkflowRun;
  node_runs: NodeRun[];
}

// ── Request types ─────────────────────────────────────────────────────────────

export interface CreateWorkflowRequest {
  name: string;
  description?: string;
}

export interface NodeInput {
  node_key: string;
  type: NodeType;
  name: string;
  config?: Record<string, unknown>;
  position?: NodePosition;
  timeout_ms?: number;
  retry_policy?: RetryPolicy;
}

export interface EdgeInput {
  source_node_key: string;
  target_node_key: string;
  condition?: string;
  metadata?: Record<string, unknown>;
}

export interface CreateVersionRequest {
  version: string;
  nodes: NodeInput[];
  edges: EdgeInput[];
}

export interface StartRunRequest {
  verification_id: string;
  input?: Record<string, unknown>;
}

// ── Validation ────────────────────────────────────────────────────────────────

export interface ValidationError {
  code: string;
  message: string;
}

export interface ValidationResult {
  valid: boolean;
  errors: ValidationError[];
  warnings: ValidationError[];
}
