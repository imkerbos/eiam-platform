import { http } from './request'

export interface DashboardStats {
  totalUsers: number
  totalOrganizations: number
  activeSessions: number
  totalApplications: number
}

export interface RecentActivity {
  id: string
  title: string
  description: string
  icon: string
  time: string
  user: string
  action: string
  resource: string
  status: string
  ip_address: string
  created_at: string
}

export interface SystemStatus {
  name: string
  status: 'success' | 'error' | 'warning'
  value: string
  details?: any
}

export interface DashboardData {
  stats: DashboardStats
  recentActivities: RecentActivity[]
  systemStatus: SystemStatus[]
}

// Get dashboard data
export const getDashboardData = () => {
  return http.get<DashboardData>('/console/dashboard')
}

// Get system health
export const getSystemHealth = () => {
  return http.get('/health')
}

// Get system statistics
export const getSystemStats = () => {
  return http.get<DashboardStats>('/console/dashboard/stats')
}

// Get recent activities
export const getRecentActivities = (limit: number = 10) => {
  return http.get<RecentActivity[]>(`/console/dashboard/activities?limit=${limit}`)
}
