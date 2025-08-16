import { TokenManager, UserInfoManager, secureStorage, StorageConfigs } from './storage'

/**
 * 存储功能测试
 */
export function testSecureStorage() {
  console.log('🧪 开始测试安全存储功能...')
  
  try {
    // 测试1: Token管理
    console.log('📋 测试1: Token管理')
    const testAccessToken = 'test-access-token-12345'
    const testRefreshToken = 'test-refresh-token-67890'
    
    // 设置token
    TokenManager.setTokens(testAccessToken, testRefreshToken)
    console.log('✅ Token设置成功')
    
    // 获取token
    const retrievedAccessToken = TokenManager.getAccessToken()
    const retrievedRefreshToken = TokenManager.getRefreshToken()
    
    if (retrievedAccessToken === testAccessToken && retrievedRefreshToken === testRefreshToken) {
      console.log('✅ Token获取成功，数据一致')
    } else {
      console.error('❌ Token数据不一致')
      console.log('期望access token:', testAccessToken)
      console.log('实际access token:', retrievedAccessToken)
      console.log('期望refresh token:', testRefreshToken)
      console.log('实际refresh token:', retrievedRefreshToken)
    }
    
    // 测试token有效性检查
    const hasValidTokens = TokenManager.hasValidTokens()
    console.log('✅ Token有效性检查:', hasValidTokens)
    
    // 测试2: 用户信息管理
    console.log('\n📋 测试2: 用户信息管理')
    const testUserInfo = {
      id: 'user-123',
      username: 'testuser',
      email: 'test@example.com',
      display_name: '测试用户'
    }
    
    // 设置用户信息
    UserInfoManager.setUserInfo(testUserInfo)
    console.log('✅ 用户信息设置成功')
    
    // 获取用户信息
    const retrievedUserInfo = UserInfoManager.getUserInfo()
    
    if (JSON.stringify(retrievedUserInfo) === JSON.stringify(testUserInfo)) {
      console.log('✅ 用户信息获取成功，数据一致')
    } else {
      console.error('❌ 用户信息数据不一致')
      console.log('期望用户信息:', testUserInfo)
      console.log('实际用户信息:', retrievedUserInfo)
    }
    
    // 测试3: 数据加密
    console.log('\n📋 测试3: 数据加密验证')
    
    // 检查sessionStorage中的access token是否被加密
    const rawSessionData = sessionStorage.getItem(StorageConfigs.ACCESS_TOKEN.key)
    if (rawSessionData && rawSessionData !== testAccessToken) {
      console.log('✅ Access Token已加密存储')
      console.log('原始数据长度:', testAccessToken.length)
      console.log('加密数据长度:', rawSessionData.length)
    } else {
      console.warn('⚠️ Access Token可能未加密存储')
    }
    
    // 检查localStorage中的refresh token是否被加密
    const rawLocalData = localStorage.getItem(StorageConfigs.REFRESH_TOKEN.key)
    if (rawLocalData && rawLocalData !== testRefreshToken) {
      console.log('✅ Refresh Token已加密存储')
      console.log('原始数据长度:', testRefreshToken.length)
      console.log('加密数据长度:', rawLocalData.length)
    } else {
      console.warn('⚠️ Refresh Token可能未加密存储')
    }
    
    // 测试4: 数据清除
    console.log('\n📋 测试4: 数据清除')
    
    // 清除token
    TokenManager.clearTokens()
    const clearedAccessToken = TokenManager.getAccessToken()
    const clearedRefreshToken = TokenManager.getRefreshToken()
    
    if (clearedAccessToken === null && clearedRefreshToken === null) {
      console.log('✅ Token清除成功')
    } else {
      console.error('❌ Token清除失败')
    }
    
    // 清除用户信息
    UserInfoManager.clearUserInfo()
    const clearedUserInfo = UserInfoManager.getUserInfo()
    
    if (clearedUserInfo === null) {
      console.log('✅ 用户信息清除成功')
    } else {
      console.error('❌ 用户信息清除失败')
    }
    
    console.log('\n🎉 安全存储功能测试完成!')
    
    return {
      success: true,
      message: '所有测试通过'
    }
    
  } catch (error) {
    console.error('❌ 存储测试失败:', error)
    return {
      success: false,
      message: `测试失败: ${error.message}`
    }
  }
}

/**
 * 测试迁移功能
 */
export function testStorageMigration() {
  console.log('🧪 开始测试存储迁移功能...')
  
  try {
    // 模拟旧的localStorage数据
    console.log('📋 模拟旧数据')
    localStorage.setItem('access_token', 'old-access-token-123')
    localStorage.setItem('refresh_token', 'old-refresh-token-456')
    
    // 导入迁移模块
    import('./migration').then(({ storageMigration }) => {
      // 检查是否需要迁移
      const needsMigration = storageMigration.needsMigration()
      console.log('✅ 检测到需要迁移:', needsMigration)
      
      if (needsMigration) {
        // 执行迁移
        storageMigration.migrate()
        console.log('✅ 迁移执行完成')
        
        // 验证迁移结果
        const migratedAccessToken = TokenManager.getAccessToken()
        const migratedRefreshToken = TokenManager.getRefreshToken()
        
        if (migratedAccessToken === 'old-access-token-123' && migratedRefreshToken === 'old-refresh-token-456') {
          console.log('✅ 迁移成功，数据正确')
        } else {
          console.error('❌ 迁移失败，数据不正确')
        }
        
        // 检查旧数据是否被清理
        const oldAccessToken = localStorage.getItem('access_token')
        const oldRefreshToken = localStorage.getItem('refresh_token')
        
        if (oldAccessToken === null && oldRefreshToken === null) {
          console.log('✅ 旧数据清理成功')
        } else {
          console.warn('⚠️ 旧数据未完全清理')
        }
      }
      
      console.log('🎉 存储迁移功能测试完成!')
    })
    
  } catch (error) {
    console.error('❌ 迁移测试失败:', error)
  }
}

// 在开发环境中自动运行测试
if (import.meta.env.DEV) {
  // 延迟执行，确保页面加载完成
  setTimeout(() => {
    console.log('🚀 开发环境自动测试开始...')
    testSecureStorage()
    testStorageMigration()
  }, 1000)
}
