#!/bin/bash

echo "=== CAS IdP Demo Test Script ==="
echo ""

# 检查EIAM后端是否运行
echo "1. 检查EIAM后端状态..."
if curl -s "http://localhost:8080/health" > /dev/null; then
    echo "✅ EIAM后端正在运行"
else
    echo "❌ EIAM后端未运行，请先启动: go run cmd/server/main.go"
    exit 1
fi

# 获取CAS服务器信息
echo ""
echo "2. 获取CAS服务器信息..."
CAS_INFO=$(curl -s "http://localhost:8080/public/cas-server-info")
if echo "$CAS_INFO" | grep -q "server_url"; then
    echo "✅ CAS服务器信息获取成功"
    echo "$CAS_INFO" | jq '.data | {server_url, protocol_version, supported_features}'
else
    echo "❌ 无法获取CAS服务器信息"
    exit 1
fi

# 测试CAS登录页面
echo ""
echo "3. 测试CAS登录页面..."
LOGIN_RESPONSE=$(curl -s "http://localhost:8080/cas/login?service=http://localhost:3001/callback")
if echo "$LOGIN_RESPONSE" | grep -q "CAS Login"; then
    echo "✅ CAS登录页面正常"
else
    echo "❌ CAS登录页面异常"
    exit 1
fi

# 测试CAS登录提交
echo ""
echo "4. 测试CAS登录提交..."
LOGIN_RESULT=$(curl -s -X POST -d "username=admin&password=admin123&service=http://localhost:3001/callback" \
    -H "Content-Type: application/x-www-form-urlencoded" \
    "http://localhost:8080/cas/login")

if echo "$LOGIN_RESULT" | grep -q "ticket"; then
    echo "✅ CAS登录成功，获得服务票据"
    TICKET=$(echo "$LOGIN_RESULT" | jq -r '.data.ticket')
    echo "票据: $TICKET"
    
    # 测试票据验证
    echo ""
    echo "5. 测试票据验证..."
    VALIDATE_RESULT=$(curl -s "http://localhost:8080/cas/validate?service=http://localhost:3001/callback&ticket=$TICKET")
    if echo "$VALIDATE_RESULT" | grep -q "yes"; then
        echo "✅ 票据验证成功 (CAS 1.0)"
        echo "响应: $VALIDATE_RESULT"
    else
        echo "❌ 票据验证失败"
    fi
    
    # 测试CAS 2.0验证
    echo ""
    echo "6. 测试CAS 2.0验证..."
    VALIDATE_2_RESULT=$(curl -s "http://localhost:8080/cas/serviceValidate?service=http://localhost:3001/callback&ticket=$TICKET")
    if echo "$VALIDATE_2_RESULT" | grep -q "authenticationSuccess"; then
        echo "✅ CAS 2.0验证成功"
        echo "响应: $VALIDATE_2_RESULT"
    else
        echo "❌ CAS 2.0验证失败"
    fi
    
else
    echo "❌ CAS登录失败"
    echo "响应: $LOGIN_RESULT"
fi

echo ""
echo "=== 测试完成 ==="
echo ""
echo "📋 下一步操作："
echo "1. 启动demo服务器: python3 run-demo-server.py"
echo "2. 打开浏览器访问: http://localhost:3001/cas-idp-demo.html"
echo "3. 按照页面提示进行CAS集成测试"
echo ""
echo "📚 详细说明请查看: CAS_DEMO_README.md"
