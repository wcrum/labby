"use client";

import React, { useEffect, useMemo, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Separator } from "@/components/ui/separator";
import { Progress } from "@/components/ui/progress";
import { Copy, Eye, EyeOff, Clock, ExternalLink, ShieldCheck, Info, Server, User, AlertCircle } from "lucide-react";
import { motion } from "framer-motion";
import { ProtectedRoute } from "@/components/auth/ProtectedRoute";
import { apiService, LabResponse } from "@/lib/api";
import { LabStartingView } from "@/components/lab/LabStartingView";
import { LabFailedView } from "@/components/lab/LabFailedView";
import { AppLayout } from "@/components/layout/AppLayout";

type Credential = {
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

// Convert backend LabResponse to frontend LabSession format
function convertLabResponse(labResponse: LabResponse): LabSession {
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

async function fetchLabSession(labId: string): Promise<LabSession> {
  const labResponse = await apiService.getLab(labId);
  return convertLabResponse(labResponse);
}



function useCountdown(to?: string, totalMs?: number) {
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
  return { ms, mins, secs, pct };
}

const MaskedSecret: React.FC<{ secret: string }> = ({ secret }) => {
  const [revealed, setRevealed] = useState(false);
  return (
    <div className="flex items-center gap-2">
      <Input type={revealed ? "text" : "password"} value={secret} readOnly className="font-mono"/>
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="secondary" size="icon" onClick={() => setRevealed((s) => !s)}>
              {revealed ? <EyeOff className="h-4 w-4"/> : <Eye className="h-4 w-4"/>}
            </Button>
          </TooltipTrigger>
          <TooltipContent>{revealed ? "Hide" : "Show"} secret</TooltipContent>
        </Tooltip>
      </TooltipProvider>
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="outline"
              size="icon"
              onClick={() => {
                try {
                  if (typeof navigator !== "undefined" && navigator.clipboard) navigator.clipboard.writeText(secret);
                } catch {}
              }}
            >
              <Copy className="h-4 w-4"/>
            </Button>
          </TooltipTrigger>
          <TooltipContent>Copy to clipboard</TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>
  );
};

function LabSessionPageContent() {
  const [labId, setLabId] = useState<string>("");

  // Read labId from URL parameters on client side
  useEffect(() => {
    if (typeof window !== 'undefined') {
      const urlParams = new URLSearchParams(window.location.search);
      const idFromUrl = urlParams.get('id');
      if (idFromUrl) {
        setLabId(idFromUrl);
      }
    }
  }, []);
  const [lab, setLab] = useState<LabSession | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [stopping, setStopping] = useState(false);

  const getBadgeVariant = (status: LabSession['status']) => {
    switch (status) {
      case 'ready':
        return 'default';
      case 'provisioning':
      case 'starting':
        return 'secondary';
      case 'error':
        return 'destructive';
      default:
        return 'secondary';
    }
  };

  const totalMs = lab?.startedAt && lab?.endsAt ? Math.max(0, new Date(lab.endsAt).getTime() - new Date(lab.startedAt).getTime()) : undefined;
  const overallCountdown = useCountdown(lab?.endsAt, totalMs);
  const progressValue = Math.min(100, Math.max(0, 100 - overallCountdown.pct));

  useEffect(() => {
    let live = true;
    (async () => {
      setLoading(true);
      setError(null);
      try {
        const data = await fetchLabSession(labId);
        if (!live) return;
        setLab(data);
      } catch (err) {
        if (!live) return;
        setError(err instanceof Error ? err.message : "Failed to load lab");
        setLab(null);
      } finally {
        if (live) {
          setLoading(false);
        }
      }
    })();
    
    const poll = setInterval(async () => {
      try {
        const data = await fetchLabSession(labId);
        if (!live) return;
        setLab(data);
        setError(null);
      } catch (err) {
        if (!live) return;
        setError(err instanceof Error ? err.message : "Failed to load lab");
        setLab(null);
      }
    }, 15000);
    
    return () => { live = false; clearInterval(poll); };
  }, [labId]);

  const handleLabReady = () => {
    // Refresh the lab data when it becomes ready
    fetchLabSession(labId).then(setLab).catch(err => {
      setError(err instanceof Error ? err.message : "Failed to load lab");
    });
  };

  const handleStopLab = async () => {
    if (!lab || !confirm("Are you sure you want to stop this lab? This will delete all resources and cannot be undone.")) {
      return;
    }

    setStopping(true);
    try {
      // Stop the lab (which will also delete and cleanup resources)
      await apiService.stopLab(lab.id);
      
      // Redirect to labs page after successful stop
      window.location.href = '/labs';
    } catch (error) {
      console.error('Failed to stop lab:', error);
      alert('Failed to stop lab. Please try again.');
    } finally {
      setStopping(false);
    }
  };

  if (loading) {
    return (
      <div className="p-6 space-y-4">
        <div className="h-8 w-64 bg-muted animate-pulse rounded" />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[0,1].map((i) => (
            <div key={i} className="h-56 bg-muted animate-pulse rounded-2xl" />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <AppLayout>
        <div className="p-6 max-w-4xl mx-auto">
          <Card className="border-destructive">
            <CardHeader>
              <CardTitle className="text-destructive flex items-center gap-2">
                <AlertCircle className="h-5 w-5" />
                Lab Not Found
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-muted-foreground">{error}</p>
              <div className="flex gap-2">
                <Button onClick={() => window.location.href = '/labs'}>
                  Browse Available Labs
                </Button>
                <Button variant="outline" onClick={() => window.location.reload()}>
                  Try Again
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </AppLayout>
    );
  }

  if (!lab) {
    return (
      <div className="p-6 space-y-4">
        <div className="h-8 w-64 bg-muted animate-pulse rounded" />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[0,1].map((i) => (
            <div key={i} className="h-56 bg-muted animate-pulse rounded-2xl" />
          ))}
        </div>
      </div>
    );
  }

  // Show starting view if lab is in starting status
  if (lab.status === "starting" || lab.status === "provisioning") {
    return (
      <AppLayout>
        <LabStartingView labId={labId} onLabReady={handleLabReady} />
      </AppLayout>
    );
  }

  // Show failed view if lab is in error status
  if (lab.status === "error") {
    return (
      <AppLayout>
        <LabFailedView 
          labId={labId} 
          lab={lab} 
          onRetry={() => window.location.reload()}
          onCleanup={() => window.location.href = '/labs'}
        />
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <div className="p-6 max-w-6xl mx-auto space-y-6">
        <header className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div className="space-y-1">
            <h1 className="text-2xl md:text-3xl font-semibold">{lab.name}</h1>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              <User className="h-4 w-4" /> 
              {lab.owner ? `${lab.owner.name} · ${lab.owner.email}` : 'Unknown Owner'}
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Badge variant={getBadgeVariant(lab.status)} className="text-sm py-1 px-2">
              <Server className="h-4 w-4 mr-1"/>{lab.status.toUpperCase()}
            </Badge>
            <div className="flex items-center gap-2 text-sm">
              <Clock className="h-4 w-4"/>
              <span>
                Ends in {String(isFinite(overallCountdown.mins) ? overallCountdown.mins : 0).padStart(2, "0")}:{String(isFinite(overallCountdown.secs) ? overallCountdown.secs : 0).padStart(2, "0")}
              </span>
            </div>
            <div className="flex items-center gap-2 ml-4">
              <Button
                variant="destructive"
                size="sm"
                onClick={handleStopLab}
                disabled={stopping}
                className="gap-2"
              >
                {stopping ? (
                  <>
                    <Clock className="h-4 w-4 animate-spin" />
                    Stopping...
                  </>
                ) : (
                  <>
                    <Server className="h-4 w-4" />
                    Stop Lab
                  </>
                )}
              </Button>
            </div>
          </div>
        </header>

        <motion.div initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.25 }}>
          <Progress value={progressValue} className="h-2" />
        </motion.div>

        <section className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {lab.credentials.map((c) => (
            <Card key={c.id} className="shadow-sm rounded-2xl">
              <CardHeader className="flex flex-row items-center justify-between space-y-0">
                <CardTitle className="text-lg flex items-center gap-2">
                  <ShieldCheck className="h-5 w-5" /> {c.label} Access
                </CardTitle>
                {c.url && (
                  <Button asChild variant="link" className="gap-1">
                    <a href={c.url.startsWith('http://') || c.url.startsWith('https://') ? c.url : `https://${c.url}`} target="_blank" rel="noreferrer">
                      Open <ExternalLink className="h-4 w-4" />
                    </a>
                  </Button>
                )}
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 gap-3">
                  <div className="grid grid-cols-3 items-center gap-2">
                    <Label className="col-span-1">Username</Label>
                    <Input className="col-span-2 font-mono" value={c.username} readOnly />
                  </div>
                  <div className="grid grid-cols-3 items-center gap-2">
                    <Label className="col-span-1">Password</Label>
                    <div className="col-span-2">
                      <MaskedSecret secret={c.password} />
                    </div>
                  </div>
                </div>
                <div className="text-xs text-muted-foreground flex items-center gap-2">
                  <Info className="h-4 w-4" /> Expires at {new Date(c.expiresAt).toLocaleString()}
                </div>
                <div className="text-sm text-muted-foreground">
                  {c.notes}
                </div>
              </CardContent>
            </Card>
          ))}
        </section>

        <Separator />

        <section className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <Card className="rounded-2xl">
            <CardHeader>
              <CardTitle className="text-base">Helpful commands</CardTitle>
            </CardHeader>
            <CardContent>
              <pre className="bg-muted p-3 rounded-xl text-xs overflow-auto"># Example: ssh into Proxmox jump host
ssh {lab.credentials.find((c) => c.id === "proxmox")?.username}@jump.lab.local

# Example: port-forward VerteX UI if on VPN
kubectl port-forward svc/vertex-ui 8443:443 -n vertex</pre>
            </CardContent>
          </Card>
          <Card className="rounded-2xl">
            <CardHeader>
              <CardTitle className="text-base">Support</CardTitle>
            </CardHeader>
            <CardContent className="text-sm space-y-2">
              <p>Having trouble? Capture the error and share with the proctor.</p>
              <ul className="list-disc ml-5 space-y-1 text-muted-foreground">
                <li>Check your VPN and DNS resolution.</li>
                <li>Try a private window for UI logins.</li>
                <li>Contact support if credentials don&apos;t work.</li>
              </ul>
            </CardContent>
          </Card>
          <Card className="rounded-2xl">
            <CardHeader>
              <CardTitle className="text-base">Session details</CardTitle>
            </CardHeader>
            <CardContent className="text-sm space-y-1 text-muted-foreground">
              <div>Lab ID: <span className="font-mono text-foreground">{lab.id}</span></div>
              <div>Started: {lab.startedAt ? new Date(lab.startedAt).toLocaleString() : "—"}</div>
              <div>Ends: {lab.endsAt ? new Date(lab.endsAt).toLocaleString() : "—"}</div>
            </CardContent>
          </Card>
        </section>
      </div>
    </AppLayout>
  );
}

export default function LabSessionPage() {
  return (
    <ProtectedRoute>
      <LabSessionPageContent />
    </ProtectedRoute>
  );
}
