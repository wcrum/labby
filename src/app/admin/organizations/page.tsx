"use client";

import React, { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { 
  Plus, 
  Users, 
  Mail, 
  Building2, 
  Calendar,
  Edit
} from "lucide-react";
import { apiService, Organization } from "@/lib/api";
import { ProtectedRoute } from "@/components/auth/ProtectedRoute";
import { AppLayout } from "@/components/layout/AppLayout";

function OrganizationsPageContent() {
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  
  // Modal states
  const [showCreateOrg, setShowCreateOrg] = useState(false);
  const [showCreateInvite, setShowCreateInvite] = useState(false);
  const [selectedOrg, setSelectedOrg] = useState<Organization | null>(null);
  const [newOrg, setNewOrg] = useState({ name: '', description: '', domain: '' });
  const [newInvite, setNewInvite] = useState({ email: '', role: 'member' });

  useEffect(() => {
    fetchOrganizations();
  }, []);

  const fetchOrganizations = async () => {
    try {
      setLoading(true);
      const orgs = await apiService.getOrganizations();
      setOrganizations(orgs);
    } catch (err) {
      console.error('Failed to fetch organizations:', err);
      setError('Failed to load organizations');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateOrg = async () => {
    try {
      const org = await apiService.createOrganization(newOrg);
      setOrganizations([...organizations, org]);
      setNewOrg({ name: '', description: '', domain: '' });
      setShowCreateOrg(false);
    } catch (error) {
      console.error('Failed to create organization:', error);
    }
  };

  const handleCreateInvite = async () => {
    try {
      const invite = await apiService.createInvite(selectedOrg?.id || 'default', newInvite);
      setNewInvite({ email: '', role: 'member' });
      setShowCreateInvite(false);
      
      // Show the invite link
      const inviteUrl = `${window.location.origin}/invite?code=${invite.id}`;
      alert(`Invitation created successfully!\n\nInvite Link: ${inviteUrl}\n\nCopy this link and share it with the user.`);
    } catch (error) {
      console.error('Failed to create invite:', error);
    }
  };

  if (loading) {
    return (
      <div className="p-6 space-y-4">
        <div className="h-8 w-64 bg-muted animate-pulse rounded" />
        <div className="h-96 bg-muted animate-pulse rounded-xl" />
      </div>
    );
  }

  return (
    <AppLayout>
      <div className="p-6 max-w-6xl mx-auto space-y-6">
        <header className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div className="space-y-1">
            <h1 className="text-3xl md:text-4xl font-semibold">Organization Management</h1>
            <p className="text-muted-foreground">Manage organizations and their members</p>
          </div>
          <Button onClick={() => setShowCreateOrg(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Create Organization
          </Button>
        </header>

        {/* Error Display */}
        {error && (
          <Card className="border-destructive">
            <CardContent className="p-4">
              <div className="flex items-center gap-2 text-destructive">
                <p className="font-medium">Error loading organizations</p>
              </div>
              <p className="text-sm text-muted-foreground mt-1">{error}</p>
            </CardContent>
          </Card>
        )}

        {/* Organizations List */}
        <div className="space-y-6">
          {organizations.map((org) => (
            <Card key={org.id} className="p-6">
              <div className="flex items-start justify-between">
                <div className="space-y-3 flex-1">
                  <div className="flex items-center gap-3">
                    <Building2 className="h-6 w-6 text-muted-foreground" />
                    <div>
                      <h3 className="text-xl font-semibold">{org.name}</h3>
                      <p className="text-muted-foreground">{org.description}</p>
                    </div>
                  </div>
                  
                  {org.domain && (
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <span>Domain: {org.domain}</span>
                    </div>
                  )}
                  
                  <div className="flex items-center gap-2 text-sm text-muted-foreground">
                    <Calendar className="h-4 w-4" />
                    <span>Created: {new Date(org.created_at).toLocaleDateString()}</span>
                  </div>
                </div>
                
                <div className="flex items-center gap-2">
                  <Button variant="outline" onClick={() => {
                    setSelectedOrg(org);
                    setShowCreateInvite(true);
                  }}>
                    <Mail className="h-4 w-4 mr-2" />
                    Send Invite
                  </Button>
                  <Button variant="outline" size="icon">
                    <Edit className="h-4 w-4" />
                  </Button>
                  <Button variant="outline" size="icon">
                    <Users className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </Card>
          ))}
        </div>

        {organizations.length === 0 && (
          <Card className="rounded-xl">
            <CardContent className="p-8 text-center">
              <Building2 className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-lg font-semibold mb-2">No Organizations</h3>
              <p className="text-muted-foreground mb-4">
                Get started by creating your first organization.
              </p>
              <Button onClick={() => setShowCreateOrg(true)}>
                <Plus className="h-4 w-4 mr-2" />
                Create Organization
              </Button>
            </CardContent>
          </Card>
        )}

        {/* Create Organization Modal */}
        {showCreateOrg && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <Card className="w-full max-w-md">
              <CardHeader>
                <CardTitle>Create Organization</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="org-name">Name</Label>
                  <Input
                    id="org-name"
                    value={newOrg.name}
                    onChange={(e) => setNewOrg({ ...newOrg, name: e.target.value })}
                    placeholder="Organization name"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="org-description">Description</Label>
                  <Input
                    id="org-description"
                    value={newOrg.description}
                    onChange={(e) => setNewOrg({ ...newOrg, description: e.target.value })}
                    placeholder="Organization description"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="org-domain">Domain (optional)</Label>
                  <Input
                    id="org-domain"
                    value={newOrg.domain}
                    onChange={(e) => setNewOrg({ ...newOrg, domain: e.target.value })}
                    placeholder="example.com"
                  />
                </div>
                <div className="flex gap-2">
                  <Button onClick={handleCreateOrg} disabled={!newOrg.name}>
                    Create
                  </Button>
                  <Button variant="outline" onClick={() => setShowCreateOrg(false)}>
                    Cancel
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        )}

        {/* Create Invite Modal */}
        {showCreateInvite && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
            <Card className="w-full max-w-md">
              <CardHeader>
                <CardTitle>Send Invitation</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="invite-email">Email</Label>
                  <Input
                    id="invite-email"
                    type="email"
                    value={newInvite.email}
                    onChange={(e) => setNewInvite({ ...newInvite, email: e.target.value })}
                    placeholder="user@example.com"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="invite-role">Role</Label>
                  <select
                    id="invite-role"
                    value={newInvite.role}
                    onChange={(e) => setNewInvite({ ...newInvite, role: e.target.value })}
                    className="w-full p-2 border rounded-md"
                  >
                    <option value="member">Member</option>
                    <option value="admin">Admin</option>
                  </select>
                </div>
                <div className="space-y-2">
                  <Label>Organization</Label>
                  <p className="text-sm text-muted-foreground">
                    {selectedOrg?.name || 'No organization selected'}
                  </p>
                </div>
                <div className="flex gap-2">
                  <Button onClick={handleCreateInvite} disabled={!newInvite.email || !selectedOrg}>
                    Send Invite
                  </Button>
                  <Button variant="outline" onClick={() => setShowCreateInvite(false)}>
                    Cancel
                  </Button>
                </div>
              </CardContent>
            </Card>
          </div>
        )}
      </div>
    </AppLayout>
  );
}

export default function OrganizationsPage() {
  return (
    <ProtectedRoute requireAdmin={true}>
      <OrganizationsPageContent />
    </ProtectedRoute>
  );
}
