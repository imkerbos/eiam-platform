package handlers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/xml"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlidp"

	"eiam-platform/config"
	"eiam-platform/internal/models"
	"eiam-platform/pkg/database"
	"eiam-platform/pkg/logger"
	"eiam-platform/pkg/utils"
)

var (
	samlIDP *samlidp.Server
)

// InitSAMLIDP 初始化SAML身份提供商
func InitSAMLIDP() error {
	// 生成或加载密钥对
	keyPair, err := generateSAMLKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate SAML key pair: %w", err)
	}

	// 获取基础URL - 从配置读取，如果配置不可用则使用默认值
	var baseURLStr string
	if config.AppConfig != nil && config.AppConfig.IdP.BaseURL != "" {
		baseURLStr = config.AppConfig.IdP.BaseURL
	} else {
		baseURLStr = "http://localhost:3000" // 默认使用前端代理地址
	}

	baseURL, err := url.Parse(baseURLStr)
	if err != nil {
		return fmt.Errorf("failed to parse base URL: %w", err)
	}

	// 创建SAML IdP服务器 - 简化版本
	samlIDP = &samlidp.Server{
		IDP: saml.IdentityProvider{
			Key:         keyPair.PrivateKey.(*rsa.PrivateKey),
			Certificate: keyPair.Leaf,
			MetadataURL: *baseURL.ResolveReference(&url.URL{Path: "/saml/metadata"}),
			SSOURL:      *baseURL.ResolveReference(&url.URL{Path: "/saml/sso"}),
			LogoutURL:   *baseURL.ResolveReference(&url.URL{Path: "/saml/sls"}),
		},
	}

	logger.Info("SAML IdP initialized successfully with crewjam/saml library")
	return nil
}

// generateSAMLKeyPair 生成SAML密钥对
func generateSAMLKeyPair() (tls.Certificate, error) {
	// 生成私钥
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}

	// 创建证书模板
	serialNumber, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{"EIAM Platform"},
			OrganizationalUnit: []string{"Identity Provider"},
			Country:            []string{"CN"},
			Province:           []string{"Beijing"},
			Locality:           []string{"Beijing"},
			CommonName:         "EIAM SAML IdP",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1年有效期
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              []string{"localhost", "eiam.local"},
	}

	// 生成证书
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return tls.Certificate{}, err
	}

	// 创建tls.Certificate
	keyPair := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  privateKey,
		Leaf:        cert,
	}

	logger.Info("SAML certificate generated successfully",
		zap.String("serial", cert.SerialNumber.String()),
		zap.Time("not_before", cert.NotBefore),
		zap.Time("not_after", cert.NotAfter),
	)

	return keyPair, nil
}

// TODO: 自定义会话和存储实现将在后续版本中添加
// 当前使用crewjam/saml的默认实现

/*
// CustomSessionProvider 自定义会话提供商 - 待实现
type CustomSessionProvider struct{}

// CustomStore 自定义存储 - 待实现
type CustomStore struct{}
*/

// SAMLMetadataHandlerIDP SAML元数据处理器 (使用crewjam/saml)
func SAMLMetadataHandlerIDP(c *gin.Context) {
	if samlIDP == nil {
		logger.Error("SAML IdP not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "SAML IdP not initialized",
		})
		return
	}

	// 使用crewjam/saml生成元数据
	metadata := samlIDP.IDP.Metadata()
	metadataXML, err := xml.MarshalIndent(metadata, "", "  ")
	if err != nil {
		logger.ErrorError("Failed to marshal SAML metadata", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to generate metadata",
		})
		return
	}

	c.Header("Content-Type", "application/xml")
	c.String(http.StatusOK, `<?xml version="1.0" encoding="UTF-8"?>`+"\n"+string(metadataXML))

	logger.Info("SAML metadata served using crewjam/saml", zap.String("ip", c.ClientIP()))
}

// SAMLSSOHandlerIDP SAML单点登录处理器 (使用crewjam/saml)
func SAMLSSOHandlerIDP(c *gin.Context) {
	if samlIDP == nil {
		logger.Error("SAML IdP not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "SAML IdP not initialized",
		})
		return
	}

	// 将gin.Context转换为http.ResponseWriter和*http.Request
	w := c.Writer
	r := c.Request

	// 使用crewjam/saml处理SSO请求
	samlIDP.ServeHTTP(w, r)

	logger.Info("SAML SSO request handled by crewjam/saml",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("ip", c.ClientIP()),
	)
}

// SAMLSLSHandlerIDP SAML单点注销处理器 (使用crewjam/saml)
func SAMLSLSHandlerIDP(c *gin.Context) {
	if samlIDP == nil {
		logger.Error("SAML IdP not initialized")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "SAML IdP not initialized",
		})
		return
	}

	// 将gin.Context转换为http.ResponseWriter和*http.Request
	w := c.Writer
	r := c.Request

	// 使用crewjam/saml处理SLS请求
	samlIDP.ServeHTTP(w, r)

	logger.Info("SAML SLS request handled by crewjam/saml",
		zap.String("method", r.Method),
		zap.String("url", r.URL.String()),
		zap.String("ip", c.ClientIP()),
	)
}

// checkUserAuthenticationFromRequest 从HTTP请求检查用户认证状态
func checkUserAuthenticationFromRequest(r *http.Request) (*models.User, bool) {
	// 尝试JWT token认证
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		jwtManager := &utils.JWTManager{}
		if claims, err := jwtManager.ValidateAccessToken(token); err == nil {
			var user models.User
			if err := database.DB.Where("id = ?", claims.UserID).First(&user).Error; err == nil {
				return &user, true
			}
		}
	}

	// 可以添加其他认证方式，如session等
	// TODO: 添加基于Cookie的会话认证

	return nil, false
}
