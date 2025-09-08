"use client";

import React, { useEffect, useState } from "react";
import { Separator } from "@/components/ui/separator";
import { Progress } from "@/components/ui/progress";
import { motion } from "framer-motion";
import { apiService } from "@/lib/api";
import { LabStartingView } from "./LabStartingView";
import { LabSession } from "@/types/lab";
import { convertLabResponse } from "@/lib/lab-utils";
import { useCountdown } from "@/hooks/useCountdown";
import { LabHeader } from "./shared/LabHeader";
import { LabCredentials } from "./shared/LabCredentials";
import { LabInfoCards } from "./shared/LabInfoCards";

async function fetchLabSession(labId: string): Promise<LabSession> {
  const labResponse = await apiService.getLab(labId);
  return convertLabResponse(labResponse);
}


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


  return (
    <div className="p-6 max-w-6xl mx-auto space-y-6">
      <LabHeader 
        lab={lab} 
        countdown={overallCountdown} 
        onStopLab={handleStopLab}
        stopping={stopping}
      />

      <motion.div initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.25 }}>
        <Progress value={progressValue} className="h-2" />
      </motion.div>

      <LabCredentials credentials={lab.credentials} />

      <Separator />

      <LabInfoCards lab={lab} />
    </div>
  );
}
