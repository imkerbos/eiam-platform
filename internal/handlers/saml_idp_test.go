package handlers

import (
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"eiam-platform/config"
)

// 设置测试环境
func setupSAMLTest() {
	gin.SetMode(gin.TestMode)

	// 设置测试配置
	config.AppConfig = &config.Config{
		IdP: config.IdPConfig{
			BaseURL: "http://localhost:3000",
		},
	}

	// 初始化SAML IdP
	InitSAMLIDP()
}

// TestSAMLMetadataHandler 测试SAML元数据处理器
func TestSAMLMetadataHandler(t *testing.T) {
	setupSAMLTest()

	// 创建测试请求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req, _ := http.NewRequest("GET", "/saml/metadata", nil)
	c.Request = req

	// 调用处理器
	SAMLMetadataHandlerIDP(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/xml", w.Header().Get("Content-Type"))

	// 验证XML格式
	var metadata interface{}
	err := xml.Unmarshal(w.Body.Bytes(), &metadata)
	assert.NoError(t, err, "Response should be valid XML")

	// 验证包含必要的SAML元素
	responseBody := w.Body.String()
	assert.Contains(t, responseBody, "EntityDescriptor")
	assert.Contains(t, responseBody, "IDPSSODescriptor")
	assert.Contains(t, responseBody, "SingleSignOnService")
	assert.Contains(t, responseBody, "SingleLogoutService")
	assert.Contains(t, responseBody, "http://localhost:3000/saml/sso")
	assert.Contains(t, responseBody, "http://localhost:3000/saml/sls")

	t.Logf("SAML Metadata Response: %s", responseBody)
}

// TestSAMLSSOHandler 测试SAML SSO处理器
func TestSAMLSSOHandler(t *testing.T) {
	setupSAMLTest()

	tests := []struct {
		name           string
		method         string
		samlRequest    string
		relayState     string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Missing SAML Request",
			method:         "GET",
			samlRequest:    "",
			relayState:     "",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "SAML IdP not initialized",
		},
		{
			name:           "GET with SAML Request",
			method:         "GET",
			samlRequest:    "PHNhbWxwOkF1dGhuUmVxdWVzdA==", // Base64编码的示例
			relayState:     "test-relay-state",
			expectedStatus: http.StatusOK,
			expectedBody:   "", // 这个会是HTML响应或重定向
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			var req *http.Request
			if tt.method == "GET" {
				url := "/saml/sso"
				if tt.samlRequest != "" {
					url += "?SAMLRequest=" + tt.samlRequest + "&RelayState=" + tt.relayState
				}
				req, _ = http.NewRequest("GET", url, nil)
			} else {
				body := strings.NewReader("SAMLRequest=" + tt.samlRequest + "&RelayState=" + tt.relayState)
				req, _ = http.NewRequest("POST", "/saml/sso", body)
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}

			c.Request = req

			// 调用处理器
			SAMLSSOHandlerIDP(c)

			// 验证响应状态
			if tt.expectedStatus == http.StatusBadRequest {
				assert.Equal(t, tt.expectedStatus, w.Code)
			} else {
				// SSO可能返回重定向或HTML表单
				assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusFound)
			}

			t.Logf("SSO Response Status: %d, Body: %s", w.Code, w.Body.String())
		})
	}
}

// TestSAMLSLSHandler 测试SAML SLS处理器
func TestSAMLSLSHandler(t *testing.T) {
	setupSAMLTest()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 创建带有SAML注销请求的测试
	req, _ := http.NewRequest("GET", "/saml/sls?SAMLRequest=PHNhbWxwOkxvZ291dFJlcXVlc3Q=&RelayState=test", nil)
	c.Request = req

	// 调用处理器
	SAMLSLSHandlerIDP(c)

	// SLS处理器应该返回成功状态或重定向
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusFound || w.Code == http.StatusBadRequest)

	t.Logf("SLS Response Status: %d, Body: %s", w.Code, w.Body.String())
}

// TestGenerateSAMLKeyPair 测试证书生成
func TestGenerateSAMLKeyPair(t *testing.T) {
	keyPair, err := generateSAMLKeyPair()

	require.NoError(t, err, "Should generate key pair without error")
	require.NotNil(t, keyPair.PrivateKey, "Private key should not be nil")
	require.NotNil(t, keyPair.Leaf, "Certificate should not be nil")

	// 验证证书字段
	cert := keyPair.Leaf
	assert.Equal(t, "EIAM SAML IdP", cert.Subject.CommonName)
	assert.Contains(t, cert.Subject.Organization, "EIAM Platform")
	assert.Contains(t, cert.Subject.OrganizationalUnit, "Identity Provider")

	// 验证证书有效期
	assert.True(t, cert.NotAfter.After(cert.NotBefore))

	t.Logf("Certificate Subject: %+v", cert.Subject)
	t.Logf("Certificate Valid From: %v To: %v", cert.NotBefore, cert.NotAfter)
}

// BenchmarkSAMLMetadata 性能测试：元数据生成
func BenchmarkSAMLMetadata(b *testing.B) {
	setupSAMLTest()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req, _ := http.NewRequest("GET", "/saml/metadata", nil)
		c.Request = req

		SAMLMetadataHandlerIDP(c)
	}
}

// TestSAMLInitialization 测试SAML初始化
func TestSAMLInitialization(t *testing.T) {
	// 测试正常初始化
	config.AppConfig = &config.Config{
		IdP: config.IdPConfig{
			BaseURL: "http://localhost:3000",
		},
	}

	err := InitSAMLIDP()
	assert.NoError(t, err, "SAML IdP should initialize successfully")
	assert.NotNil(t, samlIDP, "SAML IdP instance should be created")

	// 测试无效URL
	config.AppConfig.IdP.BaseURL = "invalid-url"
	err = InitSAMLIDP()
	assert.Error(t, err, "Should fail with invalid URL")

	// 测试空配置
	config.AppConfig = nil
	err = InitSAMLIDP()
	assert.NoError(t, err, "Should work with default configuration")
}
