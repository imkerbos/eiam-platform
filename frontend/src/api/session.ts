import { http } from './request'
import type { ApiResponse } from '@/types/api'

// 会话信息类型
export interface SessionInfo {
  session_id: string
  user_id: string
  username: string
  email: string
  display_name: string
  login_ip: string
  user_agent: string
  login_time: string
  last_activity: string
  expires_at: string
  token_id: string
}

// 获取用户会话列表
export const getUserSessions = (userId: string): Promise<ApiResponse<SessionInfo[]>> => {
  return http.get<ApiResponse<SessionInfo[]>>(`/console/sessions/users/${userId}`)
}

// 强制用户下线
export const forceLogoutUser = (userId: string): Promise<ApiResponse<null>> => {
  return http.delete<ApiResponse<null>>(`/console/sessions/users/${userId}`)
}
