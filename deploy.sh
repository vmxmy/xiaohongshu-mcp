#!/bin/bash
set -e

echo "=== 小红书MCP服务器部署脚本 ==="

# 配置变量
DEPLOY_DIR="/opt/xiaohongshu-mcp"
VERSION="v2026.01.24.2135-64dc373"
GITHUB_REPO="vmxmy/xiaohongshu-mcp"

# 1. 安装依赖
echo "步骤1: 安装依赖..."
if command -v apt &> /dev/null; then
    sudo apt update
    sudo apt install -y chromium-browser curl wget
elif command -v yum &> /dev/null; then
    sudo yum install -y chromium curl wget
fi

# 2. 安装PM2
echo "步骤2: 安装PM2..."
if ! command -v pm2 &> /dev/null; then
    if ! command -v npm &> /dev/null; then
        echo "错误: 未安装Node.js，请先安装Node.js和npm"
        echo "访问: https://nodejs.org/ 或使用包管理器安装"
        exit 1
    fi
    npm install -g pm2
fi

# 3. 创建部署目录
echo "步骤3: 创建部署目录..."
sudo mkdir -p $DEPLOY_DIR
sudo chown $USER:$USER $DEPLOY_DIR
cd $DEPLOY_DIR

# 4. 下载二进制文件
echo "步骤4: 下载MCP服务器..."
wget "https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/xiaohongshu-mcp-linux-amd64" -O xiaohongshu-mcp
chmod +x xiaohongshu-mcp

# 5. 创建目录
echo "步骤5: 创建必要目录..."
mkdir -p logs pids

# 6. 创建配置文件
echo "步骤6: 创建PM2配置文件..."
cat > ecosystem.config.js << 'EOFCONFIG'
module.exports = {
  apps: [
    {
      name: 'xiaohongshu-mcp',
      script: './xiaohongshu-mcp',
      args: '--headless=true --port=:18060',
      cwd: '/opt/xiaohongshu-mcp',
      instances: 1,
      exec_mode: 'fork',
      autorestart: true,
      watch: false,
      max_memory_restart: '500M',
      env: {
        NODE_ENV: 'production',
        PORT: '18060'
      },
      log_date_format: 'YYYY-MM-DD HH:mm:ss Z',
      error_file: './logs/mcp-error.log',
      out_file: './logs/mcp-out.log',
      log_file: './logs/mcp-combined.log',
      max_restarts: 10,
      min_uptime: '10s',
      pid_file: './pids/mcp.pid',
      kill_timeout: 5000,
      wait_ready: true,
      listen_timeout: 10000,
      merge_logs: true,
      time: true
    }
  ]
}
EOFCONFIG

# 7. 启动服务
echo "步骤7: 启动PM2服务..."
pm2 start ecosystem.config.js

# 8. 配置开机自启
echo "步骤8: 配置开机自启..."
pm2 save
echo ""
echo "请执行以下命令以启用开机自启："
pm2 startup systemd | grep "sudo"

echo ""
echo "=== 部署完成 ==="
echo ""
echo "常用命令:"
echo "  查看状态: pm2 status"
echo "  查看日志: pm2 logs xiaohongshu-mcp"
echo "  重启服务: pm2 restart xiaohongshu-mcp"
echo "  停止服务: pm2 stop xiaohongshu-mcp"
echo ""
echo "测试API:"
echo "  curl http://localhost:18060/health"
echo ""
echo "部署目录: $DEPLOY_DIR"
