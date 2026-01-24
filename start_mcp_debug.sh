#!/bin/bash

# 启动 MCP 服务器（用于 MCP Inspector 调试）

echo "=========================================="
echo "启动 MCP 服务器（调试模式）"
echo "=========================================="
echo ""

echo "配置："
echo "- 调试日志：开启"
echo "- Headless：关闭（显示浏览器）"
echo "- 端口：18060"
echo "- MCP 端点：http://127.0.0.1:18060/mcp"
echo ""

echo "日志将保存到: mcp_debug.log"
echo ""

echo "✅ 启动中..."
echo ""

# 启动服务
LOG_LEVEL=debug ./xhs-mcp -headless=false -port=:18060 2>&1 | tee mcp_debug.log
