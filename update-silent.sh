#!/bin/bash
# 小红书 MCP 静默自动更新脚本（适合 cron 定时任务）
set -e

REPO="vmxmy/xiaohongshu-mcp"
BINARY_NAME="xiaohongshu-mcp"
LOG_FILE="update.log"

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

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
        log "ERROR: 不支持的操作系统 $OS"
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
        log "ERROR: 不支持的架构 $ARCH"
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

log "系统: $OS_TYPE, 架构: $ARCH_TYPE, 平台: $PLATFORM"

# 获取最新版本
LATEST_VERSION=$(gh release list --repo $REPO --limit 1 2>/dev/null | awk '{print $3}')
if [ -z "$LATEST_VERSION" ]; then
    log "ERROR: 无法获取最新版本"
    exit 1
fi

# 检查当前版本
CURRENT_VERSION=""
FINAL_BINARY="${BINARY_NAME}${EXT}"
if [ -f "./$FINAL_BINARY" ]; then
    CURRENT_VERSION=$(./$FINAL_BINARY --version 2>&1 | grep -oP 'Version: \K[^ ]+' || echo "unknown")
fi

log "当前版本: ${CURRENT_VERSION:-未安装}"
log "最新版本: $LATEST_VERSION"

# 比较版本
if [ "$CURRENT_VERSION" = "$LATEST_VERSION" ]; then
    log "已是最新版本，无需更新"
    exit 0
fi

# 下载新版本
log "开始下载 $LATEST_VERSION ..."
DOWNLOAD_FILE="xiaohongshu-mcp-${PLATFORM}${EXT}"

# 备份旧版本
if [ -f "./$FINAL_BINARY" ]; then
    BACKUP_NAME="${FINAL_BINARY}.backup"
    mv "$FINAL_BINARY" "$BACKUP_NAME"
    log "已备份当前版本"
fi

# 下载
if gh release download "$LATEST_VERSION" \
    --repo "$REPO" \
    --pattern "$DOWNLOAD_FILE" \
    --clobber >> "$LOG_FILE" 2>&1; then

    mv "$DOWNLOAD_FILE" "$FINAL_BINARY"
    chmod +x "$FINAL_BINARY"
    log "下载成功: $LATEST_VERSION"

    # 验证
    NEW_VERSION=$(./$FINAL_BINARY --version 2>&1 | grep -oP 'Version: \K[^ ]+' || echo "unknown")
    log "验证通过: $NEW_VERSION"

    # 删除旧备份
    [ -f "${BACKUP_NAME}" ] && rm "${BACKUP_NAME}"

    # 重启 PM2 服务（如果在运行）
    if command -v pm2 &> /dev/null; then
        if pm2 describe xiaohongshu-mcp &> /dev/null; then
            log "重启 PM2 服务..."
            pm2 restart xiaohongshu-mcp >> "$LOG_FILE" 2>&1
            log "服务已重启"
        fi
    fi

    log "更新完成"
else
    log "ERROR: 下载失败"
    # 恢复备份
    if [ -f "${BACKUP_NAME}" ]; then
        mv "${BACKUP_NAME}" "$FINAL_BINARY"
        log "已恢复备份版本"
    fi
    exit 1
fi
