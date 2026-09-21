import type {
  Workflow,
  WorkflowVersion,
  WorkflowRunDetail,
  CreateWorkflowRequest,
  CreateVersionRequest,
  StartRunRequest,
  ValidationResult,
} from "./types";

const BASE = "/api/v1";

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error ?? `HTTP ${res.status}`);
  }
  return res.json() as Promise<T>;
}

// ── Workflow CRUD ─────────────────────────────────────────────────────────────

export function createWorkflow(req: CreateWorkflowRequest): Promise<Workflow> {
  return request(`${BASE}/workflows`, {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export function getWorkflow(id: string): Promise<Workflow> {
  return request(`${BASE}/workflows/${id}`);
}

// ── Versions ──────────────────────────────────────────────────────────────────

export function createVersion(
  workflowId: string,
  req: CreateVersionRequest
): Promise<WorkflowVersion> {
  return request(`${BASE}/workflows/${workflowId}/versions`, {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export function validateVersion(versionId: string): Promise<ValidationResult> {
  return request(`${BASE}/workflow-versions/${versionId}/validate`, {
    method: "POST",
  });
}

export function publishVersion(versionId: string): Promise<WorkflowVersion> {
  return request(`${BASE}/workflow-versions/${versionId}/publish`, {
    method: "POST",
  });
}

// ── Runs ──────────────────────────────────────────────────────────────────────

export function startRun(
  versionId: string,
  req: StartRunRequest
): Promise<WorkflowRunDetail> {
  return request(`${BASE}/workflow-versions/${versionId}/runs`, {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export function getRun(runId: string): Promise<WorkflowRunDetail> {
  return request(`${BASE}/workflow-runs/${runId}`);
}

export function cancelRun(runId: string): Promise<{ status: string }> {
  return request(`${BASE}/workflow-runs/${runId}/cancel`, { method: "POST" });
}
