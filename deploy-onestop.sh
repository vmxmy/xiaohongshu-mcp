#!/bin/bash
# 小红书 MCP 一键部署脚本
# 功能: 检查环境 -> 更新版本 -> 下载二进制 -> 配置 -> 启动服务

set -e

REPO="vmxmy/xiaohongshu-mcp"
BINARY_NAME="xiaohongshu-mcp"
WORK_DIR=$(pwd)

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

echo "======================================"
echo "  小红书 MCP 一键部署脚本"
echo "======================================"
echo ""

# ========================================
# 步骤 1: 检测系统环境
# ========================================
log_info "步骤 1/7: 检测系统环境"

# 检测操作系统
OS=$(uname -s)
case "$OS" in
    Linux*)
        OS_TYPE="linux"
        ;;
    Darwin*)
        OS_TYPE="darwin"
        ;;
    *)
        log_error "不支持的操作系统: $OS"
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
        log_error "不支持的架构: $ARCH"
        exit 1
        ;;
esac

PLATFORM="${OS_TYPE}-${ARCH_TYPE}"
log_info "系统: $OS_TYPE | 架构: $ARCH_TYPE | 平台: $PLATFORM"
log_info "工作目录: $WORK_DIR"
echo ""

# ========================================
# 步骤 2: 检查必要工具
# ========================================
log_info "步骤 2/7: 检查必要工具"

# 检查 gh CLI
if ! command -v gh &> /dev/null; then
    log_error "未安装 gh CLI，请先安装:"
    echo "  Ubuntu/Debian: sudo apt install gh"
    echo "  CentOS/RHEL: sudo dnf install gh"
    echo "  macOS: brew install gh"
    exit 1
fi
log_info "✓ gh CLI 已安装"

# 检查 gh 登录状态
if ! gh auth status &> /dev/null; then
    log_error "gh CLI 未登录，请运行: gh auth login"
    exit 1
fi
log_info "✓ gh CLI 已登录"

# 检查 PM2 (可选)
if command -v pm2 &> /dev/null; then
    log_info "✓ PM2 已安装"
    HAS_PM2=true
else
    log_warn "PM2 未安装，将跳过服务管理"
    HAS_PM2=false
fi
echo ""

# ========================================
# 步骤 3: 检查并获取最新版本
# ========================================
log_info "步骤 3/7: 检查最新版本"

LATEST_VERSION=$(gh release list --repo $REPO --limit 1 | awk '{print $3}')
if [ -z "$LATEST_VERSION" ]; then
    log_error "无法获取最新版本"
    exit 1
fi
log_info "最新版本: $LATEST_VERSION"

# 检查当前版本
CURRENT_VERSION=""
if [ -f "./$BINARY_NAME" ]; then
    CURRENT_VERSION=$(./$BINARY_NAME --version 2>&1 | grep -oP 'Version: \K[^ ]+' || echo "unknown")
    log_info "当前版本: $CURRENT_VERSION"

    if [ "$CURRENT_VERSION" = "$LATEST_VERSION" ]; then
        log_info "已是最新版本"
        NEED_DOWNLOAD=false
    else
        log_warn "需要更新: $CURRENT_VERSION -> $LATEST_VERSION"
        NEED_DOWNLOAD=true
    fi
else
    log_warn "未安装，需要下载"
    NEED_DOWNLOAD=true
fi
echo ""

# ========================================
# 步骤 4: 下载二进制文件
# ========================================
if [ "$NEED_DOWNLOAD" = true ]; then
    log_info "步骤 4/7: 下载二进制文件"

    DOWNLOAD_FILE="xiaohongshu-mcp-${PLATFORM}"

    # 备份旧版本
    if [ -f "./$BINARY_NAME" ]; then
        BACKUP_NAME="${BINARY_NAME}.backup"
        mv "$BINARY_NAME" "$BACKUP_NAME"
        log_info "已备份旧版本: $BACKUP_NAME"
    fi

    # 下载
    log_info "正在下载 $LATEST_VERSION ..."
    if gh release download "$LATEST_VERSION" \
        --repo "$REPO" \
        --pattern "$DOWNLOAD_FILE" \
        --clobber; then

        # 重命名
        mv "$DOWNLOAD_FILE" "$BINARY_NAME"

        # 赋予执行权限
        chmod +x "$BINARY_NAME"

        log_info "✓ 下载成功"

        # 验证
        NEW_VERSION=$(./$BINARY_NAME --version 2>&1 | grep -oP 'Version: \K[^ ]+' || echo "unknown")
        log_info "✓ 验证通过: $NEW_VERSION"

        # 删除备份
        [ -f "$BACKUP_NAME" ] && rm "$BACKUP_NAME"
    else
        log_error "下载失败"
        # 恢复备份
        [ -f "$BACKUP_NAME" ] && mv "$BACKUP_NAME" "$BINARY_NAME"
        exit 1
    fi
else
    log_info "步骤 4/7: 跳过下载（已是最新版本）"
