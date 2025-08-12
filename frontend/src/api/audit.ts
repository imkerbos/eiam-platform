import { http } from './request'
import type { PaginationParams, PaginatedResponse } from '@/types/api'

// Audit interfaces
export interface AuditLog {
  id: string
  user_id: string
  username: string
  action: string
  resource: string
  resource_id: string
  details: string
  ip_address: string
  user_agent: string
  created_at: string
}

export interface LoginLog {
  id: string
  user_id: string
  username: string
  login_type: string
  login_ip: string
  user_agent: string
  device_type: string
  location: string
  success: boolean
  fail_reason: string
  duration: number
  created_at: string
}

export interface UserSession {
  id: string
  user_id: string
  username: string
  session_id: string
  login_ip: string
  user_agent: string
  device_type: string
  location: string
  login_time: string
  last_activity: string
  expires_at: string
  is_active: boolean
}

export interface AuditListResponse extends PaginatedResponse<AuditLog> {}
export interface LoginLogListResponse extends PaginatedResponse<LoginLog> {}
export interface UserSessionListResponse extends PaginatedResponse<UserSession> {}

// Audit API methods
export const auditApi = {
  // Get operation audit logs
  getOperationLogs: (params: PaginationParams & {
    user_id?: string
    action?: string
    resource?: string
    start_date?: string
    end_date?: string
  }) => {
    const queryParams = new URLSearchParams()
    queryParams.append('page', params.page.toString())
    queryParams.append('page_size', params.page_size.toString())
    
    if (params.user_id) queryParams.append('user_id', params.user_id)
    if (params.action) queryParams.append('action', params.action)
    if (params.resource) queryParams.append('resource', params.resource)
    if (params.start_date) queryParams.append('start_date', params.start_date)
    if (params.end_date) queryParams.append('end_date', params.end_date)
    
    return http.get<AuditListResponse>(`/console/logs/audit?${queryParams.toString()}`)
  },

  // Get login audit logs
  getLoginLogs: (params: PaginationParams & {
    user_id?: string
    success?: boolean
    login_type?: string
    start_date?: string
    end_date?: string
  }) => {
    const queryParams = new URLSearchParams()
    queryParams.append('page', params.page.toString())
    queryParams.append('page_size', params.page_size.toString())
    
    if (params.user_id) queryParams.append('user_id', params.user_id)
    if (params.success !== undefined) queryParams.append('success', params.success.toString())
    if (params.login_type) queryParams.append('login_type', params.login_type)
    if (params.start_date) queryParams.append('start_date', params.start_date)
    if (params.end_date) queryParams.append('end_date', params.end_date)
    
    return http.get<LoginLogListResponse>(`/console/logs/login?${queryParams.toString()}`)
  },

  // Get online user sessions
  getOnlineSessions: (params: PaginationParams & {
    user_id?: string
    is_active?: boolean
  }) => {
    const queryParams = new URLSearchParams()
    queryParams.append('page', params.page.toString())
    queryParams.append('page_size', params.page_size.toString())
    
    if (params.user_id) queryParams.append('user_id', params.user_id)
    if (params.is_active !== undefined) queryParams.append('is_active', params.is_active.toString())
    
    return http.get<UserSessionListResponse>(`/console/sessions?${queryParams.toString()}`)
  },

  // Terminate user session
  terminateSession: (sessionId: string) => {
    return http.delete(`/console/sessions/${sessionId}`)
  },

  // Terminate all sessions for a user
  terminateAllUserSessions: (userId: string) => {
    return http.delete(`/console/sessions/user/${userId}`)
  }
}
