export type SessionStatus =
  | "CREATED"
  | "CONNECTING"
  | "CONNECTED"
  | "LISTENING"
  | "PAUSED"
  | "STOPPING"
  | "COMPLETED"
  | "FAILED"
  | "CANCELLED";

export interface VoiceSession {
  id: string;
  verification_id: string;
  workflow_run_id: string;
  status: SessionStatus;
  language?: string;
  sample_rate: number;
  channels: number;
  encoding: string;
  started_at?: string;
  ended_at?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateSessionRequest {
  verification_id: string;
  workflow_run_id: string;
  sample_rate: number;
  channels: number;
  encoding: string;
}

// WebSocket Protocol
export interface WSControlMessage {
  type: string;
  message?: string;
}
