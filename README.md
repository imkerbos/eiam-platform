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
- 7天免密码登录（Refresh Token自动续期）
- 密码重置和修改
- 智能会话管理（单设备/多设备登录）
- 自动Token刷新和重试机制

### 👥 用户管理
- 用户创建、编辑、删除
- 用户状态管理（启用/禁用/锁定）
- 密码策略配置
- 用户资料管理（头像上传）
- 用户会话监控和强制下线
- 登录日志和审计

### 🏢 组织架构
- 多级组织架构（总部、分公司、部门、小组）
- 组织关系管理
- 组织管理员分配
- 组织树形结构展示

### 🔑 权限管理
- **权限路由系统**: 基于应用/应用组的访问控制
- **权限分配**: 支持分配给用户或组织
- **系统管理员**: 系统级权限管理
- **应用访问控制**: 细粒度的应用访问权限

### 📱 应用管理
- 应用注册和管理
- 应用分组管理
- **协议支持**:
  - OAuth2配置
  - SAML配置（IdP/SP模式）
  - CAS配置
  - LDAP配置
- 应用访问统计
- 应用删除保护（关联检查）
- **IdP-initiated SSO**: 支持身份提供商发起的单点登录

### 📊 系统监控
- **Dashboard统计**: 用户数、组织数、在线用户数、应用数
- **实时监控**: 在线用户统计、活跃会话管理
- **审计日志**: 操作日志、登录日志
- **系统状态**: 数据库、Redis、API服务状态监控

### 🎨 前端功能
- **Console管理端**: 系统管理界面
- **Portal用户端**: 用户自助服务界面
- 响应式设计
- 现代化UI/UX
- 国际化支持（英文界面）
- 实时数据更新

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

## 🔄 IdP-initiated SSO流程

EIAM平台支持完整的身份提供商发起的单点登录（IdP-initiated SSO）流程，用户登录后可一键访问各种协议的第三方应用。

### 流程架构图

```mermaid
sequenceDiagram
    participant U as 用户
    participant F as 前端(3000)
    participant B as 后端(8080)
    participant SP as 第三方应用

    U->>F: 1. 登录EIAM平台
    F->>B: 2. 获取应用列表(含protocol字段)
    B-->>F: 3. 返回应用列表
    U->>F: 4. 点击SAML/CAS应用
    F->>B: 5. 跳转到 /api/v1/portal/applications/:id/launch
    B->>B: 6. 验证用户身份
    B->>B: 7. 根据protocol生成响应
    
    alt SAML应用
        B-->>F: 8. 返回SAML POST表单
        F->>SP: 9. 自动提交SAMLResponse到ACS
    else CAS应用  
        B-->>F: 8. 重定向到Service URL + ticket
        F->>SP: 9. 跳转到CAS服务
    else OAuth2/OIDC应用
        B-->>F: 8. 重定向到授权端点 + code
        F->>SP: 9. 跳转到OAuth2服务
    end
    
    SP-->>U: 10. 用户无缝登录成功
```

### 协议支持详情

#### 🔐 SAML 2.0 IdP
- **元数据端点**: `GET /saml/metadata`
- **SSO端点**: `ANY /saml/sso` (SP-initiated)
- **SLS端点**: `ANY /saml/sls` (单点注销)
- **IdP-initiated**: `GET /api/v1/portal/applications/:id/launch`
- **断言签名**: 使用RSA-SHA256算法
- **属性映射**: 支持用户属性和角色映射
- **POST Binding**: 自动表单提交到SP的ACS URL

#### 🎫 CAS 协议
- **登录端点**: `GET/POST /cas/login` 
- **票据验证**: 
  - CAS 1.0: `GET /cas/validate` (纯文本响应)
  - CAS 2.0: `GET /cas/serviceValidate` (XML/JSON响应)
- **票据管理**: 使用go-cache + 数据库持久化
- **IdP-initiated**: 生成Service Ticket并重定向
- **Gateway模式**: 支持透明认证
- **属性发布**: 支持用户属性传递

