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
import { Copy, Eye, EyeOff, Clock, ExternalLink, ShieldCheck, Info, Server, User, StopCircle } from "lucide-react";
import { motion } from "framer-motion";
import { apiService, LabResponse } from "@/lib/api";
import { LabStartingView } from "./LabStartingView";

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
    owner: labResponse.owner,
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

export function LabSessionContent({ labId }: { labId: string }) {
  const [lab, setLab] = useState<LabSession | null>(null);
  const [loading, setLoading] = useState(true);
  const [stopping, setStopping] = useState(false);

  const totalMs = lab?.startedAt && lab?.endsAt ? Math.max(0, new Date(lab.endsAt).getTime() - new Date(lab.startedAt).getTime()) : undefined;
  const overallCountdown = useCountdown(lab?.endsAt, totalMs);
  const progressValue = Math.min(100, Math.max(0, 100 - overallCountdown.pct));

  useEffect(() => {
    let live = true;
    (async () => {
      setLoading(true);
      const data = await fetchLabSession(labId);
      if (!live) return;
      setLab(data);
      setLoading(false);
    })();
    const poll = setInterval(async () => {
      const data = await fetchLabSession(labId);
      if (!live) return;
      setLab(data);
    }, 15000);
    return () => { live = false; clearInterval(poll); };
  }, [labId]);

  const handleLabReady = () => {
    // Refresh the lab data when it becomes ready
    fetchLabSession(labId).then(setLab);
  };

  const handleStopLab = async () => {
    if (!lab || stopping) return;
    
    setStopping(true);
    try {
      await apiService.stopLab(lab.id);
      // Refresh lab data to show updated status
      await fetchLabSession(labId).then(setLab);
    } catch (error) {
      console.error("Failed to stop lab:", error);
    } finally {
      setStopping(false);
    }
  };

  if (loading || !lab) {
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
    return <LabStartingView labId={labId} onLabReady={handleLabReady} />;
  }

  const getBadgeVariant = (status: LabSession['status']) => {
    switch (status) {
      case 'ready':
        return 'default';
      case 'provisioning':
      case 'starting':
        return 'secondary';
      case 'error':
      case 'expired':
        return 'destructive';
      default:
        return 'secondary';
    }
  };

  return (
    <div className="p-6 max-w-6xl mx-auto space-y-6">
      <header className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div className="space-y-1">
          <h1 className="text-2xl md:text-3xl font-semibold">{lab.name}</h1>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <User className="h-4 w-4" /> {lab.owner ? `${lab.owner.name} · ${lab.owner.email}` : 'Unknown Owner'}
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
          {lab.status === "ready" && (
            <Button
              variant="destructive"
              size="sm"
              onClick={handleStopLab}
              disabled={stopping}
              className="gap-2"
            >
              <StopCircle className="h-4 w-4" />
              {stopping ? "Stopping..." : "Stop Lab"}
            </Button>
          )}
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
  );
}
