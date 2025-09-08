"use client";

import React, { useEffect, useState } from "react";
import Link from "next/link";
import { useAuth } from "@/lib/auth";
import { apiService, Organization } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu";
import {
  NavigationMenu,
  NavigationMenuContent,
  NavigationMenuIndicator,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  NavigationMenuTrigger,
  NavigationMenuViewport,
} from "@/components/ui/navigation-menu";
import { Shield, User, LogOut, Settings, Users, FlaskConical, Trash2, Building2 } from "lucide-react";
import { ThemeToggle } from "@/components/ui/theme-toggle";
import { useTheme } from "@/lib/theme";

interface AppLayoutProps {
  children: React.ReactNode;
  showNav?: boolean;
}

export function AppLayout({ children, showNav = true }: AppLayoutProps) {
  const { user, logout, isAdmin } = useAuth();
  const { resolvedTheme } = useTheme();
  const [userOrganization, setUserOrganization] = useState<Organization | null>(null);
  const [loadingOrg, setLoadingOrg] = useState(false);

  // Fetch user's organization when user changes
  useEffect(() => {
    if (user) {
      setLoadingOrg(true);
      apiService.getUserOrganization()
        .then(org => {
          setUserOrganization(org);
        })
        .catch(error => {
          setUserOrganization(null);
        })
        .finally(() => {
          setLoadingOrg(false);
        });
    } else {
      setUserOrganization(null);
    }
  }, [user]);

  if (!user) {
    return <>{children}</>;
  }

  return (
    <div className="min-h-screen bg-background flex flex-col">
      {/* Unified Navigation Bar */}
      {showNav && (
        <nav className="border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60 relative z-[99999]">
          <div className="container mx-auto px-6 py-3">
            <div className="flex items-center justify-between">
              {/* Left side - App title and navigation */}
              <div className="flex items-center space-x-6">
                <div className="flex items-center space-x-3">
                  <Link href="/" className="hover:opacity-80 transition-opacity">
                    <img 
                      key={resolvedTheme}
                      src={resolvedTheme === "dark" 
                        ? "/SpectroCloud_Horizontal_dark-bkgd_RGB.png" 
                        : "/SpectroCloud_Horizontal_light-bkgd_RGB.png"
                      } 
                      alt="SpectroCloud" 
                      className="h-8 w-auto"
                    />
                  </Link>
                  <span className="text-sm text-muted-foreground font-medium"></span>
                </div>
                
                {/* Main navigation links */}
                <NavigationMenu>
                  <NavigationMenuList>
                    <NavigationMenuItem>
                      <NavigationMenuLink asChild>
                        <Link href="/labs" className="flex items-center gap-2 px-4 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground rounded-md">
                          <FlaskConical className="h-4 w-4" />
                          Available Labs
                        </Link>
                      </NavigationMenuLink>
                    </NavigationMenuItem>
                    
                    <NavigationMenuItem>
                      <NavigationMenuLink asChild>
                        <Link href="/" className="flex items-center gap-2 px-4 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground rounded-md">
                          <FlaskConical className="h-4 w-4" />
                          My Labs
                        </Link>
                      </NavigationMenuLink>
                    </NavigationMenuItem>
                    
                    <NavigationMenuItem>
                      <NavigationMenuLink asChild>
                        <Link href="/support" className="flex items-center gap-2 px-4 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground rounded-md">
                          <Users className="h-4 w-4" />
                          Support
                        </Link>
                      </NavigationMenuLink>
                    </NavigationMenuItem>
                    
                    
                    <NavigationMenuIndicator />
                  </NavigationMenuList>
                  <NavigationMenuViewport />
                </NavigationMenu>
              </div>

              {/* Right side - Admin Tools, User info and actions */}
              <div className="flex items-center gap-3">
                {/* Admin Tools Dropdown */}
                {isAdmin && (
                  <NavigationMenu className="admin-tools-menu">
                    <NavigationMenuList>
                      <NavigationMenuItem>
                        <NavigationMenuTrigger className="bg-[#1E7A78] text-white border-0 shadow-lg transition-all duration-300 rounded-md px-4 py-2 h-auto w-auto inline-flex items-center justify-center gap-2 font-semibold text-sm hover:shadow-xl hover:scale-105 hover:bg-[#1a6b69] focus:outline-none focus:ring-2 focus:ring-[#1E7A78] focus:ring-offset-2 data-[state=open]:bg-[#1a6b69]">
                          <Shield className="h-4 w-4 animate-pulse" />
                          <span>Admin Tools</span>
                        </NavigationMenuTrigger>
                        <NavigationMenuContent>
                          <div className="grid gap-4 p-6 w-[450px] bg-background/95 backdrop-blur-sm border border-[#1E7A78]/20 shadow-2xl rounded-md">
                            {/* Admin Dashboard */}
                            <NavigationMenuLink asChild>
                              <Link
                                href="/admin"
                                className="group block select-none space-y-2 rounded-md p-4 leading-none no-underline outline-none transition-all duration-300 bg-background/30 backdrop-blur-sm border border-[#1E7A78]/20 hover:border-[#1E7A78]/40 hover:shadow-lg hover:shadow-[#1E7A78]/10"
                              >
                                <div className="text-sm font-semibold leading-none flex items-center gap-2 text-[#1E7A78]">
                                  <Shield className="h-5 w-5 group-hover:scale-110 transition-transform duration-300" />
                                  Admin Dashboard
                                </div>
                                <p className="line-clamp-2 text-xs leading-snug text-muted-foreground">
                                  System overview, lab management, and user analytics
                                </p>
                              </Link>
                            </NavigationMenuLink>
                            
                            {/* Admin Tools Grid */}
                            <div className="grid grid-cols-2 gap-3">
                              <NavigationMenuLink asChild>
                                <Link
                                  href="/admin/organizations"
                                  className="group block select-none space-y-2 rounded-md p-4 leading-none no-underline outline-none transition-all duration-300 bg-background/30 backdrop-blur-sm border border-[#1E7A78]/20 hover:border-[#1E7A78]/40 hover:shadow-lg hover:shadow-[#1E7A78]/10"
                                >
                                  <div className="text-sm font-semibold leading-none flex items-center gap-2 text-[#1E7A78]">
                                    <Building2 className="h-5 w-5 group-hover:rotate-12 transition-transform duration-300" />
                                    Organizations
                                  </div>
                                  <p className="line-clamp-2 text-xs leading-snug text-muted-foreground">
                                    Manage organizations and create invites
                                  </p>
                                </Link>
                              </NavigationMenuLink>
                              
                              <NavigationMenuLink asChild>
                                <Link
                                  href="/admin/users"
                                  className="group block select-none space-y-2 rounded-md p-4 leading-none no-underline outline-none transition-all duration-300 bg-background/30 backdrop-blur-sm border border-[#1E7A78]/20 hover:border-[#1E7A78]/40 hover:shadow-lg hover:shadow-[#1E7A78]/10"
                                >
                                  <div className="text-sm font-semibold leading-none flex items-center gap-2 text-[#1E7A78]">
                                    <Users className="h-5 w-5 group-hover:scale-110 transition-transform duration-300" />
                                    Users
                                  </div>
                                  <p className="line-clamp-2 text-xs leading-snug text-muted-foreground">
                                    User management and role administration
                                  </p>
                                </Link>
                              </NavigationMenuLink>
                              
                              <NavigationMenuLink asChild>
                                <Link
                                  href="/admin/cleanup"
                                  className="group block select-none space-y-2 rounded-md p-4 leading-none no-underline outline-none transition-all duration-300 bg-background/30 backdrop-blur-sm border border-[#1E7A78]/20 hover:border-[#1E7A78]/40 hover:shadow-lg hover:shadow-[#1E7A78]/10"
                                >
                                  <div className="text-sm font-semibold leading-none flex items-center gap-2 text-[#1E7A78]">
                                    <Trash2 className="h-5 w-5 group-hover:rotate-12 transition-transform duration-300" />
                                    Cleanup Tools
                                  </div>
                                  <p className="line-clamp-2 text-xs leading-snug text-muted-foreground">
                                    System maintenance and cleanup tools
                                  </p>
                                </Link>
                              </NavigationMenuLink>
                            </div>
                          </div>
                        </NavigationMenuContent>
                      </NavigationMenuItem>
                      <NavigationMenuIndicator />
                    </NavigationMenuList>
                    <NavigationMenuViewport />
                  </NavigationMenu>
                )}
                
                <div className="flex items-center gap-2">
                  <Badge variant="outline" className="text-xs">
                    {user.role}
                  </Badge>
                  <span className="text-sm text-muted-foreground">
                    {user.name}
                  </span>
                </div>
                
                <ThemeToggle />
                
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button variant="ghost" size="icon">
                      <User className="h-4 w-4" />
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" className="w-64">
                    <DropdownMenuLabel>
                      <div className="flex flex-col space-y-1">
                        <p className="text-sm font-medium leading-none">{user.name}</p>
                        <p className="text-xs leading-none text-muted-foreground">
                          {user.email}
                        </p>
                        {userOrganization && (
                          <div className="flex items-center gap-1 mt-1">
                            <Building2 className="h-3 w-3 text-muted-foreground" />
                            <p className="text-xs leading-none text-muted-foreground">
                              {userOrganization.name}
                            </p>
                          </div>
                        )}
                        {loadingOrg && (
                          <p className="text-xs leading-none text-muted-foreground">
                            Loading organization...
                          </p>
                        )}
                        {!loadingOrg && !userOrganization && user && (
                          <p className="text-xs leading-none text-muted-foreground">
                            No organization
                          </p>
                        )}
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
      <main className="flex-1 relative z-10">
        {children}
      </main>
    </div>
  );
}