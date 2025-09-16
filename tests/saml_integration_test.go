package tests

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	BaseURL    = "http://localhost:3000"
	BackendURL = "http://localhost:8080"
)

// TestSAMLEndpointsIntegration 集成测试：测试所有SAML端点
func TestSAMLEndpointsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// 等待服务器启动
	waitForServer(t, BaseURL)

	t.Run("Metadata Endpoint", testMetadataEndpoint)
	t.Run("SSO Endpoint", testSSOEndpoint)
	t.Run("SLS Endpoint", testSLSEndpoint)
}

// testMetadataEndpoint 测试元数据端点
func testMetadataEndpoint(t *testing.T) {
	// 测试通过前端代理访问
	resp, err := http.Get(BaseURL + "/saml/metadata")
	require.NoError(t, err, "Should be able to access metadata endpoint")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Metadata endpoint should return 200")
	assert.Equal(t, "application/xml", resp.Header.Get("Content-Type"), "Should return XML content type")

	// 读取响应内容
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Should be able to read response body")

	// 验证XML格式
	var metadata interface{}
	err = xml.Unmarshal(body, &metadata)
	assert.NoError(t, err, "Response should be valid XML")

	// 验证SAML元数据内容
	metadataStr := string(body)
	assert.Contains(t, metadataStr, "EntityDescriptor", "Should contain EntityDescriptor")
	assert.Contains(t, metadataStr, "IDPSSODescriptor", "Should contain IDPSSODescriptor")
	assert.Contains(t, metadataStr, "SingleSignOnService", "Should contain SingleSignOnService")
	assert.Contains(t, metadataStr, "SingleLogoutService", "Should contain SingleLogoutService")
	assert.Contains(t, metadataStr, BaseURL+"/saml/sso", "Should contain correct SSO URL")
	assert.Contains(t, metadataStr, BaseURL+"/saml/sls", "Should contain correct SLS URL")

	t.Logf("Metadata XML length: %d bytes", len(body))
}

// testSSOEndpoint 测试SSO端点
func testSSOEndpoint(t *testing.T) {
	// 测试没有SAML请求的情况
	resp, err := http.Get(BaseURL + "/saml/sso")
	require.NoError(t, err, "Should be able to access SSO endpoint")
	defer resp.Body.Close()

	// SSO端点没有SAML请求时应该返回错误或重定向
	assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusFound,
		"SSO without SAML request should return 400 or redirect")

	// 测试带有模拟SAML请求
	samlRequest := createMockSAMLRequest()
	encodedRequest := base64.StdEncoding.EncodeToString([]byte(samlRequest))

	ssoURL := fmt.Sprintf("%s/saml/sso?SAMLRequest=%s&RelayState=test-state",
		BaseURL, url.QueryEscape(encodedRequest))

	resp2, err := http.Get(ssoURL)
	require.NoError(t, err, "Should be able to access SSO with SAML request")
	defer resp2.Body.Close()

	// 应该返回登录页面或处理请求
	assert.True(t, resp2.StatusCode == http.StatusOK || resp2.StatusCode == http.StatusFound,
		"SSO with SAML request should return 200 or redirect")

	t.Logf("SSO Response Status: %d", resp2.StatusCode)
}

// testSLSEndpoint 测试SLS端点
func testSLSEndpoint(t *testing.T) {
	// 测试没有SAML请求的情况
	resp, err := http.Get(BaseURL + "/saml/sls")
	require.NoError(t, err, "Should be able to access SLS endpoint")
	defer resp.Body.Close()

	// SLS端点没有SAML请求时应该返回错误
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "SLS without SAML request should return 400")

	// 测试带有模拟SAML注销请求
	logoutRequest := createMockSAMLLogoutRequest()
	encodedRequest := base64.StdEncoding.EncodeToString([]byte(logoutRequest))

	slsURL := fmt.Sprintf("%s/saml/sls?SAMLRequest=%s&RelayState=test-state",
		BaseURL, url.QueryEscape(encodedRequest))

	resp2, err := http.Get(slsURL)
	require.NoError(t, err, "Should be able to access SLS with SAML request")
	defer resp2.Body.Close()

	// 应该处理注销请求
	assert.True(t, resp2.StatusCode == http.StatusOK || resp2.StatusCode == http.StatusBadRequest,
		"SLS with SAML request should return 200 or 400")

	t.Logf("SLS Response Status: %d", resp2.StatusCode)
}

