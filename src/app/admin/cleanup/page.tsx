"use client";

import React, { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
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
import { RefreshCw, Server, Trash2, AlertTriangle, CheckCircle, Info, Settings, Zap } from "lucide-react";
import { ProtectedRoute } from "@/components/auth/ProtectedRoute";
import { AppLayout } from "@/components/layout/AppLayout";
import { apiService } from "@/lib/api";

interface ServiceType {
  type: string;
  configs: Array<{
    id: string;
    name: string;
    description: string;
    is_active: boolean;
  }>;
  parameters: Array<{
    name: string;
    description: string;
    required: boolean;
    example?: string;
  }>;
}

interface CleanupHistoryItem {
  id: string;
  serviceType: string;
  status: 'success' | 'error';
  message: string;
  timestamp: Date;
  parameters?: Record<string, string>;
}

function CleanupPageContent() {
  const [availableServices, setAvailableServices] = useState<ServiceType[]>([]);
  const [cleanupMessage, setCleanupMessage] = useState("");
  const [cleanupHistory, setCleanupHistory] = useState<CleanupHistoryItem[]>([]);
  const [loadingServices, setLoadingServices] = useState(true);
  
  // Service cleanup state (simplified)
  const [serviceConfigId, setServiceConfigId] = useState<string>("");
  const [labId, setLabId] = useState("");
  const [cleanupLoading, setCleanupLoading] = useState(false);
  
  // Confirmation dialog state
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const [confirmMessage, setConfirmMessage] = useState("");

  // Load available services on component mount
  useEffect(() => {
    const loadServices = async () => {
      try {
        const response = await apiService.getAvailableServices();
        setAvailableServices(response.available_services);
      } catch (error) {
        console.error('Failed to load available services:', error);
        setCleanupMessage(`Failed to load available services: ${error instanceof Error ? error.message : 'Unknown error'}`);
      } finally {
        setLoadingServices(false);
      }
    };

    loadServices();
  }, []);

  // Get selected service config details
  const selectedServiceConfig = availableServices
    .flatMap(s => s.configs.map(config => ({ ...config, serviceType: s.type })))
    .find(config => config.id === serviceConfigId);

  // Handle service cleanup
  const handleServiceCleanup = async () => {
    if (!serviceConfigId) {
      setCleanupMessage("Please select a service configuration");
      return;
    }

    if (!labId.trim()) {
      setCleanupMessage("Please enter a lab UUID");
      return;
    }

    const message = `Are you sure you want to cleanup service config "${serviceConfigId}" for lab "${labId.trim()}"? This will permanently delete resources and cannot be undone.`;
    setConfirmMessage(message);
    setShowConfirmDialog(true);
  };

  const handleConfirmCleanup = async () => {
    setShowConfirmDialog(false);
    setCleanupLoading(true);
    setCleanupMessage("");

    try {
      const response = await apiService.cleanupServiceByID(serviceConfigId, labId.trim());
      setCleanupMessage(response.message);

      // Add to history
      setCleanupHistory(prev => [{
        id: Date.now().toString(),
        serviceType: response.service_type,
        status: 'success',
        message: response.message,
        timestamp: new Date(),
        parameters: { lab_id: labId.trim(), service_config_id: serviceConfigId }
      }, ...prev]);

      // Reset form
      setServiceConfigId("");
      setLabId("");
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      setCleanupMessage(`Cleanup failed: ${errorMessage}`);

      // Add to history
      setCleanupHistory(prev => [{
        id: Date.now().toString(),
        serviceType: selectedServiceConfig?.serviceType || 'unknown',
        status: 'error',
        message: errorMessage,
        timestamp: new Date(),
        parameters: { lab_id: labId.trim(), service_config_id: serviceConfigId }
      }, ...prev]);
    } finally {
      setCleanupLoading(false);
    }
  };


  return (
    <AppLayout>
      <div className="p-6 max-w-6xl mx-auto space-y-6">
        <header className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div className="space-y-1">
            <h1 className="text-3xl md:text-4xl font-semibold">Admin Cleanup Tools</h1>
            <p className="text-muted-foreground">Clean up resources by service type and lab UUID</p>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => window.location.reload()}
              disabled={cleanupLoading}
            >
              <RefreshCw className="h-4 w-4 mr-2" />
              Refresh
            </Button>
          </div>
        </header>

        {/* Service Cleanup Section */}
        <Card className="rounded-xl">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Trash2 className="h-5 w-5" />
              Service Cleanup
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {loadingServices ? (
              <div className="flex items-center justify-center py-8">
                <RefreshCw className="h-6 w-6 animate-spin mr-2" />
                Loading available services...
              </div>
            ) : (
              <>
                {/* Service Config Selection */}
                <div className="space-y-2">
                  <Label htmlFor="service-config">Service Configuration *</Label>
                  <Select value={serviceConfigId} onValueChange={setServiceConfigId}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select a service configuration to cleanup" />
                    </SelectTrigger>
                    <SelectContent>
                      {availableServices.map((service) => 
                        service.configs.map((config) => (
                          <SelectItem key={config.id} value={config.id}>
                            <div className="flex items-center gap-2">
                              <span className="font-medium">{config.name}</span>
                              <Badge variant="secondary" className="text-xs">
                                {service.type}
                              </Badge>
                              {!config.is_active && (
                                <Badge variant="outline" className="text-xs">Inactive</Badge>
                              )}
                            </div>
                          </SelectItem>
                        ))
                      )}
                    </SelectContent>
                  </Select>
                </div>

                {/* Lab ID */}
                <div className="space-y-2">
                  <Label htmlFor="lab-id">Lab UUID *</Label>
                  <Input
                    id="lab-id"
                    placeholder="Enter lab UUID (e.g., abc123)"
                    value={labId}
                    onChange={(e) => setLabId(e.target.value)}
                  />
                  <p className="text-sm text-muted-foreground">
                    Resource names will be auto-constructed from this lab UUID
                  </p>
                </div>

                {/* Service Information */}
                {selectedServiceConfig && (
                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <div className="flex items-start gap-2">
                      <Info className="h-4 w-4 text-blue-600 mt-0.5" />
                      <div className="space-y-1">
                        <p className="text-sm font-medium text-blue-900">
                          {selectedServiceConfig.name} ({selectedServiceConfig.serviceType})
                        </p>
                        <p className="text-sm text-blue-700">
                          {selectedServiceConfig.description}
                        </p>
                        <div className="text-xs text-blue-700">
                          <p className="font-medium">Auto-constructed resources:</p>
                          <ul className="mt-1 space-y-1">
                            <li>• project_name: {getAutoConstructedExample('project_name', labId || 'uuid')}</li>
                            <li>• user_email: {getAutoConstructedExample('user_email', labId || 'uuid')}</li>
                            <li>• api_key_name: {getAutoConstructedExample('api_key_name', labId || 'uuid')}</li>
                          </ul>
                        </div>
                      </div>
                    </div>
                  </div>
                )}

                {/* Cleanup Button */}
                <div className="flex gap-2">
                  <Button
                    onClick={handleServiceCleanup}
                    disabled={!serviceConfigId || !labId.trim() || cleanupLoading}
                    className="flex-1"
                  >
                    {cleanupLoading ? (
                      <>
                        <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
                        Cleaning up...
                      </>
                    ) : (
                      <>
                        <Trash2 className="h-4 w-4 mr-2" />
                        Cleanup {selectedServiceConfig?.name || 'Service'}
                      </>
                    )}
                  </Button>
                </div>
              </>
            )}
          </CardContent>
        </Card>


        {/* Status Message */}
        {cleanupMessage && (
          <Card className={`rounded-xl ${cleanupMessage.includes('failed') || cleanupMessage.includes('error') ? 'border-red-200 bg-red-50' : 'border-green-200 bg-green-50'}`}>
            <CardContent className="pt-6">
              <div className="flex items-center gap-2">
                {cleanupMessage.includes('failed') || cleanupMessage.includes('error') ? (
                  <AlertTriangle className="h-4 w-4 text-red-600" />
                ) : (
                  <CheckCircle className="h-4 w-4 text-green-600" />
                )}
                <p className={`text-sm ${cleanupMessage.includes('failed') || cleanupMessage.includes('error') ? 'text-red-800' : 'text-green-800'}`}>
                  {cleanupMessage}
                </p>
              </div>
            </CardContent>
          </Card>
        )}

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
                {cleanupHistory.slice(0, 10).map((item) => (
                  <div
                    key={item.id}
                    className={`flex items-start gap-3 p-3 rounded-lg border ${
                      item.status === 'success' ? 'border-green-200 bg-green-50' : 'border-red-200 bg-red-50'
                    }`}
                  >
                    {item.status === 'success' ? (
                      <CheckCircle className="h-4 w-4 text-green-600 mt-0.5" />
                    ) : (
                      <AlertTriangle className="h-4 w-4 text-red-600 mt-0.5" />
                    )}
                    <div className="flex-1 space-y-1">
                      <div className="flex items-center gap-2">
                        <Badge variant="outline" className="text-xs">
                          {item.serviceType}
                        </Badge>
                        <span className="text-xs text-muted-foreground">
                          {item.timestamp.toLocaleString()}
                        </span>
                      </div>
                      <p className={`text-sm ${item.status === 'success' ? 'text-green-800' : 'text-red-800'}`}>
                        {item.message}
                      </p>
                      {item.parameters && Object.keys(item.parameters).length > 0 && (
                        <div className="text-xs text-muted-foreground">
                          Parameters: {Object.entries(item.parameters).map(([key, value]) => `${key}=${value}`).join(', ')}
                        </div>
                      )}
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
                <li>Always verify the service type and parameters before cleanup</li>
                <li>Ensure no active users are using the resources</li>
                <li>Consider backing up important data before cleanup</li>
                <li>Use bulk cleanup operations during maintenance windows</li>
                <li>Service-specific cleanup will remove all associated resources</li>
                <li>Custom parameters allow fine-grained control over cleanup operations</li>
              </ul>
            </div>
          </CardContent>
        </Card>

        {/* Confirmation Dialog */}
        <AlertDialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Confirm Cleanup</AlertDialogTitle>
              <AlertDialogDescription>
                {confirmMessage}
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel onClick={() => setShowConfirmDialog(false)}>
                Cancel
              </AlertDialogCancel>
              <AlertDialogAction onClick={handleConfirmCleanup}>
                Continue
              </AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </AppLayout>
  );
}

// Helper function to get auto-constructed resource examples
function getAutoConstructedExample(paramName: string, labId: string): string {
  switch (paramName) {
    case 'project_name':
      return `lab-${labId}`;
    case 'user_email':
      return `lab+${labId}@spectrocloud.com`;
    case 'api_key_name':
      return `lab-${labId}-api-key`;
    case 'tenant_id':
      return `tenant-${labId}`;
    case 'username':
      return `lab-${labId}@pve`;
    case 'pool_name':
      return `lab-${labId}-pool`;
    case 'workspace_name':
      return `lab-${labId}-workspace`;
    default:
      return `lab-${labId}`;
  }
}

export default function CleanupPage() {
  return (
    <ProtectedRoute requireAdmin={true}>
      <CleanupPageContent />
    </ProtectedRoute>
  );
}