// Shared lab credentials display component

import React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { PasswordToggleFieldWithCopy } from "@/components/ui/password-toggle-field";
import { ShieldCheck, ExternalLink, Info } from "lucide-react";
import { Credential } from "@/types/lab";

interface LabCredentialsProps {
  credentials: Credential[];
}

export function LabCredentials({ credentials }: LabCredentialsProps) {
  if (credentials.length === 0) {
    return null;
  }

  return (
    <section className="grid grid-cols-1 md:grid-cols-2 gap-6">
      {credentials.map((cred) => (
        <Card key={cred.id} className="shadow-sm rounded-2xl">
          <CardHeader className="flex flex-row items-center justify-between space-y-0">
            <CardTitle className="text-lg flex items-center gap-2">
              <ShieldCheck className="h-5 w-5" /> {cred.label} Access
            </CardTitle>
            {cred.url && (
              <Button asChild variant="link" className="gap-1">
                <a 
                  href={cred.url.startsWith('http://') || cred.url.startsWith('https://') ? cred.url : `https://${cred.url}`} 
                  target="_blank" 
                  rel="noreferrer"
                >
                  Open <ExternalLink className="h-4 w-4" />
                </a>
              </Button>
            )}
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-1 gap-3">
              <div className="grid grid-cols-3 items-center gap-2">
                <Label className="col-span-1">Username</Label>
                <Input className="col-span-2 font-mono" value={cred.username} readOnly />
              </div>
              <div className="grid grid-cols-3 items-center gap-2">
                <Label className="col-span-1">Password</Label>
                <div className="col-span-2">
                  <PasswordToggleFieldWithCopy 
                    value={cred.password} 
                    placeholder="Password"
                    className="w-full"
                  />
                </div>
              </div>
            </div>
            <div className="text-xs text-muted-foreground flex items-center gap-2">
              <Info className="h-4 w-4" /> Expires at {new Date(cred.expiresAt).toLocaleString()}
            </div>
            {cred.notes && (
              <div className="text-sm text-muted-foreground">
                {cred.notes}
              </div>
            )}
          </CardContent>
        </Card>
      ))}
    </section>
  );
}
