# EIAM Platform 登录功能使用指南

## 🎯 功能概述

EIAM Platform 已成功实现了完整的登录流程，包括：

- ✅ 用户名/密码登录
- ✅ JWT令牌管理（Access Token + Refresh Token）
- ✅ OTP双因素认证（可选）
- ✅ 账户锁定机制
- ✅ 登录日志记录
- ✅ 前后端完整串联

## 🚀 快速开始

### 1. 启动服务

#### 方式一：使用启动脚本（推荐）
```bash
./start.sh
```

#### 方式二：分别启动
```bash
# 启动后端服务
go run cmd/server/main.go

# 启动前端服务（新终端）
cd frontend && npm run dev
```

### 2. 初始化数据

首次使用需要创建默认管理员账户：

```bash
# 运行数据库迁移
make migrate

# 创建默认管理员账户
make seed
```

### 3. 访问应用

- **前端应用**: http://localhost:3000
- **后端API**: http://localhost:8080
- **健康检查**: http://localhost:8080/health
- **MD5登录测试**: http://localhost:3000/test-md5-login.html

## 🔐 默认账户

系统会自动创建一个默认管理员账户：

- **用户名**: `admin`
- **密码**: `admin123`
- **邮箱**: `admin@eiam-platform.com`

⚠️ **重要提醒**: 首次登录后请立即修改默认密码！

## 📡 API接口

### 登录接口

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123",
  "otp_code": "123456"  // 可选，如果用户启用了OTP
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": {
      "id": "user_admin",
      "username": "admin",
      "email": "admin@eiam-platform.com",
      "display_name": "System Administrator",
      "status": "active",
      "email_verified": true,
      "phone_verified": false,
      "enable_otp": false,
      "last_login_at": "2025-08-12T11:23:35.18834+08:00",
      "last_login_ip": "::1"
    },
    "require_otp": false
  }
}
```

### 刷新令牌接口

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

### 登出接口

```http
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
```

## 🧪 测试登录

### 1. 使用MD5登录测试页面

访问 `test-md5-login.html` 文件进行MD5加密和登录测试：

```bash
# 在浏览器中打开
open test-md5-login.html
```

这个测试页面可以：
- 查看MD5加密过程
- 测试MD5加密后的登录功能
- 验证错误密码被正确拒绝

### 2. 使用原始测试页面

访问 `test-login.html` 文件进行API测试：

```bash
# 在浏览器中打开
open test-login.html
```

### 2. 使用curl命令

```bash
# 测试登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | jq .

# 测试健康检查
curl http://localhost:8080/health | jq .
```

### 3. 使用前端应用

1. 访问 http://localhost:3000
2. 使用默认账户登录：
   - 用户名: `admin`
   - 密码: `admin123`
3. 登录成功后会自动跳转到管理控制台

## 🔒 安全特性

### 密码传输安全
- **前端密码加密**: 使用MD5加密传输密码
- **加密算法**: MD5哈希算法
- **传输安全**: 密码在前端加密后传输到后端
- **统一格式**: 前后端使用相同的MD5格式

### 密码存储安全
- 使用MD5哈希存储密码
- 支持密码复杂度验证
- 密码失败次数限制（5次后锁定30分钟）

### 令牌安全
- JWT令牌加密
- Access Token有效期：1小时
- Refresh Token有效期：7天
- 支持令牌刷新机制

### 账户保护
- 账户状态检查
- 登录IP记录
- 登录日志审计
- 账户锁定机制

### OTP双因素认证
- 支持TOTP（基于时间的一次性密码）
- 可选的OTP验证
- 备用验证码支持

## 📊 登录流程

### 标准登录流程
1. 用户输入用户名和密码
2. 系统验证用户存在性和状态
3. 验证密码正确性
4. 检查账户锁定状态
5. 生成JWT令牌
6. 更新登录信息
7. 记录登录日志
8. 返回令牌和用户信息

### OTP验证流程
1. 用户输入用户名和密码
2. 系统验证基础认证
3. 检查用户是否启用OTP
4. 如果需要OTP，返回要求OTP的响应
5. 用户输入OTP验证码
6. 验证OTP正确性
7. 完成登录流程

## 🐛 故障排除

### 常见问题

#### 1. 数据库连接失败
```bash
# 检查数据库配置
cat config/config.yaml

# 检查数据库服务
mysql -u root -p -e "SHOW DATABASES;"
```

#### 2. 登录失败
- 检查用户名和密码是否正确
- 确认账户状态是否为active
- 检查账户是否被锁定
- 查看后端日志获取详细错误信息

#### 3. 前端无法连接后端
- 确认后端服务在8080端口运行
- 检查Vite代理配置
- 确认CORS配置正确

#### 4. 令牌验证失败
- 检查令牌是否过期
- 确认令牌格式正确
- 验证JWT密钥配置

### 5. 账户被锁定
如果账户因多次登录失败被锁定，可以使用以下命令重置：

```bash
# 重置admin用户锁定状态
make reset-user

# 或者指定用户名重置
go run cmd/reset-user/main.go -username admin
```

### 日志查看

```bash
# 查看后端日志
tail -f logs/service.log

# 查看访问日志
tail -f logs/access.log

# 查看错误日志
tail -f logs/error.log
```

## 🔧 配置说明

### JWT配置
```yaml
jwt:
  secret: "your-secret-key-here"
  access_token_expire: 3600    # 1小时
  refresh_token_expire: 604800 # 7天
  issuer: "eiam-platform"
```

### 登录配置
```yaml
login:
  default_method: "password"
  enable_otp: true
  enable_third_party: false
```

## 📈 监控指标

系统会记录以下登录相关指标：

- 登录成功/失败次数
- 登录响应时间
- 用户活跃度
- 账户锁定统计
- OTP使用情况

## 🔄 下一步开发

### 已实现功能
- ✅ 基础登录认证
- ✅ JWT令牌管理
- ✅ 用户管理基础结构
- ✅ 组织架构模型
- ✅ 角色权限模型

### 待开发功能
- 🔄 用户管理CRUD接口
- 🔄 组织架构管理
- 🔄 角色权限管理
- 🔄 应用管理
- 🔄 完整的OTP实现
- 🔄 第三方登录（OAuth2/SAML）

---

**EIAM Platform** - 企业级身份认证与访问管理平台
