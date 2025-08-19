import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types/api'
import { login as loginAPI, logout as logoutAPI, refreshToken as refreshTokenAPI, getCurrentUser as getCurrentUserAPI } from '@/api/auth'
import { TokenManager, UserInfoManager } from '@/utils/storage'
import { useSiteStore } from '@/stores/site'

export const useUserStore = defineStore('user', () => {
  // State - 从安全存储中初始化数据
  const user = ref<User | null>(UserInfoManager.getUserInfo())
  const token = ref<string | null>(TokenManager.getAccessToken())
  const refreshToken = ref<string | null>(TokenManager.getRefreshToken())
  const sessionId = ref<string | null>(null)
  
  console.log('用户store初始化:', {
    user: user.value,
    token: !!token.value,
    userRoles: user.value?.roles
  })
  
  // 调试：直接从sessionStorage读取
  const rawUserInfo = sessionStorage.getItem('user_info')
  console.log('直接从sessionStorage读取:', rawUserInfo)
  if (rawUserInfo) {
    try {
      const parsedUser = JSON.parse(rawUserInfo)
      console.log('解析后的用户信息:', parsedUser)
    } catch (e) {
      console.error('解析用户信息失败:', e)
    }
  }

  // Getters
  const isLoggedIn = computed(() => !!token.value && !!user.value)
  const currentUser = computed(() => user.value)
  const userRoles = computed(() => user.value?.roles || [])
  const isAdmin = computed(() => userRoles.value.includes('admin') || userRoles.value.includes('SYSTEM_ADMIN'))

  // Actions
  const setUser = (userData: User) => {
    console.log('设置用户信息:', userData)
    user.value = userData
    // 同步到安全存储
    UserInfoManager.setUserInfo(userData)
    console.log('用户信息已存储到安全存储')
  }

  const setToken = (accessToken: string, refreshTokenValue: string) => {
    token.value = accessToken
    refreshToken.value = refreshTokenValue
    // 使用安全存储替代直接操作localStorage
    TokenManager.setTokens(accessToken, refreshTokenValue)
  }

  const clearAuth = () => {
    user.value = null
    token.value = null
    refreshToken.value = null
    sessionId.value = null
    // 使用安全存储清除所有认证数据
    TokenManager.clearTokens()
    UserInfoManager.clearUserInfo()
  }

  const login = async (username: string, password: string, otpCode?: string) => {
    try {
      // 不加密密码，直接发送明文（HTTPS保证传输安全）
      const response = await loginAPI({
        username,
        password: password,
        otp_code: otpCode
      })
      
      // response is already the data from the API (due to interceptor)
      const { access_token, refresh_token, user: userData, session_id } = response
      setToken(access_token, refresh_token)
      setUser(userData)
      sessionId.value = session_id || null
      
      // 登录成功，使用响应中的用户信息，不需要再次获取
      console.log('登录成功，用户信息:', userData)
      
      // 登录成功后加载站点配置
      try {
        const siteStore = useSiteStore()
        await siteStore.loadSiteSettings()
      } catch (error) {
        console.warn('Failed to load site settings after login:', error)
        // 站点配置加载失败不影响登录流程
      }
      
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
      // response已经是经过interceptor处理的数据，直接使用
      setUser(response)
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
    sessionId,
    
    // Getters
    isLoggedIn,
    currentUser,
    userRoles,
    isAdmin,
    
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
