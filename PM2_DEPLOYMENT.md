# PM2 部署指南 - 小红书MCP服务器

## 环境要求

- Linux系统 (Ubuntu 20.04+, CentOS 7+, Debian 10+)
- Node.js 14+ (PM2需要)
- Chromium/Chrome浏览器（无头模式）
- 二进制文件：xiaohongshu-mcp

## 安装步骤

### 1. 安装PM2

```bash
# 使用npm全局安装PM2
npm install -g pm2

# 验证安装
pm2 -v
```

### 2. 安装Chromium（无头浏览器）

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install -y chromium-browser
# 或者
sudo apt install -y chromium
```

**CentOS/RHEL:**
```bash
sudo yum install -y chromium
```

**验证安装:**
```bash
which chromium-browser  # Ubuntu
which chromium          # CentOS
```

### 3. 下载并配置MCP服务器

```bash
# 创建部署目录
sudo mkdir -p /opt/xiaohongshu-mcp
cd /opt/xiaohongshu-mcp

# 下载对应平台的二进制文件
# 从 https://github.com/vmxmy/xiaohongshu-mcp/releases/latest
wget https://github.com/vmxmy/xiaohongshu-mcp/releases/download/v2026.01.24.2135-64dc373/xiaohongshu-mcp-linux-amd64

# 重命名并添加执行权限
mv xiaohongshu-mcp-linux-amd64 xiaohongshu-mcp
chmod +x xiaohongshu-mcp

# 创建必要的目录
mkdir -p logs pids

# 测试运行
./xiaohongshu-mcp --help
```

### 4. 配置PM2

```bash
# 复制ecosystem.config.js到部署目录
cp ecosystem.config.js /opt/xiaohongshu-mcp/

# 编辑配置文件，修改cwd路径
nano ecosystem.config.js
# 将 cwd: '/path/to/xiaohongshu-mcp' 改为 cwd: '/opt/xiaohongshu-mcp'
```

**如果Chromium路径非标准，需要设置环境变量:**
```javascript
env: {
  NODE_ENV: 'production',
  PORT: '18060',
  ROD_BROWSER_BIN: '/usr/bin/chromium-browser'  // 取消注释并设置正确路径
}
```

## PM2 常用命令

### 启动服务

```bash
# 使用配置文件启动
pm2 start ecosystem.config.js

# 或直接启动（不推荐）
pm2 start ./xiaohongshu-mcp --name xiaohongshu-mcp -- --headless=true --port=:18060
```

### 管理服务

```bash
# 查看状态
pm2 status
pm2 list

# 查看详细信息
pm2 show xiaohongshu-mcp

# 查看日志
pm2 logs xiaohongshu-mcp          # 实时日志
pm2 logs xiaohongshu-mcp --lines 100  # 最近100行
pm2 logs --err                     # 只看错误日志

# 停止服务
pm2 stop xiaohongshu-mcp

# 重启服务
pm2 restart xiaohongshu-mcp

# 重载配置（零停机）
pm2 reload xiaohongshu-mcp

# 删除服务
pm2 delete xiaohongshu-mcp

# 清空日志
pm2 flush
```

### 监控

```bash
# 实时监控面板
pm2 monit

# Web监控界面（可选）
pm2 web
# 访问 http://localhost:9615
```

## 开机自启动

### 1. 生成PM2启动脚本

```bash
# 生成systemd启动脚本
pm2 startup systemd

# 执行输出的命令（类似下面的格式）
sudo env PATH=$PATH:/usr/bin pm2 startup systemd -u $USER --hp $HOME

# 保存当前PM2进程列表
pm2 save
```

### 2. 验证自启动

```bash
# 重启系统测试
sudo reboot

# 重启后检查
pm2 list
```

### 3. 禁用自启动

```bash
pm2 unstartup systemd
```

## 配置文件说明

### ecosystem.config.js 关键参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `name` | 应用名称 | xiaohongshu-mcp |
| `script` | 执行文件路径 | ./xiaohongshu-mcp |
| `args` | 命令行参数 | --headless=true --port=:18060 |
| `cwd` | 工作目录 | 需要修改为实际路径 |
| `instances` | 实例数量 | 1（建议保持1个实例） |
| `autorestart` | 自动重启 | true |
| `max_memory_restart` | 内存超限重启 | 500M |
| `error_file` | 错误日志路径 | ./logs/mcp-error.log |
| `out_file` | 标准输出日志 | ./logs/mcp-out.log |

## 日志管理

### 日志位置

```bash
cd /opt/xiaohongshu-mcp/logs
ls -lh
# mcp-error.log    - 错误日志
# mcp-out.log      - 标准输出
# mcp-combined.log - 合并日志
```

### 日志轮转

安装PM2日志轮转模块：

```bash
pm2 install pm2-logrotate

# 配置日志轮转
pm2 set pm2-logrotate:max_size 10M        # 单个日志文件最大10MB
pm2 set pm2-logrotate:retain 30           # 保留30个日志文件
pm2 set pm2-logrotate:compress true       # 压缩旧日志
pm2 set pm2-logrotate:dateFormat YYYY-MM-DD_HH-mm-ss
pm2 set pm2-logrotate:rotateModule true
```

## 健康检查

### 检查服务状态

```bash
# 检查进程
pm2 status

