"use client";

import React, { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

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
import { AlertCircle } from "lucide-react";
import { motion } from "framer-motion";
import { ProtectedRoute } from "@/components/auth/ProtectedRoute";
import { apiService } from "@/lib/api";
import { LabStartingView } from "@/components/lab/LabStartingView";
import { LabFailedView } from "@/components/lab/LabFailedView";
import { AppLayout } from "@/components/layout/AppLayout";
import { LabSession } from "@/types/lab";
import { convertLabResponse } from "@/lib/lab-utils";
import { useCountdown } from "@/hooks/useCountdown";
import { LabHeader } from "@/components/lab/shared/LabHeader";
import { LabCredentials } from "@/components/lab/shared/LabCredentials";
import { LabInfoCards } from "@/components/lab/shared/LabInfoCards";

async function fetchLabSession(labId: string): Promise<LabSession> {
  const labResponse = await apiService.getLab(labId);
  return convertLabResponse(labResponse);
}


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
  const [showStopDialog, setShowStopDialog] = useState(false);


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
    if (!lab) return;

    setStopping(true);
    try {
      // Stop the lab and cleanup all resources
      await apiService.stopLab(lab.id);
      
      // Redirect to labs page after successful stop
      window.location.href = '/labs';
    } catch (error) {
      console.error('Failed to stop lab:', error);
      setError('Failed to stop lab. Please try again.');
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
      </div>

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
