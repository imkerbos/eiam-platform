import { http } from './request'
import type { LoginRequest, LoginResponse, RefreshTokenRequest, RefreshTokenResponse } from '@/types/api'

// Login API
export const login = (data: LoginRequest): Promise<LoginResponse> => {
  return http.post<LoginResponse>('/auth/login', data)
}

// Refresh token API
export const refreshToken = (data: RefreshTokenRequest): Promise<RefreshTokenResponse> => {
  return http.post<RefreshTokenResponse>('/auth/refresh', data)
}

// Logout API
export const logout = (): Promise<void> => {
  return http.post('/auth/logout')
}

// Get current user info
export const getCurrentUser = (): Promise<any> => {
  return http.get('/auth/me')
}
