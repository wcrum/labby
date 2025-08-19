"use client";

import React from "react";
import Link from "next/link";
import { useAuth } from "@/lib/auth";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import { Shield, User, LogOut, Settings, Users, FlaskConical } from "lucide-react";

interface AppLayoutProps {
  children: React.ReactNode;
  showNav?: boolean;
}

export function AppLayout({ children, showNav = true }: AppLayoutProps) {
  const { user, logout, isAdmin } = useAuth();

  if (!user) {
    return <>{children}</>;
  }

  return (
    <div className="min-h-screen bg-background flex flex-col">
      {/* Unified Navigation Bar */}
      {showNav && (
        <nav className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
          <div className="container mx-auto px-6 py-3">
            <div className="flex items-center justify-between">
              {/* Left side - App title and navigation */}
              <div className="flex items-center space-x-6">
                <div className="flex items-center space-x-3">
                  <img 
                    src="/SpectroCloud_Horizontal_light-bkgd_RGB.png" 
                    alt="SpectroCloud" 
                    className="h-8 w-auto"
                  />
                  <span className="text-sm text-muted-foreground font-medium"></span>
                </div>
                
                {/* Main navigation links */}
                <div className="flex items-center space-x-4">
                  <Button variant="ghost" asChild>
                    <Link href="/" className="flex items-center gap-2">
                      <FlaskConical className="h-4 w-4" />
                      My Labs
                    </Link>
                  </Button>
                  
                  {isAdmin && (
                    <Button variant="ghost" asChild>
                      <Link href="/admin" className="flex items-center gap-2">
                        <Shield className="h-4 w-4" />
                        Admin Dashboard
                      </Link>
                    </Button>
                  )}
                  
                  <Button variant="ghost" asChild>
                    <Link href="/support" className="flex items-center gap-2">
                      <Users className="h-4 w-4" />
                      Support
                    </Link>
                  </Button>
                </div>
              </div>

              {/* Right side - User info and actions */}
              <div className="flex items-center gap-3">
                <div className="flex items-center gap-2">
                  <Badge variant="outline" className="text-xs">
                    {user.role}
                  </Badge>
                  <span className="text-sm text-muted-foreground">
                    {user.name}
                  </span>
                </div>
                
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon">
                      <User className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-56">
                    <DropdownMenuLabel>
                      <div className="flex flex-col space-y-1">
                        <p className="text-sm font-medium leading-none">{user.name}</p>
                        <p className="text-xs leading-none text-muted-foreground">
                          {user.email}
                        </p>
                      </div>
                    </DropdownMenuLabel>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem>
                      <Settings className="mr-2 h-4 w-4" />
                      <span>Settings</span>
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem onClick={logout}>
                      <LogOut className="mr-2 h-4 w-4" />
                      <span>Log out</span>
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            </div>
          </div>
        </nav>
      )}

      {/* Main content area */}
      <main className="flex-1">
        {children}
      </main>
    </div>
  );
}
