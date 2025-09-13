import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { message } from 'ant-design-vue'
import type { ApiResponse } from '@/types/api'
import router from '@/router'
import { TokenManager } from '@/utils/storage'

// Create axios instance
const request: AxiosInstance = axios.create({
  baseURL: '/api/v1', // 使用代理访问后端
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// Request interceptor
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // Add trade ID
    const tradeId = `frontend_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    config.headers['X-Trade-ID'] = tradeId

    // Add authorization token - 使用安全存储获取token
    const token = TokenManager.getAccessToken()
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Track if we're currently refreshing token to avoid multiple simultaneous refresh attempts
let isRefreshing = false
let failedQueue: Array<{ resolve: (value?: any) => void; reject: (reason?: any) => void }> = []

const processQueue = (error: any, token?: string) => {
  failedQueue.forEach(({ resolve, reject }) => {
    if (error) {
      reject(error)
    } else {
      resolve(token)
    }
  })
  
  failedQueue = []
}

// Response interceptor with automatic token refresh
request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const { data } = response
    
    // Handle successful response
    if (data.code === 200 || data.code === 201) {
      return data.data || data
    }
    
    // Handle business error - preserve error details for component handling
    const error = new Error(data.message || 'Request failed')
    // Attach additional error information
    ;(error as any).code = data.code
    ;(error as any).data = data.data
    return Promise.reject(error)
  },
  async (error) => {
    const originalRequest = error.config
    
    // Handle HTTP error
    if (error.response) {
      const { status, data } = error.response
      
      if (status === 401 && !originalRequest._retry) {
        if (isRefreshing) {
          // If we're already refreshing, queue this request
          return new Promise((resolve, reject) => {
            failedQueue.push({ resolve, reject })
          }).then(() => {
            // Retry the original request with the new token - 使用安全存储获取token
            const token = TokenManager.getAccessToken()
            if (token) {
              originalRequest.headers.Authorization = `Bearer ${token}`
            }
            return request(originalRequest)
          }).catch(err => {
            return Promise.reject(err)
          })
        }

        originalRequest._retry = true
        isRefreshing = true

        const refreshToken = TokenManager.getRefreshToken()
        
        if (refreshToken) {
          try {
            // Attempt to refresh the token
            const response = await axios.post('/api/v1/console/auth/refresh', {
              refresh_token: refreshToken
            })
            
            const { access_token, refresh_token: newRefreshToken } = response.data.data
            
            // Update stored tokens - 使用安全存储
            TokenManager.setTokens(access_token, newRefreshToken)
            
            // Update request header
            originalRequest.headers.Authorization = `Bearer ${access_token}`
            
            // Process queued requests
            processQueue(null, access_token)
            
            isRefreshing = false
            
            // Retry the original request
            return request(originalRequest)
            
          } catch (refreshError) {
            // Refresh failed, clear tokens and redirect to login
            processQueue(refreshError, undefined)
            isRefreshing = false
            
            // 使用安全存储清除token
            TokenManager.clearTokens()
            
            // Only redirect if not already on login page
            if (router.currentRoute.value.path !== '/login') {
              message.error('登录已过期，请重新登录')
              router.push('/login')
            }
            
            return Promise.reject(refreshError)
          }
        } else {
          // No refresh token, redirect to login
          isRefreshing = false
          
          if (router.currentRoute.value.path !== '/login') {
            message.error('请先登录')
            router.push('/login')
          }
          
          return Promise.reject(error)
        }
      }
      
      // Handle other HTTP status codes
      switch (status) {
        case 403:
          // 权限不足时，直接清除token并跳转到登录页
          // 这通常表示token中的角色信息已损坏
          TokenManager.clearTokens()
          message.error('登录已过期，请重新登录')
          router.push('/login')
          break
        case 404:
          message.error('请求的资源不存在')
          break
        case 500:
          message.error('服务器内部错误，请稍后重试')
          break
        default:
          message.error(data?.message || '请求失败')
      }
    } else if (error.request) {
      message.error('网络错误，请检查网络连接')
    } else {
      message.error('请求配置错误')
    }
    
    return Promise.reject(error)
  }
)

// Request methods
export const http = {
  get: <T = any>(url: string, config?: AxiosRequestConfig): Promise<T> => {
    return request.get(url, config)
  },
  
  post: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> => {
    return request.post(url, data, config)
  },
  
  put: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> => {
    return request.put(url, data, config)
  },
  
  delete: <T = any>(url: string, config?: AxiosRequestConfig): Promise<T> => {
    return request.delete(url, config)
  },
  
  patch: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> => {
    return request.patch(url, data, config)
  }
}

export default request
