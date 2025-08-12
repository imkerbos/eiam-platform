package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"eiam-platform/internal/models"
	"eiam-platform/pkg/utils"
)

type ProfileHandler struct {
	db *gorm.DB
}

func NewProfileHandler(db *gorm.DB) *ProfileHandler {
	return &ProfileHandler{db: db}
}

// GetProfileHandler 获取用户个人资料
func (h *ProfileHandler) GetProfileHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user models.User
	if err := h.db.Preload("Organization").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"id":                user.ID,
			"username":          user.Username,
			"email":             user.Email,
			"display_name":      user.DisplayName,
			"phone":             user.Phone,
			"avatar":            user.Avatar,
			"status":            user.Status,
			"email_verified":    user.EmailVerified,
			"phone_verified":    user.PhoneVerified,
			"enable_otp":        user.EnableOTP,
			"organization_id":   user.OrganizationID,
			"organization_name": user.Organization.Name,
			"created_at":        user.CreatedAt,
			"updated_at":        user.UpdatedAt,
		},
	})
}

// UpdateProfileHandler 更新用户个人资料
func (h *ProfileHandler) UpdateProfileHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		DisplayName string `json:"display_name" binding:"required"`
		Phone       string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 更新用户信息
	updates := map[string]interface{}{
		"display_name": req.DisplayName,
		"phone":        req.Phone,
		"updated_at":   time.Now(),
	}

	if err := h.db.Model(&user).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{"message": "Profile updated successfully"},
	})
}

// UploadAvatarHandler 上传头像
func (h *ProfileHandler) UploadAvatarHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// 验证文件类型
	allowedTypes := []string{"image/jpeg", "image/png", "image/gif"}
	contentType := file.Header.Get("Content-Type")
	isAllowed := false
	for _, allowedType := range allowedTypes {
		if contentType == allowedType {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only JPG, PNG and GIF files are allowed"})
		return
	}

	// 验证文件大小 (2MB)
	if file.Size > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size must be less than 2MB"})
		return
	}

	// 创建上传目录
	uploadDir := "uploads/avatars"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s_%s%s", userID, uuid.New().String()[:8], ext)
	filepath := filepath.Join(uploadDir, filename)

	// 保存文件
	if err := c.SaveUploadedFile(file, filepath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// 更新用户头像路径
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 删除旧头像文件
	if user.Avatar != "" && user.Avatar != filepath {
		os.Remove(user.Avatar)
	}

	// 更新数据库
	if err := h.db.Model(&user).Update("avatar", filepath).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update avatar"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"message": "Avatar uploaded successfully",
			"avatar":  filepath,
		},
	})
}

// ChangePasswordHandler 修改密码
func (h *ProfileHandler) ChangePasswordHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 验证当前密码
	if !utils.CheckPassword(req.CurrentPassword, user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is incorrect"})
		return
	}

	// 加密新密码
	hashedPassword, err := utils.HashPassword(req.NewPassword, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// 更新密码
	if err := h.db.Model(&user).Update("password", hashedPassword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{"message": "Password changed successfully"},
	})
}

// VerifyEmailHandler 发送邮箱验证
func (h *ProfileHandler) VerifyEmailHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.EmailVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is already verified"})
		return
	}

	// TODO: 实现邮箱验证逻辑
	// 1. 生成验证码
	// 2. 发送验证邮件
	// 3. 保存验证码到数据库

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{"message": "Verification email sent successfully"},
	})
}

// SetupOTPHandler 设置2FA
func (h *ProfileHandler) SetupOTPHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		VerificationCode string `json:"verification_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// TODO: 验证OTP代码
	// 这里应该验证用户提供的验证码是否正确

	// 启用2FA
	if err := h.db.Model(&user).Update("enable_otp", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable 2FA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{"message": "2FA enabled successfully"},
	})
}

// DisableOTPHandler 禁用2FA
func (h *ProfileHandler) DisableOTPHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if !user.EnableOTP {
		c.JSON(http.StatusBadRequest, gin.H{"error": "2FA is not enabled"})
		return
	}

	// 禁用2FA
	if err := h.db.Model(&user).Update("enable_otp", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable 2FA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{"message": "2FA disabled successfully"},
	})
}

// GetBackupCodesHandler 获取备份码
func (h *ProfileHandler) GetBackupCodesHandler(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if !user.EnableOTP {
		c.JSON(http.StatusBadRequest, gin.H{"error": "2FA is not enabled"})
		return
	}

	// TODO: 生成或获取备份码
	backupCodes := []string{
		"12345678",
		"87654321",
		"11111111",
		"22222222",
		"33333333",
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{
			"backup_codes": backupCodes,
			"message":      "Backup codes generated successfully",
		},
	})
}
