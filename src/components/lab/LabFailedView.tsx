import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Separator } from '@/components/ui/separator';
import { AlertTriangle, RefreshCw, Trash2, Clock, User } from 'lucide-react';
import { apiService } from '@/lib/api';
import { LabSession } from '@/app/lab/page';

interface LabFailedViewProps {
  labId: string;
  lab: LabSession;
  onRetry?: () => void;
  onCleanup?: () => void;
}

export function LabFailedView({ labId, lab, onCleanup }: LabFailedViewProps) {
  const [isCleaningUp, setIsCleaningUp] = useState(false);
  const [isRetrying, setIsRetrying] = useState(false);

  const handleCleanup = async () => {
    if (!lab || isCleaningUp) return;
    
    setIsCleaningUp(true);
    try {
      await apiService.cleanupFailedLab(labId);
      onCleanup?.();
    } catch (error) {
      console.error('Failed to cleanup lab:', error);
    } finally {
      setIsCleaningUp(false);
    }
  };

  const handleRetry = async () => {
    if (!lab || isRetrying) return;
    
    setIsRetrying(true);
    try {
      // For now, we'll just refresh the page to retry
      // In a real implementation, you might want to call an API to restart the lab
      window.location.reload();
    } catch (error) {
      console.error('Failed to retry lab:', error);
    } finally {
      setIsRetrying(false);
    }
  };

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-6">
      <header className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div className="space-y-1">
          <h1 className="text-2xl md:text-3xl font-semibold text-destructive">{lab.name}</h1>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <User className="h-4 w-4" /> {lab.owner.name} · {lab.owner.email}
          </div>
        </div>
        <div className="flex items-center gap-3">
          <Badge variant="destructive" className="text-sm py-1 px-2">
            <AlertTriangle className="h-4 w-4 mr-1"/>FAILED
          </Badge>
          <div className="flex items-center gap-2 text-sm">
            <Clock className="h-4 w-4"/>
            <span>Failed at {new Date(lab.startedAt || Date.now()).toLocaleString()}</span>
          </div>
        </div>
      </header>

      <Alert variant="destructive">
        <AlertTriangle className="h-4 w-4" />
        <AlertDescription>
          This lab failed during provisioning. The lab environment may be in an inconsistent state and should be cleaned up.
        </AlertDescription>
      </Alert>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card className="border-destructive/20">
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-destructive" />
              Lab Status
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <div className="flex justify-between">
                <span className="text-sm font-medium">Status:</span>
                <Badge variant="destructive">Failed</Badge>
              </div>
              <div className="flex justify-between">
                <span className="text-sm font-medium">Started:</span>
                <span className="text-sm text-muted-foreground">
                  {new Date(lab.startedAt || Date.now()).toLocaleString()}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-sm font-medium">Status:</span>
                <span className="text-sm text-muted-foreground">
                  Failed
                </span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card className="border-destructive/20">
          <CardHeader>
            <CardTitle className="text-lg flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-destructive" />
              Actions
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-3">
              <Button
                onClick={handleRetry}
                disabled={isRetrying}
                className="w-full gap-2"
                variant="outline"
              >
                <RefreshCw className={`h-4 w-4 ${isRetrying ? 'animate-spin' : ''}`} />
                {isRetrying ? 'Retrying...' : 'Retry Lab Creation'}
              </Button>
              
              <Button
                onClick={handleCleanup}
                disabled={isCleaningUp}
                className="w-full gap-2"
                variant="destructive"
              >
                <Trash2 className={`h-4 w-4 ${isCleaningUp ? 'animate-spin' : ''}`} />
                {isCleaningUp ? 'Cleaning Up...' : 'Clean Up Failed Lab'}
              </Button>
            </div>
            
            <div className="text-xs text-muted-foreground space-y-1">
              <p><strong>Retry:</strong> Attempt to recreate the lab environment</p>
              <p><strong>Clean Up:</strong> Remove all resources and start fresh</p>
            </div>
          </CardContent>
        </Card>
      </div>

      <Separator />

      <Card>
        <CardHeader>
          <CardTitle className="text-lg">What happened?</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="text-sm text-muted-foreground space-y-2">
            <p>
              The lab failed during the provisioning process. This could be due to:
            </p>
            <ul className="list-disc list-inside space-y-1 ml-4">
              <li>Network connectivity issues</li>
              <li>Resource allocation failures</li>
              <li>Service configuration errors</li>
              <li>Temporary infrastructure problems</li>
            </ul>
            <p>
              We recommend cleaning up the failed lab to ensure no orphaned resources remain, 
              then retrying the lab creation.
            </p>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
