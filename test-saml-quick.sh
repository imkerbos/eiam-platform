#!/bin/bash

# SAML快速测试脚本
echo "🧪 SAML快速测试"
echo "================"

# 检查服务器状态
echo "1. 检查服务器状态..."
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ 后端服务器运行正常"
else
    echo "❌ 后端服务器未运行"
    exit 1
fi

# 测试SAML元数据端点
echo "2. 测试SAML元数据端点..."
response=$(curl -s -w "%{http_code}" -o /tmp/saml_metadata.xml http://localhost:3000/saml/metadata)
if [ "$response" = "200" ]; then
    echo "✅ 元数据端点正常 (通过前端代理)"
    # 检查XML内容
    if grep -q "EntityDescriptor" /tmp/saml_metadata.xml; then
        echo "✅ 元数据包含正确的SAML内容"
    else
        echo "❌ 元数据内容不正确"
    fi
else
    echo "❌ 元数据端点失败: HTTP $response"
fi

# 测试直接后端访问
echo "3. 测试后端直接访问..."
response=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:8080/saml/metadata)
if [ "$response" = "200" ]; then
    echo "✅ 后端直接访问正常"
else
    echo "❌ 后端直接访问失败: HTTP $response"
fi

# 测试SSO端点
echo "4. 测试SSO端点..."
response=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:3000/saml/sso)
if [ "$response" = "400" ]; then
    echo "✅ SSO端点正常响应 (400是预期的，因为没有SAML请求)"
else
    echo "⚠️  SSO端点响应: HTTP $response"
fi

echo ""
echo "🎉 快速测试完成！"
echo "💡 运行完整测试: ./scripts/run-saml-tests.sh"
