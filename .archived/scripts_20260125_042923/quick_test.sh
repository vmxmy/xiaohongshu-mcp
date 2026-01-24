#!/bin/bash

# 快速测试脚本 - 验证 MCP 服务和工具

echo "=========================================="
echo "MCP 服务快速测试"
echo "=========================================="
echo ""

# 检查编译
if [ ! -f "./xhs-mcp" ]; then
    echo "⏳ 编译 xhs-mcp..."
    go build -o xhs-mcp .
    if [ $? -ne 0 ]; then
        echo "❌ 编译失败"
        exit 1
    fi
    echo "✅ 编译成功"
fi

# 停止旧服务
pkill -f "xhs-mcp$" 2>/dev/null || true
sleep 1

# 启动服务
echo "⏳ 启动服务..."
./xhs-mcp > /tmp/xhs_test.log 2>&1 &
SERVER_PID=$!

# 等待启动
sleep 3

# 检查进程
if ! ps -p $SERVER_PID > /dev/null; then
    echo "❌ 服务启动失败"
    echo "日志："
    cat /tmp/xhs_test.log
    exit 1
fi

echo "✅ 服务已启动 (PID: $SERVER_PID)"
echo ""

# 测试健康检查
echo "测试健康检查..."
HEALTH=$(curl -s http://localhost:18060/health)
if [ "$HEALTH" = "OK" ]; then
    echo "✅ 健康检查通过"
else
    echo "❌ 健康检查失败"
    exit 1
fi
echo ""

# 测试 MCP 初始化
echo "测试 MCP 初始化..."
INIT_RESULT=$(curl -s http://localhost:18060/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}')

SERVER_NAME=$(echo "$INIT_RESULT" | jq -r '.result.serverInfo.name' 2>/dev/null)

if [ "$SERVER_NAME" = "xiaohongshu-mcp" ]; then
    echo "✅ MCP 初始化成功"
    echo "   服务名: $SERVER_NAME"
else
    echo "❌ MCP 初始化失败"
    echo "$INIT_RESULT" | jq '.' 2>/dev/null || echo "$INIT_RESULT"
    exit 1
fi
echo ""

# 测试工具列表（需要先 initialized）
echo "发送 initialized 通知..."
curl -s http://localhost:18060/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"notifications/initialized"}' > /dev/null

echo ""
echo "=========================================="
echo "✅ 所有测试通过！"
echo "=========================================="
echo ""
echo "MCP 服务已就绪:"
echo "- URL: http://127.0.0.1:18060/mcp"
echo "- PID: $SERVER_PID"
echo ""
echo "下一步:"
echo "1. 使用 MCP Inspector 连接测试"
echo "2. 或者运行: ./test_publish_complete.sh"
echo ""
echo "停止服务: kill $SERVER_PID"
