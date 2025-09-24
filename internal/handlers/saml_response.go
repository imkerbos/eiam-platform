package handlers

import (
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"eiam-platform/internal/models"
	"eiam-platform/pkg/database"
	"eiam-platform/pkg/logger"
	"eiam-platform/pkg/utils"
)

// createSAMLResponseXML 创建SAML响应XML（简化版本）
func createSAMLResponseXML(user *models.User, app *models.Application) (string, error) {
	now := time.Now()

	// 生成唯一ID
	responseID := "_" + utils.GenerateTradeIDString("response")
	assertionID := "_" + utils.GenerateTradeIDString("assertion")

	// 时间格式化
	issueInstant := now.UTC().Format("2006-01-02T15:04:05Z")
	notBefore := now.Add(-5 * time.Minute).UTC().Format("2006-01-02T15:04:05Z")
	notOnOrAfter := now.Add(5 * time.Minute).UTC().Format("2006-01-02T15:04:05Z")
	authnInstant := now.UTC().Format("2006-01-02T15:04:05Z")

	// 获取IdP信息
	var issuerURL string
	if samlIDP != nil {
		issuerURL = samlIDP.IDP.MetadataURL.String()
	} else {
		issuerURL = "http://localhost:3000/saml/metadata"
	}

	// 获取用户角色
	var userRoles []models.Role
	var roleAttributes strings.Builder
	if err := database.DB.Table("user_roles").
		Select("roles.*").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", user.ID).
		Find(&userRoles).Error; err == nil {

		for _, role := range userRoles {
			roleAttributes.WriteString(fmt.Sprintf(`
			<saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">%s</saml:AttributeValue>`, role.Name))
		}
	}

	// 构建属性语句
	var attributeStatement string
	if user.Email != "" || user.DisplayName != "" || roleAttributes.Len() > 0 {
		attributeStatement = `
		<saml:AttributeStatement>`

		// 邮箱属性
		if user.Email != "" {
			attributeStatement += fmt.Sprintf(`
			<saml:Attribute Name="http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:uri">
				<saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">%s</saml:AttributeValue>
			</saml:Attribute>`, user.Email)
		}

		// 显示名称属性
		if user.DisplayName != "" {
			attributeStatement += fmt.Sprintf(`
			<saml:Attribute Name="http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:uri">
				<saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">%s</saml:AttributeValue>
			</saml:Attribute>`, user.DisplayName)
		}

		// 用户名属性
		attributeStatement += fmt.Sprintf(`
		<saml:Attribute Name="http://schemas.xmlsoap.org/ws/2005/05/identity/claims/nameidentifier" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:uri">
			<saml:AttributeValue xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="xs:string">%s</saml:AttributeValue>
		</saml:Attribute>`, user.Username)

		// 角色属性
		if roleAttributes.Len() > 0 {
			attributeStatement += fmt.Sprintf(`
			<saml:Attribute Name="http://schemas.microsoft.com/ws/2008/06/identity/claims/role" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:uri">%s
			</saml:Attribute>`, roleAttributes.String())
		}

		attributeStatement += `
		</saml:AttributeStatement>`
	}

	// 构建完整的SAML响应XML
	samlResponseXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" 
				xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" 
				ID="%s" 
				Version="2.0" 
				IssueInstant="%s" 
				Destination="%s">
	<saml:Issuer>%s</saml:Issuer>
	<samlp:Status>
		<samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/>
	</samlp:Status>
	<saml:Assertion xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" 
					xmlns:xs="http://www.w3.org/2001/XMLSchema" 
					ID="%s" 
					Version="2.0" 
					IssueInstant="%s">
		<saml:Issuer>%s</saml:Issuer>
		<saml:Subject>
			<saml:NameID Format="urn:oasis:names:tc:SAML:2.0:nameid-format:persistent">%s</saml:NameID>
			<saml:SubjectConfirmation Method="urn:oasis:names:tc:SAML:2.0:cm:bearer">
				<saml:SubjectConfirmationData NotOnOrAfter="%s" Recipient="%s"/>
			</saml:SubjectConfirmation>
		</saml:Subject>
		<saml:Conditions NotBefore="%s" NotOnOrAfter="%s">
			<saml:AudienceRestriction>
				<saml:Audience>%s</saml:Audience>
			</saml:AudienceRestriction>
		</saml:Conditions>
		<saml:AuthnStatement AuthnInstant="%s">
			<saml:AuthnContext>
				<saml:AuthnContextClassRef>urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport</saml:AuthnContextClassRef>
			</saml:AuthnContext>
		</saml:AuthnStatement>%s
	</saml:Assertion>
</samlp:Response>`,
		responseID,         // Response ID
		issueInstant,       // Response IssueInstant
		app.AcsURL,         // Destination
		issuerURL,          // Response Issuer
		assertionID,        // Assertion ID
		issueInstant,       // Assertion IssueInstant
		issuerURL,          // Assertion Issuer
		user.Username,      // NameID
		notOnOrAfter,       // SubjectConfirmationData NotOnOrAfter
		app.AcsURL,         // SubjectConfirmationData Recipient
		notBefore,          // Conditions NotBefore
		notOnOrAfter,       // Conditions NotOnOrAfter
		app.EntityID,       // Audience
		authnInstant,       // AuthnStatement AuthnInstant
		attributeStatement, // AttributeStatement
	)

	logger.Info("SAML response XML created",
		zap.String("response_id", responseID),
		zap.String("assertion_id", assertionID),
		zap.String("username", user.Username),
		zap.String("audience", app.EntityID),
	)

	return samlResponseXML, nil
}
