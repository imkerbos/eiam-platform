import { TokenManager, UserInfoManager } from './storage'

/**
 * 存储迁移工具
 * 负责将旧的localStorage数据迁移到新的安全存储系统
 */
export class StorageMigration {
  private readonly MIGRATION_KEY = 'storage_migration_completed'

  /**
   * 检查是否需要迁移
   */
  needsMigration(): boolean {
    // 如果已经迁移过，则不需要再次迁移
    if (localStorage.getItem(this.MIGRATION_KEY)) {
      return false
    }

    // 检查是否存在旧的存储数据
    const hasOldToken = localStorage.getItem('access_token') !== null
    const hasOldRefreshToken = localStorage.getItem('refresh_token') !== null
    
    return hasOldToken || hasOldRefreshToken
  }

  /**
   * 执行迁移
   */
  migrate(): void {
    try {
      console.log('开始迁移存储数据到安全存储...')
      
      // 迁移token
      this.migrateTokens()
      
      // 迁移用户信息（如果存在的话）
      this.migrateUserInfo()
      
      // 清理旧数据
      this.cleanupOldData()
      
      // 标记迁移完成
      localStorage.setItem(this.MIGRATION_KEY, new Date().toISOString())
      
      console.log('存储数据迁移完成')
    } catch (error) {
      console.error('存储数据迁移失败:', error)
      // 迁移失败时，不清理旧数据，以确保用户不会丢失登录状态
    }
  }

  /**
   * 迁移token数据
   */
  private migrateTokens(): void {
    const accessToken = localStorage.getItem('access_token')
    const refreshToken = localStorage.getItem('refresh_token')
    
    if (accessToken && refreshToken) {
      console.log('迁移token数据...')
      TokenManager.setTokens(accessToken, refreshToken)
    }
  }

  /**
   * 迁移用户信息
   */
  private migrateUserInfo(): void {
    // 检查是否有存储的用户信息（可能以其他key存储）
    const userInfoKeys = ['user_info', 'current_user', 'user_data']
    
    for (const key of userInfoKeys) {
      const userInfo = localStorage.getItem(key)
      if (userInfo) {
        try {
          const userData = JSON.parse(userInfo)
          console.log(`迁移用户信息数据 (${key})...`)
          UserInfoManager.setUserInfo(userData)
          break // 只迁移第一个找到的用户信息
        } catch (error) {
          console.warn(`解析用户信息失败 (${key}):`, error)
        }
      }
    }
  }

  /**
   * 清理旧的存储数据
   */
  private cleanupOldData(): void {
    console.log('清理旧的存储数据...')
    
    // 清理token
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    
    // 清理可能的用户信息
    const keysToClean = ['user_info', 'current_user', 'user_data']
    keysToClean.forEach(key => {
      localStorage.removeItem(key)
    })
  }

  /**
   * 强制重新迁移（开发调试用）
   */
  forceMigration(): void {
    localStorage.removeItem(this.MIGRATION_KEY)
    this.migrate()
  }
}

// 创建迁移实例
export const storageMigration = new StorageMigration()

/**
 * 自动执行迁移检查和迁移
 * 在应用启动时调用
 */
export function initStorageMigration(): void {
  try {
    if (storageMigration.needsMigration()) {
      console.log('检测到需要迁移存储数据')
      storageMigration.migrate()
    }
  } catch (error) {
    console.error('存储迁移初始化失败:', error)
  }
}
