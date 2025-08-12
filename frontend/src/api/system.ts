import { http } from './request'

// System settings interfaces
export interface SiteSettings {
  site_name: string
  site_url: string
  contact_email: string
  support_email: string
  description: string
  logo: string
  favicon: string
  footer_text: string
  maintenance_mode: boolean
}

export interface SecuritySettings {
  // Password Policy
  min_password_length: number
  max_password_length: number
  password_expiry_days: number
  require_uppercase: boolean
  require_lowercase: boolean
  require_numbers: boolean
  require_special_chars: boolean
  password_history_count: number
  
  // Session Management
  session_timeout: number
  max_concurrent_sessions: number
  remember_me_days: number
  
  // 2FA Configuration
  enable_2fa: boolean
  require_2fa_for_admins: boolean
  allow_backup_codes: boolean
  enable_totp: boolean
  enable_sms: boolean
  enable_email: boolean
  
  // Login Security
  max_login_attempts: number
  lockout_duration: number
  enable_ip_whitelist: boolean
  enable_geolocation: boolean
  enable_device_fingerprinting: boolean
  notify_failed_logins: boolean
  notify_new_devices: boolean
  notify_password_changes: boolean
}

// System API methods
export const systemApi = {
  // Get all system settings
  getSystemSettings: (category?: string) => {
    const params = category ? `?category=${category}` : ''
    return http.get<Record<string, any>>(`/console/system/settings${params}`)
  },

  // Update system settings
  updateSystemSettings: (settings: Record<string, any>) => {
    return http.put('/console/system/settings', settings)
  },

  // Get site settings
  getSiteSettings: () => {
    return http.get<SiteSettings>('/console/system/site-settings')
  },

  // Get security settings
  getSecuritySettings: () => {
    return http.get<SecuritySettings>('/console/system/security-settings')
  }
}
