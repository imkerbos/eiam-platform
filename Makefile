# EIAM Platform Makefile

.PHONY: help build run migrate test clean

# 默认目标
help:
	@echo "EIAM Platform 管理命令:"
	@echo "  build     - 构建项目"
	@echo "  run       - 运行服务器"
	@echo "  migrate   - 运行数据库迁移"
	@echo "  test      - 运行测试"
	@echo "  clean     - 清理构建文件"

# 构建项目
build:
	@echo "构建 EIAM Platform..."
	go build -o build/eiam-platform cmd/server/main.go
	@echo "构建完成: build/eiam-platform"

# 运行服务器
run:
	@echo "启动 EIAM Platform 服务器..."
	go run cmd/server/main.go

# 运行数据库迁移
migrate:
	@echo "运行数据库迁移..."
	go run cmd/migrate/main.go

# 运行数据库种子
seed:
	@echo "运行数据库种子..."
	go run cmd/seed/main.go

# 运行测试
test:
	@echo "运行测试..."
	go test ./...

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -rf build/
	@echo "清理完成"

# 安装依赖
deps:
	@echo "安装依赖..."
	go mod tidy
	go mod download

# 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...

# 代码检查
lint:
	@echo "代码检查..."
	golangci-lint run

# 开发模式（热重载）
dev:
	@echo "开发模式启动..."
	air

# 生产构建
build-prod:
	@echo "生产环境构建..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o build/eiam-platform cmd/server/main.go
	@echo "生产构建完成: build/eiam-platform"
