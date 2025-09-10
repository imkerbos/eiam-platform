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


// 权限路由相关类型定义
export interface PermissionRoute {
  id: string
  name: string
  code: string
  description?: string
  applications: string[]
  application_groups: string[]
  status: string
  created_at: string
  updated_at: string
}

export interface CreatePermissionRouteRequest {
  name: string
  code: string
  description?: string
  applications: string[]
  application_groups: string[]
  status?: string
}

export interface UpdatePermissionRouteRequest {
  name?: string
  description?: string
  applications?: string[]
  application_groups?: string[]
  status?: string
}

export interface PermissionRouteAssignment {
  id: string
  permission_route_id: string
  permission_name: string
  permission_code: string
  assignee_type: string
  assignee_id: string
  assignee_name: string
  status: string
  assigned_at: string
}

export interface AssignPermissionRouteRequest {
  permission_route_id: string
  assignee_type: string
  assignee_id: string
  status?: string
}

// 权限管理API
export const permissionApi = {
  // 获取权限列表（传统权限）
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

  // 创建权限（传统权限）
  createPermission: (data: CreatePermissionRequest) => {
    return http.post<Permission>('/console/permissions', data)
  },

  // 更新权限（传统权限）
  updatePermission: (id: string, data: UpdatePermissionRequest) => {
    return http.put<Permission>(`/console/permissions/${id}`, data)
  },

  // 删除权限（传统权限）
  deletePermission: (id: string) => {
    return http.delete(`/console/permissions/${id}`)
  },

  // 获取权限路由列表
  getPermissionRoutes: (params?: {
    page?: number
    page_size?: number
    search?: string
    status?: string
  }) => {
    return http.get<PaginatedResponse<PermissionRoute>>('/console/permission-routes', { params })
  },

  // 创建权限路由
  createPermissionRoute: (data: CreatePermissionRouteRequest) => {
    return http.post<PermissionRoute>('/console/permission-routes', data)
  },

  // 更新权限路由
  updatePermissionRoute: (id: string, data: UpdatePermissionRouteRequest) => {
    return http.put<PermissionRoute>(`/console/permission-routes/${id}`, data)
  },

  // 删除权限路由
  deletePermissionRoute: (id: string) => {
    return http.delete(`/console/permission-routes/${id}`)
  },

  // 获取权限路由分配列表
  getPermissionRouteAssignments: (params?: {
    page?: number
    page_size?: number
  }) => {
    return http.get<PaginatedResponse<PermissionRouteAssignment>>('/console/permission-route-assignments', { params })
  },

  // 分配权限路由
  assignPermissionRoute: (data: AssignPermissionRouteRequest) => {
    return http.post('/console/permission-route-assignments', data)
  },

  // 移除权限路由分配
  removePermissionRouteAssignment: (assigneeType: string, assigneeId: string, permissionRouteId: string) => {
    return http.delete(`/console/permission-route-assignments/${assigneeType}/${assigneeId}/${permissionRouteId}`)
  },

}
