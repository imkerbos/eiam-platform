import { http } from './request'
import type { PaginationParams, PaginatedResponse } from '@/types/api'

// Application types
export interface Application {
  id: string
  name: string
  description: string
  type: string
  status: string
  url: string
  logo?: string
  group?: string
  group_color?: string
}

export interface ApplicationGroup {
  id: string
  name: string
  description: string
  color: string
  applications: Application[]
}

export interface ApplicationListResponse extends PaginatedResponse<Application> {}
export interface ApplicationGroupListResponse extends PaginatedResponse<ApplicationGroup> {}

// Application API methods
export const applicationApi = {
  // Get user applications (for portal)
  getUserApplications: () => {
    return http.get<ApplicationGroup[]>('/portal/applications')
  },

  // Get all applications (for console)
  getApplications: (params: PaginationParams & {
    search?: string
    type?: string
    status?: string
    group_id?: string
  }) => {
    const queryParams = new URLSearchParams()
    queryParams.append('page', params.page.toString())
    queryParams.append('page_size', params.page_size.toString())
    
    if (params.search) queryParams.append('search', params.search)
    if (params.type) queryParams.append('type', params.type)
    if (params.status) queryParams.append('status', params.status)
    if (params.group_id) queryParams.append('group_id', params.group_id)
    
    return http.get<ApplicationListResponse>(`/console/applications?${queryParams.toString()}`)
  },

  // Get application groups
  getApplicationGroups: (params: PaginationParams) => {
    const queryParams = new URLSearchParams()
    queryParams.append('page', params.page.toString())
    queryParams.append('page_size', params.page_size.toString())
    
    return http.get<ApplicationGroupListResponse>(`/console/application-groups?${queryParams.toString()}`)
  }
}
