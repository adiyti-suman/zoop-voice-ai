import { Verification, CreateVerificationRequest } from "./types";

const API_BASE_URL = "http://localhost:8081/api/v1";

export async function createVerification(
  req: CreateVerificationRequest,
  idempotencyKey?: string
): Promise<Verification> {
  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };
  if (idempotencyKey) {
    headers["Idempotency-Key"] = idempotencyKey;
  }

  const res = await fetch(`${API_BASE_URL}/verifications`, {
    method: "POST",
    headers,
    body: JSON.stringify(req),
  });

  if (!res.ok) {
    throw new Error(`Failed to create verification: ${res.statusText}`);
  }
  return res.json();
}

export async function getVerification(id: string): Promise<Verification> {
  const res = await fetch(`${API_BASE_URL}/verifications/${id}`);
  if (!res.ok) {
    throw new Error(`Failed to fetch verification: ${res.statusText}`);
  }
  return res.json();
}
