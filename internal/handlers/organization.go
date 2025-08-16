package handlers

import (
	"net/http"
	"strconv"

	"eiam-platform/internal/models"
	"eiam-platform/pkg/database"
	"eiam-platform/pkg/i18n"
	"eiam-platform/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// CreateOrganizationRequest 创建组织请求
type CreateOrganizationRequest struct {
	Name        string `json:"name" binding:"required" validate:"required,max=100"`
	Code        string `json:"code" binding:"required" validate:"required,max=50"`
	Type        int    `json:"type" binding:"required"`
	ParentID    string `json:"parent_id"`
	Description string `json:"description" validate:"max=500"`
	Manager     string `json:"manager"`
	Location    string `json:"location" validate:"max=200"`
	Phone       string `json:"phone" validate:"max=50"`
	Email       string `json:"email" validate:"omitempty,email"`
	Status      int    `json:"status"`
}

// UpdateOrganizationRequest 更新组织请求
type UpdateOrganizationRequest struct {
	Name        string `json:"name" validate:"omitempty,max=100"`
	Code        string `json:"code" validate:"omitempty,max=50"`
	Type        *int   `json:"type"`
	ParentID    string `json:"parent_id"`
	Description string `json:"description" validate:"max=500"`
	Manager     string `json:"manager"`
	Location    string `json:"location" validate:"max=200"`
	Phone       string `json:"phone" validate:"max=50"`
	Email       string `json:"email" validate:"omitempty,email"`
	Status      *int   `json:"status"`
}

// OrganizationInfo 组织信息
type OrganizationInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Code        string `json:"code"`
	Type        int    `json:"type"`
	TypeName    string `json:"type_name"`
	ParentID    string `json:"parent_id"`
	Level       int    `json:"level"`
	Path        string `json:"path"`
	Sort        int    `json:"sort"`
	Description string `json:"description"`
	Manager     string `json:"manager"`      // 保持兼容性，显示manager名称
	ManagerID   string `json:"manager_id"`   // 新增：manager的用户ID
	ManagerName string `json:"manager_name"` // 新增：manager的显示名称
	Location    string `json:"location"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Status      int    `json:"status"`
	StatusName  string `json:"status_name"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// OrganizationTreeInfo organization tree info for response
type OrganizationTreeInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Code        string                 `json:"code"`
	Type        int                    `json:"type"`
	TypeName    string                 `json:"type_name"`
	ParentID    string                 `json:"parent_id"`
	Level       int                    `json:"level"`
	Path        string                 `json:"path"`
	Sort        int                    `json:"sort"`
	Description string                 `json:"description"`
	Manager     string                 `json:"manager"`
	Location    string                 `json:"location"`
	Phone       string                 `json:"phone"`
	Email       string                 `json:"email"`
	Status      int                    `json:"status"`
	StatusName  string                 `json:"status_name"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
	Children    []OrganizationTreeInfo `json:"children"`
}

// OrganizationListResponse 组织列表响应
type OrganizationListResponse struct {
	Items      []OrganizationInfo `json:"items"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// GetOrganizationsHandler 获取组织列表
func GetOrganizationsHandler(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	search := c.Query("search")
	status := c.Query("status")
	orgType := c.Query("type")
	parentID := c.Query("parent_id")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 构建查询
	query := database.DB.Model(&models.Organization{})

	// 搜索条件
	if search != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 状态过滤
	if status != "" {
		if statusInt, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", statusInt)
		}
	}

	// 类型过滤
	if orgType != "" {
		if typeInt, err := strconv.Atoi(orgType); err == nil {
			query = query.Where("type = ?", typeInt)
		}
	}

	// 父级过滤
	if parentID != "" {
		query = query.Where("parent_id = ?", parentID)
	}

	// 获取总数
	var total int64
	query.Count(&total)

	// 计算总页数
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	// 分页查询
	offset := (page - 1) * pageSize
	var organizations []models.Organization
	err := query.Offset(offset).Limit(pageSize).Order("sort ASC, created_at DESC").Find(&organizations).Error
	if err != nil {
		logger.ErrorError("Failed to get organizations", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 转换为响应格式
	items := make([]OrganizationInfo, len(organizations))
	for i, org := range organizations {
		parentID := ""
		if org.ParentID != nil {
			parentID = *org.ParentID
		}

		// Get manager username if manager ID exists
		managerID := org.Manager
		managerName := ""
		if org.Manager != "" {
			var managerUser models.User
			if err := database.DB.Select("username, display_name").Where("id = ?", org.Manager).First(&managerUser).Error; err == nil {
				if managerUser.DisplayName != "" {
					managerName = managerUser.DisplayName
				} else {
					managerName = managerUser.Username
				}
			}
		}

		items[i] = OrganizationInfo{
			ID:          org.ID,
			Name:        org.Name,
			Code:        org.Code,
			Type:        int(org.Type),
			TypeName:    org.Type.String(),
			ParentID:    parentID,
			Level:       org.Level,
			Path:        org.Path,
			Sort:        org.Sort,
			Description: org.Description,
			Manager:     managerName, // 显示名称，向后兼容
			ManagerID:   managerID,   // 用户ID，用于编辑
			ManagerName: managerName, // 显示名称，明确字段
			Location:    org.Location,
			Phone:       org.Phone,
			Email:       org.Email,
			Status:      int(org.Status),
			StatusName:  org.Status.String(),
			CreatedAt:   org.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   org.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.Success,
		"data": OrganizationListResponse{
			Items:      items,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

// GetOrganizationsTreeHandler 获取组织树形结构
func GetOrganizationsTreeHandler(c *gin.Context) {
	search := c.Query("search")
	status := c.Query("status")

	// 构建查询
	query := database.DB.Model(&models.Organization{})

	// 搜索条件
	if search != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 状态过滤
	if status != "" {
		if statusInt, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", statusInt)
		}
	}

	// 只查询顶级组织（没有父级或父级为空）
	query = query.Where("parent_id IS NULL OR parent_id = ''")

	var organizations []models.Organization
	err := query.Order("sort ASC, created_at DESC").Find(&organizations).Error
	if err != nil {
		logger.ErrorError("Failed to get organizations tree", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 递归加载子组织
	for i := range organizations {
		loadChildrenRecursively(&organizations[i], search, status)
	}

	// 转换为响应格式
	items := make([]OrganizationTreeInfo, len(organizations))
	for i, org := range organizations {
		items[i] = convertToTreeInfo(org)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.Success,
		"data":    items,
	})
}

// CreateOrganizationHandler 创建组织
func CreateOrganizationHandler(c *gin.Context) {
	var req CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.InvalidRequestData,
			"data":    nil,
		})
		return
	}

	// 检查组织代码是否已存在
	var existingOrg models.Organization
	if err := database.DB.Where("code = ?", req.Code).First(&existingOrg).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.OrganizationExists,
			"data":    nil,
		})
		return
	}

	// 检查父级组织是否存在
	if req.ParentID != "" {
		if err := database.DB.Where("id = ?", req.ParentID).First(&models.Organization{}).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": i18n.OrganizationNotFound,
				"data":    nil,
			})
			return
		}
	}

	// 检查管理员是否存在
	if req.Manager != "" {
		var managerUser models.User
		if err := database.DB.Where("id = ?", req.Manager).First(&managerUser).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": i18n.UserNotFound,
				"data":    nil,
			})
			return
		}
	}

	// 设置默认状态
	if req.Status == 0 {
		req.Status = int(models.StatusActive)
	}

	// 计算组织层级和路径
	level := 1
	path := "/"
	if req.ParentID != "" {
		var parentOrg models.Organization
		if err := database.DB.Where("id = ?", req.ParentID).First(&parentOrg).Error; err == nil {
			level = parentOrg.Level + 1
			path = parentOrg.Path + parentOrg.ID + "/"
		}
	}

	// 创建组织
	var parentID *string
	if req.ParentID != "" {
		parentID = &req.ParentID
	}
	organization := models.Organization{
		Name:        req.Name,
		Code:        req.Code,
		Type:        models.OrganizationType(req.Type),
		ParentID:    parentID,
		Level:       level,
		Path:        path,
		Description: req.Description,
		Manager:     req.Manager,
		Location:    req.Location,
		Phone:       req.Phone,
		Email:       req.Email,
		Status:      models.Status(req.Status),
	}

	if err := database.DB.Create(&organization).Error; err != nil {
		logger.ErrorError("Failed to create organization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	logger.Info("Organization created successfully",
		zap.String("org_id", organization.ID),
		zap.String("org_name", organization.Name),
		zap.String("created_by", c.GetString("user_id")),
	)

	var parentIDStr string
	if organization.ParentID != nil {
		parentIDStr = *organization.ParentID
	}
	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": i18n.SuccessCreated,
		"data": OrganizationInfo{
			ID:          organization.ID,
			Name:        organization.Name,
			Code:        organization.Code,
			Type:        int(organization.Type),
			TypeName:    organization.Type.String(),
			ParentID:    parentIDStr,
			Level:       organization.Level,
			Path:        organization.Path,
			Sort:        organization.Sort,
			Description: organization.Description,
			Manager:     organization.Manager,
			Location:    organization.Location,
			Phone:       organization.Phone,
			Email:       organization.Email,
			Status:      int(organization.Status),
			StatusName:  organization.Status.String(),
			CreatedAt:   organization.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   organization.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

// UpdateOrganizationHandler 更新组织
func UpdateOrganizationHandler(c *gin.Context) {
	orgID := c.Param("id")

	var req UpdateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.InvalidRequestData,
			"data":    nil,
		})
		return
	}

	// 检查组织是否存在
	var organization models.Organization
	if err := database.DB.Where("id = ?", orgID).First(&organization).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "Organization not found",
				"data":    nil,
			})
			return
		}
		logger.ErrorError("Failed to get organization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 更新字段
	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Code != "" {
		// 检查代码是否已被其他组织使用
		var existingOrg models.Organization
		if err := database.DB.Where("code = ? AND id != ?", req.Code, orgID).First(&existingOrg).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": i18n.OrganizationExists,
				"data":    nil,
			})
			return
		}
		updates["code"] = req.Code
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.ParentID != "" {
		// 检查父级组织是否存在
		var parentOrg models.Organization
		if err := database.DB.Where("id = ?", req.ParentID).First(&parentOrg).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": i18n.OrganizationNotFound,
				"data":    nil,
			})
			return
		}
		updates["parent_id"] = req.ParentID
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Manager != "" {
		// 检查管理员是否存在
		var managerUser models.User
		if err := database.DB.Where("id = ?", req.Manager).First(&managerUser).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": i18n.ManagerNotFound,
				"data":    nil,
			})
			return
		}
		updates["manager"] = req.Manager
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	// 执行更新
	if err := database.DB.Model(&organization).Updates(updates).Error; err != nil {
		logger.ErrorError("Failed to update organization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	logger.Info("Organization updated successfully",
		zap.String("org_id", organization.ID),
		zap.String("updated_by", c.GetString("user_id")),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.SuccessUpdated,
		"data":    nil,
	})
}

