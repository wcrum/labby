"use client";

import React, { useEffect, useState } from "react";
import { apiService, LabResponse } from "@/lib/api";

import { LabSessionContent } from "./LabSessionContent";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Loader2, AlertCircle, RefreshCw, FlaskConical } from "lucide-react";

export function LabManager() {
  const [labs, setLabs] = useState<LabResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const fetchLabs = async () => {
    try {
      setError("");
      const userLabs = await apiService.getUserLabs();
      setLabs(userLabs || []); // Ensure labs is always an array
    } catch (err: any) {
      console.error("Failed to fetch labs:", err);
      
      // Check if it's an authentication error
      if (err.message && err.message.includes("Invalid token")) {
        setError("Authentication failed. Please log in again.");
        // Redirect to login or clear auth state
        apiService.clearToken();
        window.location.href = "/";
      } else {
        setError("Failed to load labs. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  };

  const refreshLabs = async () => {
    setLoading(true);
    await fetchLabs();
  };

  useEffect(() => {
    fetchLabs();
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="flex items-center space-x-2">
          <Loader2 className="h-6 w-6 animate-spin" />
          <span>Loading your lab sessions...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-red-500" />
              Error Loading Labs
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
            <Button onClick={refreshLabs} className="w-full">
              <RefreshCw className="mr-2 h-4 w-4" />
              Try Again
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Ensure labs is always an array before checking length
  const labsArray = Array.isArray(labs) ? labs : [];

  // If user has no labs, show message to create from template
  if (labsArray.length === 0) {
    return (
      <div className="flex items-center justify-center p-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FlaskConical className="h-5 w-5" />
              No Lab Sessions
            </CardTitle>
            <CardDescription>
              You don&apos;t have any active lab sessions. Start a new lab from available templates.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <Button asChild className="w-full">
              <a href="/labs">
                <FlaskConical className="mr-2 h-4 w-4" />
                Browse Available Labs
              </a>
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  // If user has exactly one lab, show it
  if (labsArray.length === 1) {
    const lab = labsArray[0];
    return <LabSessionContent labId={lab.id} />;
  }

  // If user has multiple labs (shouldn't happen with single lab policy)
  // Show a warning and let them choose
  return (
    <div className="flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <AlertCircle className="h-5 w-5 text-yellow-500" />
            Multiple Labs Found
          </CardTitle>
          <CardDescription>
            You have multiple lab sessions. Please select one to continue.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <Alert>
            <AlertDescription>
              You currently have {labsArray.length} lab sessions. Only one active lab is allowed per user.
            </AlertDescription>
          </Alert>
          
          <div className="space-y-2">
            {labsArray.map((lab) => (
              <Card key={lab.id} className="p-3">
                <div className="flex items-center justify-between">
                  <div>
                    <h4 className="font-medium">{lab.name}</h4>
                    <p className="text-sm text-muted-foreground">
                      Status: {lab.status} • Created: {new Date(lab.started_at).toLocaleDateString()}
                    </p>
                  </div>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => {
                      // Navigate to the selected lab
                      window.location.href = `/lab?id=${lab.id}`;
                    }}
                  >
                    Open
                  </Button>
                </div>
              </Card>
            ))}
          </div>

          <Button onClick={refreshLabs} variant="outline" className="w-full">
            <RefreshCw className="mr-2 h-4 w-4" />
            Refresh
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
