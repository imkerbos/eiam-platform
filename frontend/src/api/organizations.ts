import { http } from './request'
import type { Organization, PaginationParams, PaginatedResponse } from '@/types/api'

// Organization API interfaces
export interface CreateOrganizationRequest {
  name: string
  code: string
  type: number
  parent_id?: string
  description?: string
  manager?: string
  location?: string
  phone?: string
  email?: string
  status?: number
}

export interface UpdateOrganizationRequest {
  name?: string
  code?: string
  type?: number
  parent_id?: string
  description?: string
  manager?: string
  location?: string
  phone?: string
  email?: string
  status?: number
}

export interface OrganizationListResponse extends PaginatedResponse<Organization> {}

// Organization API methods
export const organizationApi = {
  // Get organization list
  getOrganizations: (params: PaginationParams & {
    search?: string
    status?: string
    type?: string
    parent_id?: string
  }) => {
    const queryParams = new URLSearchParams()
    queryParams.append('page', params.page.toString())
    queryParams.append('page_size', params.page_size.toString())
    
    if (params.search) queryParams.append('search', params.search)
    if (params.status) queryParams.append('status', params.status)
    if (params.type) queryParams.append('type', params.type)
    if (params.parent_id) queryParams.append('parent_id', params.parent_id)
    
    return http.get<OrganizationListResponse>(`/console/organizations?${queryParams.toString()}`)
  },

  // Get organization tree
  getOrganizationsTree: (params?: {
    search?: string
    status?: string
  }) => {
    const queryParams = new URLSearchParams()
    queryParams.append('tree', 'true')
    
    if (params?.search) queryParams.append('search', params.search)
    if (params?.status) queryParams.append('status', params.status)
    
    return http.get<any[]>(`/console/organizations/tree?${queryParams.toString()}`)
  },

  // Get single organization
  getOrganization: (id: string) => {
    return http.get<Organization>(`/console/organizations/${id}`)
  },

  // Create organization
  createOrganization: (data: CreateOrganizationRequest) => {
    return http.post<Organization>('/console/organizations', data)
  },

  // Update organization
  updateOrganization: (id: string, data: UpdateOrganizationRequest) => {
    return http.put(`/console/organizations/${id}`, data)
  },

  // Delete organization
  deleteOrganization: (id: string) => {
    return http.delete(`/console/organizations/${id}`)
  }
}
