// API Response types
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
  trade_id?: string
}

// User types
export interface User {
  id: string
  username: string
  email: string
  display_name: string
  phone?: string
  avatar?: string
  gender: number
  birthday?: string
  status: number
  last_login_at?: string
  last_login_ip?: string
  login_count: number
  failed_count: number
  locked_until?: string
  email_verified: boolean
  phone_verified: boolean
  enable_otp: boolean
  organization_id?: string
  organization?: Organization
  created_at: string
  updated_at: string
}

export interface UserProfile {
  id: string
  user_id: string
  key: string
  value: string
  description?: string
}

// Organization types
export interface Organization {
  id: string
  name: string
  code: string
  type: number
  type_name?: string
  parent_id?: string
  level: number
  path: string
  sort: number
  description?: string
  manager?: string        // 向后兼容，显示名称
  manager_id?: string     // 新增：manager的用户ID
  manager_name?: string   // 新增：manager的显示名称
  location?: string
  phone?: string
  email?: string
  status: number
  status_name?: string
  created_at: string
  updated_at: string
}

// Role types
export interface Role {
  id: string
  name: string
  code: string
  description?: string
  type: string
  is_system: boolean
  scope: string
  scope_id?: string
  status: number
  created_at: string
  updated_at: string
}

// Permission types
export interface Permission {
  id: string
  name: string
  code: string
  resource: string
  action: string
  description?: string
  category: string
  application_id?: string
  is_system: boolean
  status: number
  created_at: string
  updated_at: string
}

// Application types
export interface Application {
  id: string
  name: string
  code: string
  description?: string
  type: string
  client_id: string
  client_secret: string
  redirect_uris: string[]
  scopes: string[]
  logo_url?: string
  homepage_url?: string
  status: number
  created_at: string
  updated_at: string
}

// Login types
export interface LoginRequest {
  username: string
  password: string
  otp_code?: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
  user: User
  require_otp: boolean
}

export interface RefreshTokenRequest {
  refresh_token: string
}

export interface RefreshTokenResponse {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
}

// Pagination types
export interface PaginationParams {
  page: number
  page_size: number
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}
