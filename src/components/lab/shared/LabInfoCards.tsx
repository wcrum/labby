// Shared lab information cards component

import React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { LabSession } from "@/types/lab";

interface LabInfoCardsProps {
  lab: LabSession;
}

export function LabInfoCards({ lab }: LabInfoCardsProps) {
  return (
    <section className="grid grid-cols-1 md:grid-cols-3 gap-4">
      <Card className="rounded-2xl">
        <CardHeader>
          <CardTitle className="text-base">Helpful commands</CardTitle>
        </CardHeader>
        <CardContent>
          <pre className="bg-muted p-3 rounded-xl text-xs overflow-auto">
            {`# Example: ssh into Proxmox jump host
ssh ${lab.credentials.find((c) => c.id === "proxmox")?.username || "user"}@jump.lab.local

# Example: port-forward VerteX UI if on VPN
kubectl port-forward svc/vertex-ui 8443:443 -n vertex`}
          </pre>
        </CardContent>
      </Card>
      
      <Card className="rounded-2xl">
        <CardHeader>
          <CardTitle className="text-base">Support</CardTitle>
        </CardHeader>
        <CardContent className="text-sm space-y-2">
          <p>Having trouble? Capture the error and share with the proctor.</p>
          <ul className="list-disc ml-5 space-y-1 text-muted-foreground">
            <li>Check your VPN and DNS resolution.</li>
            <li>Try a private window for UI logins.</li>
            <li>Contact support if credentials don&apos;t work.</li>
          </ul>
        </CardContent>
      </Card>
      
      <Card className="rounded-2xl">
        <CardHeader>
          <CardTitle className="text-base">Session details</CardTitle>
        </CardHeader>
        <CardContent className="text-sm space-y-1 text-muted-foreground">
          <div>Lab ID: <span className="font-mono text-foreground">{lab.id}</span></div>
          <div>Started: {lab.startedAt ? new Date(lab.startedAt).toLocaleString() : "—"}</div>
          <div>Ends: {lab.endsAt ? new Date(lab.endsAt).toLocaleString() : "—"}</div>
        </CardContent>
      </Card>
    </section>
  );
}
