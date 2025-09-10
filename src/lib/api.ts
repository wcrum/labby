// API service for communicating with the backend

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export type UserRole = 'user' | 'admin';

export interface User {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  organization_id?: string;
  created_at: string;
  updated_at: string;
}

export interface UserWithOrganization {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  organization?: Organization;
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
  used_services?: ServiceTemplate[];
}

export interface LoginRequest {
  email: string;
  invite_code?: string;
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
  type?: string; // Optional since it's enriched from ServiceConfig
  description: string;
  service_id: string; // Required for enriched services
  logo?: string; // Optional since it's enriched from ServiceConfig
  config?: Record<string, string>; // Optional for backward compatibility
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

export interface ServiceConfig {
  id: string;
  name: string;
  type: string;
  description: string;
  config: Record<string, string>;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ServiceLimit {
  id: string;
  service_id: string;
  max_labs: number;
  max_duration: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface ServiceUsage {
  service_id: string;
  active_labs: number;
  limit: number;
}

export interface Organization {
  id: string;
  name: string;
  description: string;
  domain: string;
  created_at: string;
  updated_at: string;
}

export interface Invite {
  id: string;
  organization_id: string;
  email: string;
  invited_by: string;
  role: string;
  status: string;
  expires_at: string;
  created_at: string;
  accepted_at?: string;
  usage_limit?: number;
  usage_count: number;
  last_used_at?: string;
  used_by: string[];
}

export interface CreateInviteRequest {
  email: string;
  role: string;
  usage_limit?: number;
}

export interface InviteUsageStats {
  invite_id: string;
  email: string;
  organization_name: string;
  usage_count: number;
  usage_limit?: number;
  last_used_at?: string;
  status: string;
  created_at: string;
  expires_at: string;
  used_by: UserInfo[];
}

export interface UserInfo {
  id: string;
  email: string;
  name: string;
  used_at: string;
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

      // Handle 204 No Content responses (common for DELETE operations)
      if (response.status === 204) {
        return {} as T;
      }

      return await response.json();
    } catch (error) {
      console.error('API request failed:', error);
      throw error;
    }
  }

