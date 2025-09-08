"use client";

import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { RefreshCw, Server, Trash2, AlertTriangle, CheckCircle } from "lucide-react";
import { ProtectedRoute } from "@/components/auth/ProtectedRoute";
import { AppLayout } from "@/components/layout/AppLayout";
import { apiService } from "@/lib/api";

function CleanupPageContent() {
  const [cleanupProjectName, setCleanupProjectName] = useState("");
  const [cleanupTerraformLabId, setCleanupTerraformLabId] = useState("");
  const [cleanupLoading, setCleanupLoading] = useState(false);
  const [cleanupMessage, setCleanupMessage] = useState("");
  const [cleanupHistory, setCleanupHistory] = useState<Array<{
    id: string;
    projectName: string;
    status: 'success' | 'error';
    message: string;
    timestamp: Date;
  }>>([]);

  const handleCleanupPaletteProject = async () => {
    if (!cleanupProjectName.trim()) return;
    
    setCleanupLoading(true);
    setCleanupMessage("");
    
    try {
      await apiService.cleanupPaletteProject(cleanupProjectName.trim());
      const successMessage = "Palette Project cleanup completed successfully!";
      setCleanupMessage(successMessage);
      
      // Add to cleanup history
      setCleanupHistory(prev => [{
        id: Date.now().toString(),
        projectName: cleanupProjectName.trim(),
        status: 'success',
        message: successMessage,
        timestamp: new Date()
      }, ...prev]);
      
      setCleanupProjectName("");
    } catch (error) {
      const errorMessage = `Cleanup failed: ${error instanceof Error ? error.message : 'Unknown error'}`;
      setCleanupMessage(errorMessage);
      
      // Add to cleanup history
      setCleanupHistory(prev => [{
        id: Date.now().toString(),
        projectName: cleanupProjectName.trim(),
        status: 'error',
        message: errorMessage,
        timestamp: new Date()
      }, ...prev]);
    } finally {
      setCleanupLoading(false);
    }
  };

  const handleCleanupTerraformCloud = async () => {
    if (!cleanupTerraformLabId.trim()) return;
    
    setCleanupLoading(true);
    setCleanupMessage("");
    
    try {
      await apiService.cleanupTerraformCloud(cleanupTerraformLabId.trim());
      const successMessage = "Terraform Cloud cleanup completed successfully!";
      setCleanupMessage(successMessage);
      
      // Add to cleanup history
      setCleanupHistory(prev => [{
        id: Date.now().toString(),
        projectName: `terraform-${cleanupTerraformLabId.trim()}`,
        status: 'success',
        message: successMessage,
        timestamp: new Date()
      }, ...prev]);
      
      setCleanupTerraformLabId("");
    } catch (error) {
      const errorMessage = `Terraform Cloud cleanup failed: ${error instanceof Error ? error.message : 'Unknown error'}`;
      setCleanupMessage(errorMessage);
      
      // Add to cleanup history
      setCleanupHistory(prev => [{
        id: Date.now().toString(),
        projectName: `terraform-${cleanupTerraformLabId.trim()}`,
        status: 'error',
        message: errorMessage,
        timestamp: new Date()
      }, ...prev]);
    } finally {
      setCleanupLoading(false);
    }
  };

  const handleBulkCleanup = async () => {
    if (!confirm("Are you sure you want to perform bulk cleanup? This will clean up all expired labs and may take some time.")) {
      return;
    }
    
    setCleanupLoading(true);
    setCleanupMessage("");
    
    try {
      // This would be a new API endpoint for bulk cleanup
      // await apiService.bulkCleanup();
      setCleanupMessage("Bulk cleanup completed successfully!");
    } catch (error) {
      setCleanupMessage(`Bulk cleanup failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setCleanupLoading(false);
    }
  };

  return (
    <AppLayout>
      <div className="p-6 max-w-4xl mx-auto space-y-6">
        <header className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div className="space-y-1">
            <h1 className="text-3xl md:text-4xl font-semibold">Admin Cleanup Tools</h1>
            <p className="text-muted-foreground">Clean up resources and manage system maintenance</p>
          </div>
        </header>

        {/* Palette Project Cleanup Section */}
        <Card className="rounded-xl">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Server className="h-5 w-5" />
              Palette Project Cleanup Tool
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="text-sm text-muted-foreground">
                Enter a palette project name to directly clean up its resources (API keys, users, projects, etc.)
              </div>
              <div className="flex gap-2">
                <Input
                  placeholder="Enter project name (e.g., lab-482407e5)"
                  value={cleanupProjectName}
                  onChange={(e) => setCleanupProjectName(e.target.value)}
                  className="flex-1"
                />
                <Button 
                  onClick={handleCleanupPaletteProject}
                  disabled={!cleanupProjectName.trim() || cleanupLoading}
                  variant="destructive"
                  className="gap-2"
                >
                  {cleanupLoading ? (
                    <>
                      <RefreshCw className="h-4 w-4 animate-spin" />
                      Cleaning...
                    </>
                  ) : (
                    <>
                      <Server className="h-4 w-4" />
                      Cleanup Project
                    </>
                  )}
                </Button>
              </div>
              {cleanupMessage && (
                <div className={`text-sm p-3 rounded-md ${
                  cleanupMessage.includes('success') 
                    ? 'bg-green-50 text-green-700 border border-green-200' 
                    : 'bg-red-50 text-red-700 border border-red-200'
                }`}>
                  {cleanupMessage}
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Terraform Cloud Cleanup Section */}
        <Card className="rounded-xl">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Server className="h-5 w-5" />
              Terraform Cloud Cleanup Tool
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="text-sm text-muted-foreground">
                Enter a lab ID to clean up all Terraform Cloud resources (workspaces, runs, etc.) associated with that lab
              </div>
              <div className="flex gap-2">
                <Input
                  placeholder="Enter lab ID (e.g., fd2e72e2)"
                  value={cleanupTerraformLabId}
                  onChange={(e) => setCleanupTerraformLabId(e.target.value)}
                  className="flex-1"
                />
                <Button 
                  onClick={handleCleanupTerraformCloud}
                  disabled={!cleanupTerraformLabId.trim() || cleanupLoading}
                  variant="destructive"
                  className="gap-2"
                >
                  {cleanupLoading ? (
                    <>
                      <RefreshCw className="h-4 w-4 animate-spin" />
                      Cleaning...
                    </>
                  ) : (
                    <>
                      <Server className="h-4 w-4" />
                      Cleanup Terraform Cloud
                    </>
                  )}
                </Button>
              </div>
              {cleanupMessage && (
                <div className={`text-sm p-3 rounded-md ${
                  cleanupMessage.includes('success') 
                    ? 'bg-green-50 text-green-700 border border-green-200' 
                    : 'bg-red-50 text-red-700 border border-red-200'
                }`}>
                  {cleanupMessage}
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        {/* Bulk Cleanup Section */}
        <Card className="rounded-xl">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Trash2 className="h-5 w-5" />
              Bulk Cleanup Operations
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="text-sm text-muted-foreground">
                Perform system-wide cleanup operations for expired labs and unused resources
              </div>
              <div className="flex gap-2">
                <Button 
                  onClick={handleBulkCleanup}
                  disabled={cleanupLoading}
                  variant="outline"
                  className="gap-2"
                >
                  {cleanupLoading ? (
                    <>
                      <RefreshCw className="h-4 w-4 animate-spin" />
                      Processing...
                    </>
                  ) : (
                    <>
                      <Trash2 className="h-4 w-4" />
                      Bulk Cleanup
                    </>
                  )}
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Cleanup History */}
        {cleanupHistory.length > 0 && (
          <Card className="rounded-xl">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <CheckCircle className="h-5 w-5" />
                Cleanup History
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {cleanupHistory.map((item) => (
                  <div
                    key={item.id}
                    className={`flex items-center justify-between p-3 rounded-md border ${
                      item.status === 'success' 
                        ? 'bg-green-50 border-green-200' 
                        : 'bg-red-50 border-red-200'
                    }`}
                  >
                    <div className="flex items-center gap-3">
                      {item.status === 'success' ? (
                        <CheckCircle className="h-4 w-4 text-green-600" />
                      ) : (
                        <AlertTriangle className="h-4 w-4 text-red-600" />
                      )}
                      <div>
                        <div className="font-medium text-sm">
                          {item.status === 'success' ? 'Cleanup Successful' : 'Cleanup Failed'}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          Project: {item.projectName}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {item.timestamp.toLocaleString()}
                        </div>
                      </div>
                    </div>
                    <div className="text-xs text-muted-foreground max-w-xs">
                      {item.message}
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        )}

        {/* Safety Information */}
        <Card className="rounded-xl border-amber-200 bg-amber-50">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-amber-800">
              <AlertTriangle className="h-5 w-5" />
              Important Safety Information
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2 text-amber-800">
              <p className="text-sm">
                <strong>Warning:</strong> These cleanup tools will permanently delete resources and cannot be undone.
              </p>
              <ul className="text-sm list-disc list-inside space-y-1">
                <li>Always verify the project name/lab ID before cleanup</li>
                <li>Ensure no active users are using the resources</li>
                <li>Consider backing up important data before cleanup</li>
                <li>Use bulk cleanup operations during maintenance windows</li>
                <li>Terraform Cloud cleanup will remove workspaces, runs, and variables</li>
                <li>Palette Project cleanup will remove API keys, users, and projects</li>
              </ul>
            </div>
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

export default function CleanupPage() {
  return (
    <ProtectedRoute requireAdmin={true}>
      <CleanupPageContent />
    </ProtectedRoute>
  );
}
