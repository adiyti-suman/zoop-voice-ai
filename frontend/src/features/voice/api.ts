import type { VoiceSession, CreateSessionRequest } from "./types";

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

export function createVoiceSession(req: CreateSessionRequest): Promise<VoiceSession> {
  return request(`${BASE}/voice/sessions`, {
    method: "POST",
    body: JSON.stringify(req),
  });
}

export function getVoiceSession(id: string): Promise<VoiceSession> {
  return request(`${BASE}/voice/sessions/${id}`);
}

export function pauseVoiceSession(id: string): Promise<{ status: string }> {
  return request(`${BASE}/voice/sessions/${id}/pause`, { method: "POST" });
}

export function resumeVoiceSession(id: string): Promise<{ status: string }> {
  return request(`${BASE}/voice/sessions/${id}/resume`, { method: "POST" });
}

export function stopVoiceSession(id: string): Promise<{ status: string }> {
  return request(`${BASE}/voice/sessions/${id}/stop`, { method: "POST" });
}

export function getWebSocketURL(sessionId: string): string {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  // For local development with Next.js rewrites, or direct backend hit:
  return `${protocol}//${window.location.host}${BASE}/voice/sessions/${sessionId}/stream`;
}
