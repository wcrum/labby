// Shared hook for fetching and managing lab data

import { useEffect, useState } from "react";
import { apiService, LabResponse } from "@/lib/api";
import { LabSession } from "@/types/lab";
import { convertLabResponse } from "@/lib/lab-utils";

export interface UseLabDataOptions {
  pollInterval?: number;
  autoFetch?: boolean;
}

export function useLabData(labId: string, options: UseLabDataOptions = {}) {
  const { pollInterval = 15000, autoFetch = true } = options;
  const [lab, setLab] = useState<LabSession | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchLab = async () => {
    try {
      const labResponse = await apiService.getLab(labId);
      const labSession = convertLabResponse(labResponse);
      setLab(labSession);
      setError(null);
      return labSession;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to load lab";
      setError(errorMessage);
      setLab(null);
      throw err;
    }
  };

  useEffect(() => {
    if (!labId || !autoFetch) return;

    let live = true;
    
    const initialFetch = async () => {
      setLoading(true);
      try {
        await fetchLab();
      } finally {
        if (live) {
          setLoading(false);
        }
      }
    };

    initialFetch();
    
    const poll = setInterval(async () => {
      if (live) {
        try {
          await fetchLab();
        } catch (err) {
          // Error is already handled in fetchLab
        }
      }
    }, pollInterval);
    
    return () => { 
      live = false; 
      clearInterval(poll); 
    };
  }, [labId, pollInterval, autoFetch]);

  return {
    lab,
    loading,
    error,
    refetch: fetchLab,
    setLab,
    setError
  };
}

// Hook for fetching all labs (admin functionality)
export function useAllLabsData(options: UseLabDataOptions = {}) {
  const { pollInterval = 10000, autoFetch = true } = options;
  const [labs, setLabs] = useState<LabSession[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchAllLabs = async () => {
    try {
      const labResponses = await apiService.getAllLabs();
      const labSessions = labResponses.map(convertLabResponse);
      setLabs(labSessions);
      setError(null);
      return labSessions;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : "Failed to load labs";
      setError(errorMessage);
      setLabs([]);
      throw err;
    }
  };

  useEffect(() => {
    if (!autoFetch) return;

    let live = true;
    
    const initialFetch = async () => {
      setLoading(true);
      try {
        await fetchAllLabs();
      } finally {
        if (live) {
          setLoading(false);
        }
      }
    };

    initialFetch();
    
    const poll = setInterval(async () => {
      if (live) {
        try {
          await fetchAllLabs();
        } catch (err) {
          // Error is already handled in fetchAllLabs
        }
      }
    }, pollInterval);
    
    return () => { 
      live = false; 
      clearInterval(poll); 
    };
  }, [pollInterval, autoFetch]);

  return {
    labs,
    loading,
    error,
    refetch: fetchAllLabs,
    setLabs,
    setError
  };
}
