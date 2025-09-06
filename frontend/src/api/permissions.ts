import { http } from './request'
import type { PaginatedResponse } from '@/types/api'

// 权限相关类型定义
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
  status: string
  created_at: string
  updated_at: string
}

export interface CreatePermissionRequest {
  name: string
  code: string
  resource: string
  action: string
  description?: string
  category: string
  application_id?: string
  is_system?: boolean
}

export interface UpdatePermissionRequest {
  name?: string
  resource?: string
  action?: string
  description?: string
  category?: string
  application_id?: string
  status?: string
}

export interface Role {
  id: string
  name: string
  code: string
  description?: string
  type: string
  is_system: boolean
  scope: string
  status: string
  created_at: string
  updated_at: string
}

export interface CreateRoleRequest {
  name: string
  code: string
  description?: string
  type?: string
  scope?: string
}

export interface UpdateRoleRequest {
  name?: string
  description?: string
  type?: string
  scope?: string
  status?: string
}

export interface RoleAssignment {
  id: string
  user_id: string
  username: string
  display_name: string
  email: string
  role_id: string
  role_name: string
  role_code: string
  status: string
  assigned_at: string
}

export interface AssignRoleRequest {
  user_id: string
  role_id: string
}

// 权限管理API
export const permissionApi = {
  // 获取权限列表
  getPermissions: (params?: {
    page?: number
    page_size?: number
    search?: string
    status?: string
    category?: string
    resource?: string
  }) => {
    return http.get<PaginatedResponse<Permission>>('/console/permissions', { params })
  },

  // 创建权限
  createPermission: (data: CreatePermissionRequest) => {
    return http.post<Permission>('/console/permissions', data)
  },

  // 更新权限
  updatePermission: (id: string, data: UpdatePermissionRequest) => {
    return http.put<Permission>(`/console/permissions/${id}`, data)
  },

  // 删除权限
  deletePermission: (id: string) => {
    return http.delete(`/console/permissions/${id}`)
  },

  // 获取角色列表
  getRoles: (params?: {
    page?: number
    page_size?: number
    search?: string
    status?: string
    type?: string
  }) => {
    return http.get<PaginatedResponse<Role>>('/console/roles', { params })
  },

  // 创建角色
  createRole: (data: CreateRoleRequest) => {
    return http.post<Role>('/console/roles', data)
  },

  // 更新角色
  updateRole: (id: string, data: UpdateRoleRequest) => {
    return http.put<Role>(`/console/roles/${id}`, data)
  },

  // 删除角色
  deleteRole: (id: string) => {
    return http.delete(`/console/roles/${id}`)
  },

  // 获取角色分配列表
  getRoleAssignments: (params?: {
    page?: number
    page_size?: number
    search?: string
    role_id?: string
    user_id?: string
  }) => {
    return http.get<PaginatedResponse<RoleAssignment>>('/console/role-assignments', { params })
  },

  // 分配角色给用户
  assignRole: (data: AssignRoleRequest) => {
    return http.post('/console/role-assignments', data)
  },

  // 移除用户角色
  removeRole: (userId: string, roleId: string) => {
    return http.delete(`/console/role-assignments/${userId}/${roleId}`)
  }
}