// GetOrganizationHandler 获取单个组织
func GetOrganizationHandler(c *gin.Context) {
	orgID := c.Param("id")

	var organization models.Organization
	if err := database.DB.Where("id = ?", orgID).First(&organization).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "Organization not found",
				"data":    nil,
			})
			return
		}
		logger.ErrorError("Failed to get organization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	var parentIDStr string
	if organization.ParentID != nil {
		parentIDStr = *organization.ParentID
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.Success,
		"data": OrganizationInfo{
			ID:          organization.ID,
			Name:        organization.Name,
			Code:        organization.Code,
			Type:        int(organization.Type),
			TypeName:    organization.Type.String(),
			ParentID:    parentIDStr,
			Level:       organization.Level,
			Path:        organization.Path,
			Sort:        organization.Sort,
			Description: organization.Description,
			Manager:     organization.Manager,
			Location:    organization.Location,
			Phone:       organization.Phone,
			Email:       organization.Email,
			Status:      int(organization.Status),
			StatusName:  organization.Status.String(),
			CreatedAt:   organization.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   organization.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

// DeleteOrganizationHandler 删除组织
func DeleteOrganizationHandler(c *gin.Context) {
	orgID := c.Param("id")

	// 检查组织是否存在
	var organization models.Organization
	if err := database.DB.Where("id = ?", orgID).First(&organization).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "Organization not found",
				"data":    nil,
			})
			return
		}
		logger.ErrorError("Failed to get organization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	// 检查是否有子组织
	var childCount int64
	database.DB.Model(&models.Organization{}).Where("parent_id = ?", orgID).Count(&childCount)
	if childCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.CannotDeleteOrgWithChildren,
			"data":    nil,
		})
		return
	}

	// 检查是否有用户属于该组织
	var userCount int64
	database.DB.Model(&models.User{}).Where("organization_id = ?", orgID).Count(&userCount)
	if userCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": i18n.CannotDeleteOrgWithUsers,
			"data":    nil,
		})
		return
	}

	// 软删除组织
	if err := database.DB.Delete(&organization).Error; err != nil {
		logger.ErrorError("Failed to delete organization", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": i18n.InternalServerError,
			"data":    nil,
		})
		return
	}

	logger.Info("Organization deleted successfully",
		zap.String("org_id", organization.ID),
		zap.String("deleted_by", c.GetString("user_id")),
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": i18n.SuccessDeleted,
		"data":    nil,
	})
}

