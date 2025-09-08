"use client";

import React, { useState } from "react";
import { Separator } from "@/components/ui/separator";
import { Progress } from "@/components/ui/progress";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { motion } from "framer-motion";
import { apiService } from "@/lib/api";
import { LabStartingView } from "./LabStartingView";
import { LabSession } from "@/types/lab";
import { useCountdown } from "@/hooks/useCountdown";
import { useLabData } from "@/hooks/useLabData";
import { LabHeader } from "./shared/LabHeader";
import { LabCredentials } from "./shared/LabCredentials";
import { LabInfoCards } from "./shared/LabInfoCards";
import { LabPageSkeleton } from "@/components/ui/loading-skeleton";

export function LabSessionContent({ labId }: { labId: string }) {
  const [stopping, setStopping] = useState(false);
  const [showStopDialog, setShowStopDialog] = useState(false);

  // Use the existing useLabData hook instead of manual fetching
  const { lab, loading, error, refetch } = useLabData(labId, { 
    pollInterval: 15000,
    autoFetch: true 
  });

  const totalMs = lab?.startedAt && lab?.endsAt ? Math.max(0, new Date(lab.endsAt).getTime() - new Date(lab.startedAt).getTime()) : undefined;
  const overallCountdown = useCountdown(lab?.endsAt, totalMs);
  const progressValue = Math.min(100, Math.max(0, 100 - overallCountdown.pct));

  const handleLabReady = () => {
    // Refresh the lab data when it becomes ready
    refetch();
  };

  const handleStopLab = async () => {
    if (!lab || stopping) return;
    
    setStopping(true);
    try {
      await apiService.stopLab(lab.id);
      // Refresh lab data to show updated status
      await refetch();
      setShowStopDialog(false);
    } catch (error) {
      console.error("Failed to stop lab:", error);
    } finally {
      setStopping(false);
    }
  };

  if (loading || !lab) {
    return <LabPageSkeleton />;
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
        onStopLab={() => setShowStopDialog(true)}
        stopping={stopping}
        showStopButton={true}
      />

      <motion.div initial={{ opacity: 0, y: 8 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.25 }}>
        <Progress value={progressValue} className="h-2" />
      </motion.div>

      <LabCredentials credentials={lab.credentials} />

      <Separator />

      <LabInfoCards lab={lab} />

      {/* Stop Lab Confirmation Dialog */}
      <AlertDialog open={showStopDialog} onOpenChange={setShowStopDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Stop Lab</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to stop this lab? This will stop the lab, cleanup all resources, and cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction 
              onClick={handleStopLab}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Stop Lab
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}