import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { message } from 'ant-design-vue'
import type { ApiResponse } from '@/types/api'

// Create axios instance
const request: AxiosInstance = axios.create({
  baseURL: '/api/v1',
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

    // Add authorization token
    const token = localStorage.getItem('access_token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const { data } = response
    
    // Handle successful response
    if (data.code === 200 || data.code === 201) {
      return data.data || data
    }
    
    // Handle business error - don't show message here, let the component handle it
    const errorMessage = data.message || 'Request failed'
    return Promise.reject(new Error(errorMessage))
  },
  (error) => {
    // Handle HTTP error
    if (error.response) {
      const { status, data } = error.response
      
      switch (status) {
        case 401:
          // Don't show message here, let the login component handle it
          break
        case 403:
          message.error('Forbidden, insufficient permissions')
          break
        case 404:
          message.error('Resource not found')
          break
        case 500:
          message.error('Internal server error')
          break
        default:
          message.error(data?.message || 'Request failed')
      }
    } else if (error.request) {
      message.error('Network error, please check your connection')
    } else {
      message.error('Request configuration error')
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
