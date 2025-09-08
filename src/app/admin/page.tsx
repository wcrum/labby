"use client";

import React, { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Progress } from "@/components/ui/progress";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { PasswordToggleFieldWithCopy } from "@/components/ui/password-toggle-field";
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
import { 
  Copy, 
  RefreshCw, 
  ExternalLink, 
  ShieldCheck, 
  User, 
  Users,
  Activity,
  Calendar,
  AlertTriangle
} from "lucide-react";
import { motion } from "framer-motion";
import { LabSession } from "@/types/lab";
import { ProtectedRoute } from "@/components/auth/ProtectedRoute";
import { AppLayout } from "@/components/layout/AppLayout";
import { apiService, ServiceUsage, ServiceConfig, ServiceLimit } from "@/lib/api";
import { convertLabResponses } from "@/lib/lab-utils";
import { useAllCountdowns } from "@/hooks/useCountdown";



// Fetch all lab sessions from the API
async function fetchAllLabSessions(): Promise<LabSession[]> {
  try {
    const labs = await apiService.getAllLabs();
    return convertLabResponses(labs);
  } catch (error) {
    console.error('Failed to fetch labs:', error);
    throw error; // Re-throw the error so the component can handle it
  }
}



function AdminPageContent() {
  const [labs, setLabs] = useState<LabSession[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm] = useState("");
  const [statusFilter] = useState<string>("all");
  const [serviceUsage, setServiceUsage] = useState<ServiceUsage[]>([]);
  const [serviceConfigs, setServiceConfigs] = useState<ServiceConfig[]>([]);
  const [serviceLimits, setServiceLimits] = useState<ServiceLimit[]>([]);
  const [showStopDialog, setShowStopDialog] = useState(false);
  const [selectedLab, setSelectedLab] = useState<LabSession | null>(null);

  const handleStopLab = async () => {
    if (!selectedLab) return;

    try {
      // Stop the lab first
      await apiService.adminStopLab(selectedLab.id);
      
      // Then cleanup any remaining resources
      try {
        await apiService.cleanupLab(selectedLab.id);
      } catch (cleanupError) {
        console.warn('Cleanup failed (lab may already be cleaned up):', cleanupError);
      }
      
      // Finally delete the lab record
      try {
        await apiService.adminDeleteLab(selectedLab.id);
      } catch (deleteError) {
        console.warn('Delete failed (lab may already be deleted):', deleteError);
      }
      
      // Refresh the labs list
      const data = await fetchAllLabSessions();
      setLabs(data);
      setShowStopDialog(false);
      setSelectedLab(null);
    } catch (error) {
      console.error('Failed to stop lab:', error);
      setError('Failed to stop lab');
    }
  };

  // Get countdown data for all labs using a single hook
  const labCountdowns = useAllCountdowns(labs);

  useEffect(() => {
    let live = true;
    (async () => {
      setLoading(true);
      setError(null);
      try {
        const [labsData, usageData, configsData, limitsData] = await Promise.all([
          fetchAllLabSessions(),
          apiService.getServiceUsage(),
          apiService.getServiceConfigs(),
          apiService.getServiceLimits()
        ]);
        if (!live) return;
        setLabs(labsData);
        setServiceUsage(usageData);
        setServiceConfigs(configsData);
        setServiceLimits(limitsData);
      } catch (err) {
        if (!live) return;
        console.error('Failed to fetch data:', err);
        setError(err instanceof Error ? err.message : 'Failed to load data');
        setLabs([]);
      } finally {
        if (live) {
          setLoading(false);
        }
      }
    })();
    
    const poll = setInterval(async () => {
      try {
        const [labsData, usageData] = await Promise.all([
          fetchAllLabSessions(),
          apiService.getServiceUsage()
        ]);
        if (!live) return;
        setLabs(labsData);
        setServiceUsage(usageData);
        setError(null);
      } catch (err) {
        if (!live) return;
        console.error('Failed to fetch data during poll:', err);
        setError(err instanceof Error ? err.message : 'Failed to load data');
      }
    }, 10000); // Poll every 10 seconds for more responsive updates
    
    return () => { live = false; clearInterval(poll); };
  }, []);

  const filteredLabs = labs.filter(lab => {
    const matchesSearch = lab.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         (lab.owner?.name || "Unknown").toLowerCase().includes(searchTerm.toLowerCase()) ||
                         (lab.owner?.email || "unknown").toLowerCase().includes(searchTerm.toLowerCase());
    const matchesStatus = statusFilter === "all" || lab.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  const stats = {
    total: labs.length,
    ready: labs.filter(l => l.status === "ready").length,
    provisioning: labs.filter(l => l.status === "provisioning").length,
    error: labs.filter(l => l.status === "error").length,
    expired: labs.filter(l => l.status === "expired").length,
  };



  if (loading) {
    return (
      <div className="p-6 space-y-4">
        <div className="h-8 w-64 bg-muted animate-pulse rounded" />
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                  {[0,1,2,3].map((_, i) => (
          <div key={i} className="h-24 bg-muted animate-pulse rounded-xl" />
        ))}
        </div>
        <div className="grid grid-cols-1 gap-4">
          {[0,1,2].map((_, i) => (
            <div key={i} className="h-32 bg-muted animate-pulse rounded-xl" />
          ))}
        </div>
      </div>
    );
  }

  return (
    <AppLayout>
      <div className="p-6 max-w-6xl mx-auto space-y-6">
        <header className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div className="space-y-1">
            <h1 className="text-3xl md:text-4xl font-semibold">Admin Dashboard</h1>
            <p className="text-muted-foreground">Monitor and manage lab sessions</p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={async () => {
                setLoading(true);
                setError(null);
                try {
                  const data = await fetchAllLabSessions();
                  setLabs(data);
                } catch (err) {
                  console.error('Failed to refresh labs:', err);
                  setError(err instanceof Error ? err.message : 'Failed to load labs');
                } finally {
                  setLoading(false);
                }
              }}
              disabled={loading}
              className="gap-2"
            >
              <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
            <Badge variant="outline" className="gap-1">
              <Users className="h-4 w-4" />
              {stats.total} Active Lab{stats.total !== 1 ? 's' : ''}
            </Badge>
          </div>
        </header>

        {/* Error Display */}
        {error && (
          <Card className="border-destructive">
            <CardContent className="p-4">
              <div className="flex items-center gap-2 text-destructive">
                <AlertTriangle className="h-5 w-5" />
                <p className="font-medium">Error loading labs</p>
              </div>
              <p className="text-sm text-muted-foreground mt-1">{error}</p>
            </CardContent>
          </Card>
        )}

        {/* Statistics Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card className="rounded-xl">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Total Labs</p>
                  <p className="text-2xl font-bold">{stats.total}</p>
                </div>
                <Users className="h-8 w-8 text-muted-foreground" />
              </div>
            </CardContent>
          </Card>
          <Card className="rounded-xl">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Ready</p>
                  <p className="text-2xl font-bold text-green-600">{stats.ready}</p>
                </div>
                <Activity className="h-8 w-8 text-green-600" />
              </div>
            </CardContent>
          </Card>
          <Card className="rounded-xl">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Provisioning</p>
                  <p className="text-2xl font-bold text-blue-600">{stats.provisioning}</p>
                </div>
                <RefreshCw className="h-8 w-8 text-blue-600" />
              </div>
            </CardContent>
          </Card>
          <Card className="rounded-xl">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Errors & Expired</p>
                  <p className="text-2xl font-bold text-red-600">{stats.error + stats.expired}</p>
                </div>
                <ShieldCheck className="h-8 w-8 text-red-600" />
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Service Usage Section */}
        <Card className="rounded-xl">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-5 w-5" />
              Service Usage & Limits
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {serviceUsage.length > 0 ? (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {serviceUsage.map((usage) => {
                    const config = serviceConfigs.find(c => c.id === usage.service_id);
                    const limit = serviceLimits.find(l => l.service_id === usage.service_id);
                    const usagePercentage = limit ? (usage.active_labs / limit.max_labs) * 100 : 0;
                    const isNearLimit = usagePercentage >= 80;
                    const isAtLimit = usagePercentage >= 100;

                    return (
                      <Card key={usage.service_id} className="p-4">
                        <div className="space-y-3">
                          <div className="flex items-center justify-between">
                            <div>
                              <h4 className="font-medium text-sm">
                                {config?.name || usage.service_id}
                              </h4>
                              <p className="text-xs text-muted-foreground">
                                {config?.type || 'Unknown Type'}
                              </p>
                            </div>
                            <Badge 
                              variant={isAtLimit ? "destructive" : isNearLimit ? "secondary" : "default"}
                              className="text-xs"
                            >
                              {usage.active_labs}/{limit?.max_labs || '∞'}
                            </Badge>
                          </div>
                          
                          <div className="space-y-2">
                            <div className="flex justify-between text-xs">
                              <span className="text-muted-foreground">Usage</span>
                              <span className="font-medium">{usage.active_labs} active labs</span>
                            </div>
                            <Progress 
                              value={Math.min(usagePercentage, 100)} 
                              className={`h-2 ${isAtLimit ? 'bg-red-100' : isNearLimit ? 'bg-yellow-100' : ''}`}
                            />
                            {limit && (
                              <div className="text-xs text-muted-foreground">
                                Limit: {limit.max_labs} labs, {limit.max_duration} min duration
                              </div>
                            )}
                            {isAtLimit && (
                              <div className="text-xs text-red-600 font-medium">
                                ⚠️ Service at capacity
                              </div>
                            )}
                            {isNearLimit && !isAtLimit && (
                              <div className="text-xs text-yellow-600 font-medium">
                                ⚠️ Approaching limit
                              </div>
                            )}
                          </div>
                        </div>
                      </Card>
                    );
                  })}
                </div>
              ) : (
                <div className="text-center py-8">
                  <Activity className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                  <p className="text-muted-foreground">No service usage data available</p>
                </div>
              )}
            </div>
          </CardContent>
        </Card>


      {/* Lab Overview */}
      {filteredLabs.map((lab) => {
        const labIndex = labs.findIndex(l => l.id === lab.id);
        const { countdown, progressValue } = labCountdowns[labIndex] || { countdown: { mins: 0, secs: 0 }, progressValue: 0 };

        return (
          <motion.div
            key={lab.id}
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.25 }}
          >
            <Card className="rounded-xl">
              <CardHeader>
                <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
                  <div className="space-y-2">
                    <div className="flex items-center gap-3">
                      <h3 className="text-xl font-semibold">{lab.name}</h3>
                      <Badge variant={lab.status === "ready" ? "default" : lab.status === "provisioning" ? "secondary" : "destructive"}>
                        {lab.status.toUpperCase()}
                      </Badge>
                    </div>
                    <div className="flex items-center gap-4 text-sm text-muted-foreground">
                      <div className="flex items-center gap-1">
                        <User className="h-4 w-4" />
                        {lab.owner?.name || "Unknown"} ({lab.owner?.email || "unknown"})
                      </div>
                      <div className="flex items-center gap-1">
                        <Calendar className="h-4 w-4" />
                        Started {lab.startedAt ? new Date(lab.startedAt).toLocaleString() : "—"}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <div className="text-right">
                      <div className="text-lg font-medium">
                        {String(isFinite(countdown.mins) ? countdown.mins : 0).padStart(2, "0")}:{String(isFinite(countdown.secs) ? countdown.secs : 0).padStart(2, "0")}
                      </div>
                      <div className="text-xs text-muted-foreground">remaining</div>
                    </div>
                    <div className="flex items-center gap-2">
                      <Button asChild variant="outline" size="sm">
                        <a href={`/lab?id=${lab.id}`} target="_blank" rel="noreferrer">
                          View Details
                        </a>
                      </Button>
                      {lab.status === 'ready' && (
                        <Button
                          variant="destructive"
                          size="sm"
                          onClick={() => {
                            setSelectedLab(lab);
                            setShowStopDialog(true);
                          }}
                        >
                          Stop Lab
                        </Button>
                      )}
                    </div>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                <Progress value={progressValue} className="h-2" />
                
                {lab.credentials.length > 0 && (
                  <div className="space-y-4">
                    <h4 className="text-lg font-medium">Access Credentials</h4>
                    <div className="rounded-md border">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead>Service</TableHead>
                            <TableHead>Username</TableHead>
                            <TableHead>Password</TableHead>
                            <TableHead>URL</TableHead>
                            <TableHead>Notes</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {lab.credentials.map((cred) => (
                            <TableRow key={cred.id}>
                              <TableCell className="font-medium">
                                <div className="flex items-center gap-2">
                                  <ShieldCheck className="h-4 w-4" />
                                  {cred.label}
                                </div>
                              </TableCell>
                              <TableCell>
                                <div className="flex items-center gap-2">
                                  <Input value={cred.username} readOnly className="font-mono text-sm h-8" />
                                  <TooltipProvider>
                                    <Tooltip>
                                      <TooltipTrigger asChild>
                                        <Button
                                          variant="outline"
                                          size="icon"
                                          className="h-8 w-8"
                                          onClick={() => {
                                            try {
                                              if (typeof navigator !== "undefined" && navigator.clipboard) navigator.clipboard.writeText(cred.username);
                                            } catch {}
                                          }}
                                        >
                                          <Copy className="h-4 w-4"/>
                                        </Button>
                                      </TooltipTrigger>
                                      <TooltipContent>Copy username</TooltipContent>
                                    </Tooltip>
                                  </TooltipProvider>
                                </div>
                              </TableCell>
                              <TableCell>
                                <PasswordToggleFieldWithCopy 
                                  value={cred.password} 
                                  placeholder="Password"
                                  className="w-full max-w-xs"
                                />
                              </TableCell>
                              <TableCell>
                                {cred.url ? (
                                  <Button asChild variant="ghost" size="sm">
                                    <a href={cred.url.startsWith('http://') || cred.url.startsWith('https://') ? cred.url : `https://${cred.url}`} target="_blank" rel="noreferrer" className="flex items-center gap-1">
                                      <ExternalLink className="h-4 w-4" />
                                      Open
                                    </a>
                                  </Button>
                                ) : (
                                  <span className="text-muted-foreground">—</span>
                                )}
                              </TableCell>
                              <TableCell className="text-sm text-muted-foreground max-w-xs">
                                {cred.notes || "—"}
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </div>
                  </div>
                )}
              </CardContent>
            </Card>
          </motion.div>
        );
      })}

      {filteredLabs.length === 0 && (
        <Card className="rounded-xl">
          <CardContent className="p-8 text-center">
            <p className="text-muted-foreground">No labs found matching your criteria.</p>
          </CardContent>
        </Card>
      )}

      {/* Stop Lab Confirmation Dialog */}
      <AlertDialog open={showStopDialog} onOpenChange={setShowStopDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Stop Lab</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to stop lab &quot;{selectedLab?.name}&quot;? This will delete all resources and cannot be undone.
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
    </AppLayout>
  );
}

export default function AdminPage() {
  return (
    <ProtectedRoute requireAdmin={true}>
      <AdminPageContent />
    </ProtectedRoute>
  );
}
