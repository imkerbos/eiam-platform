import { http } from './request'
import type { User, PaginationParams, PaginatedResponse } from '@/types/api'

// User API interfaces
export interface CreateUserRequest {
  username: string
  email: string
  display_name: string
  phone?: string
  password: string
  organization_id: string
  status?: number
  enable_otp?: boolean
  email_verified?: boolean
  phone_verified?: boolean
}

export interface UpdateUserRequest {
  display_name?: string
  phone?: string
  organization_id?: string
  status?: number
  enable_otp?: boolean
  email_verified?: boolean
  phone_verified?: boolean
}

export interface UserListResponse extends PaginatedResponse<User> {}

// User API methods
export const userApi = {
  // Get user list
  getUsers: (params: PaginationParams & {
    search?: string
    status?: string
    organization_id?: string
  }) => {
    const queryParams = new URLSearchParams()
    queryParams.append('page', params.page.toString())
    queryParams.append('page_size', params.page_size.toString())
    
    if (params.search) queryParams.append('search', params.search)
    if (params.status) queryParams.append('status', params.status)
    if (params.organization_id) queryParams.append('organization_id', params.organization_id)
    
    return http.get<UserListResponse>(`/console/users?${queryParams.toString()}`)
  },

  // Get single user
  getUser: (id: string) => {
    return http.get<User>(`/console/users/${id}`)
  },

  // Create user
  createUser: (data: CreateUserRequest) => {
    return http.post<User>('/console/users', data)
  },

  // Update user
  updateUser: (id: string, data: UpdateUserRequest) => {
    return http.put(`/console/users/${id}`, data)
  },

  // Delete user
  deleteUser: (id: string) => {
    return http.delete(`/console/users/${id}`)
  }
}
