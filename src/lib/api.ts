// API service for communicating with the backend

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export type UserRole = 'user' | 'admin';

export interface User {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  created_at: string;
  updated_at: string;
}

export interface Credential {
  id: string;
  lab_id: string;
  label: string;
  username: string;
  password: string;
  url?: string;
  expires_at: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface Lab {
  id: string;
  name: string;
  status: 'provisioning' | 'ready' | 'error' | 'expired';
  owner_id: string;
  started_at: string;
  ends_at: string;
  created_at: string;
  updated_at: string;
  credentials: Credential[];
}

export interface LabResponse {
  id: string;
  name: string;
  status: 'provisioning' | 'ready' | 'error' | 'expired';
  owner: User;
  started_at: string;
  ends_at: string;
  credentials: Credential[];
}

export interface LoginRequest {
  email: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface CreateLabRequest {
  name?: string; // Optional since backend will generate UUID
  owner_id: string;
  duration: number;
}

export interface ServiceTemplate {
  name: string;
  type: string;
  description: string;
  config: Record<string, string>;
}

export interface LabTemplate {
  name: string;
  id: string;
  description: string;
  expiration_duration: string;
  owner: string;
  created_at: string;
  services: ServiceTemplate[];
}

class ApiService {
  private token: string | null = null;

  setToken(token: string) {
    this.token = token;
    if (typeof window !== 'undefined') {
      localStorage.setItem('auth_token', token);
    }
  }

  getToken(): string | null {
    if (!this.token && typeof window !== 'undefined') {
      this.token = localStorage.getItem('auth_token');
    }
    return this.token;
  }

  clearToken() {
    this.token = null;
    if (typeof window !== 'undefined') {
      localStorage.removeItem('auth_token');
    }
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {}
  ): Promise<T> {
    const url = `${API_BASE_URL}${endpoint}`;
    const token = this.getToken();

    const config: RequestInit = {
      headers: {
        'Content-Type': 'application/json',
        ...(token && { Authorization: `Bearer ${token}` }),
        ...options.headers,
      },
      ...options,
    };

    try {
      const response = await fetch(url, config);
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error('API request failed:', error);
      throw error;
    }
  }

  // Authentication
  async login(email: string): Promise<LoginResponse> {
    const response = await this.request<LoginResponse>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email }),
    });
    
    this.setToken(response.token);
    return response;
  }

  // Labs
  async createLab(data: CreateLabRequest): Promise<Lab> {
    return this.request<Lab>('/api/labs', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getLab(labId: string): Promise<LabResponse> {
    return this.request<LabResponse>(`/api/labs/${labId}`);
  }

  async getLabProgress(labId: string): Promise<{
    lab_id: string;
    overall: number;
    current_step: string;
    services: Array<{
      name: string;
      description: string;
      status: string;
      progress: number;
      steps: Array<{
        name: string;
        status: string;
        message: string;
        started_at?: string;
        completed_at?: string;
        error?: string;
      }>;
      started_at?: string;
      completed_at?: string;
      error?: string;
    }>;
    logs: string[];
    updated_at: string;
  }> {
    return this.request(`/api/labs/${labId}/progress`);
  }

  async getUserLabs(): Promise<LabResponse[]> {
    return this.request<LabResponse[]>('/api/labs');
  }

  async deleteLab(labId: string): Promise<void> {
    await this.request(`/api/labs/${labId}`, {
      method: 'DELETE',
    });
  }

  async stopLab(labId: string): Promise<void> {
    await this.request(`/api/labs/${labId}/stop`, {
      method: 'POST',
    });
  }

  async adminStopLab(labId: string): Promise<void> {
    await this.request(`/api/admin/labs/${labId}/stop`, {
      method: 'POST',
    });
  }

  async adminDeleteLab(labId: string): Promise<void> {
    await this.request(`/api/admin/labs/${labId}`, {
      method: 'DELETE',
    });
  }

  async cleanupLab(labId: string): Promise<void> {
    await this.request(`/api/admin/labs/${labId}/cleanup`, {
      method: 'POST',
    });
  }

  // Template management
  async getTemplates(): Promise<LabTemplate[]> {
    return this.request<LabTemplate[]>('/api/templates');
  }

  async getTemplate(templateId: string): Promise<LabTemplate> {
    return this.request<LabTemplate>(`/api/templates/${templateId}`);
  }

  async createLabFromTemplate(templateId: string): Promise<Lab> {
    return this.request<Lab>(`/api/templates/${templateId}/labs`, {
      method: 'POST',
    });
  }

  async cleanupPaletteProject(projectName: string): Promise<void> {
    await this.request(`/api/admin/palette-project/cleanup`, {
      method: 'POST',
      body: JSON.stringify({ project_name: projectName }),
    });
  }



  // Admin endpoints
  async getAllLabs(): Promise<LabResponse[]> {
    return this.request<LabResponse[]>('/api/admin/labs');
  }

  async getAllUsers(): Promise<User[]> {
    return this.request<User[]>('/api/admin/users');
  }

  async createUser(data: { email: string; name: string; role: UserRole }): Promise<User> {
    return this.request<User>('/api/admin/users', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateUserRole(userId: string, role: UserRole): Promise<void> {
    await this.request(`/api/admin/users/${userId}/role`, {
      method: 'PUT',
      body: JSON.stringify({ role }),
    });
  }

  async deleteUser(userId: string): Promise<void> {
    await this.request(`/api/admin/users/${userId}`, {
      method: 'DELETE',
    });
  }

  // Get current user (validate token)
  async getCurrentUser(): Promise<User> {
    return this.request<User>('/api/auth/me');
  }

  // Health check
  async healthCheck(): Promise<{ status: string; message: string }> {
    return this.request<{ status: string; message: string }>('/health');
  }
}

export const apiService = new ApiService();
