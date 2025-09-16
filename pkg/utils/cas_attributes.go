package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"eiam-platform/internal/models"
)

// CASAttributeMapping CAS属性映射配置
type CASAttributeMapping struct {
	// 默认属性映射
	DefaultAttributes map[string]string `json:"default_attributes"`
	// 自定义属性映射
	CustomAttributes map[string]string `json:"custom_attributes"`
	// 是否包含所有用户字段
	IncludeAllFields bool `json:"include_all_fields"`
}

// GetDefaultCASAttributeMapping 获取默认的CAS属性映射
func GetDefaultCASAttributeMapping() CASAttributeMapping {
	return CASAttributeMapping{
		DefaultAttributes: map[string]string{
			"email":       "email",
			"displayName": "display_name",
			"username":    "username",
			"userId":      "id",
		},
		CustomAttributes: make(map[string]string),
		IncludeAllFields: false,
	}
}

// ParseAttributeMapping 解析属性映射配置
func ParseAttributeMapping(mappingJSON string) (CASAttributeMapping, error) {
	if mappingJSON == "" {
		return GetDefaultCASAttributeMapping(), nil
	}

	var mapping CASAttributeMapping
	if err := json.Unmarshal([]byte(mappingJSON), &mapping); err != nil {
		return GetDefaultCASAttributeMapping(), fmt.Errorf("failed to parse attribute mapping: %w", err)
	}

	// 如果DefaultAttributes为空，使用默认值
	if mapping.DefaultAttributes == nil {
		mapping.DefaultAttributes = GetDefaultCASAttributeMapping().DefaultAttributes
	}

	// 如果CustomAttributes为空，初始化为空map
	if mapping.CustomAttributes == nil {
		mapping.CustomAttributes = make(map[string]string)
	}

	return mapping, nil
}

// BuildUserAttributes 根据映射配置构建用户属性
func BuildUserAttributes(user *models.User, mapping CASAttributeMapping) map[string]interface{} {
	attributes := make(map[string]interface{})

	// 添加默认属性
	for casAttr, userField := range mapping.DefaultAttributes {
		value := getUserFieldValue(user, userField)
		if value != nil {
			attributes[casAttr] = value
		}
	}

	// 添加自定义属性
	for casAttr, userField := range mapping.CustomAttributes {
		value := getUserFieldValue(user, userField)
		if value != nil {
			attributes[casAttr] = value
		}
	}

	// 如果配置了包含所有字段，添加其他用户字段
	if mapping.IncludeAllFields {
		// 添加其他可能的用户字段
		additionalFields := map[string]interface{}{
			"phone":          user.Phone,
			"status":         user.Status,
			"createdAt":      user.CreatedAt,
			"lastLoginAt":    user.LastLoginAt,
			"organizationId": user.OrganizationID,
		}

		for field, value := range additionalFields {
			if value != nil && value != "" {
				// 检查是否已经存在，避免重复
				if _, exists := attributes[field]; !exists {
					attributes[field] = value
				}
			}
		}
	}

	return attributes
}

// getUserFieldValue 根据字段名获取用户字段值
func getUserFieldValue(user *models.User, field string) interface{} {
	switch strings.ToLower(field) {
	case "id", "user_id":
		return user.ID
	case "username", "user_name":
		return user.Username
	case "email", "email_address":
		return user.Email
	case "display_name", "displayname", "full_name":
		return user.DisplayName
	case "phone", "phone_number":
		return user.Phone
	case "status", "user_status":
		return user.Status
	case "created_at", "createdat":
		return user.CreatedAt
	case "updated_at", "updatedat":
		return user.UpdatedAt
	case "last_login_at", "lastloginat":
		return user.LastLoginAt
	default:
		// 对于未知字段，返回nil
		return nil
	}
}

// FormatAttributesForCAS 格式化属性为CAS XML格式
func FormatAttributesForCAS(attributes map[string]interface{}) string {
	if len(attributes) == 0 {
		return ""
	}

	var parts []string
	for key, value := range attributes {
		if value != nil {
			parts = append(parts, fmt.Sprintf("            <cas:%s>%v</cas:%s>", key, value, key))
		}
	}

	if len(parts) == 0 {
		return ""
	}

	return fmt.Sprintf(`
        <cas:attributes>
%s
        </cas:attributes>`, strings.Join(parts, "\n"))
}

// FormatAttributesForJSON 格式化属性为JSON格式
func FormatAttributesForJSON(attributes map[string]interface{}) map[string]interface{} {
	// 过滤掉nil值
	result := make(map[string]interface{})
	for key, value := range attributes {
		if value != nil {
			result[key] = value
		}
	}
	return result
}
