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
  const [services, setServices] = useState<Array<{
    name: string;
    description: string;
    status: string;
    progress: number;
    steps: StartupStep[];
    started_at?: string;
    completed_at?: string;
    error?: string;
  }>>([]);
  const [overallProgress, setOverallProgress] = useState(0);
  const [logs, setLogs] = useState<string[]>([]);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [isCompleting, setIsCompleting] = useState(false);


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

  const fetchProgress = useCallback(async () => {
    try {
      const progressData = await apiService.getLabProgress(labId);
      
      if (progressData) {
        setOverallProgress(progressData.overall || 0);
        setLogs(progressData.logs || []);
        
        // Convert services data
        const servicesData = progressData.services?.map((service) => ({
          name: service.name,
          description: service.description,
          status: service.status,
          progress: service.progress,
          steps: service.steps.map((step, index) => ({
            id: String(index + 1),
            name: step.name,
            status: step.status as "pending" | "running" | "completed" | "failed",
            message: step.message || "",
            timestamp: step.started_at || "",
          })),
          started_at: service.started_at,
          completed_at: service.completed_at,
          error: service.error,
        })) || [];
        
        setServices(servicesData);
        
        // Check if all services are completed or overall progress is 100%
        const allServicesCompleted = servicesData.length > 0 && servicesData.every(service => service.status === 'completed');
        const progressComplete = (progressData.overall || 0) >= 100;
        
        if (allServicesCompleted || progressComplete) {
          if (!isCompleting) {
            setIsCompleting(true);
            console.log("Lab appears complete, verifying status...");
          }
          
          // Double-check lab status before redirecting
          try {
            const labData = await apiService.getLab(labId);
            if (labData.status === 'ready') {
              console.log("Lab confirmed ready, redirecting...");
              onLabReady();
              return;
            }
          } catch (statusErr) {
            console.error("Failed to verify lab status:", statusErr);
          }
        }
      }
    } catch (err) {
      console.error("Failed to fetch progress:", err);
    }
  }, [labId, onLabReady]);

  const refreshStatus = async () => {
    setIsRefreshing(true);
    await fetchLabStatus();
    setIsRefreshing(false);
  };







  // Check lab status and progress periodically
  useEffect(() => {
    const statusInterval = setInterval(fetchLabStatus, 5000);
    const progressInterval = setInterval(fetchProgress, 2000);
    
    // Fetch initial progress
    fetchProgress();
    
    return () => {
      clearInterval(statusInterval);
      clearInterval(progressInterval);
    };
  }, [labId, fetchLabStatus, fetchProgress, isCompleting]);



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

  const completedServices = services.filter(service => service.status === 'completed').length;
  const totalServices = services.length;

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-6">
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
              Service {completedServices} of {totalServices} completed
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <Progress value={overallProgress} className="h-3" />
              <div className="flex justify-between text-sm">
                <span>{isCompleting ? "Completing..." : overallProgress >= 100 ? "Finalizing..." : "Initializing..."}</span>
                <span>{Math.round(overallProgress)}%</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Services */}
          <Card>
            <CardHeader>
              <CardTitle>Services</CardTitle>
              <CardDescription>
                Individual service progress and status
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {services.map((service) => (
                  <div key={service.name} className="border rounded-lg p-4">
                    <div className="flex items-center justify-between mb-3">
                      <div>
                        <h4 className="font-medium text-sm">{service.name}</h4>
                        <p className="text-xs text-muted-foreground">{service.description}</p>
                      </div>
                      <Badge 
                        variant={
                          service.status === 'completed' ? 'default' :
                          service.status === 'running' ? 'secondary' :
                          service.status === 'failed' ? 'destructive' : 'outline'
                        }
                        className="text-xs"
                      >
                        {service.status}
                      </Badge>
                    </div>
                    
                    <Progress value={service.progress} className="h-2 mb-3" />
                    
                    <div className="space-y-2">
                      {service.steps.map((step) => (
                        <div key={step.id} className="flex items-start gap-2 text-xs">
                          {getStepIcon(step.status)}
                          <div className="flex-1">
                            <div className="flex items-center justify-between">
                              <span className="font-medium">{step.name}</span>
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
                            {step.message && (
                              <p className="text-muted-foreground mt-1">{step.message}</p>
                            )}
                          </div>
                        </div>
                      ))}
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
            {isCompleting ? (
              <>
                🎉 Lab setup complete! Redirecting to your lab environment...
              </>
            ) : (
              <>
                Your lab is currently being provisioned. You&apos;ll be automatically redirected once it&apos;s ready.
                This process typically takes 3-5 minutes depending on system load.
              </>
            )}
          </AlertDescription>
        </Alert>
      </div>
  );
}