// TestSAMLProxyIntegration 测试前端代理功能
func TestSAMLProxyIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	waitForServer(t, BaseURL)

	// 测试代理是否正确转发请求
	endpoints := []string{
		"/saml/metadata",
		"/saml/sso",
		"/saml/sls",
	}

	for _, endpoint := range endpoints {
		t.Run("Proxy "+endpoint, func(t *testing.T) {
			// 通过前端代理访问
			proxyResp, err := http.Get(BaseURL + endpoint)
			require.NoError(t, err, "Should access via proxy")
			defer proxyResp.Body.Close()

			// 直接访问后端（如果可用）
			backendResp, err := http.Get(BackendURL + endpoint)
			if err == nil {
				defer backendResp.Body.Close()

				// 比较响应状态码（可能不完全相同，但都应该是有效响应）
				assert.True(t, proxyResp.StatusCode >= 200 && proxyResp.StatusCode < 500,
					"Proxy should return valid HTTP status")
				assert.True(t, backendResp.StatusCode >= 200 && backendResp.StatusCode < 500,
					"Backend should return valid HTTP status")
			}

			t.Logf("Endpoint %s - Proxy Status: %d", endpoint, proxyResp.StatusCode)
		})
	}
}

// TestSAMLWorkflow 测试完整的SAML工作流程
func TestSAMLWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	waitForServer(t, BaseURL)

	// 1. 获取IdP元数据
	t.Run("Step 1: Get IdP Metadata", func(t *testing.T) {
		resp, err := http.Get(BaseURL + "/saml/metadata")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		// 元数据应该包含必要信息
		metadata := string(body)
		assert.Contains(t, metadata, "EntityDescriptor")
		assert.Contains(t, metadata, BaseURL+"/saml/sso")
	})

	// 2. 模拟SP发起SAML请求
	t.Run("Step 2: SP Initiated SAML Request", func(t *testing.T) {
		samlRequest := createMockSAMLRequest()
		encodedRequest := base64.StdEncoding.EncodeToString([]byte(samlRequest))

		ssoURL := fmt.Sprintf("%s/saml/sso?SAMLRequest=%s&RelayState=test-workflow",
			BaseURL, url.QueryEscape(encodedRequest))

		resp, err := http.Get(ssoURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		// 应该处理请求（可能重定向到登录或返回错误）
		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 500)
	})
}

// 辅助函数

// waitForServer 等待服务器启动
func waitForServer(t *testing.T, baseURL string) {
	timeout := time.After(30 * time.Second)
	tick := time.Tick(500 * time.Millisecond)

	for {
		select {
		case <-timeout:
			t.Fatal("Server did not start within 30 seconds")
		case <-tick:
			resp, err := http.Get(baseURL + "/health")
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return
				}
			}
		}
	}
}

// createMockSAMLRequest 创建模拟的SAML认证请求
func createMockSAMLRequest() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<samlp:AuthnRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"
                    xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"
                    ID="test-request-id"
                    Version="2.0"
                    IssueInstant="2025-09-16T09:00:00Z"
                    AssertionConsumerServiceURL="http://localhost:3001/saml/acs"
                    Destination="` + BaseURL + `/saml/sso">
    <saml:Issuer>http://localhost:3001</saml:Issuer>
    <samlp:NameIDPolicy Format="urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress" AllowCreate="true"/>
</samlp:AuthnRequest>`
}

// createMockSAMLLogoutRequest 创建模拟的SAML注销请求
func createMockSAMLLogoutRequest() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<samlp:LogoutRequest xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol"
                     xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion"
                     ID="test-logout-id"
                     Version="2.0"
                     IssueInstant="2025-09-16T09:00:00Z"
                     Destination="` + BaseURL + `/saml/sls">
    <saml:Issuer>http://localhost:3001</saml:Issuer>
    <saml:NameID Format="urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress">test@example.com</saml:NameID>
</samlp:LogoutRequest>`
}
