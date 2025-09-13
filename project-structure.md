# EIAM Platform 项目结构

## 整体架构
```
eiam-platform/
├── cmd/                     # 应用程序入口
│   ├── server/             # 主服务器
│   ├── migrate/            # 数据库迁移工具
│   ├── seed/               # 数据种子
│   ├── reset-user/         # 用户重置工具
│   └── update-password/    # 密码更新工具
├── config/                 # 配置文件
│   ├── config.go          # 配置结构定义
│   └── config.yaml        # 配置文件
├── internal/               # 内部包
│   ├── handlers/          # API处理器
│   │   ├── auth.go        # 认证处理器
│   │   ├── user.go        # 用户管理
│   │   ├── organization.go # 组织管理
│   │   ├── system_setting.go # 系统设置
│   │   └── password_policy.go # 密码策略
│   ├── middleware/        # 中间件
│   │   ├── auth.go        # 认证中间件
│   │   └── common.go      # 通用中间件
│   ├── models/            # 数据模型
│   │   ├── user.go        # 用户模型
│   │   ├── organization.go # 组织模型
│   │   ├── application.go # 应用模型
│   │   └── base.go        # 基础模型
│   └── router/            # 路由配置
│       └── router.go      # 路由定义
├── pkg/                   # 公共包
│   ├── database/          # 数据库连接
│   ├── redis/             # Redis连接
│   ├── logger/            # 日志配置
│   ├── session/           # 会话管理
│   ├── utils/             # 工具函数
│   └── i18n/              # 国际化
├── frontend/              # 前端项目
│   ├── src/
│   │   ├── views/         # 页面组件
│   │   │   ├── console/   # 管理端页面
│   │   │   └── portal/    # 用户端页面
│   │   ├── stores/        # 状态管理
│   │   ├── api/           # API接口
│   │   ├── components/    # 公共组件
│   │   ├── router/        # 路由配置
│   │   ├── types/         # TypeScript类型
│   │   └── utils/         # 工具函数
│   ├── public/            # 静态资源
│   ├── package.json       # 依赖配置
│   └── vite.config.ts     # 构建配置
├── migrations/            # 数据库迁移文件
├── uploads/               # 上传文件目录
├── logs/                  # 日志文件目录
├── go.mod                 # Go模块配置
├── go.sum                 # Go依赖锁定
├── README.md              # 项目说明
├── TODO.md                # 开发计划
└── project-structure.md   # 项目结构说明
```

## 技术栈

### 后端技术栈
- **Go 1.21+** - 编程语言
- **Gin** - Web框架
- **GORM** - ORM框架
- **MySQL/PostgreSQL** - 数据库
- **Redis** - 缓存和会话存储
- **JWT** - 身份认证
- **Zap** - 日志框架
- **Viper** - 配置管理

### 前端技术栈
- **Vue 3** - 渐进式JavaScript框架
- **TypeScript** - 类型安全的JavaScript
- **Vite** - 快速构建工具
- **Vue Router** - 路由管理
- **Pinia** - 状态管理
- **Ant Design Vue** - UI组件库
- **Axios** - HTTP客户端
- **Day.js** - 日期处理

### 开发工具
- **ESLint** - 代码检查
- **Prettier** - 代码格式化
- **Git** - 版本控制

## 开发流程建议

### 1. 开发环境
```bash
# 后端开发
cd backend
go run cmd/server/main.go

# 前端开发（新终端）
cd frontend
npm run dev
```

### 2. API接口规范
- 统一使用RESTful API
- 请求/响应格式标准化
- 错误处理统一化
- 接口文档自动化

### 3. 认证流程
1. 用户登录 → 获取JWT Token
2. 前端存储Token（localStorage/sessionStorage）
3. 请求时携带Token
4. 后端验证Token
5. Token过期 → 自动刷新

## 学习路径建议

### 第一阶段：基础概念
1. 理解前后端分离概念
2. 学习HTTP协议和RESTful API
3. 掌握基本的JavaScript/TypeScript

### 第二阶段：Vue 3基础
1. Vue 3组件系统
2. 响应式数据
3. 生命周期钩子
4. 模板语法

### 第三阶段：项目实战
1. 路由配置
2. 状态管理
3. API调用
4. 组件开发

### 第四阶段：进阶优化
1. 性能优化
2. 代码规范
3. 测试编写
4. 部署上线

## 快速开始指南

### 1. 创建前端项目
```bash
# 使用Vite创建Vue项目
npm create vue@latest frontend
cd frontend
npm install

# 安装依赖
npm install element-plus @element-plus/icons-vue
npm install axios dayjs lodash-es
npm install -D @types/lodash-es
```

### 2. 配置开发环境
```bash
# 后端API地址配置
# frontend/.env.development
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

### 3. 启动开发服务器
```bash
# 后端
cd backend
go run cmd/server/main.go

# 前端（新终端）
cd frontend
npm run dev
```

## 核心功能模块

### 1. 认证系统 ✅
- 用户名/邮箱 + 密码登录
- OTP双因素认证
- JWT令牌管理（Access Token + Refresh Token）
- 7天免密码登录（Refresh Token自动续期）
- 智能会话管理（单设备/多设备登录）
- 自动Token刷新和重试机制

### 2. 用户管理 ✅
- 用户创建、编辑、删除
- 用户状态管理（启用/禁用/锁定）
- 密码策略配置
- 用户资料管理（头像上传）
- 用户会话监控和强制下线
- 登录日志和审计

### 3. 组织架构 ✅
- 多级组织架构（总部、分公司、部门、小组）
- 组织关系管理
- 组织管理员分配
- 组织树形结构展示

### 4. 权限管理 ✅
- **权限路由系统**: 基于应用/应用组的访问控制
- **权限分配**: 支持分配给用户或组织
- **系统管理员**: 系统级权限管理
- **应用访问控制**: 细粒度的应用访问权限

### 5. 应用管理 ✅
- 应用注册和管理
- 应用分组管理
- **协议支持**:
  - OAuth2配置
  - SAML配置（IdP/SP模式）
  - CAS配置
  - LDAP配置
- 应用访问统计
- 应用删除保护（关联检查）

### 6. 系统监控 ✅
- **Dashboard统计**: 用户数、组织数、在线用户数、应用数
- **实时监控**: 在线用户统计、活跃会话管理
- **审计日志**: 操作日志、登录日志
- **系统状态**: 数据库、Redis、API服务状态监控

### 7. 前端界面 ✅
- **Console管理端**: 系统管理界面
- **Portal用户端**: 用户自助服务界面
- 响应式设计
- 现代化UI/UX
- 国际化支持（英文界面）
- 实时数据更新

## 部署建议

### 开发环境
- 使用Docker Compose
- 热重载开发
- 本地数据库

### 生产环境
- 前后端分离部署
- Nginx反向代理
- HTTPS配置
- 数据库集群

## 学习资源

### 官方文档
- [Vue 3官方文档](https://cn.vuejs.org/)
- [Element Plus文档](https://element-plus.org/zh-CN/)
- [Vite官方文档](https://cn.vitejs.dev/)

### 推荐教程
- Vue 3 + TypeScript实战
- 前后端分离项目开发
- 企业级前端架构

### 实践项目
- 管理后台系统
- 用户权限管理
- API接口设计
