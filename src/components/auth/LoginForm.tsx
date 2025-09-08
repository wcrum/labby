"use client";

import React, { useState } from "react";
import { useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Loader2, Mail } from "lucide-react";
import { useAuth } from "@/lib/auth";
import { useTheme } from "@/lib/theme";

export function LoginForm() {
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");
  const [showErrorDialog, setShowErrorDialog] = useState(false);
  const { login, isLoading } = useAuth();
  const { resolvedTheme } = useTheme();
  const searchParams = useSearchParams();
  const inviteCode = searchParams.get('code');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!email.trim()) {
      setError("Please enter your email address");
      return;
    }

    try {
      await login(email, inviteCode || undefined);
    } catch {
      setError("Login failed. Please try again.");
      setShowErrorDialog(true);
    }
  };

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
            <CardTitle className="text-2xl font-bold">Welcome to Spectro Lab</CardTitle>
            <CardDescription>
              {inviteCode ? (
                <>You&apos;ve been invited to join an organization. Enter your email to get started.</>
              ) : (
                <>Enter your email to access your lab sessions</>
              )}
            </CardDescription>
          </div>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="email">Email Address</Label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="email"
                  type="email"
                  placeholder="Enter your email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="pl-10"
                  disabled={isLoading}
                />
              </div>
            </div>


            <Button 
              type="submit" 
              className="w-full" 
              disabled={isLoading}
            >
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Signing in...
                </>
              ) : (
                "Sign In"
              )}
            </Button>
          </form>

          <div className="mt-6 text-center text-sm text-muted-foreground">
            <p>No password required - just enter your email to get started</p>

          </div>
        </CardContent>
      </Card>

      <AlertDialog open={showErrorDialog} onOpenChange={setShowErrorDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Login Failed</AlertDialogTitle>
            <AlertDialogDescription>
              {error}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogAction onClick={() => setShowErrorDialog(false)}>
              Try Again
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