// loadChildrenRecursively 递归加载子组织
func loadChildrenRecursively(org *models.Organization, search, status string) {
	query := database.DB.Model(&models.Organization{}).Where("parent_id = ?", org.ID)

	// 搜索条件
	if search != "" {
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 状态过滤
	if status != "" {
		if statusInt, err := strconv.Atoi(status); err == nil {
			query = query.Where("status = ?", statusInt)
		}
	}

	var children []models.Organization
	if err := query.Order("sort ASC, created_at DESC").Find(&children).Error; err == nil {
		org.Children = children
		// 递归加载子组织的子组织
		for i := range children {
			loadChildrenRecursively(&children[i], search, status)
		}
	}
}

// convertToTreeInfo 转换为树形信息
func convertToTreeInfo(org models.Organization) OrganizationTreeInfo {
	parentID := ""
	if org.ParentID != nil {
		parentID = *org.ParentID
	}

	children := make([]OrganizationTreeInfo, len(org.Children))
	for i, child := range org.Children {
		children[i] = convertToTreeInfo(child)
	}

	return OrganizationTreeInfo{
		ID:          org.ID,
		Name:        org.Name,
		Code:        org.Code,
		Type:        int(org.Type),
		TypeName:    org.Type.String(),
		ParentID:    parentID,
		Level:       org.Level,
		Path:        org.Path,
		Sort:        org.Sort,
		Description: org.Description,
		Manager:     org.Manager,
		Location:    org.Location,
		Phone:       org.Phone,
		Email:       org.Email,
		Status:      int(org.Status),
		StatusName:  org.Status.String(),
		CreatedAt:   org.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   org.UpdatedAt.Format("2006-01-02 15:04:05"),
		Children:    children,
	}
}
