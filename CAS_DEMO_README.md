  b # CAS IdP Demo - EIAM Platform

这是一个完整的CAS IdP（Identity Provider）演示应用，用于测试EIAM平台的CAS Service Provider功能。

## 🚀 快速开始

### 1. 启动EIAM平台
确保EIAM平台后端正在运行：
```bash
go run cmd/server/main.go
```

### 2. 启动Demo服务器
```bash
python3 run-demo-server.py
```

或者直接打开HTML文件：
```bash
open cas-idp-demo.html
```

### 3. 访问Demo
打开浏览器访问：http://localhost:3001/cas-idp-demo.html

## 🔧 配置说明

### EIAM服务器配置
- **EIAM Server URL**: `http://localhost:3000` (通过前端代理访问)
- **Service URL**: `http://localhost:3001/callback` (默认)

### 回调URL配置
Demo服务器的回调URL是：`http://localhost:3001/callback`

**注意**: 
- 使用端口3001避免与Node.js前端(端口3000)冲突
- EIAM Server URL通过前端代理访问，实际后端运行在8080端口

## 📋 测试流程

### 1. 获取CAS服务器信息
- 点击 "Get CAS Server Info" 按钮
- 查看EIAM平台的CAS端点信息
- 确认支持的协议版本和功能

### 2. 开始CAS登录
- 点击 "Start CAS Login" 按钮
- 系统会打开新的窗口显示EIAM登录页面
- 使用管理员账户登录（admin/admin123）

### 3. 验证服务票据
- 登录成功后，会获得一个服务票据（Service Ticket）
- 点击 "Validate Ticket" 按钮验证票据
- 查看用户信息和认证结果

### 4. 注销
- 点击 "Logout" 按钮
- 系统会重定向到EIAM注销页面

## 🎯 功能特性

### ✅ 已实现功能
- **CAS 1.0支持**: `/cas/validate` 端点
- **CAS 2.0支持**: `/cas/serviceValidate` 端点
- **自动票据验证**: 支持XML响应解析
- **用户信息显示**: 显示认证用户的详细信息
- **实时日志**: 显示所有操作步骤和结果
- **错误处理**: 完善的错误提示和处理

### 🔄 测试场景
1. **正常登录流程**: 用户通过CAS登录并验证票据
2. **票据验证**: 验证CAS 2.0 XML响应格式
3. **用户信息获取**: 从CAS响应中提取用户属性
4. **注销流程**: 完整的单点注销功能

## 📊 日志说明

Demo会记录以下操作：
- ✅ 成功操作（绿色）
- ❌ 错误信息（红色）
- ℹ️ 一般信息（蓝色）

## 🛠️ 技术实现

### 前端技术
- **纯HTML/CSS/JavaScript**: 无框架依赖
- **Fetch API**: 用于HTTP请求
- **DOMParser**: 解析CAS XML响应
- **响应式设计**: 支持移动端访问

### CAS协议支持
- **CAS 1.0**: 纯文本响应格式
- **CAS 2.0**: XML响应格式
- **Service Ticket**: 服务票据验证
- **User Attributes**: 用户属性传递

## 🔍 故障排除

### 常见问题

1. **无法连接到EIAM服务器**
   - 检查EIAM后端是否正在运行
   - 确认端口8080没有被其他服务占用
   - 检查防火墙设置

2. **票据验证失败**
   - 确认Service URL配置正确
   - 检查EIAM中是否注册了对应的应用
   - 查看浏览器控制台错误信息

3. **用户信息为空**
   - 检查CAS响应XML格式
   - 确认用户属性配置正确
   - 查看EIAM用户数据

### 调试技巧
- 打开浏览器开发者工具查看网络请求
- 检查EIAM后端日志
- 使用curl命令测试CAS端点

## 📝 示例命令

### 测试CAS登录
```bash
curl "http://localhost:3000/cas/login?service=http://localhost:3001/callback"
```

### 测试票据验证
```bash
curl "http://localhost:3000/cas/serviceValidate?service=http://localhost:3001/callback&ticket=ST-xxx"
```

### 获取CAS服务器信息
```bash
curl "http://localhost:3000/public/cas-server-info"
```

## 🎉 成功标志

当看到以下信息时，说明CAS集成成功：
- ✅ CAS服务器信息获取成功
- ✅ 服务票据生成成功
- ✅ 票据验证成功
- ✅ 用户信息显示正确
- ✅ 注销功能正常

## 📚 相关文档

- [CAS协议规范](https://apereo.github.io/cas/6.6.x/protocol/CAS-Protocol.html)
- [EIAM平台文档](./README.md)
- [CAS Server Info API](./README.md#cas服务端地址获取方式)

---

**注意**: 这是一个演示应用，仅用于测试EIAM平台的CAS功能。在生产环境中，请使用正式的CAS客户端库。