# 检查端口
ss -tlnp | grep 18060
# 或
netstat -tlnp | grep 18060

# 测试API
curl http://localhost:18060/health
# 或
curl -X POST http://localhost:18060/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

### 性能监控

```bash
# CPU和内存使用
pm2 monit

# 详细指标
pm2 show xiaohongshu-mcp
```

## 故障排查

### 服务无法启动

1. **检查二进制文件权限**
```bash
ls -l /opt/xiaohongshu-mcp/xiaohongshu-mcp
chmod +x /opt/xiaohongshu-mcp/xiaohongshu-mcp
```

2. **检查Chromium是否安装**
```bash
which chromium-browser
chromium-browser --version
```

3. **查看错误日志**
```bash
pm2 logs xiaohongshu-mcp --err --lines 50
```

### 端口被占用

```bash
# 查看端口占用
sudo lsof -i :18060

# 修改端口
# 编辑 ecosystem.config.js，修改 args 中的 --port 参数
```

### 内存溢出

```bash
# 增加内存限制
# 编辑 ecosystem.config.js
max_memory_restart: '1G'  # 改为1GB
```

## 安全建议

### 1. 防火墙配置

```bash
# Ubuntu/Debian (ufw)
sudo ufw allow 18060/tcp
sudo ufw status

# CentOS/RHEL (firewalld)
sudo firewall-cmd --permanent --add-port=18060/tcp
sudo firewall-cmd --reload
```

### 2. 反向代理（可选）

使用Nginx作为反向代理：

```nginx
server {
    listen 80;
    server_name mcp.example.com;

    location / {
        proxy_pass http://localhost:18060;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
        
        # 超时设置
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}
```

### 3. 限制访问

```bash
# 只监听本地
# 修改 ecosystem.config.js 中的 args
args: '--headless=true --port=localhost:18060'
```

## 升级服务

```bash
# 1. 下载新版本
cd /opt/xiaohongshu-mcp
wget https://github.com/vmxmy/xiaohongshu-mcp/releases/download/NEW_VERSION/xiaohongshu-mcp-linux-amd64 -O xiaohongshu-mcp.new

# 2. 备份当前版本
mv xiaohongshu-mcp xiaohongshu-mcp.backup

# 3. 使用新版本
mv xiaohongshu-mcp.new xiaohongshu-mcp
chmod +x xiaohongshu-mcp

# 4. 重启服务
pm2 restart xiaohongshu-mcp

# 5. 验证
pm2 logs xiaohongshu-mcp --lines 20

# 6. 如果有问题，回滚
# mv xiaohongshu-mcp.backup xiaohongshu-mcp
# pm2 restart xiaohongshu-mcp
```

## 完整部署脚本

保存为 `deploy.sh`:

```bash
#!/bin/bash
set -e

echo "=== 小红书MCP服务器部署脚本 ==="

# 1. 安装依赖
echo "步骤1: 安装依赖..."
sudo apt update
sudo apt install -y chromium-browser curl wget

# 2. 安装PM2
echo "步骤2: 安装PM2..."
if ! command -v pm2 &> /dev/null; then
    npm install -g pm2
fi

# 3. 创建部署目录
echo "步骤3: 创建部署目录..."
DEPLOY_DIR="/opt/xiaohongshu-mcp"
sudo mkdir -p $DEPLOY_DIR
sudo chown $USER:$USER $DEPLOY_DIR
cd $DEPLOY_DIR

# 4. 下载二进制文件
echo "步骤4: 下载MCP服务器..."
VERSION="v2026.01.24.2135-64dc373"
wget "https://github.com/vmxmy/xiaohongshu-mcp/releases/download/${VERSION}/xiaohongshu-mcp-linux-amd64" -O xiaohongshu-mcp
chmod +x xiaohongshu-mcp

# 5. 创建目录
echo "步骤5: 创建必要目录..."
mkdir -p logs pids

# 6. 下载配置文件
echo "步骤6: 下载配置文件..."
wget "https://raw.githubusercontent.com/vmxmy/xiaohongshu-mcp/main/ecosystem.config.js" -O ecosystem.config.js
# 修改cwd路径
sed -i "s|/path/to/xiaohongshu-mcp|$DEPLOY_DIR|g" ecosystem.config.js

# 7. 启动服务
echo "步骤7: 启动PM2服务..."
pm2 start ecosystem.config.js

# 8. 配置开机自启
echo "步骤8: 配置开机自启..."
pm2 save
pm2 startup systemd

echo ""
echo "=== 部署完成 ==="
echo "查看状态: pm2 status"
echo "查看日志: pm2 logs xiaohongshu-mcp"
echo "测试API: curl http://localhost:18060/health"
```

使用方法：
```bash
chmod +x deploy.sh
./deploy.sh
```

## 参考文档

- PM2官方文档: https://pm2.keymetrics.io/docs/
- 小红书MCP项目: https://github.com/vmxmy/xiaohongshu-mcp
