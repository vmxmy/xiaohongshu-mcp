#!/bin/bash
set -e

REPO="vmxmy/xiaohongshu-mcp"
BINARY_NAME="xiaohongshu-mcp"

echo "=== 小红书 MCP 自动更新脚本 ==="
echo ""

# 检测操作系统
OS=$(uname -s)
case "$OS" in
    Linux*)
        OS_TYPE="linux"
        ;;
    Darwin*)
        OS_TYPE="darwin"
        ;;
    MINGW*|MSYS*|CYGWIN*)
        OS_TYPE="windows"
        ;;
    *)
        echo "❌ 不支持的操作系统: $OS"
        exit 1
        ;;
esac

# 检测架构
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64)
        ARCH_TYPE="amd64"
        ;;
    aarch64|arm64)
        ARCH_TYPE="arm64"
        ;;
    *)
        echo "❌ 不支持的架构: $ARCH"
        exit 1
        ;;
esac

# 组合平台标识
PLATFORM="${OS_TYPE}-${ARCH_TYPE}"

# Windows 需要 .exe 后缀
EXT=""
if [ "$OS_TYPE" = "windows" ]; then
    EXT=".exe"
fi

echo "📍 检测到系统: $OS ($OS_TYPE)"
echo "📍 检测到架构: $ARCH ($ARCH_TYPE)"
echo "📍 目标平台: $PLATFORM"
echo ""

# 获取最新版本
echo "🔍 检查最新版本..."
LATEST_VERSION=$(gh release list --repo $REPO --limit 1 | awk '{print $1}')

if [ -z "$LATEST_VERSION" ]; then
    echo "❌ 无法获取最新版本，请确保已登录 gh CLI"
    exit 1
fi

echo "✅ 最新版本: $LATEST_VERSION"
echo ""

# 检查当前版本
CURRENT_VERSION=""
FINAL_BINARY="${BINARY_NAME}${EXT}"
if [ -f "./$FINAL_BINARY" ]; then
    CURRENT_VERSION=$(./$FINAL_BINARY --version 2>&1 | grep -oP 'Version: \K[^ ]+' || echo "unknown")
    echo "📦 当前版本: $CURRENT_VERSION"
else
    echo "📦 当前版本: 未安装"
fi
echo ""

# 比较版本
if [ "$CURRENT_VERSION" = "$LATEST_VERSION" ]; then
    echo "✅ 已是最新版本，无需更新"
    exit 0
fi

# 下载新版本
echo "⬇️  开始下载 $LATEST_VERSION ..."
DOWNLOAD_FILE="xiaohongshu-mcp-${PLATFORM}${EXT}"
FINAL_BINARY="${BINARY_NAME}${EXT}"

# 备份旧版本
if [ -f "./$FINAL_BINARY" ]; then
    BACKUP_NAME="${FINAL_BINARY}.backup.$(date +%Y%m%d%H%M%S)"
    echo "💾 备份当前版本到: $BACKUP_NAME"
    mv "$FINAL_BINARY" "$BACKUP_NAME"
fi

# 下载新版本
gh release download "$LATEST_VERSION" \
    --repo "$REPO" \
    --pattern "$DOWNLOAD_FILE" \
    --clobber

# 重命名并赋予权限
if [ -f "$DOWNLOAD_FILE" ]; then
    mv "$DOWNLOAD_FILE" "$FINAL_BINARY"
    chmod +x "$FINAL_BINARY"
    echo "✅ 下载成功"
else
    echo "❌ 下载失败: 找不到文件 $DOWNLOAD_FILE"
    echo "💡 可用的文件:"
    gh release view "$LATEST_VERSION" --repo "$REPO" | grep "xiaohongshu-mcp"
    # 恢复备份
    if [ -f "$BACKUP_NAME" ]; then
        echo "🔄 恢复备份版本..."
        mv "$BACKUP_NAME" "$FINAL_BINARY"
    fi
    exit 1
fi

# 验证新版本
echo ""
echo "🔍 验证新版本..."
NEW_VERSION=$(./$FINAL_BINARY --version 2>&1 | grep -oP 'Version: \K[^ ]+' || echo "unknown")
echo "✅ 当前版本: $NEW_VERSION"

# 清理备份（可选）
if [ -f "$BACKUP_NAME" ]; then
    echo ""
    read -p "是否删除备份文件 $BACKUP_NAME? (y/n) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm "$BACKUP_NAME"
        echo "🗑️  已删除备份"
    else
        echo "💾 备份保留: $BACKUP_NAME"
    fi
fi

echo ""
echo "=== 更新完成 ==="
echo ""
echo "如果使用 PM2 管理服务，请运行:"
echo "  pm2 restart xiaohongshu-mcp"