  // Authentication
  async login(email: string, inviteCode?: string): Promise<LoginResponse> {
    const requestBody: LoginRequest = { email };
    if (inviteCode) {
      requestBody.invite_code = inviteCode;
    }
    
    const response = await this.request<LoginResponse>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify(requestBody),
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

  async cleanupFailedLab(labId: string): Promise<void> {
    await this.request(`/api/labs/${labId}/cleanup`, {
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

  async cleanupTerraformCloud(labId: string): Promise<void> {
    await this.request(`/api/admin/terraform-cloud/cleanup`, {
      method: 'POST',
      body: JSON.stringify({ lab_id: labId }),
    });
  }

  async cleanupGuacamole(labId: string): Promise<void> {
    await this.request(`/api/admin/guacamole/cleanup`, {
      method: 'POST',
      body: JSON.stringify({ lab_id: labId }),
    });
  }

  // New flexible cleanup API methods
  async getAvailableServices(): Promise<{
    available_services: Array<{
      type: string;
      configs: Array<{
        id: string;
        name: string;
        description: string;
        is_active: boolean;
      }>;
      parameters: Array<{
        name: string;
        description: string;
        required: boolean;
        example?: string;
      }>;
    }>;
    usage: {
      endpoint: string;
      method: string;
      required_fields: string[];
      optional_fields: string[];
      example: Record<string, unknown>;
    };
  }> {
    return this.request('/api/admin/cleanup/services', {
      method: 'GET',
    });
  }

  async cleanupService(request: {
    service_type: string;
    service_config_id?: string;
    lab_id?: string;
    parameters?: Record<string, string>;
  }): Promise<{
    message: string;
    service_type: string;
    lab_id?: string;
    parameters?: Record<string, string>;
  }> {
    return this.request('/api/admin/cleanup/service', {
      method: 'POST',
      body: JSON.stringify(request),
    });
  }

  // Simplified cleanup by lab UUID only - auto-constructs all resource names
  async cleanupByLab(labId: string): Promise<{
    message: string;
    lab_id: string;
    results: Record<string, string>;
    errors: Record<string, string>;
    successful: number;
    failed: number;
  }> {
    return this.request('/api/admin/cleanup/lab', {
      method: 'POST',
      body: JSON.stringify({ lab_id: labId }),
    });
  }

  // Service-specific cleanup by service config ID and lab UUID only
  async cleanupServiceByID(serviceConfigId: string, labId: string): Promise<{
    message: string;
    service_config_id: string;
    service_type: string;
    lab_id: string;
    auto_constructed_resources: Record<string, string>;
  }> {
    return this.request('/api/admin/cleanup/service-by-id', {
      method: 'POST',
      body: JSON.stringify({ 
        service_config_id: serviceConfigId,
        lab_id: labId 
      }),
    });
  }



  // Admin endpoints
  async getAllLabs(): Promise<LabResponse[]> {
    return this.request<LabResponse[]>('/api/admin/labs');
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

  // Service management
  async getServiceConfigs(): Promise<ServiceConfig[]> {
    return this.request<ServiceConfig[]>('/api/admin/service-configs');
  }

  async createServiceConfig(config: Omit<ServiceConfig, 'id' | 'created_at' | 'updated_at'>): Promise<ServiceConfig> {
    return this.request<ServiceConfig>('/api/admin/service-configs', {
      method: 'POST',
      body: JSON.stringify(config),
    });
  }

  async updateServiceConfig(id: string, config: Partial<ServiceConfig>): Promise<ServiceConfig> {
    return this.request<ServiceConfig>(`/api/admin/service-configs/${id}`, {
      method: 'PUT',
      body: JSON.stringify(config),
    });
  }

  async deleteServiceConfig(id: string): Promise<void> {
    await this.request(`/api/admin/service-configs/${id}`, {
      method: 'DELETE',
    });
  }

  async getServiceLimits(): Promise<ServiceLimit[]> {
    return this.request<ServiceLimit[]>('/api/admin/service-limits');
  }

  async createServiceLimit(limit: Omit<ServiceLimit, 'id' | 'created_at' | 'updated_at'>): Promise<ServiceLimit> {
    return this.request<ServiceLimit>('/api/admin/service-limits', {
      method: 'POST',
      body: JSON.stringify(limit),
    });
  }

  async updateServiceLimit(id: string, limit: Partial<ServiceLimit>): Promise<ServiceLimit> {
    return this.request<ServiceLimit>(`/api/admin/service-limits/${id}`, {
      method: 'PUT',
      body: JSON.stringify(limit),
    });
  }

  async deleteServiceLimit(id: string): Promise<void> {
    await this.request(`/api/admin/service-limits/${id}`, {
      method: 'DELETE',
    });
  }

  async getServiceUsage(): Promise<ServiceUsage[]> {
    return this.request<ServiceUsage[]>('/api/admin/service-usage');
  }

  // Get current user (validate token)
  async getCurrentUser(): Promise<User> {
    return this.request<User>('/api/auth/me');
  }

  // Health check
  async healthCheck(): Promise<{ status: string; message: string }> {
    return this.request<{ status: string; message: string }>('/health');
  }

  // Organization management
  async getOrganizations(): Promise<Organization[]> {
    return this.request<Organization[]>('/api/admin/organizations');
  }

  async createOrganization(data: { name: string; description?: string; domain?: string }): Promise<Organization> {
    return this.request<Organization>('/api/admin/organizations', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async createInvite(organizationId: string, data: CreateInviteRequest): Promise<Invite> {
    return this.request<Invite>(`/api/admin/organizations/${organizationId}/invites`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  // Get all invites across all organizations (admin)
  async getAllInvites(): Promise<Invite[]> {
    return this.request<Invite[]>('/api/admin/invites');
  }

  // Get invite usage statistics (admin)
  async getInviteUsageStats(): Promise<InviteUsageStats[]> {
    return this.request<InviteUsageStats[]>('/api/admin/invites/usage');
  }

  // Get invites for a specific organization (admin)
  async getOrganizationInvites(organizationId: string): Promise<Invite[]> {
    return this.request<Invite[]>(`/api/admin/organizations/${organizationId}/invites`);
  }

  // Get user's organization
  async getUserOrganization(): Promise<Organization | null> {
    try {
      return await this.request<Organization>('/api/user/organization');
    } catch (error) {
      // If user has no organization (404), return null
      // Other errors should be logged but still return null
      return null;
    }
  }

  // Get all users (admin only)
  async getAllUsers(): Promise<UserWithOrganization[]> {
    return this.request<UserWithOrganization[]>('/api/admin/users');
  }
}

export const apiService = new ApiService();
