import { http } from './request'

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
