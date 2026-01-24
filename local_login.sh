#!/bin/bash
set -e

echo "=== 小红书登录 - 生成远程同步命令 ==="
echo ""

# 1. 执行登录
echo "步骤1: 开始登录（请扫码）..."
./bin/xiaohongshu-login

# 检查登录是否成功
if [ ! -f cookies.json ]; then
    echo "❌ 错误: cookies.json 文件不存在，登录可能失败"
    exit 1
fi

echo ""
echo "✅ 登录成功！"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📋 请复制以下命令，粘贴到远程服务器执行："
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# 2. 生成可复制粘贴的 curl 命令
COOKIES_BASE64=$(base64 -i cookies.json | tr -d '\n')

cat << EOF
curl -X POST http://localhost:18060/mcp \\
  -H "Content-Type: application/json" \\
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "sync_cookies",
      "arguments": {
        "cookies_base64": "${COOKIES_BASE64}"
      }
    }
  }'
EOF

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "💡 使用说明："
echo "1. 复制上面的 curl 命令"
echo "2. SSH 登录到远程服务器"
echo "3. 粘贴并执行"
echo "4. 如果远程 MCP 端口不是 18060，请修改 localhost:18060"
echo ""
