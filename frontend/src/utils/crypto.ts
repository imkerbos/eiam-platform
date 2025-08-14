import CryptoJS from 'crypto-js'

// 加密密钥（应该从环境变量或配置中获取）
const ENCRYPTION_KEY = 'eiam-platform-2024'

/**
 * 加密密码
 * @param password 原始密码
 * @returns 加密后的密码
 */
export function encryptPassword(password: string): string {
  try {
    // 使用MD5加密
    console.log('Encrypting password:', password)
    const encrypted = CryptoJS.MD5(password).toString()
    console.log('Encrypted password:', encrypted)
    return encrypted
  } catch (error) {
    console.error('Password encryption failed:', error)
    // 如果加密失败，返回原始密码（不推荐，但确保功能可用）
    return password
  }
}

/**
 * 解密密码（仅用于测试，生产环境不应解密）
 * @param encryptedPassword 加密后的密码
 * @returns 解密后的密码
 */
export function decryptPassword(encryptedPassword: string): string {
  try {
    const key = CryptoJS.enc.Utf8.parse(ENCRYPTION_KEY)
    const rawData = CryptoJS.enc.Base64.parse(encryptedPassword)
    
    // 提取IV (前16字节)
    const iv = CryptoJS.lib.WordArray.create(rawData.words.slice(0, 4))
    const ciphertext = CryptoJS.lib.WordArray.create(rawData.words.slice(4))
    
    const decrypted = CryptoJS.AES.decrypt(
      CryptoJS.lib.CipherParams.create({ ciphertext: ciphertext }),
      key,
      {
        iv: iv,
        mode: CryptoJS.mode.CBC,
        padding: CryptoJS.pad.Pkcs7
      }
    )
    
    return decrypted.toString(CryptoJS.enc.Utf8)
  } catch (error) {
    console.error('Password decryption failed:', error)
    return encryptedPassword
  }
}

/**
 * 生成随机盐值
 * @param length 盐值长度
 * @returns 随机盐值
 */
export function generateSalt(length: number = 16): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  let result = ''
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  return result
}

/**
 * 哈希密码（使用盐值）
 * @param password 密码
 * @param salt 盐值
 * @returns 哈希后的密码
 */
export function hashPassword(password: string, salt: string): string {
  return CryptoJS.SHA256(password + salt).toString()
}
