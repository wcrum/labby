"use client";

import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Loader2, Play, Clock, Shield } from "lucide-react";
import { apiService } from "@/lib/api";
import { useAuth } from "@/lib/auth";

interface LabStarterProps {
  onLabCreated: () => void;
}

export function LabStarter({ onLabCreated }: LabStarterProps) {
  const { user } = useAuth();
  const [duration, setDuration] = useState("60");
  const [isCreating, setIsCreating] = useState(false);
  const [error, setError] = useState("");

  const durationOptions = [
    { value: "30", label: "30 minutes" },
    { value: "60", label: "1 hour" },
    { value: "120", label: "2 hours" },
    { value: "240", label: "4 hours" },
    { value: "480", label: "8 hours" },
  ];

  const handleCreateLab = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setIsCreating(true);

    try {
      if (!user) {
        throw new Error("User not authenticated");
      }
      await apiService.createLab({
        owner_id: user.id,
        duration: parseInt(duration),
      });
      onLabCreated();
    } catch (err) {
      setError("Failed to create lab. Please try again.");
      console.error("Lab creation failed:", err);
    } finally {
      setIsCreating(false);
    }
  };

  return (
    <div className="flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center space-y-2">
          <div className="mx-auto w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center">
            <Shield className="w-6 h-6 text-blue-600" />
          </div>
          <CardTitle className="text-2xl font-bold">Start Your Lab Session</CardTitle>
          <CardDescription>
            Create a new lab environment to begin your hands-on experience
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleCreateLab} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="duration">Session Duration</Label>
              <Select value={duration} onValueChange={setDuration} disabled={isCreating}>
                <SelectTrigger>
                  <SelectValue placeholder="Select duration" />
                </SelectTrigger>
                <SelectContent>
                  {durationOptions.map((option) => (
                    <SelectItem key={option.value} value={option.value}>
                      <div className="flex items-center gap-2">
                        <Clock className="h-4 w-4" />
                        {option.label}
                      </div>
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            <Button 
              type="submit" 
              className="w-full" 
              disabled={isCreating}
            >
              {isCreating ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Creating Lab...
                </>
              ) : (
                <>
                  <Play className="mr-2 h-4 w-4" />
                  Start Lab Session
                </>
              )}
            </Button>
          </form>

          <div className="mt-6 p-4 bg-muted rounded-lg">
            <h4 className="font-medium mb-2">What you&apos;ll get:</h4>
            <ul className="text-sm text-muted-foreground space-y-1">
              <li>• Access to VerteX (Spectro Cloud control plane)</li>
              <li>• Proxmox VE UI for cluster management</li>
              <li>• Kubernetes cluster access</li>
              <li>• OpenShift console access</li>
              <li>• Secure credentials with rotation capability</li>
            </ul>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
