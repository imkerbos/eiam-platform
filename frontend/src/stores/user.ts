import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types/api'
import { login as loginAPI, logout as logoutAPI, refreshToken as refreshTokenAPI, getCurrentUser as getCurrentUserAPI } from '@/api/auth'
import { encryptPassword } from '@/utils/crypto'

export const useUserStore = defineStore('user', () => {
  // State
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('access_token'))
  const refreshToken = ref<string | null>(localStorage.getItem('refresh_token'))

  // Getters
  const isLoggedIn = computed(() => !!token.value && !!user.value)
  const currentUser = computed(() => user.value)

  // Actions
  const setUser = (userData: User) => {
    user.value = userData
  }

  const setToken = (accessToken: string, refreshTokenValue: string) => {
    token.value = accessToken
    refreshToken.value = refreshTokenValue
    localStorage.setItem('access_token', accessToken)
    localStorage.setItem('refresh_token', refreshTokenValue)
  }

  const clearAuth = () => {
    user.value = null
    token.value = null
    refreshToken.value = null
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }

  const login = async (username: string, password: string, otpCode?: string) => {
    try {
      // 加密密码
      const encryptedPassword = encryptPassword(password)
      
      const response = await loginAPI({
        username,
        password: encryptedPassword,
        otp_code: otpCode
      })
      
      // response is already the data from the API (due to interceptor)
      const { access_token, refresh_token, user: userData } = response
      setToken(access_token, refresh_token)
      setUser(userData)
      
      return response
    } catch (error) {
      throw error
    }
  }

  const logout = async () => {
    try {
      if (token.value) {
        await logoutAPI()
      }
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      clearAuth()
    }
  }

  const refreshTokenAction = async () => {
    try {
      if (!refreshToken.value) {
        throw new Error('No refresh token')
      }
      
      const response = await refreshTokenAPI({
        refresh_token: refreshToken.value
      })
      
      const { access_token, refresh_token } = response
      setToken(access_token, refresh_token)
      
      return response
    } catch (error) {
      clearAuth()
      throw error
    }
  }

  const getCurrentUser = async () => {
    try {
      const response = await getCurrentUserAPI()
      setUser(response.data)
      return response
    } catch (error) {
      throw error
    }
  }

  return {
    // State
    user,
    token,
    refreshToken,
    
    // Getters
    isLoggedIn,
    currentUser,
    
    // Actions
    setUser,
    setToken,
    clearAuth,
    login,
    logout,
    refreshTokenAction,
    getCurrentUser
  }
})
