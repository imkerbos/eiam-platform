# EIAM Platform Frontend

基于 Vue 3 + TypeScript + Ant Design Vue 的企业级身份认证与访问管理平台前端。

## 技术栈

- **框架**: Vue 3 + TypeScript
- **UI组件库**: Ant Design Vue 4.x
- **路由**: Vue Router 4
- **状态管理**: Pinia
- **HTTP客户端**: Axios
- **构建工具**: Vite
- **代码规范**: ESLint + Prettier

## 快速开始

### 1. 安装依赖

```bash
cd frontend
npm install
```

### 2. 启动开发服务器

```bash
npm run dev
```

前端将在 http://localhost:3000 启动

### 3. 构建生产版本

```bash
npm run build
```

### 4. 预览生产版本

```bash
npm run preview
```

## 项目结构

```
frontend/
├── src/
│   ├── api/           # API接口
│   ├── assets/        # 静态资源
│   ├── components/    # 公共组件
│   ├── router/        # 路由配置
│   ├── stores/        # 状态管理
│   ├── styles/        # 全局样式
│   ├── types/         # TypeScript类型定义
│   ├── utils/         # 工具函数
│   ├── views/         # 页面组件
│   ├── App.vue        # 根组件
│   └── main.ts        # 入口文件
├── public/            # 公共资源
├── index.html         # HTML模板
├── package.json       # 依赖配置
├── tsconfig.json      # TypeScript配置
├── vite.config.ts     # Vite配置
└── README.md          # 项目说明
```

## 功能特性

### 1. 用户认证
- 用户名/密码登录
- OTP双因子认证
- JWT Token管理
- 自动Token刷新

### 2. 管理控制台
- 用户管理
- 组织架构管理
- 角色权限管理
- 应用管理
- 系统监控

### 3. 用户门户
- 个人信息管理
- 应用访问
- 安全设置

## 开发指南

### 1. 添加新页面

1. 在 `src/views/` 下创建页面组件
2. 在 `src/router/index.ts` 中添加路由配置
3. 在菜单中添加对应的导航项

### 2. 添加新API

1. 在 `src/types/api.ts` 中定义类型
2. 在 `src/api/` 下创建API模块
3. 在组件中调用API

### 3. 状态管理

使用 Pinia 进行状态管理：

```typescript
// stores/example.ts
import { defineStore } from 'pinia'

export const useExampleStore = defineStore('example', () => {
  const state = ref({})
  const actions = {}
  return { state, actions }
})
```

### 4. 组件开发

使用 Vue 3 Composition API：

```vue
<template>
  <div>{{ message }}</div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const message = ref('Hello World')
</script>
```

## 环境配置

### 开发环境

创建 `.env.development` 文件：

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_APP_TITLE=EIAM Platform (Dev)
```

### 生产环境

创建 `.env.production` 文件：

```env
VITE_API_BASE_URL=https://api.eiam.com/api/v1
VITE_APP_TITLE=EIAM Platform
```

## 代码规范

### 1. 命名规范

- 组件名：PascalCase (如: UserList.vue)
- 文件名：kebab-case (如: user-list.vue)
- 变量名：camelCase (如: userName)
- 常量名：UPPER_SNAKE_CASE (如: API_BASE_URL)

### 2. 目录规范

- 页面组件放在 `views/` 目录
- 公共组件放在 `components/` 目录
- API接口放在 `api/` 目录
- 类型定义放在 `types/` 目录

### 3. 代码检查

```bash
# 代码检查
npm run lint

# 代码格式化
npm run format
```

## 部署

### 1. 构建

```bash
npm run build
```

### 2. 部署到服务器

将 `dist/` 目录下的文件部署到Web服务器。

### 3. Nginx配置示例

```nginx
server {
    listen 80;
    server_name your-domain.com;
    root /path/to/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://backend-server:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 常见问题

### 1. 跨域问题

开发环境已配置代理，生产环境需要配置Nginx代理。

### 2. Token过期

系统会自动刷新Token，如果刷新失败会跳转到登录页。

### 3. 路由问题

确保所有路由都在 `router/index.ts` 中正确配置。

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交代码
4. 创建 Pull Request

## 许可证

MIT License
