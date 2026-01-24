#!/bin/bash
set -e

echo "========================================"
echo "小红书 MCP 登录工具"
echo "========================================"
echo ""

# 检查登录程序是否存在
LOGIN_BIN=""
if [ -f "./xiaohongshu-login" ]; then
    LOGIN_BIN="./xiaohongshu-login"
elif [ -f "./bin/xiaohongshu-login" ]; then
    LOGIN_BIN="./bin/xiaohongshu-login"
elif [ -f "./cmd/login/login" ]; then
    LOGIN_BIN="./cmd/login/login"
else
    echo "❌ 未找到登录程序"
    echo ""
    echo "请先编译登录工具："
    echo "  go build -o xiaohongshu-login ./cmd/login"
    exit 1
fi

echo "🚀 启动登录程序..."
echo ""
echo "📋 操作步骤:"
echo "  1. 等待浏览器窗口打开"
echo "  2. 扫描二维码登录小红书"
echo "  3. 登录成功后 cookies 自动保存"
echo "  4. 浏览器窗口会自动关闭"
echo ""

# 执行登录
$LOGIN_BIN

# 检查登录结果
COOKIE_FILE=""
if [ -f "cookies.json" ]; then
    COOKIE_FILE="cookies.json"
elif [ -f "$HOME/.xiaohongshu/cookies.json" ]; then
    COOKIE_FILE="$HOME/.xiaohongshu/cookies.json"
fi

if [ -z "$COOKIE_FILE" ]; then
    echo ""
    echo "❌ 登录失败: 未找到 cookies 文件"
    exit 1
fi

echo ""
echo "✅ 登录成功！"
echo ""
echo "📁 Cookies 保存位置: $COOKIE_FILE"
echo ""

# 询问是否需要同步到远程
read -p "是否需要同步到远程服务器？(y/N): " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📋 请复制以下命令到远程服务器执行："
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    COOKIES_BASE64=$(base64 -i "$COOKIE_FILE" | tr -d '\n')

    cat <<EOF
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
    echo "2. SSH 登录远程服务器"
    echo "3. 粘贴并执行"
    echo "4. 如果远程端口不是 18060，请修改命令中的端口号"
    echo ""
fi

echo "✅ 完成"
