#!/bin/bash
set -e

echo "=== 小红书登录并同步 Cookies ==="
echo ""

# 1. 执行登录
echo "步骤1: 开始登录..."
./bin/xiaohongshu-login

# 检查登录是否成功
if [ ! -f cookies.json ]; then
    echo "❌ 错误: cookies.json 文件不存在，登录可能失败"
    exit 1
fi

echo "✅ 登录成功，cookies 已保存到 cookies.json"
echo ""

# 2. 读取 cookies 并编码为 base64
echo "步骤2: 读取 cookies..."
COOKIES_BASE64=$(base64 -i cookies.json)

# 3. 调用 MCP sync_cookies 工具
echo "步骤3: 同步 cookies 到 MCP 服务器..."
RESPONSE=$(curl -s http://localhost:18060/mcp -X POST \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "sync_cookies",
      "arguments": {
        "cookies_base64": "'"$COOKIES_BASE64"'"
      }
    }
  }')

# 检查响应
if echo "$RESPONSE" | jq -e '.result' > /dev/null 2>&1; then
    echo "✅ Cookies 同步成功！"
    echo ""
    echo "$RESPONSE" | jq -r '.result.content[0].text'
else
    echo "❌ 同步失败，响应："
    echo "$RESPONSE" | jq '.'
    exit 1
fi

echo ""
echo "=== 完成 ==="
echo "现在你可以使用 MCP 工具了（无需重启服务器）"