#### 🔑 OAuth2/OIDC
- **授权端点**: 标准OAuth2授权码流程
- **令牌端点**: 支持多种grant类型
- **用户信息端点**: 获取用户详细信息
- **IdP-initiated**: 生成授权码并重定向
- **PKCE支持**: 增强安全性（计划中）

### 核心端点

| 端点 | 方法 | 描述 |
|------|------|------|
| `/api/v1/portal/applications` | GET | 获取用户应用列表（含protocol字段） |
| `/api/v1/portal/applications/:id/launch` | GET | 启动应用（IdP-initiated SSO） |
| `/saml/metadata` | GET | SAML IdP元数据 |
| `/saml/sso` | ANY | SAML SSO端点 |
| `/cas/login` | GET/POST | CAS登录端点 |
| `/cas/serviceValidate` | GET | CAS票据验证 |
| `/health` | GET | 系统健康检查 |

### 技术实现

#### SAML实现
- **库**: `github.com/crewjam/saml` - 成熟的Go SAML库
- **证书管理**: 自动生成RSA-2048密钥对和X.509证书
- **签名算法**: RSA-SHA256
- **断言生成**: 支持NameID、属性声明、角色映射
- **POST Binding**: 自动HTML表单提交

#### CAS实现  
- **票据管理**: `github.com/patrickmn/go-cache` + 数据库持久化
- **版本支持**: CAS 1.0/2.0协议
- **响应格式**: 支持XML和JSON格式
- **安全特性**: 票据过期、一次性使用、服务URL验证

#### 前端集成
- **智能路由**: 根据应用protocol字段选择启动方式
- **用户体验**: 加载提示、错误处理、无缝跳转
- **类型安全**: 完整的TypeScript类型定义

## 📊 API文档

### 认证相关

#### Console登录
```http
POST /api/v1/console/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123",
  "otp_code": "123456"  // 可选
}
```

#### Portal登录
```http
POST /api/v1/portal/auth/login
Content-Type: application/json

{
  "username": "user",
  "password": "password123"
}
```

#### 刷新令牌
```http
POST /api/v1/console/auth/refresh
Content-Type: application/json

{
  "refresh_token": "your_refresh_token"
}
```

#### 登出
```http
POST /api/v1/console/auth/logout
Authorization: Bearer <access_token>
```

### 用户管理

#### 获取用户列表
```http
GET /api/v1/console/users?page=1&page_size=10&search=keyword
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
  "display_name": "New User",
  "organization_id": "org-uuid",
  "password": "password123"
}
```

#### 获取用户会话
```http
GET /api/v1/console/sessions?page=1&page_size=10
Authorization: Bearer <access_token>
```

### 应用管理

#### 获取应用列表
```http
GET /api/v1/console/applications?page=1&page_size=10
Authorization: Bearer <access_token>
```

#### 创建应用
```http
POST /api/v1/console/applications
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "My App",
  "description": "Application description",
  "group_id": "group-uuid",
  "protocol": "oauth2",
  "config": {
    "client_id": "app_client_id",
    "client_secret": "app_client_secret",
    "redirect_uris": "https://app.com/callback"
  }
}
```

### 权限管理

#### 获取权限路由
```http
GET /api/v1/console/permission-routes?page=1&page_size=10
Authorization: Bearer <access_token>
```

#### 创建权限路由
```http
POST /api/v1/console/permission-routes
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "name": "App Access Route",
  "code": "APP_ACCESS",
  "description": "Access to specific applications",
  "application_ids": ["app-uuid-1", "app-uuid-2"]
}
```

#### 分配权限路由
```http
POST /api/v1/console/permission-route-assignments
Authorization: Bearer <access_token>
Content-Type: application/json

{
  "permission_route_id": "route-uuid",
  "assignee_type": "user",  // "user" 或 "organization"
  "assignee_id": "user-uuid"
}
```

### 系统监控

#### 获取Dashboard数据
```http
GET /api/v1/console/dashboard
Authorization: Bearer <access_token>
```

#### 获取审计日志
```http
GET /api/v1/console/logs/audit?page=1&page_size=10
Authorization: Bearer <access_token>
```

#### 获取登录日志
```http
GET /api/v1/console/logs/login?page=1&page_size=10
Authorization: Bearer <access_token>
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