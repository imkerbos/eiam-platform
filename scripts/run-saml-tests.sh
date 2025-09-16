#!/bin/bash

# SAML测试运行脚本
# 用法: ./scripts/run-saml-tests.sh [unit|integration|performance|all]

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查依赖
check_dependencies() {
    log_info "检查测试依赖..."
    
    # 检查go
    if ! command -v go &> /dev/null; then
        log_error "Go未安装"
        exit 1
    fi
    
    # 检查测试包
    if ! go list github.com/stretchr/testify/assert &> /dev/null; then
        log_info "安装测试依赖..."
        go get github.com/stretchr/testify/assert
        go get github.com/stretchr/testify/require
    fi
    
    log_success "依赖检查完成"
}

# 检查服务器状态
check_server() {
    log_info "检查服务器状态..."
    
    if curl -s http://localhost:8080/health > /dev/null; then
        log_success "后端服务器运行正常 (8080)"
    else
        log_warning "后端服务器未运行，请先启动: go run cmd/server/main.go"
    fi
    
    if curl -s http://localhost:3000/health > /dev/null; then
        log_success "前端代理运行正常 (3000)"
    else
        log_warning "前端代理未运行，请先启动: cd frontend && npm run dev"
    fi
}

# 运行单元测试
run_unit_tests() {
    log_info "运行SAML单元测试..."
    
    echo "========================================="
    echo "           SAML 单元测试"
    echo "========================================="
    
    if go test -v ./internal/handlers -run "TestSAML" -timeout 30s; then
        log_success "单元测试通过"
    else
        log_error "单元测试失败"
        return 1
    fi
}

# 运行集成测试
run_integration_tests() {
    log_info "运行SAML集成测试..."
    
    echo "========================================="
    echo "           SAML 集成测试"
    echo "========================================="
    
    if go test -v ./tests -run "TestSAML.*Integration" -timeout 60s; then
        log_success "集成测试通过"
    else
        log_error "集成测试失败"
        return 1
    fi
}

# 运行性能测试
run_performance_tests() {
    log_info "运行SAML性能测试..."
    
    echo "========================================="
    echo "           SAML 性能测试"
    echo "========================================="
    
    # 运行基准测试
    log_info "运行基准测试..."
    go test -v ./tests -run "^$" -bench "BenchmarkSAML" -benchtime=10s -timeout 120s
    
    # 运行并发测试
    log_info "运行并发测试..."
    if go test -v ./tests -run "TestSAMLConcurrentAccess" -timeout 120s; then
        log_success "并发测试通过"
    else
        log_error "并发测试失败"
        return 1
    fi
    
    # 运行负载测试
    log_info "运行负载测试..."
    if go test -v ./tests -run "TestSAMLLoadTest" -timeout 120s; then
        log_success "负载测试通过"
    else
        log_error "负载测试失败"
        return 1
    fi
}

# 运行工作流测试
run_workflow_tests() {
    log_info "运行SAML工作流测试..."
    
    echo "========================================="
    echo "           SAML 工作流测试"
    echo "========================================="
    
    if go test -v ./tests -run "TestSAMLWorkflow" -timeout 60s; then
        log_success "工作流测试通过"
    else
        log_error "工作流测试失败"
        return 1
    fi
}

# 生成测试报告
generate_report() {
    log_info "生成测试报告..."
    
    local report_dir="test-reports"
    mkdir -p $report_dir
    
    # 生成覆盖率报告
    log_info "生成代码覆盖率报告..."
    go test -coverprofile=$report_dir/saml-coverage.out ./internal/handlers -run "TestSAML"
    go tool cover -html=$report_dir/saml-coverage.out -o $report_dir/saml-coverage.html
    
    log_success "测试报告已生成: $report_dir/saml-coverage.html"
}

# 清理测试环境
cleanup() {
    log_info "清理测试环境..."
    # 这里可以添加清理逻辑
    log_success "清理完成"
}

# 主函数
main() {
    local test_type=${1:-"all"}
    
    echo "========================================="
    echo "           SAML 测试套件"
    echo "========================================="
    
    # 检查依赖
    check_dependencies
    
    # 检查服务器
    check_server
    
    case $test_type in
        "unit")
            run_unit_tests
            ;;
        "integration")
            run_integration_tests
            ;;
        "performance")
            run_performance_tests
            ;;
        "workflow")
            run_workflow_tests
            ;;
        "all")
            log_info "运行所有SAML测试..."
            
            # 按顺序运行所有测试
            run_unit_tests || exit 1
            run_integration_tests || exit 1
            run_workflow_tests || exit 1
            run_performance_tests || exit 1
            
            # 生成报告
            generate_report
            
            log_success "所有SAML测试完成！"
            ;;
        *)
            echo "用法: $0 [unit|integration|performance|workflow|all]"
            echo ""
            echo "测试类型:"
            echo "  unit        - 运行单元测试"
            echo "  integration - 运行集成测试"
            echo "  performance - 运行性能测试"
            echo "  workflow    - 运行工作流测试"
            echo "  all         - 运行所有测试 (默认)"
            exit 1
            ;;
    esac
}

# 捕获退出信号进行清理
trap cleanup EXIT

# 运行主函数
main "$@"
