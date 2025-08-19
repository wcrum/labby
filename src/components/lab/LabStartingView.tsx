"use client";

import React, { useState, useEffect, useCallback } from "react";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Loader2, RefreshCw, CheckCircle, XCircle, Clock, Server } from "lucide-react";
import { apiService } from "@/lib/api";

interface LabStartingViewProps {
  labId: string;
  onLabReady: () => void;
}

interface StartupStep {
  id: string;
  name: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  message: string;
  timestamp: string;
}

export function LabStartingView({ labId, onLabReady }: LabStartingViewProps) {
  const [startupSteps, setStartupSteps] = useState<StartupStep[]>([
    {
      id: "1",
      name: "Initializing Infrastructure",
      status: "running",
      message: "Setting up virtual machines and networking...",
      timestamp: new Date().toISOString(),
    },
    {
      id: "2",
      name: "Deploying Spectro Cloud Platform",
      status: "pending",
      message: "Waiting to start...",
      timestamp: "",
    },
    {
      id: "3",
      name: "Configuring Kubernetes Cluster",
      status: "pending",
      message: "Waiting to start...",
      timestamp: "",
    },
    {
      id: "4",
      name: "Setting up OpenShift Environment",
      status: "pending",
      message: "Waiting to start...",
      timestamp: "",
    },
    {
      id: "5",
      name: "Generating Access Credentials",
      status: "pending",
      message: "Waiting to start...",
      timestamp: "",
    },
    {
      id: "6",
      name: "Final Configuration",
      status: "pending",
      message: "Waiting to start...",
      timestamp: "",
    },
  ]);
  const [overallProgress, setOverallProgress] = useState(15);
  const [logs, setLogs] = useState<string[]>([
    "[INFO] Starting lab initialization...",
    "[INFO] Checking system requirements...",
    "[INFO] Allocating resources for lab environment...",
    "[INFO] Creating virtual network infrastructure...",
    "[INFO] Deploying base virtual machines...",
  ]);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const fetchLabStatus = useCallback(async () => {
    try {
      const labData = await apiService.getLab(labId);
      
      // If lab is ready, notify parent
      if (labData.status === 'ready') {
        onLabReady();
        return;
      }
    } catch (err) {
      console.error("Failed to fetch lab status:", err);
    }
  }, [labId, onLabReady]);

  const refreshStatus = async () => {
    setIsRefreshing(true);
    await fetchLabStatus();
    setIsRefreshing(false);
  };

  // Simulate startup progress
  useEffect(() => {
    const interval = setInterval(() => {
      setOverallProgress(prev => {
        if (prev >= 100) return 100;
        return prev + Math.random() * 5;
      });
    }, 2000);

    return () => clearInterval(interval);
  }, []);

  // Simulate step progression
  useEffect(() => {
    const stepInterval = setInterval(() => {
      setStartupSteps(prev => {
        const newSteps = [...prev];
        const runningIndex = newSteps.findIndex(step => step.status === 'running');
        
        if (runningIndex !== -1) {
          // Complete current step
          newSteps[runningIndex] = {
            ...newSteps[runningIndex],
            status: 'completed',
            message: 'Step completed successfully',
            timestamp: new Date().toISOString(),
          };
          
          // Start next step if available
          const nextIndex = runningIndex + 1;
          if (nextIndex < newSteps.length) {
            newSteps[nextIndex] = {
              ...newSteps[nextIndex],
              status: 'running',
              message: getStepMessage(newSteps[nextIndex].name),
              timestamp: new Date().toISOString(),
            };
          }
        }
        
        return newSteps;
      });
    }, 3000);

    return () => clearInterval(stepInterval);
  }, []);

  // Simulate log updates
  useEffect(() => {
    const logInterval = setInterval(() => {
      setLogs(prev => {
        const newLogs = [...prev];
        const logMessages = [
          "[INFO] Configuring network interfaces...",
          "[INFO] Installing required packages...",
          "[INFO] Setting up firewall rules...",
          "[INFO] Configuring storage volumes...",
          "[INFO] Starting container runtime...",
          "[INFO] Deploying Spectro Cloud components...",
          "[INFO] Initializing Kubernetes control plane...",
          "[INFO] Joining worker nodes to cluster...",
          "[INFO] Installing OpenShift operators...",
          "[INFO] Configuring authentication...",
          "[INFO] Generating SSL certificates...",
          "[INFO] Setting up monitoring...",
          "[INFO] Creating user accounts...",
          "[INFO] Finalizing configuration...",
        ];
        
        const randomMessage = logMessages[Math.floor(Math.random() * logMessages.length)];
        newLogs.push(`[${new Date().toLocaleTimeString()}] ${randomMessage}`);
        
        // Keep only last 20 logs
        return newLogs.slice(-20);
      });
    }, 1500);

    return () => clearInterval(logInterval);
  }, []);

  // Check lab status periodically
  useEffect(() => {
    const statusInterval = setInterval(fetchLabStatus, 5000);
    return () => clearInterval(statusInterval);
  }, [labId, fetchLabStatus]);

  const getStepMessage = (stepName: string): string => {
    const messages: Record<string, string> = {
      "Initializing Infrastructure": "Setting up virtual machines and networking...",
      "Deploying Spectro Cloud Platform": "Installing Spectro Cloud components...",
      "Configuring Kubernetes Cluster": "Setting up Kubernetes control plane...",
      "Setting up OpenShift Environment": "Deploying OpenShift operators...",
      "Generating Access Credentials": "Creating user accounts and credentials...",
      "Final Configuration": "Applying final settings and configurations...",
    };
    return messages[stepName] || "Processing...";
  };

  const getStepIcon = (status: StartupStep['status']) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-green-500" />;
      case 'running':
        return <Loader2 className="h-4 w-4 text-blue-500 animate-spin" />;
      case 'failed':
        return <XCircle className="h-4 w-4 text-red-500" />;
      default:
        return <Clock className="h-4 w-4 text-gray-400" />;
    }
  };

  const completedSteps = startupSteps.filter(step => step.status === 'completed').length;
  const totalSteps = startupSteps.length;

  return (
    <div className="min-h-screen bg-background p-4">
      <div className="max-w-4xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Lab Starting</h1>
            <p className="text-muted-foreground">
              Your lab environment is being prepared. This may take a few minutes.
            </p>
          </div>
          <Button onClick={refreshStatus} disabled={isRefreshing} variant="outline">
            <RefreshCw className={`mr-2 h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`} />
            Refresh
          </Button>
        </div>

        {/* Progress Overview */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Server className="h-5 w-5" />
              Overall Progress
            </CardTitle>
            <CardDescription>
              Step {completedSteps} of {totalSteps} completed
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <Progress value={overallProgress} className="h-3" />
              <div className="flex justify-between text-sm">
                <span>Initializing...</span>
                <span>{Math.round(overallProgress)}%</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Startup Steps */}
          <Card>
            <CardHeader>
              <CardTitle>Startup Steps</CardTitle>
              <CardDescription>
                Current progress of lab initialization
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {startupSteps.map((step) => (
                  <div key={step.id} className="flex items-start gap-3 p-3 rounded-lg border">
                    {getStepIcon(step.status)}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center justify-between">
                        <h4 className="font-medium text-sm">{step.name}</h4>
                        <Badge 
                          variant={
                            step.status === 'completed' ? 'default' :
                            step.status === 'running' ? 'secondary' :
                            step.status === 'failed' ? 'destructive' : 'outline'
                          }
                          className="text-xs"
                        >
                          {step.status}
                        </Badge>
                      </div>
                      <p className="text-sm text-muted-foreground mt-1">
                        {step.message}
                      </p>
                      {step.timestamp && (
                        <p className="text-xs text-muted-foreground mt-1">
                          {new Date(step.timestamp).toLocaleTimeString()}
                        </p>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          {/* Live Logs */}
          <Card>
            <CardHeader>
              <CardTitle>Live Logs</CardTitle>
              <CardDescription>
                Real-time startup logs and system messages
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="bg-black text-green-400 p-4 rounded-lg font-mono text-sm h-96 overflow-y-auto">
                {logs.map((log, index) => (
                  <div key={index} className="mb-1">
                    {log}
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Status Alert */}
        <Alert>
          <Clock className="h-4 w-4" />
          <AlertDescription>
            Your lab is currently being provisioned. You&apos;ll be automatically redirected once it&apos;s ready.
            This process typically takes 3-5 minutes depending on system load.
          </AlertDescription>
        </Alert>
      </div>
    </div>
  );
}
