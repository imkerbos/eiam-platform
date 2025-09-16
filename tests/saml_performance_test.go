package tests

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// BenchmarkSAMLMetadataEndpoint 性能测试：元数据端点
func BenchmarkSAMLMetadataEndpoint(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark tests in short mode")
	}

	// 确保服务器运行
	ensureServerRunning(b)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		client := &http.Client{Timeout: 10 * time.Second}
		for pb.Next() {
			resp, err := client.Get(BaseURL + "/saml/metadata")
			if err != nil {
				b.Error("Failed to get metadata:", err)
				continue
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				b.Errorf("Expected status 200, got %d", resp.StatusCode)
			}
		}
	})
}

// BenchmarkSAMLSSOEndpoint 性能测试：SSO端点
func BenchmarkSAMLSSOEndpoint(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping benchmark tests in short mode")
	}

	ensureServerRunning(b)

	// 准备SAML请求
	samlRequest := createMockSAMLRequest()
	encodedRequest := base64.StdEncoding.EncodeToString([]byte(samlRequest))
	ssoURL := fmt.Sprintf("%s/saml/sso?SAMLRequest=%s&RelayState=bench-test",
		BaseURL, url.QueryEscape(encodedRequest))

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		client := &http.Client{Timeout: 10 * time.Second}
		for pb.Next() {
			resp, err := client.Get(ssoURL)
			if err != nil {
				b.Error("Failed to access SSO:", err)
				continue
			}
			resp.Body.Close()
		}
	})
}

// TestSAMLConcurrentAccess 并发访问测试
func TestSAMLConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent tests in short mode")
	}

	ensureServerRunning(t)

	const numGoroutines = 50
	const requestsPerGoroutine = 10

	var wg sync.WaitGroup
	results := make(chan TestResult, numGoroutines*requestsPerGoroutine)

	// 启动多个goroutine并发访问
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			client := &http.Client{Timeout: 10 * time.Second}

			for j := 0; j < requestsPerGoroutine; j++ {
				start := time.Now()
				resp, err := client.Get(BaseURL + "/saml/metadata")
				duration := time.Since(start)

				result := TestResult{
					GoroutineID: goroutineID,
					RequestID:   j,
					Duration:    duration,
					Success:     err == nil && resp != nil,
				}

				if resp != nil {
					result.StatusCode = resp.StatusCode
					resp.Body.Close()
				}

				results <- result
			}
		}(i)
	}

	// 等待所有goroutine完成
	wg.Wait()
	close(results)

	// 分析结果
	var (
		totalRequests   int
		successRequests int
		totalDuration   time.Duration
		maxDuration     time.Duration
		minDuration     time.Duration = time.Hour
	)

	for result := range results {
		totalRequests++
		if result.Success && result.StatusCode == http.StatusOK {
			successRequests++
		}

		totalDuration += result.Duration
		if result.Duration > maxDuration {
			maxDuration = result.Duration
		}
		if result.Duration < minDuration {
			minDuration = result.Duration
		}
	}

	// 验证结果
	successRate := float64(successRequests) / float64(totalRequests) * 100
	avgDuration := totalDuration / time.Duration(totalRequests)

	assert.Greater(t, successRate, 95.0, "Success rate should be > 95%")
	assert.Less(t, avgDuration, 2*time.Second, "Average response time should be < 2s")

	t.Logf("Concurrent Test Results:")
	t.Logf("  Total Requests: %d", totalRequests)
	t.Logf("  Successful Requests: %d (%.2f%%)", successRequests, successRate)
	t.Logf("  Average Duration: %v", avgDuration)
	t.Logf("  Min Duration: %v", minDuration)
	t.Logf("  Max Duration: %v", maxDuration)
}

// TestSAMLLoadTest 负载测试
func TestSAMLLoadTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load tests in short mode")
	}

	ensureServerRunning(t)

	// 负载测试参数
	const (
		testDuration   = 30 * time.Second
		maxConcurrency = 20
	)

	var (
		requestCount int64
		errorCount   int64
		mu           sync.Mutex
		wg           sync.WaitGroup
	)

	// 启动时间
	startTime := time.Now()
	endTime := startTime.Add(testDuration)

	// 启动多个worker
	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			client := &http.Client{Timeout: 5 * time.Second}

			for time.Now().Before(endTime) {
				resp, err := client.Get(BaseURL + "/saml/metadata")

				mu.Lock()
				requestCount++
				if err != nil || (resp != nil && resp.StatusCode != http.StatusOK) {
					errorCount++
				}
				mu.Unlock()

				if resp != nil {
					resp.Body.Close()
				}

				// 短暂休息避免过度负载
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// 计算结果
	actualDuration := time.Since(startTime)
	rps := float64(requestCount) / actualDuration.Seconds()
	errorRate := float64(errorCount) / float64(requestCount) * 100

	t.Logf("Load Test Results:")
	t.Logf("  Test Duration: %v", actualDuration)
	t.Logf("  Total Requests: %d", requestCount)
	t.Logf("  Error Count: %d (%.2f%%)", errorCount, errorRate)
	t.Logf("  Requests per Second: %.2f", rps)

	// 验证性能指标
	assert.Greater(t, rps, 10.0, "Should handle > 10 requests per second")
	assert.Less(t, errorRate, 5.0, "Error rate should be < 5%")
}

// TestResult 测试结果结构
type TestResult struct {
	GoroutineID int
	RequestID   int
	Duration    time.Duration
	StatusCode  int
	Success     bool
}

// ensureServerRunning 确保服务器运行
func ensureServerRunning(t testing.TB) {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(BaseURL + "/health")
	if err != nil {
		t.Fatalf("Server is not running: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Server health check failed: status %d", resp.StatusCode)
	}
}
