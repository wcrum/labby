"use client";

import React, { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import { 
  RefreshCw, 
  Mail, 
  Users, 
  Calendar,
  ExternalLink,
  Building2,
  Activity,
  ArrowLeft
} from "lucide-react";
import { motion } from "framer-motion";
import { apiService, Invite, InviteUsageStats } from "@/lib/api";
import { ProtectedRoute } from "@/components/auth/ProtectedRoute";
import { AppLayout } from "@/components/layout/AppLayout";
import { AdminPageSkeleton } from "@/components/ui/loading-skeleton";

function InvitesPageContent() {
  const [invites, setInvites] = useState<Invite[]>([]);
  const [usageStats, setUsageStats] = useState<InviteUsageStats[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<'invites' | 'usage'>('invites');
  const [searchTerm, setSearchTerm] = useState("");

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      setLoading(true);
      setError(null);
      const [invitesData, statsData] = await Promise.all([
        apiService.getAllInvites(),
        apiService.getInviteUsageStats()
      ]);
      setInvites(invitesData);
      setUsageStats(statsData);
    } catch (err) {
      console.error('Failed to fetch invites data:', err);
      setError('Failed to load invites data');
    } finally {
      setLoading(false);
    }
  };

  const filteredInvites = invites.filter(invite => 
    invite.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
    invite.organization_id.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const filteredStats = (usageStats || []).filter(stat => 
    stat.email.toLowerCase().includes(searchTerm.toLowerCase()) ||
    stat.organization_name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'pending':
        return <Badge variant="default">Pending</Badge>;
      case 'accepted':
        return <Badge variant="secondary">Accepted</Badge>;
      case 'expired':
        return <Badge variant="destructive">Expired</Badge>;
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const getUsageBadge = (usageCount: number, usageLimit?: number) => {
    if (!usageLimit) {
      return <Badge variant="outline">{usageCount} uses</Badge>;
    }
    
    const percentage = (usageCount / usageLimit) * 100;
    if (percentage >= 100) {
      return <Badge variant="destructive">{usageCount}/{usageLimit}</Badge>;
    } else if (percentage >= 80) {
      return <Badge variant="secondary">{usageCount}/{usageLimit}</Badge>;
    } else {
      return <Badge variant="default">{usageCount}/{usageLimit}</Badge>;
    }
  };

  if (loading) {
    return <AdminPageSkeleton />;
  }

  return (
    <AppLayout>
      <div className="p-6 max-w-7xl mx-auto space-y-6">
        <header className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div className="space-y-1">
            <div className="flex items-center gap-2">
              <Button variant="ghost" size="sm" asChild>
                <a href="/admin/organizations">
                  <ArrowLeft className="h-4 w-4 mr-2" />
                  Back to Organizations
                </a>
              </Button>
            </div>
            <h1 className="text-3xl md:text-4xl font-semibold">Invite Management</h1>
            <p className="text-muted-foreground">Manage and monitor all organization invites</p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={fetchData}
              disabled={loading}
              className="gap-2"
            >
              <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
              Refresh
            </Button>
            <div className="flex rounded-md border">
              <Button
                variant={viewMode === 'invites' ? 'default' : 'ghost'}
                size="sm"
                onClick={() => setViewMode('invites')}
                className="rounded-r-none"
              >
                <Mail className="h-4 w-4 mr-2" />
                Invites
              </Button>
              <Button
                variant={viewMode === 'usage' ? 'default' : 'ghost'}
                size="sm"
                onClick={() => setViewMode('usage')}
                className="rounded-l-none"
              >
                <Activity className="h-4 w-4 mr-2" />
                Usage Stats
              </Button>
            </div>
          </div>
        </header>

        {/* Error Display */}
        {error && (
          <Card className="border-destructive">
            <CardContent className="p-4">
              <div className="flex items-center gap-2 text-destructive">
                <p className="font-medium">Error loading invites</p>
              </div>
              <p className="text-sm text-muted-foreground mt-1">{error}</p>
            </CardContent>
          </Card>
        )}

        {/* Search */}
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center gap-2">
              <Input
                placeholder="Search by email or organization..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="max-w-sm"
              />
            </div>
          </CardContent>
        </Card>

        {/* Statistics Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <Card className="rounded-xl">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Total Invites</p>
                  <p className="text-2xl font-bold">{invites.length}</p>
                </div>
                <Mail className="h-8 w-8 text-muted-foreground" />
              </div>
            </CardContent>
          </Card>
          <Card className="rounded-xl">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Pending</p>
                  <p className="text-2xl font-bold text-blue-600">
                    {invites.filter(i => i.status === 'pending').length}
                  </p>
                </div>
                <Calendar className="h-8 w-8 text-blue-600" />
              </div>
            </CardContent>
          </Card>
          <Card className="rounded-xl">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Accepted</p>
                  <p className="text-2xl font-bold text-green-600">
                    {invites.filter(i => i.status === 'accepted').length}
                  </p>
                </div>
                <Users className="h-8 w-8 text-green-600" />
              </div>
            </CardContent>
          </Card>
          <Card className="rounded-xl">
            <CardContent className="p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-muted-foreground">Total Usage</p>
                  <p className="text-2xl font-bold text-purple-600">
                    {invites.reduce((sum, invite) => sum + invite.usage_count, 0)}
                  </p>
                </div>
                <Activity className="h-8 w-8 text-purple-600" />
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Invites Table */}
        {viewMode === 'invites' && (
          <Card className="rounded-xl">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Mail className="h-5 w-5" />
                All Invites
              </CardTitle>
            </CardHeader>
            <CardContent>
              {filteredInvites.length > 0 ? (
                <div className="rounded-md border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Email</TableHead>
                        <TableHead>Organization</TableHead>
                        <TableHead>Role</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Usage</TableHead>
                        <TableHead>Created</TableHead>
                        <TableHead>Expires</TableHead>
                        <TableHead>Actions</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {filteredInvites.map((invite) => (
                        <TableRow key={invite.id}>
                          <TableCell className="font-medium">{invite.email}</TableCell>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <Building2 className="h-4 w-4 text-muted-foreground" />
                              {invite.organization_id}
                            </div>
                          </TableCell>
                          <TableCell>
                            <Badge variant="outline">{invite.role}</Badge>
                          </TableCell>
                          <TableCell>{getStatusBadge(invite.status)}</TableCell>
                          <TableCell>{getUsageBadge(invite.usage_count, invite.usage_limit)}</TableCell>
                          <TableCell className="text-sm text-muted-foreground">
                            {new Date(invite.created_at).toLocaleDateString()}
                          </TableCell>
                          <TableCell className="text-sm text-muted-foreground">
                            {new Date(invite.expires_at).toLocaleDateString()}
                          </TableCell>
                          <TableCell>
                            <Button variant="ghost" size="sm" asChild>
                              <a href={`/invite?code=${invite.id}`} target="_blank" rel="noreferrer">
                                <ExternalLink className="h-4 w-4" />
                              </a>
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              ) : (
                <div className="text-center py-8">
                  <Mail className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                  <p className="text-muted-foreground">No invites found</p>
                </div>
              )}
            </CardContent>
          </Card>
        )}

        {/* Usage Statistics Table */}
        {viewMode === 'usage' && (
          <Card className="rounded-xl">
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Activity className="h-5 w-5" />
                Usage Statistics
              </CardTitle>
            </CardHeader>
            <CardContent>
              {filteredStats.length > 0 ? (
                <div className="rounded-md border">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>Email</TableHead>
                        <TableHead>Organization</TableHead>
                        <TableHead>Usage</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Last Used</TableHead>
                        <TableHead>Created</TableHead>
                        <TableHead>Expires</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {filteredStats.map((stat) => (
                        <TableRow key={stat.invite_id}>
                          <TableCell className="font-medium">{stat.email}</TableCell>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              <Building2 className="h-4 w-4 text-muted-foreground" />
                              {stat.organization_name}
                            </div>
                          </TableCell>
                          <TableCell>{getUsageBadge(stat.usage_count, stat.usage_limit)}</TableCell>
                          <TableCell>{getStatusBadge(stat.status)}</TableCell>
                          <TableCell className="text-sm text-muted-foreground">
                            {stat.last_used_at ? new Date(stat.last_used_at).toLocaleDateString() : 'Never'}
                          </TableCell>
                          <TableCell className="text-sm text-muted-foreground">
                            {new Date(stat.created_at).toLocaleDateString()}
                          </TableCell>
                          <TableCell className="text-sm text-muted-foreground">
                            {new Date(stat.expires_at).toLocaleDateString()}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              ) : (
                <div className="text-center py-8">
                  <Activity className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                  <p className="text-muted-foreground">No usage statistics available</p>
                </div>
              )}
            </CardContent>
          </Card>
        )}
      </div>
    </AppLayout>
  );
}

export default function InvitesPage() {
  return (
    <ProtectedRoute requireAdmin={true}>
      <InvitesPageContent />
    </ProtectedRoute>
  );
}
