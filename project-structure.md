# EIAM Platform 项目结构

## 整体架构
```
eiam-platform/
├── backend/                 # 后端项目（当前目录）
│   ├── cmd/
│   ├── config/
│   ├── internal/
│   ├── pkg/
│   ├── go.mod
│   └── go.sum
├── frontend/                # 前端项目（新建）
│   ├── src/
│   ├── public/
│   ├── package.json
│   └── vite.config.ts
├── docs/                    # 项目文档
├── docker-compose.yml       # 开发环境
└── README.md
```

## 前端技术栈推荐

### 核心框架
- **Vue 3** - 渐进式JavaScript框架
- **TypeScript** - 类型安全的JavaScript
- **Vite** - 快速构建工具
- **Vue Router** - 路由管理
- **Pinia** - 状态管理

### UI组件库
- **Element Plus** - Vue 3的UI组件库
- **@element-plus/icons-vue** - 图标库

### 工具库
- **Axios** - HTTP客户端
- **Day.js** - 日期处理
- **Lodash-es** - 工具函数库

### 开发工具
- **ESLint** - 代码检查
- **Prettier** - 代码格式化
- **Husky** - Git钩子

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

## 项目特色功能

### 1. 用户管理
- 用户登录/注册
- 用户信息管理
- 角色权限控制

### 2. 组织管理
- 组织架构树
- 部门管理
- 用户分配

### 3. 应用管理
- 应用列表
- 权限配置
- 单点登录

### 4. 系统监控
- 登录日志
- 操作审计
- 系统状态

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
