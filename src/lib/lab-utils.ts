// Utility functions for lab data conversion and manipulation

import { LabResponse } from "@/lib/api";
import { LabSession, Credential } from "@/types/lab";

// Convert backend LabResponse to frontend LabSession format
export function convertLabResponse(labResponse: LabResponse): LabSession {
  return {
    id: labResponse.id,
    name: labResponse.name,
    status: labResponse.status,
    startedAt: labResponse.started_at,
    endsAt: labResponse.ends_at,
    owner: labResponse.owner || { name: "Unknown", email: "unknown" },
    credentials: labResponse.credentials.map(cred => ({
      id: cred.id,
      label: cred.label,
      username: cred.username,
      password: cred.password,
      url: cred.url,
      expiresAt: cred.expires_at,
      notes: cred.notes,
    })),
  };
}

// Convert multiple lab responses
export function convertLabResponses(labResponses: LabResponse[]): LabSession[] {
  return labResponses.map(convertLabResponse);
}
