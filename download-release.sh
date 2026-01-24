#!/bin/bash

echo "=== 使用 gh CLI 下载 Release ==="

# 1. 验证 gh 登录状态
echo "步骤1: 验证登录状态..."
gh auth status

# 2. 下载 Linux amd64 二进制文件
echo ""
echo "步骤2: 下载二进制文件..."
gh release download v2026.01.24.2135-64dc373 \
  --repo vmxmy/xiaohongshu-mcp \
  --pattern "xiaohongshu-mcp-linux-amd64"

# 3. 重命名并添加执行权限
echo ""
echo "步骤3: 设置执行权限..."
if [ -f "xiaohongshu-mcp-linux-amd64" ]; then
    mv xiaohongshu-mcp-linux-amd64 xiaohongshu-mcp
    chmod +x xiaohongshu-mcp
    echo "✅ 下载成功: xiaohongshu-mcp"
    ls -lh xiaohongshu-mcp
else
    echo "❌ 下载失败"
    exit 1
fi

# 4. 验证文件
echo ""
echo "步骤4: 验证文件..."
file xiaohongshu-mcp
./xiaohongshu-mcp --help || echo "注意: 可能需要安装依赖"

echo ""
echo "=== 下载完成 ==="
