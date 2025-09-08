// Shared lab header component

import React from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Clock, Server, User, StopCircle } from "lucide-react";
import { LabSession, getLabBadgeVariant } from "@/types/lab";
import { CountdownResult } from "@/hooks/useCountdown";

interface LabHeaderProps {
  lab: LabSession;
  countdown: CountdownResult;
  onStopLab?: () => void;
  stopping?: boolean;
  showStopButton?: boolean;
}

export function LabHeader({ 
  lab, 
  countdown, 
  onStopLab, 
  stopping = false, 
  showStopButton = true 
}: LabHeaderProps) {
  return (
    <header className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
      <div className="space-y-1">
        <h1 className="text-2xl md:text-3xl font-semibold">{lab.name}</h1>
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <User className="h-4 w-4" /> 
          {lab.owner ? `${lab.owner.name} · ${lab.owner.email}` : 'Unknown Owner'}
        </div>
      </div>
      <div className="flex items-center gap-3">
        <Badge variant={getLabBadgeVariant(lab.status)} className="text-sm py-1 px-2">
          <Server className="h-4 w-4 mr-1"/>{lab.status.toUpperCase()}
        </Badge>
        <div className="flex items-center gap-2 text-sm">
          <Clock className="h-4 w-4"/>
          <span>
            Ends in {String(isFinite(countdown.mins) ? countdown.mins : 0).padStart(2, "0")}:{String(isFinite(countdown.secs) ? countdown.secs : 0).padStart(2, "0")}
          </span>
        </div>
        {showStopButton && lab.status === "ready" && onStopLab && (
          <Button
            variant="destructive"
            size="sm"
            onClick={onStopLab}
            disabled={stopping}
            className="gap-2"
          >
            <StopCircle className="h-4 w-4" />
            {stopping ? "Stopping..." : "Stop Lab"}
          </Button>
        )}
      </div>
    </header>
  );
}
