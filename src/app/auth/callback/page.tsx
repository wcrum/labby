"use client";

import { useEffect, useState, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuth } from '@/lib/auth';
import { apiService } from '@/lib/api';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2, AlertCircle, CheckCircle, Home } from 'lucide-react';
import { useTheme } from '@/lib/theme';

function AuthCallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { refreshUser } = useAuth();
  const { resolvedTheme } = useTheme();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const handleCallback = async () => {
      try {
        const code = searchParams.get('code');
        const state = searchParams.get('state');
        const errorParam = searchParams.get('error');

        if (errorParam) {
          setError(`Authentication failed: ${errorParam}`);
          setIsLoading(false);
          return;
        }

        if (!code || !state) {
          setError('Missing authentication parameters');
          setIsLoading(false);
          return;
        }

        // Handle OIDC callback
        await apiService.oidcCallback(code, state);
        
        // Refresh user data
        await refreshUser();
        
        // Redirect to main page
        router.push('/');
      } catch (err) {
        console.error('OIDC callback error:', err);
        setError(err instanceof Error ? err.message : 'Authentication failed');
        setIsLoading(false);
      }
    };

    handleCallback();
  }, [searchParams, router, refreshUser]);

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background p-4">
        <Card className="w-full max-w-md bg-background border-border shadow-lg">
          <CardHeader className="text-center space-y-4">
            <div className="mx-auto">
              <img 
                src={resolvedTheme === "dark" 
                  ? "/SpectroCloud_Horizontal_dark-bkgd_RGB.png" 
                  : "/SpectroCloud_Horizontal_light-bkgd_RGB.png"
                } 
                alt="SpectroCloud" 
                className="h-12 w-auto"
              />
            </div>
            <div className="space-y-2">
              <CardTitle className="text-2xl font-bold">Completing Authentication</CardTitle>
              <CardDescription>
                Please wait while we complete your sign-in process
              </CardDescription>
            </div>
          </CardHeader>
          <CardContent className="text-center">
            <div className="flex items-center justify-center space-x-2">
              <Loader2 className="h-6 w-6 animate-spin text-primary" />
              <span className="text-muted-foreground">Processing...</span>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background p-4">
        <Card className="w-full max-w-md bg-background border-border shadow-lg">
          <CardHeader className="text-center space-y-4">
            <div className="mx-auto">
              <img 
                src={resolvedTheme === "dark" 
                  ? "/SpectroCloud_Horizontal_dark-bkgd_RGB.png" 
                  : "/SpectroCloud_Horizontal_light-bkgd_RGB.png"
                } 
                alt="SpectroCloud" 
                className="h-12 w-auto"
              />
            </div>
            <div className="space-y-2">
              <CardTitle className="text-2xl font-bold">Authentication Failed</CardTitle>
              <CardDescription>
                We encountered an issue while signing you in
              </CardDescription>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <Alert variant="destructive">
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                {error}
              </AlertDescription>
            </Alert>
            
            <div className="text-center">
              <Button
                onClick={() => router.push('/')}
                className="w-full"
                variant="default"
              >
                <Home className="mr-2 h-4 w-4" />
                Return to Home
              </Button>
            </div>
            
            <div className="text-center text-sm text-muted-foreground">
              <p>You can try signing in again from the home page</p>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return null;
}

export default function AuthCallback() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center bg-background p-4">
        <Card className="w-full max-w-md bg-background border-border shadow-lg">
          <CardContent className="text-center py-8">
            <div className="flex items-center justify-center space-x-2">
              <Loader2 className="h-6 w-6 animate-spin text-primary" />
              <span className="text-muted-foreground">Loading...</span>
            </div>
          </CardContent>
        </Card>
      </div>
    }>
      <AuthCallbackContent />
    </Suspense>
  );
}
