#!/bin/bash

# 小红书 MCP 工具构建脚本
# 用于编译登录工具和MCP服务

set -e

echo "🚀 开始编译小红书 MCP 工具..."
echo ""

# 获取版本信息
VERSION=${1:-"v1.0.0"}
BUILD_TIME=$(date +"%Y-%m-%d %H:%M:%S")
PLATFORM=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# 转换架构名称
if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
fi

OUTPUT_DIR="./bin"
mkdir -p "$OUTPUT_DIR"

echo "📦 编译信息:"
echo "  版本: $VERSION"
echo "  时间: $BUILD_TIME"
echo "  平台: $PLATFORM-$ARCH"
echo ""

# 1. 编译登录工具
echo "🔨 编译登录工具..."
go build -o "$OUTPUT_DIR/xiaohongshu-login" \
    -ldflags "-X 'main.Version=$VERSION' -X 'main.BuildTime=$BUILD_TIME'" \
    ./cmd/login/main.go

if [ $? -eq 0 ]; then
    echo "✅ 登录工具编译成功: $OUTPUT_DIR/xiaohongshu-login"
    ls -lh "$OUTPUT_DIR/xiaohongshu-login"
else
    echo "❌ 登录工具编译失败"
    exit 1
fi

echo ""

# 2. 编译MCP服务
echo "🔨 编译MCP服务..."
go build -o "$OUTPUT_DIR/xiaohongshu-mcp" \
    -ldflags "-X 'main.Version=$VERSION' -X 'main.BuildTime=$BUILD_TIME'" \
    .

if [ $? -eq 0 ]; then
    echo "✅ MCP服务编译成功: $OUTPUT_DIR/xiaohongshu-mcp"
    ls -lh "$OUTPUT_DIR/xiaohongshu-mcp"
else
    echo "❌ MCP服务编译失败"
    exit 1
fi

echo ""
echo "🎉 编译完成！"
echo ""
echo "📝 使用说明:"
echo "  1. 首次使用，先运行登录工具:"
echo "     $OUTPUT_DIR/xiaohongshu-login"
echo ""
echo "  2. 登录成功后，启动MCP服务:"
echo "     $OUTPUT_DIR/xiaohongshu-mcp"
echo ""
echo "  3. 配置MCP客户端，添加以下配置:"
echo "     {\"command\": \"$(pwd)/$OUTPUT_DIR/xiaohongshu-mcp\"}"
echo ""
