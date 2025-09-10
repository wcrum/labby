"use client";

import React, { createContext, useContext, useEffect, useState } from 'react';
import { apiService, User } from './api';

interface AuthContextType {
  user: User | null;
  isLoading: boolean;
  login: (email: string, inviteCode?: string) => Promise<void>;
  oidcLogin: (inviteCode?: string) => void;
  logout: () => void;
  refreshUser: () => Promise<void>;
  isAdmin: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Check if user is already logged in on mount
    const token = apiService.getToken();
    if (token) {
      // Validate the token and restore user data
      restoreUserSession();
    } else {
      setIsLoading(false);
    }
  }, []);

  const restoreUserSession = async () => {
    try {
      const userData = await apiService.getCurrentUser();
      setUser(userData);
    } catch (error) {
      console.error('Failed to restore user session:', error);
      // Token is invalid, clear it
      apiService.clearToken();
    } finally {
      setIsLoading(false);
    }
  };

  const login = async (email: string, inviteCode?: string) => {
    try {
      setIsLoading(true);
      const response = await apiService.login(email, inviteCode);
      setUser(response.user);
    } catch (error) {
      console.error('Login failed:', error);
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const oidcLogin = (inviteCode?: string) => {
    const loginURL = apiService.getOIDCLoginURL(inviteCode);
    window.location.href = loginURL;
  };

  const logout = () => {
    apiService.clearToken();
    setUser(null);
  };

  const refreshUser = async () => {
    try {
      const userData = await apiService.getCurrentUser();
      setUser(userData);
    } catch (error) {
      console.error('Failed to refresh user data:', error);
      throw error;
    }
  };

  const isAdmin = user?.role === 'admin';

  return (
    <AuthContext.Provider value={{ user, isLoading, login, oidcLogin, logout, refreshUser, isAdmin }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
