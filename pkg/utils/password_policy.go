package utils

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// PasswordPolicy 密码策略配置
type PasswordPolicy struct {
	MinLength         int  `json:"min_length"`
	MaxLength         int  `json:"max_length"`
	RequireUppercase  bool `json:"require_uppercase"`
	RequireLowercase  bool `json:"require_lowercase"`
	RequireNumbers    bool `json:"require_numbers"`
	RequireSpecial    bool `json:"require_special"`
	HistoryCount      int  `json:"history_count"`
	ExpiryDays        int  `json:"expiry_days"`
	PreventCommon     bool `json:"prevent_common"`
	PreventUsername   bool `json:"prevent_username"`
}

// PasswordValidationResult 密码验证结果
type PasswordValidationResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// DefaultPasswordPolicy 默认密码策略
func DefaultPasswordPolicy() *PasswordPolicy {
	return &PasswordPolicy{
		MinLength:        8,
		MaxLength:        128,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumbers:   true,
		RequireSpecial:   true,
		HistoryCount:     5,
		ExpiryDays:       90,
		PreventCommon:    true,
		PreventUsername:  true,
	}
}

// ValidatePassword 验证密码是否符合策略
func ValidatePassword(password string, policy *PasswordPolicy, username string, passwordHistory []string) *PasswordValidationResult {
	result := &PasswordValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// 检查密码长度
	if len(password) < policy.MinLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("密码长度不能少于%d个字符", policy.MinLength))
	}

	if len(password) > policy.MaxLength {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("密码长度不能超过%d个字符", policy.MaxLength))
	}

	// 检查大写字母
	if policy.RequireUppercase {
		hasUpper := false
		for _, char := range password {
			if unicode.IsUpper(char) {
				hasUpper = true
				break
			}
		}
		if !hasUpper {
			result.Valid = false
			result.Errors = append(result.Errors, "密码必须包含至少一个大写字母")
		}
	}

	// 检查小写字母
	if policy.RequireLowercase {
		hasLower := false
		for _, char := range password {
			if unicode.IsLower(char) {
				hasLower = true
				break
			}
		}
		if !hasLower {
			result.Valid = false
			result.Errors = append(result.Errors, "密码必须包含至少一个小写字母")
		}
	}

	// 检查数字
	if policy.RequireNumbers {
		hasNumber := false
		for _, char := range password {
			if unicode.IsDigit(char) {
				hasNumber = true
				break
			}
		}
		if !hasNumber {
			result.Valid = false
			result.Errors = append(result.Errors, "密码必须包含至少一个数字")
		}
	}

	// 检查特殊字符
	if policy.RequireSpecial {
		specialChars := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`)
		if !specialChars.MatchString(password) {
			result.Valid = false
			result.Errors = append(result.Errors, "密码必须包含至少一个特殊字符")
		}
	}

	// 检查是否包含用户名
	if policy.PreventUsername && username != "" {
		if strings.Contains(strings.ToLower(password), strings.ToLower(username)) {
			result.Valid = false
			result.Errors = append(result.Errors, "密码不能包含用户名")
		}
	}

	// 检查常见密码
	if policy.PreventCommon {
		commonPasswords := []string{
			"password", "123456", "12345678", "qwerty", "abc123",
			"password123", "admin", "admin123", "root", "root123",
			"test", "test123", "guest", "guest123", "user", "user123",
		}
		lowerPassword := strings.ToLower(password)
		for _, common := range commonPasswords {
			if lowerPassword == common {
				result.Valid = false
				result.Errors = append(result.Errors, "不能使用常见密码")
				break
			}
		}
	}

	// 检查密码历史
	if policy.HistoryCount > 0 && len(passwordHistory) > 0 {
		for _, oldPassword := range passwordHistory {
			if CheckPassword(password, oldPassword) {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf("不能使用最近%d次使用过的密码", policy.HistoryCount))
				break
			}
		}
	}

	// 检查密码强度
	strength := CalculatePasswordStrength(password)
	if strength < 3 {
		result.Warnings = append(result.Warnings, "密码强度较弱，建议使用更复杂的密码")
	}

	return result
}

// CalculatePasswordStrength 计算密码强度 (1-5)
func CalculatePasswordStrength(password string) int {
	strength := 0

	// 长度加分
	if len(password) >= 8 {
		strength++
	}
	if len(password) >= 12 {
		strength++
	}

	// 字符类型加分
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsLower(char) {
			hasLower = true
		} else if unicode.IsDigit(char) {
			hasNumber = true
		} else {
			hasSpecial = true
		}
	}

	if hasUpper {
		strength++
	}
	if hasLower {
		strength++
	}
	if hasNumber {
		strength++
	}
	if hasSpecial {
		strength++
	}

	// 避免重复字符
	hasRepeating := false
	for i := 0; i < len(password)-2; i++ {
		if password[i] == password[i+1] && password[i] == password[i+2] {
			hasRepeating = true
			break
		}
	}
	if !hasRepeating {
		strength++
	}

	// 限制最大强度为5
	if strength > 5 {
		strength = 5
	}

	return strength
}

// GetPasswordStrengthText 获取密码强度文本描述
func GetPasswordStrengthText(strength int) string {
	switch strength {
	case 1:
		return "很弱"
	case 2:
		return "弱"
	case 3:
		return "中等"
	case 4:
		return "强"
	case 5:
		return "很强"
	default:
		return "未知"
	}
}

// GetPasswordStrengthColor 获取密码强度颜色
func GetPasswordStrengthColor(strength int) string {
	switch strength {
	case 1, 2:
		return "red"
	case 3:
		return "orange"
	case 4:
		return "blue"
	case 5:
		return "green"
	default:
		return "gray"
	}
}

// GenerateStrongPassword 生成强密码
func GenerateStrongPassword(policy *PasswordPolicy) (string, error) {
	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numbers   = "0123456789"
		special   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	)

	var chars string
	var password strings.Builder

	// 确保包含必需的字符类型
	if policy.RequireLowercase {
		password.WriteByte(lowercase[RandomInt(0, len(lowercase)-1)])
		chars += lowercase
	}
	if policy.RequireUppercase {
		password.WriteByte(uppercase[RandomInt(0, len(uppercase)-1)])
		chars += uppercase
	}
	if policy.RequireNumbers {
		password.WriteByte(numbers[RandomInt(0, len(numbers)-1)])
		chars += numbers
	}
	if policy.RequireSpecial {
		password.WriteByte(special[RandomInt(0, len(special)-1)])
		chars += special
	}

	// 如果chars为空，使用所有字符类型
	if chars == "" {
		chars = lowercase + uppercase + numbers + special
	}

	// 填充剩余长度
	remainingLength := policy.MinLength - password.Len()
	for i := 0; i < remainingLength; i++ {
		if len(chars) > 0 {
			password.WriteByte(chars[RandomInt(0, len(chars)-1)])
		}
	}

	// 使用Fisher-Yates洗牌算法打乱密码字符顺序
	result := []byte(password.String())
	n := len(result)
	for i := n - 1; i > 0; i-- {
		j := RandomInt(0, i)
		if j <= i && j >= 0 {
			result[i], result[j] = result[j], result[i]
		}
	}

	return string(result), nil
}

// RandomInt 生成随机整数
func RandomInt(min, max int) int {
	if min >= max {
		return min
	}
	// 确保结果在 [min, max] 范围内
	result := min + int(RandomUint32())%(max-min+1)
	if result > max {
		result = max
	}
	return result
}

// RandomUint32 生成随机uint32
func RandomUint32() uint32 {
	bytes := make([]byte, 4)
	_, err := rand.Read(bytes)
	if err != nil {
		return 0
	}
	return uint32(bytes[0])<<24 | uint32(bytes[1])<<16 | uint32(bytes[2])<<8 | uint32(bytes[3])
}
