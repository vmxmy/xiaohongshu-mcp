#!/bin/bash

echo "======================================"
echo "启动小红书登录程序（有头浏览器）"
echo "======================================"
echo ""

# 检查登录程序是否存在
if [ ! -f "./bin/xiaohongshu-login" ]; then
    echo "❌ 未找到登录程序: ./bin/xiaohongshu-login"
    exit 1
fi

echo "🚀 启动有头浏览器登录..."
echo ""
echo "📋 操作步骤:"
echo "  1. 等待浏览器窗口打开"
echo "  2. 在浏览器中扫描二维码登录"
echo "  3. 登录成功后，cookies 会自动保存"
echo "  4. 浏览器窗口会自动关闭"
echo ""

# 启动有头浏览器登录（不使用 headless 模式）
./bin/xiaohongshu-login

echo ""
echo "✅ 登录程序已退出"
echo ""
echo "📝 接下来需要同步 cookies 到远程服务器（如果需要）"
