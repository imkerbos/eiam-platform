package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hash password using bcrypt (recommended) or MD5 (legacy)
func HashPassword(password string, cost int) (string, error) {
	// 默认使用bcrypt加密
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// HashPasswordMD5 hash password using MD5 (legacy support)
func HashPasswordMD5(password string) string {
	hash := md5.Sum([]byte(password))
	return fmt.Sprintf("%x", hash)
}

// CheckPassword verify password (supports both bcrypt and MD5)
func CheckPassword(password, hash string) bool {
	// 检查是否是bcrypt哈希（以$2a$, $2b$, $2x$, $2y$开头）
	if strings.HasPrefix(hash, "$2a$") || strings.HasPrefix(hash, "$2b$") ||
		strings.HasPrefix(hash, "$2x$") || strings.HasPrefix(hash, "$2y$") {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
		return err == nil
	}

	// 兼容旧的MD5哈希
	if len(hash) == 32 {
		// 如果传入的密码已经是MD5格式，直接比较
		if len(password) == 32 {
			return password == hash
		}
		// 否则计算密码的MD5值
		passwordHash := md5.Sum([]byte(password))
		passwordHashStr := fmt.Sprintf("%x", passwordHash)
		return passwordHashStr == hash
	}

	return false
}

// ShouldUpgradePassword 检查密码是否需要升级到bcrypt
func ShouldUpgradePassword(hash string) bool {
	// 如果是MD5哈希（32位十六进制），建议升级
	return len(hash) == 32 && !strings.HasPrefix(hash, "$2")
}

// UpgradePasswordHash 将密码哈希升级到bcrypt
func UpgradePasswordHash(plainPassword string) (string, error) {
	return HashPassword(plainPassword, bcrypt.DefaultCost)
}

// GenerateSalt generate random salt
func GenerateSalt(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateRandomString generate random string
func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// HashWithSalt hash with salt
func HashWithSalt(data, salt string) string {
	h := sha256.New()
	h.Write([]byte(data + salt))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// GenerateSecretKey generate secret key
func GenerateSecretKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// GenerateClientSecret generate client secret
func GenerateClientSecret() (string, error) {
	return GenerateSecretKey(32)
}

// GenerateClientID generate client ID
func GenerateClientID() (string, error) {
	return GenerateRandomString(16)
}

// DecryptPassword decrypt password from frontend
func DecryptPassword(encryptedPassword string) (string, error) {
	// 加密密钥（应该与前端保持一致）
	key := []byte("eiam-platform-2024")

	// 如果密钥长度不是32字节，需要填充或截断
	if len(key) < 32 {
		// 填充到32字节
		paddedKey := make([]byte, 32)
		copy(paddedKey, key)
		key = paddedKey
	} else if len(key) > 32 {
		// 截断到32字节
		key = key[:32]
	}

	// 解码base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %v", err)
	}

	// 创建AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	// 检查密文长度
	if len(ciphertext) < aes.BlockSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// 提取IV
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	// 创建CBC解密器
	mode := cipher.NewCBCDecrypter(block, iv)

	// 解密
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// 去除填充
	plaintext = removePKCS7Padding(plaintext)

	return string(plaintext), nil
}

// removePKCS7Padding removes PKCS7 padding
func removePKCS7Padding(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return data
	}

	padding := int(data[length-1])
	if padding > length {
		return data
	}

	return data[:length-padding]
}
