import CryptoJS from 'crypto-js'

// 存储加密密钥（实际项目中应该从环境变量获取）
const STORAGE_KEY = import.meta.env.VITE_STORAGE_KEY || 'eiam-platform-2024'

/**
 * 存储类型枚举
 */
export enum StorageType {
  SESSION = 'session',
  LOCAL = 'local'
}

/**
 * 存储项配置
 */
interface StorageConfig {
  key: string
  type: StorageType
  encrypt?: boolean
  expireTime?: number // 过期时间（毫秒）
}

/**
 * 存储数据结构
 */
interface StorageData<T = any> {
  data: T
  timestamp: number
  expireTime?: number
}

/**
 * 安全存储工具类
 */
class SecureStorage {
  /**
   * 加密数据
   */
  private encrypt(data: string): string {
    try {
      return CryptoJS.AES.encrypt(data, STORAGE_KEY).toString()
    } catch (error) {
      console.error('数据加密失败:', error)
      return data
    }
  }

  /**
   * 解密数据
   */
  private decrypt(encryptedData: string): string {
    try {
      const bytes = CryptoJS.AES.decrypt(encryptedData, STORAGE_KEY)
      return bytes.toString(CryptoJS.enc.Utf8)
    } catch (error) {
      console.error('数据解密失败:', error)
      return encryptedData
    }
  }

  /**
   * 获取存储引擎
   */
  private getStorage(type: StorageType): Storage {
    return type === StorageType.SESSION ? sessionStorage : localStorage
  }

  /**
   * 设置数据
   */
  setItem<T>(config: StorageConfig, value: T): void {
    try {
      const storage = this.getStorage(config.type)
      const now = Date.now()
      
      const storageData: StorageData<T> = {
        data: value,
        timestamp: now,
        expireTime: config.expireTime ? now + config.expireTime : undefined
      }

      let dataString = JSON.stringify(storageData)
      
      // 如果需要加密
      if (config.encrypt) {
        dataString = this.encrypt(dataString)
      }

      storage.setItem(config.key, dataString)
    } catch (error) {
      console.error('存储数据失败:', error)
    }
  }

  /**
   * 获取数据
   */
  getItem<T>(config: StorageConfig): T | null {
    try {
      const storage = this.getStorage(config.type)
      let dataString = storage.getItem(config.key)
      
      if (!dataString) {
        return null
      }

      // 如果数据是加密的
      if (config.encrypt) {
        dataString = this.decrypt(dataString)
      }

      if (!dataString) {
        return null
      }

      const storageData: StorageData<T> = JSON.parse(dataString)
      
      // 检查是否过期
      if (storageData.expireTime && Date.now() > storageData.expireTime) {
        this.removeItem(config)
        return null
      }

      return storageData.data
    } catch (error) {
      console.error('获取存储数据失败:', error)
      return null
    }
  }

  /**
   * 移除数据
   */
  removeItem(config: StorageConfig): void {
    try {
      const storage = this.getStorage(config.type)
      storage.removeItem(config.key)
    } catch (error) {
      console.error('移除存储数据失败:', error)
    }
  }

  /**
   * 清空指定类型的存储
   */
  clear(type: StorageType): void {
    try {
      const storage = this.getStorage(type)
      storage.clear()
    } catch (error) {
      console.error('清空存储失败:', error)
    }
  }

  /**
   * 检查数据是否存在且未过期
   */
  hasValidItem(config: StorageConfig): boolean {
    return this.getItem(config) !== null
  }
}

// 创建存储实例
export const secureStorage = new SecureStorage()

/**
 * 预定义的存储配置
 */
export const StorageConfigs = {
  // Access Token - 使用sessionStorage，页面关闭后自动清除，更安全
  ACCESS_TOKEN: {
    key: 'access_token',
    type: StorageType.SESSION,
    encrypt: true,
    expireTime: 24 * 60 * 60 * 1000 // 24小时
  } as StorageConfig,

  // Refresh Token - 使用localStorage，但加密存储
  REFRESH_TOKEN: {
    key: 'refresh_token', 
    type: StorageType.LOCAL,
    encrypt: true,
    expireTime: 7 * 24 * 60 * 60 * 1000 // 7天
  } as StorageConfig,

  // 用户基本信息 - 使用sessionStorage，不包含敏感信息
  USER_INFO: {
    key: 'user_info',
    type: StorageType.SESSION,
    encrypt: false // 用户基本信息不包含敏感数据，可以不加密以提高性能
  } as StorageConfig,

  // 用户偏好设置 - 使用localStorage，持久化存储
  USER_PREFERENCES: {
    key: 'user_preferences',
    type: StorageType.LOCAL,
    encrypt: false
  } as StorageConfig
}

/**
 * Token管理工具
 */
export const TokenManager = {
  /**
   * 设置Token
   */
  setTokens(accessToken: string, refreshToken: string): void {
    secureStorage.setItem(StorageConfigs.ACCESS_TOKEN, accessToken)
    secureStorage.setItem(StorageConfigs.REFRESH_TOKEN, refreshToken)
  },

  /**
   * 获取Access Token
   */
  getAccessToken(): string | null {
    return secureStorage.getItem<string>(StorageConfigs.ACCESS_TOKEN)
  },

  /**
   * 获取Refresh Token
   */
  getRefreshToken(): string | null {
    return secureStorage.getItem<string>(StorageConfigs.REFRESH_TOKEN)
  },

  /**
   * 清除所有Token
   */
  clearTokens(): void {
    secureStorage.removeItem(StorageConfigs.ACCESS_TOKEN)
    secureStorage.removeItem(StorageConfigs.REFRESH_TOKEN)
  },

  /**
   * 检查Token是否有效
   */
  hasValidTokens(): boolean {
    return secureStorage.hasValidItem(StorageConfigs.ACCESS_TOKEN) && 
           secureStorage.hasValidItem(StorageConfigs.REFRESH_TOKEN)
  }
}

/**
 * 用户信息管理工具
 */
export const UserInfoManager = {
  /**
   * 设置用户信息
   */
  setUserInfo(userInfo: any): void {
    secureStorage.setItem(StorageConfigs.USER_INFO, userInfo)
  },

  /**
   * 获取用户信息
   */
  getUserInfo(): any | null {
    return secureStorage.getItem(StorageConfigs.USER_INFO)
  },

  /**
   * 清除用户信息
   */
  clearUserInfo(): void {
    secureStorage.removeItem(StorageConfigs.USER_INFO)
  },

  /**
   * 设置用户偏好
   */
  setPreferences(preferences: any): void {
    secureStorage.setItem(StorageConfigs.USER_PREFERENCES, preferences)
  },

  /**
   * 获取用户偏好
   */
  getPreferences(): any | null {
    return secureStorage.getItem(StorageConfigs.USER_PREFERENCES)
  }
}
