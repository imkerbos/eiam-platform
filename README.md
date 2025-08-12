# EIAM Platform

基于Go Gin框架开发的企业级身份认证与访问管理（EIAM）SSO平台，支持多种认证协议和现代化的身份管理功能。

## 🚀 技术栈

### 后端
- **框架**: Go Gin
- **配置管理**: Viper
- **日志**: Zap (JSON格式，按日期切割)
- **数据库**: GORM (支持MySQL/PostgreSQL)
- **缓存**: Go-Redis
- **认证**: JWT (Access Token + Refresh Token)
- **密码加密**: bcrypt
- **请求追踪**: TradeID

### 前端
- **框架**: Vue 3 + TypeScript
- **UI组件库**: Ant Design Vue
- **构建工具**: Vite
- **状态管理**: Pinia
- **路由**: Vue Router
- **HTTP客户端**: Axios
- **工具库**: Day.js, Lodash-es

## 📁 项目结构

```
eiam-platform/
├── cmd/                    # 应用程序入口
│   ├── server/            # 主服务器
│   └── migrate/           # 数据库迁移工具
├── config/                # 配置文件
├── internal/              # 内部包
│   ├── handlers/          # API处理器
│   ├── middleware/        # 中间件
│   ├── models/            # 数据模型
│   └── router/            # 路由配置
├── pkg/                   # 公共包
│   ├── database/          # 数据库连接
│   ├── redis/             # Redis连接
│   ├── logger/            # 日志配置
│   ├── utils/             # 工具函数
│   └── i18n/              # 国际化
├── frontend/              # 前端项目
│   ├── src/
│   │   ├── views/         # 页面组件
│   │   ├── stores/        # 状态管理
│   │   ├── api/           # API接口
│   │   └── types/         # TypeScript类型
│   └── package.json
├── migrations/            # 数据库迁移文件
├── docs/                  # 文档
└── static/                # 静态文件
```

## 🛠️ 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- MySQL 8.0+ 或 PostgreSQL 13+
- Redis 6.0+

### 1. 克隆项目

```bash
git clone <repository-url>
cd eiam-platform
```

### 2. 配置环境

复制环境配置文件：

```bash
cp env.example .env
```

编辑 `.env` 文件，配置数据库和Redis连接信息：

```env
# Database
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASSWORD=123456
DB_NAME=eiam_platform

# Redis
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Server
SERVER_PORT=8080
SERVER_MODE=debug
```

### 3. 安装依赖

#### 后端依赖
```bash
go mod tidy
```

#### 前端依赖
```bash
cd frontend
npm install
```

### 4. 数据库迁移

```bash
go run cmd/migrate/main.go
```

### 5. 启动服务

#### 方式一：使用启动脚本（推荐）
```bash
./start.sh
```

#### 方式二：分别启动

**启动后端服务：**
```bash
go run cmd/server/main.go
```

**启动前端开发服务器：**
```bash
cd frontend
npm run dev
```

### 6. 访问应用

- **前端应用**: http://localhost:3000
- **后端API**: http://localhost:8080
- **健康检查**: http://localhost:8080/health

## 📋 功能特性

### 🔐 认证系统
- 用户名/邮箱 + 密码登录
- OTP双因素认证
- JWT令牌管理（Access Token + Refresh Token）
- 密码重置
- 会话管理

### 👥 用户管理
- 用户创建、编辑、删除
- 用户状态管理
- 密码策略
- 用户资料管理

### 🏢 组织架构
- 多级组织架构（总部、分公司、部门、小组）
- 组织关系管理
- 组织管理员

### 🔑 角色权限
- 基于角色的访问控制（RBAC）
- 权限管理
- 角色分配
- 权限继承

### 📱 应用管理
- 应用注册和管理
- 应用分组
- OAuth2配置
- SAML配置
- 应用访问控制

### 🎨 前端功能
- **Console管理端**: 系统管理界面
- **Portal用户端**: 用户自助服务界面
- 响应式设计
- 现代化UI/UX

## 🔧 开发指南

### 后端开发

#### 添加新的API端点

1. 在 `internal/handlers/` 中添加处理器
2. 在 `internal/router/` 中注册路由
3. 在 `internal/models/` 中定义数据模型

#### 数据库迁移

```bash
# 运行迁移
go run cmd/migrate/main.go

# 添加新的迁移文件
# 在 internal/models/ 中添加新模型
```

### 前端开发

#### 添加新页面

1. 在 `frontend/src/views/` 中创建Vue组件
2. 在 `frontend/src/router/index.ts` 中添加路由
3. 在 `frontend/src/types/api.ts` 中定义TypeScript类型

#### 开发命令

```bash
cd frontend

# 开发模式
npm run dev

# 构建生产版本
npm run build

# 代码检查
npm run lint

# 类型检查
npm run type-check
```

## 📊 API文档

### 认证相关

#### 登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password",
  "otp_code": "123456"  // 可选
}
```

#### 刷新令牌
```http
POST /api/v1/auth/refresh
Authorization: Bearer <refresh_token>
```

### 用户管理

#### 获取用户列表
```http
GET /api/v1/console/users?page=1&size=10
Authorization: Bearer <access_token>
```

#### 创建用户
```http
POST /api/v1/console/users
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "username": "newuser",
  "email": "user@example.com",
  "displayName": "New User",
  "organizationId": "1"
}
```

## 🔒 安全特性

- JWT令牌加密
- 密码bcrypt加密
- CORS配置
- 请求频率限制
- SQL注入防护
- XSS防护

## 📝 日志系统

- JSON格式日志
- 多级别日志（DEBUG, INFO, WARN, ERROR）
- 按日期自动切割
- 分离服务日志、访问日志、错误日志
- 支持stdout和文件输出

## 🌐 国际化

- 支持多语言
- 集中化消息管理
- 默认英文界面

## 🚀 部署

### Docker部署

```bash
# 构建镜像
docker build -t eiam-platform .

# 运行容器
docker run -d -p 8080:8080 eiam-platform
```

### 生产环境配置

1. 修改 `config/config.yaml` 中的生产环境配置
2. 设置环境变量
3. 配置反向代理（Nginx）
4. 配置SSL证书

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 支持

如有问题或建议，请提交 Issue 或联系开发团队。

---

**EIAM Platform** - 企业级身份认证与访问管理平台