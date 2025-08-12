package handlers

import (
	"net/http"
	"strconv"

	"eiam-platform/internal/models"
	"eiam-platform/pkg/database"
	"eiam-platform/pkg/i18n"
	"eiam-platform/pkg/logger"
	"eiam-platform/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username       string `json:"username" binding:"required" validate:"required,min=3,max=50"`
	Email          string `json:"email" binding:"required" validate:"required,email"`
	DisplayName    string `json:"display_name" binding:"required" validate:"required,max=100"`
	Phone          string `json:"phone" validate:"omitempty,len=11"`
	Password       string `json:"password" binding:"required" validate:"required,min=6"`
	OrganizationID string `json:"organization_id" validate:"required"`
	Status         int    `json:"status"`
	EnableOTP      bool   `json:"enable_otp"`
	EmailVerified  bool   `json:"email_verified"`
	PhoneVerified  bool   `json:"phone_verified"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	DisplayName    string `json:"display_name" validate:"omitempty,max=100"`
	Phone          string `json:"phone" validate:"omitempty,len=11"`
	OrganizationID string `json:"organization_id"`
	Status         *int   `json:"status"`
	EnableOTP      *bool  `json:"enable_otp"`
	EmailVerified  *bool  `json:"email_verified"`
	PhoneVerified  *bool  `json:"phone_verified"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Items      []UserInfo `json:"items"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}

// GetUsersHandler 获取用户列表
func GetUsersHandler(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")
	status := c.Query("status")
	organizationID := c.Query("organization_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询
	query := database.DB.Model(&models.User{})

	// 搜索条件
	if search != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR display_name LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 状态过滤
	if status != "" {
		if statusInt, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", statusInt)
		}
	}

	// 组织过滤
	if organizationID != "" {
		query = query.Where("organization_id = ?", organizationID)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 计算总页数
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	// 分页查询
	offset := (page - 1) * pageSize
	var users []models.User
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error
	if err != nil {
		logger.ErrorError("Failed to get users", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 转换为响应格式
	items := make([]UserInfo, len(users))
	for i, user := range users {
		items[i] = UserInfo{
			ID:            user.ID,
			Username:      user.Username,
			Email:         user.Email,
			DisplayName:   user.DisplayName,
			Avatar:        user.Avatar,
			Status:        user.Status.String(),
			EmailVerified: user.EmailVerified,
			PhoneVerified: user.PhoneVerified,
			EnableOTP:     user.EnableOTP,
		}
		if user.LastLoginAt != nil {
			items[i].LastLoginAt = *user.LastLoginAt
		}
		items[i].LastLoginIP = user.LastLoginIP
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.Success,
		"data": UserListResponse{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

// CreateUserHandler 创建用户
func CreateUserHandler(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.InvalidRequestData,
			"data":    nil,
		})
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := database.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.UserExists,
			"data":    nil,
		})
		return
	}

	// 检查邮箱是否已存在
	if err := database.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.EmailExists,
			"data":    nil,
		})
		return
	}

	// 检查组织是否存在
	var organization models.Organization
	logger.Info("Checking organization", zap.String("org_id", req.OrganizationID))
	if err := database.DB.Where("id = ?", req.OrganizationID).First(&organization).Error; err != nil {
		logger.ErrorError("Organization not found", zap.String("org_id", req.OrganizationID), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.OrganizationNotFound,
			"data":    nil,
		})
		return
	}
	logger.Info("Organization found", zap.String("org_id", organization.ID), zap.String("org_name", organization.Name))

	// 加密密码
	hashedPassword, err := utils.HashPassword(req.Password, 12)
	if err != nil {
		logger.ErrorError("Failed to hash password", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 生成盐值
	salt, err := utils.GenerateSalt(16)
	if err != nil {
		logger.ErrorError("Failed to generate salt", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 设置默认状态
	if req.Status == 0 {
		req.Status = int(models.StatusActive)
	}

	// 创建用户
	user := models.User{
		Username:           req.Username,
		Email:              req.Email,
		DisplayName:        req.DisplayName,
		Phone:              req.Phone,
		Password:           hashedPassword,
		Salt:               salt,
		OrganizationID:     req.OrganizationID,
		Status:             models.Status(req.Status),
		EmailVerified:      req.EmailVerified,
		PhoneVerified:      req.PhoneVerified,
		EnableOTP:          req.EnableOTP,
		MustChangePassword: true, // 新用户必须修改密码
	}

	if err := database.DB.Create(&user).Error; err != nil {
		logger.ErrorError("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	logger.Info("User created successfully",
		zap.String("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("created_by", c.GetString("user_id")),
	)

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": i18n.SuccessCreated,
		"data": UserInfo{
			ID:            user.ID,
			Username:      user.Username,
			Email:         user.Email,
			DisplayName:   user.DisplayName,
			Avatar:        user.Avatar,
			Status:        user.Status.String(),
			EmailVerified: user.EmailVerified,
			PhoneVerified: user.PhoneVerified,
			EnableOTP:     user.EnableOTP,
		},
	})
}

// GetUserHandler 获取单个用户
func GetUserHandler(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": i18n.UserNotFound,
				"data":    nil,
			})
			return
		}
		logger.ErrorError("Failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	userInfo := UserInfo{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		Avatar:        user.Avatar,
		Status:        user.Status.String(),
		EmailVerified: user.EmailVerified,
		PhoneVerified: user.PhoneVerified,
		EnableOTP:     user.EnableOTP,
		LastLoginIP:   user.LastLoginIP,
	}
	if user.LastLoginAt != nil {
		userInfo.LastLoginAt = *user.LastLoginAt
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.Success,
		"data":    userInfo,
	})
}

// UpdateUserHandler 更新用户
func UpdateUserHandler(c *gin.Context) {
	userID := c.Param("id")

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.InvalidRequestData,
			"data":    nil,
		})
		return
	}

	// 检查用户是否存在
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": i18n.UserNotFound,
				"data":    nil,
			})
			return
		}
		logger.ErrorError("Failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 更新字段
	updates := make(map[string]interface{})

	if req.DisplayName != "" {
		updates["display_name"] = req.DisplayName
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.OrganizationID != "" {
		// 检查组织是否存在
		var organization models.Organization
		if err := database.DB.Where("id = ?", req.OrganizationID).First(&organization).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": i18n.OrganizationNotFound,
				"data":    nil,
			})
			return
		}
		updates["organization_id"] = req.OrganizationID
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.EnableOTP != nil {
		updates["enable_otp"] = *req.EnableOTP
	}
	if req.EmailVerified != nil {
		updates["email_verified"] = *req.EmailVerified
	}
	if req.PhoneVerified != nil {
		updates["phone_verified"] = *req.PhoneVerified
	}

	// 执行更新
	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		logger.ErrorError("Failed to update user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	logger.Info("User updated successfully",
		zap.String("user_id", user.ID),
		zap.String("updated_by", c.GetString("user_id")),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.SuccessUpdated,
		"data":    nil,
	})
}

// DeleteUserHandler 删除用户
func DeleteUserHandler(c *gin.Context) {
	userID := c.Param("id")

	// 检查用户是否存在
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": i18n.UserNotFound,
				"data":    nil,
			})
			return
		}
		logger.ErrorError("Failed to get user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 软删除用户
	if err := database.DB.Delete(&user).Error; err != nil {
		logger.ErrorError("Failed to delete user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	logger.Info("User deleted successfully",
		zap.String("user_id", user.ID),
		zap.String("deleted_by", c.GetString("user_id")),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.SuccessDeleted,
		"data":    nil,
	})
}
