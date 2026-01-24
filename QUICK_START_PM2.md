# PM2快速启动指南

## 一键部署（推荐）

```bash
# 1. 下载部署脚本
wget https://raw.githubusercontent.com/vmxmy/xiaohongshu-mcp/main/deploy.sh

# 2. 添加执行权限
chmod +x deploy.sh

# 3. 运行部署
./deploy.sh

# 4. 按提示执行开机自启命令（会输出具体命令）
sudo env PATH=$PATH:/usr/bin pm2 startup systemd -u $USER --hp $HOME

# 5. 验证运行
pm2 status
pm2 logs xiaohongshu-mcp
```

## 手动部署

### 1. 安装依赖

```bash
# Ubuntu/Debian
sudo apt update && sudo apt install -y chromium-browser nodejs npm

# CentOS/RHEL  
sudo yum install -y chromium nodejs npm

# 安装PM2
sudo npm install -g pm2
```

### 2. 下载MCP服务器

```bash
# 创建目录
sudo mkdir -p /opt/xiaohongshu-mcp
cd /opt/xiaohongshu-mcp

# 下载（选择对应架构）
# AMD64
wget https://github.com/vmxmy/xiaohongshu-mcp/releases/download/v2026.01.24.2135-64dc373/xiaohongshu-mcp-linux-amd64 -O xiaohongshu-mcp

# ARM64
# wget https://github.com/vmxmy/xiaohongshu-mcp/releases/download/v2026.01.24.2135-64dc373/xiaohongshu-mcp-linux-arm64 -O xiaohongshu-mcp

chmod +x xiaohongshu-mcp
mkdir -p logs pids
```

### 3. 下载配置文件

```bash
wget https://raw.githubusercontent.com/vmxmy/xiaohongshu-mcp/main/ecosystem.config.js
```

### 4. 启动服务

```bash
pm2 start ecosystem.config.js
pm2 save
```

### 5. 配置开机自启

```bash
pm2 startup systemd
# 执行输出的sudo命令
```

## 常用命令

```bash
# 查看状态
pm2 status

# 查看日志
pm2 logs xiaohongshu-mcp

# 重启服务
pm2 restart xiaohongshu-mcp

# 停止服务
pm2 stop xiaohongshu-mcp

# 实时监控
pm2 monit
```

## 验证运行

```bash
# 检查进程
pm2 status

# 检查端口
ss -tlnp | grep 18060

# 测试API
curl -X POST http://localhost:18060/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

## 目录结构

```
/opt/xiaohongshu-mcp/
├── xiaohongshu-mcp          # 主程序
├── ecosystem.config.js      # PM2配置
├── cookies.json             # 登录凭证（首次登录后生成）
├── logs/                    # 日志目录
│   ├─�� mcp-error.log
│   ├── mcp-out.log
│   └── mcp-combined.log
└── pids/                    # 进程ID文件
    └── mcp.pid
```

## 故障排查

### 服务无法启动

```bash
# 查看详细日志
pm2 logs xiaohongshu-mcp --lines 50

# 检查二进制文件
ls -l /opt/xiaohongshu-mcp/xiaohongshu-mcp
./xiaohongshu-mcp --help

# 检查Chromium
which chromium-browser
```

### 端口被占用

```bash
# 查看占用
sudo lsof -i :18060

# 修改端口
# 编辑 ecosystem.config.js
# args: '--headless=true --port=:18061'
pm2 restart xiaohongshu-mcp
```

### 查看完整文档

```bash
# 下载完整部署文档
wget https://raw.githubusercontent.com/vmxmy/xiaohongshu-mcp/main/PM2_DEPLOYMENT.md
```

## 更新服务

```bash
cd /opt/xiaohongshu-mcp

# 备份当前版本
mv xiaohongshu-mcp xiaohongshu-mcp.backup

# 下载新版本
wget https://github.com/vmxmy/xiaohongshu-mcp/releases/download/NEW_VERSION/xiaohongshu-mcp-linux-amd64 -O xiaohongshu-mcp
chmod +x xiaohongshu-mcp

# 重启
pm2 restart xiaohongshu-mcp
```

## 完全卸载

```bash
# 停止并删除服务
pm2 stop xiaohongshu-mcp
pm2 delete xiaohongshu-mcp
pm2 save

# 删除部署目录
sudo rm -rf /opt/xiaohongshu-mcp

# 禁用开机自启
pm2 unstartup systemd
```
