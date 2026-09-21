export type VerificationStatus =
  | "CREATED"
  | "COLLECTING"
  | "PROCESSING"
  | "EVALUATING"
  | "REVIEW_REQUIRED"
  | "REVIEWING"
  | "MORE_INFORMATION_REQUIRED"
  | "APPROVED"
  | "REJECTED"
  | "FAILED"
  | "CANCELLED"
  | "EXPIRED"
  | "COMPLETED";

export interface Verification {
  id: string;
  external_id?: string;
  type: string;
  status: VerificationStatus;
  subject_id?: string;
  workflow_id?: string;
  workflow_version?: string;
  language?: string;
  locale?: string;
  metadata: Record<string, any>;
  created_at: string;
  updated_at: string;
  started_at?: string;
  completed_at?: string;
  expires_at?: string;
}

export interface CreateVerificationRequest {
  type: string;
  subject_id?: string;
  workflow_id?: string;
  workflow_version?: string;
  language?: string;
  locale?: string;
  metadata?: Record<string, any>;
}