fi
echo ""

# ========================================
# 步骤 5: 检查配置文件
# ========================================
log_info "步骤 5/7: 检查配置文件"

# 检查 ecosystem.config.js
if [ ! -f "ecosystem.config.js" ]; then
    log_warn "未找到 ecosystem.config.js，创建默认配置..."

    cat > ecosystem.config.js << 'EOFCONFIG'
module.exports = {
  apps: [
    {
      name: 'xiaohongshu-mcp',
      script: './xiaohongshu-mcp',
      args: '--headless=true --port=:18060',
      cwd: process.cwd(),

      // 实例配置
      instances: 1,
      exec_mode: 'fork',

      // 自动重启
      autorestart: true,
      watch: false,
      max_memory_restart: '500M',

      // 环境变量
      env: {
        NODE_ENV: 'production',
        PORT: '18060',
      },

      // 日志配置
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z',
      error_file: './logs/mcp-error.log',
      out_file: './logs/mcp-out.log',
      log_file: './logs/mcp-combined.log',

      // 日志轮转
      max_restarts: 10,
      min_uptime: '10s',

      // 进程ID文件
      pid_file: './pids/mcp.pid',

      // 优雅退出
      kill_timeout: 5000,
      wait_ready: true,
      listen_timeout: 10000,

      // 错误处理
      merge_logs: true,
      time: true
    }
  ]
}
EOFCONFIG

    log_info "✓ 已创建 ecosystem.config.js"
else
    log_info "✓ ecosystem.config.js 已存在"

    # 更新 cwd 路径
    if grep -q "cwd:.*'/.*'" ecosystem.config.js; then
        log_warn "检测到硬编码路径，正在更新..."
        if [[ "$OS_TYPE" == "darwin" ]]; then
            # macOS 使用不同的 sed 语法
            sed -i '' "s|cwd: '.*'|cwd: '$WORK_DIR'|g" ecosystem.config.js
        else
            sed -i "s|cwd: '.*'|cwd: '$WORK_DIR'|g" ecosystem.config.js
        fi
        log_info "✓ 已更新工作目录: $WORK_DIR"
    fi
fi

# 创建必要目录
mkdir -p logs pids
log_info "✓ 已创建日志和进程目录"
echo ""

# ========================================
# 步骤 6: 检查浏览器
# ========================================
log_info "步骤 6/7: 检查浏览器环境"

if [ "$OS_TYPE" = "linux" ]; then
    # 检查 Chromium
    if command -v chromium &> /dev/null; then
        BROWSER_PATH=$(which chromium)
        log_info "✓ 已安装 Chromium: $BROWSER_PATH"
    elif command -v chromium-browser &> /dev/null; then
        BROWSER_PATH=$(which chromium-browser)
        log_info "✓ 已安装 Chromium: $BROWSER_PATH"
    elif command -v google-chrome &> /dev/null; then
        BROWSER_PATH=$(which google-chrome)
        log_info "✓ 已安装 Chrome: $BROWSER_PATH"
    else
        log_warn "未检测到浏览器，尝试安装 Chromium..."
        if command -v apt &> /dev/null; then
            sudo apt update && sudo apt install -y chromium chromium-driver
            log_info "✓ 已安装 Chromium"
        else
            log_warn "无法自动安装浏览器，请手动安装 chromium 或 google-chrome"
        fi
    fi
else
    log_info "macOS 环境，Rod 会自动下载浏览器"
fi
echo ""

# ========================================
# 步骤 7: 启动服务
# ========================================
log_info "步骤 7/7: 启动服务"

if [ "$HAS_PM2" = true ]; then
    # 检查服务是否已运行
    if pm2 describe xiaohongshu-mcp &> /dev/null; then
        log_info "服务已运行，正在重启..."
        pm2 restart xiaohongshu-mcp
        log_info "✓ 服务已重启"
    else
        log_info "正在启动服务..."
        pm2 start ecosystem.config.js
        log_info "✓ 服务已启动"
    fi

    # 保存 PM2 配置
    pm2 save
    log_info "✓ PM2 配置已保存"

    echo ""
    log_info "查看服务状态:"
    pm2 status

    echo ""
    log_info "查看实时日志:"
    echo "  pm2 logs xiaohongshu-mcp"
else
    log_warn "未安装 PM2，手动启动服务:"
    echo "  ./$BINARY_NAME --headless=true --port=:18060"
fi

echo ""
echo "======================================"
echo "  部署完成！"
echo "======================================"
echo ""
echo "服务信息:"
echo "  版本: $LATEST_VERSION"
echo "  端口: 18060"
echo "  工作目录: $WORK_DIR"
echo ""
echo "常用命令:"
echo "  pm2 status              # 查看服务状态"
echo "  pm2 logs xiaohongshu-mcp  # 查看日志"
echo "  pm2 restart xiaohongshu-mcp  # 重启服务"
echo "  pm2 stop xiaohongshu-mcp     # 停止服务"
echo ""
