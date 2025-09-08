"use client";

import React, { useEffect, useState, useCallback, Suspense } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { CheckCircle, AlertCircle, Clock, Building2, User } from "lucide-react";
import { useAuth } from "@/lib/auth";
import { AppLayout } from "@/components/layout/AppLayout";
import { useSearchParams } from "next/navigation";

interface Invite {
  id: string;
  organization_id: string;
  email: string;
  invited_by: string;
  role: string;
  status: string;
  expires_at: string;
  created_at: string;
  accepted_at?: string;
}

interface Organization {
  id: string;
  name: string;
  description: string;
  domain: string;
  created_at: string;
  updated_at: string;
}

function InvitePageContent() {
  const { user, refreshUser } = useAuth();
  const searchParams = useSearchParams();
  const inviteCode = searchParams.get('code');
  
  const [invite, setInvite] = useState<Invite | null>(null);
  const [organization, setOrganization] = useState<Organization | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [accepting, setAccepting] = useState(false);
  const [accepted, setAccepted] = useState(false);

  const fetchInvite = useCallback(async () => {
    if (!inviteCode) {
      setError('No invite code provided');
      setLoading(false);
      return;
    }
    
    try {
      const response = await fetch(`/api/invites/${inviteCode}`);
      if (!response.ok) {
        throw new Error('Invite not found or has expired');
      }
      
      const inviteData: Invite = await response.json();
      setInvite(inviteData);
      
      // Fetch organization details
      if (inviteData.organization_id) {
        const orgResponse = await fetch(`/api/admin/organizations/${inviteData.organization_id}`);
        if (orgResponse.ok) {
          const orgData = await orgResponse.json();
          setOrganization(orgData.organization);
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load invite');
    } finally {
      setLoading(false);
    }
  }, [inviteCode]);

  useEffect(() => {
    fetchInvite();
  }, [fetchInvite]);

  const handleAcceptInvite = async () => {
    if (!user || !invite) return;

    setAccepting(true);
    try {
      const requestBody = {
        invite_id: invite.id,
        user_id: user.id,
      };
      

      const response = await fetch(`/api/invites/${invite.id}/accept`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestBody),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Failed to accept invite: ${response.status} ${errorText}`);
      }

      const result = await response.json();

      setAccepted(true);
      
      // Refresh user data to get updated organization information
      try {
        await refreshUser();
      } catch (error) {
        // Fallback to page reload if refresh fails
        setTimeout(() => {
          window.location.reload();
        }, 2000);
      }
    } catch (err) {
      console.error('DEBUG: Invite acceptance error:', err);
      setError(err instanceof Error ? err.message : 'Failed to accept invite');
    } finally {
      setAccepting(false);
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

  if (error) {
    return (
      <AppLayout>
        <div className="p-6 max-w-2xl mx-auto">
          <Card className="border-destructive">
            <CardHeader>
              <CardTitle className="text-destructive flex items-center gap-2">
                <AlertCircle className="h-5 w-5" />
                Invite Error
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <AlertDescription>{error}</AlertDescription>
              <div className="flex gap-2">
                <Button onClick={() => window.location.href = '/'}>
                  Go Home
                </Button>
                <Button variant="outline" onClick={() => window.location.reload()}>
                  Try Again
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </AppLayout>
    );
  }

  if (!invite) {
    return (
      <AppLayout>
        <div className="p-6 max-w-2xl mx-auto">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <AlertCircle className="h-5 w-5" />
                Invite Not Found
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground">
                This invitation could not be found or may have expired.
              </p>
            </CardContent>
          </Card>
        </div>
      </AppLayout>
    );
  }

  if (accepted) {
    return (
      <AppLayout>
        <div className="p-6 max-w-2xl mx-auto">
          <Card className="border-green-200 bg-green-50">
            <CardHeader>
              <CardTitle className="text-green-800 flex items-center gap-2">
                <CheckCircle className="h-5 w-5" />
                Invite Accepted!
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <p className="text-green-700">
                You have successfully joined {organization?.name || 'the organization'}!
              </p>
              <Button onClick={() => window.location.href = '/'}>
                Go to Dashboard
              </Button>
            </CardContent>
          </Card>
        </div>
      </AppLayout>
    );
  }

  const isExpired = new Date(invite.expires_at) < new Date();
  const isPending = invite.status === 'pending' && !isExpired;

  return (
    <AppLayout>
      <div className="p-6 max-w-2xl mx-auto">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Building2 className="h-5 w-5" />
              Organization Invitation
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            {organization && (
              <div className="space-y-2">
                <h3 className="text-lg font-semibold">{organization.name}</h3>
                <p className="text-muted-foreground">{organization.description}</p>
                {organization.domain && (
                  <p className="text-sm text-muted-foreground">
                    Domain: {organization.domain}
                  </p>
                )}
              </div>
            )}

            <div className="space-y-4">
              <div className="flex items-center gap-2">
                <User className="h-4 w-4" />
                <span className="text-sm text-muted-foreground">
                  Invited as: <span className="font-medium capitalize">{invite.role}</span>
                </span>
              </div>
              
              <div className="flex items-center gap-2">
                <Clock className="h-4 w-4" />
                <span className="text-sm text-muted-foreground">
                  Expires: {new Date(invite.expires_at).toLocaleString()}
                </span>
              </div>
            </div>

            {isExpired && (
              <Alert className="border-amber-200 bg-amber-50">
                <AlertCircle className="h-4 w-4 text-amber-600" />
                <AlertDescription className="text-amber-800">
                  This invitation has expired and can no longer be accepted.
                </AlertDescription>
              </Alert>
            )}

            {!isPending && (
              <Alert>
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>
                  This invitation cannot be accepted at this time.
                </AlertDescription>
              </Alert>
            )}

            {isPending && (
              <div className="space-y-4">
                {!user ? (
                  <div className="space-y-2">
                    <p className="text-sm text-muted-foreground">
                      You need to be logged in to accept this invitation.
                    </p>
                    <Button onClick={() => window.location.href = `/?code=${inviteCode}`}>
                      Log In
                    </Button>
                  </div>
                ) : (
                  <div className="space-y-2">
                    <p className="text-sm text-muted-foreground">
                      Welcome, {user.name}! Click below to accept this invitation.
                    </p>
                    <Button 
                      onClick={handleAcceptInvite}
                      disabled={accepting}
                      className="w-full"
                    >
                      {accepting ? 'Accepting...' : 'Accept Invitation'}
                    </Button>
                  </div>
                )}
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </AppLayout>
  );
}

export default function InvitePage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <InvitePageContent />
    </Suspense>
  );
}
