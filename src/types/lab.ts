// Shared types for lab-related functionality

export type Credential = {
  id: string;
  label: string;
  username: string;
  password: string;
  url?: string;
  expiresAt: string;
  notes?: string;
};

export type LabSession = {
  id: string;
  name: string;
  status: "provisioning" | "ready" | "error" | "expired" | "starting";
  startedAt?: string;
  endsAt?: string;
  owner: { name: string; email: string };
  credentials: Credential[];
};

export type LabStatus = LabSession['status'];

// Utility function to get badge variant for lab status
export function getLabBadgeVariant(status: LabStatus) {
  switch (status) {
    case 'ready':
      return 'default' as const;
    case 'provisioning':
    case 'starting':
      return 'secondary' as const;
    case 'error':
    case 'expired':
      return 'destructive' as const;
    default:
      return 'secondary' as const;
  }
}
