import { http } from './request'

// 分页相关类型
export interface PaginationParams {
  page: number
  page_size: number
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  total_pages: number
  current_page: number
  page_size: number
}

export interface DashboardStats {
  totalUsers: number
  totalOrganizations: number
  activeSessions: number
  totalApplications: number
}

export interface RecentActivity {
  id: string
  title: string
  description: string
  icon: string
  time: string
  user: string
  action: string
  resource: string
  status: string
  ip_address: string
  created_at: string
}

export interface SystemStatus {
  name: string
  status: 'success' | 'error' | 'warning'
  value: string
  details?: any
}

export interface DashboardData {
  stats: DashboardStats
  recentActivities: RecentActivity[]
  systemStatus: SystemStatus[]
}

export interface SystemSettings {
  site_name: string
  site_url: string
  contact_email: string
  support_email: string
  description: string
  logo_url?: string
}

export interface SecuritySettings {
  min_password_length: number
  max_password_length: number
  password_expiry_days: number
  require_uppercase: boolean
  require_lowercase: boolean
  require_numbers: boolean
  require_special_chars: boolean
  password_history_count: number
  session_timeout: number
  max_concurrent_sessions: number
  remember_me_days: number
  enable_2fa: boolean
  require_2fa_for_admins: boolean
  allow_backup_codes: boolean
  enable_totp: boolean
  enable_sms: boolean
  enable_email: boolean
  max_login_attempts: number
  lockout_duration: number
}

export interface AdministratorInfo {
  id: string
  username: string
  display_name: string
  email: string
  role: string
  role_code: string
  status: string
  created_at: string
}

export interface AdministratorListResponse extends PaginatedResponse<AdministratorInfo> {}

export interface AssignRoleRequest {
  user_id: string
  role_id: string
}

export interface RoleInfo {
  id: string
  name: string
  code: string
  description: string
  type: string
  status: string
  created_at: string
}

export interface RoleListResponse extends PaginatedResponse<RoleInfo> {}

// System API methods
export const systemApi = {
  // Get dashboard data
  getDashboardData: () => {
    return http.get<DashboardData>('/console/dashboard')
  },

  // Get system health
  getSystemHealth: () => {
    return http.get('/health')
  },

  // Get system statistics
  getSystemStats: () => {
    return http.get<DashboardStats>('/console/dashboard/stats')
  },

  // Get recent activities
  getRecentActivities: (limit: number = 10) => {
    return http.get<RecentActivity[]>(`/console/dashboard/activities?limit=${limit}`)
  },

  // Get top login users
  getTopLoginUsers: () => {
    return http.get<any[]>('/console/dashboard/top-users')
  },

  // Get top login applications
  getTopLoginApplications: () => {
    return http.get<any[]>('/console/dashboard/top-applications')
  },

  // Get administrators
  getAdministrators: (params: PaginationParams) => {
    const queryParams = new URLSearchParams()
    queryParams.append('page', params.page.toString())
    queryParams.append('page_size', params.page_size.toString())

    return http.get<AdministratorListResponse>(`/console/administrators?${queryParams.toString()}`)
  },

  // Assign administrator role
  assignAdministratorRole: (data: AssignRoleRequest) => {
    return http.post('/console/administrators/assign', data)
  },

  // Remove administrator role
  removeAdministratorRole: (userID: string, roleID: string) => {
    return http.delete(`/console/administrators/${userID}/${roleID}`)
  },

  // Get roles
  getRoles: (params: PaginationParams & {
    search?: string
    status?: string
    type?: string
  }) => {
    const queryParams = new URLSearchParams()
    queryParams.append('page', params.page.toString())
    queryParams.append('page_size', params.page_size.toString())

    if (params.search) queryParams.append('search', params.search)
    if (params.status) queryParams.append('status', params.status)
    if (params.type) queryParams.append('type', params.type)

    return http.get<RoleListResponse>(`/console/roles?${queryParams.toString()}`)
  },

  // Create role
  createRole: (data: any) => {
    return http.post('/console/roles', data)
  },

  // Update role
  updateRole: (id: string, data: any) => {
    return http.put(`/console/roles/${id}`, data)
  },

  // Delete role
  deleteRole: (id: string) => {
    return http.delete(`/console/roles/${id}`)
  },

  // Get system settings
  getSystemSettings: () => {
    return http.get<SystemSettings>('/console/system/settings')
  },

  // Update system settings
  updateSystemSettings: (data: Partial<SystemSettings>) => {
    return http.put<SystemSettings>('/console/system/settings', data)
  },

  // Get site settings (authenticated)
  getSiteSettings: () => {
    return http.get<SystemSettings>('/console/system/site-settings')
  },

  // Update site settings
  updateSiteSettings: (data: Partial<SystemSettings>) => {
    return http.put<SystemSettings>('/console/system/site-settings', data)
  },

  // Get public site info (no authentication required)
  getPublicSiteInfo: () => {
    return http.get<{site_name: string, logo_url: string}>('/public/site-info')
  },

  // Get security settings
  getSecuritySettings: () => {
    return http.get<SecuritySettings>('/console/system/security-settings')
  },

  // Update security settings
  updateSecuritySettings: (data: Partial<SecuritySettings>) => {
    return http.put<SecuritySettings>('/console/system/security-settings', data)
  },

  // Upload logo
  uploadLogo: (formData: FormData) => {
    return http.post('/console/system/upload-logo', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      }
    })
  }
}

// Legacy exports for backward compatibility
export const getDashboardData = systemApi.getDashboardData
export const getSystemHealth = systemApi.getSystemHealth
export const getSystemStats = systemApi.getSystemStats
export const getRecentActivities = systemApi.getRecentActivities
