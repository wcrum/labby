"use client";

import { Suspense } from "react";
import { useAuth } from "@/lib/auth";

import { LabManager } from "@/components/lab/LabManager";
import { LoginForm } from "@/components/auth/LoginForm";
import { AppLayout } from "@/components/layout/AppLayout";

function HomeContent() {
  const { user } = useAuth();

  if (!user) {
    return <LoginForm />;
  }

  return (
    <AppLayout>
      <div className="container mx-auto px-6 py-6">
        <div className="mb-6">
          <h2 className="text-2xl font-bold mb-2">My Lab Sessions</h2>
          <p className="text-muted-foreground">
            Access your current lab session or start a new one
          </p>
        </div>
        
        {/* Full Width Lab Manager */}
        <div className="w-full">
          <LabManager />
        </div>
      </div>
    </AppLayout>
  );
}

export default function Home() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <HomeContent />
    </Suspense>
  );
}
