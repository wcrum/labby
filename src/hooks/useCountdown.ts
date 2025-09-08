// Shared countdown hook for lab timers

import { useEffect, useState, useMemo } from "react";

export interface CountdownResult {
  ms: number;
  mins: number;
  secs: number;
  pct: number;
  progressValue: number;
}

export function useCountdown(to?: string, totalMs?: number): CountdownResult {
  const [ms, setMs] = useState(0);
  
  useEffect(() => {
    if (!to) return;
    const t = new Date(to).getTime();
    setMs(Math.max(0, t - Date.now()));
    const i = setInterval(() => setMs(Math.max(0, t - Date.now())), 1000);
    return () => clearInterval(i);
  }, [to]);
  
  const mins = Math.floor(ms / 60000);
  const secs = Math.floor((ms % 60000) / 1000);
  
  const pct = useMemo(() => {
    const total = totalMs && totalMs > 0 ? totalMs : 60 * 60 * 1000;
    return Math.min(100, Math.max(0, (ms / total) * 100));
  }, [ms, totalMs]);
  
  const progressValue = Math.min(100, Math.max(0, 100 - pct));
  
  return { ms, mins, secs, pct, progressValue };
}

// Hook for multiple countdowns (used in admin page)
export function useAllCountdowns(labs: Array<{ endsAt?: string; startedAt?: string }>) {
  const [currentTime, setCurrentTime] = useState(Date.now());

  useEffect(() => {
    const interval = setInterval(() => setCurrentTime(Date.now()), 1000);
    return () => clearInterval(interval);
  }, []);

  return labs.map(lab => {
    if (!lab.endsAt) {
      return { countdown: { mins: 0, secs: 0, pct: 0 }, progressValue: 0 };
    }

    const endTime = new Date(lab.endsAt).getTime();
    const ms = Math.max(0, endTime - currentTime);
    const mins = Math.floor(ms / 60000);
    const secs = Math.floor((ms % 60000) / 1000);
    
    const totalMs = lab.startedAt && lab.endsAt ? Math.max(0, new Date(lab.endsAt).getTime() - new Date(lab.startedAt).getTime()) : 60 * 60 * 1000;
    const pct = Math.min(100, Math.max(0, (ms / totalMs) * 100));
    const progressValue = Math.min(100, Math.max(0, 100 - pct));

    return { countdown: { mins, secs, pct }, progressValue };
  });
}
