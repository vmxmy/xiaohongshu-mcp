# 自动更新脚本使用说明

本项目提供两个自动更新脚本，帮助你在远程服务器上自动检查和更新到最新版本。

## 脚本说明

### 1. `update.sh` - 交互式更新

**特点**:
- ✅ 显示详细的更新过程
- ✅ 版本对比和确认
- ✅ 询问是否删除备份
- ✅ 适合手动执行

**使用方法**:
```bash
# 下载脚本
wget https://raw.githubusercontent.com/vmxmy/xiaohongshu-mcp/main/update.sh
chmod +x update.sh

# 运行更新
./update.sh
```

### 2. `update-silent.sh` - 静默更新

**特点**:
- ✅ 静默执行，无需交互
- ✅ 自动备份和恢复
- ✅ 日志记录（update.log）
- ✅ 自动重启 PM2 服务
- ✅ 适合 cron 定时任务

**使用方法**:
```bash
# 下载脚本
wget https://raw.githubusercontent.com/vmxmy/xiaohongshu-mcp/main/update-silent.sh
chmod +x update-silent.sh

# 手动运行
./update-silent.sh

# 或添加到 cron（每天凌晨2点检查更新）
crontab -e
# 添加以下行:
0 2 * * * cd /path/to/xiaohongshu-mcp && ./update-silent.sh
```

## 前置要求

两个脚本都需要：
1. ✅ 已安装 `gh` CLI 工具
2. ✅ 已使用 `gh auth login` 登录

### 安装 gh CLI

**Ubuntu/Debian**:
```bash
sudo apt update
sudo apt install gh
gh auth login
```

**CentOS/RHEL**:
```bash
sudo dnf install gh
gh auth login
```

## 完整部署示例

### 首次部署

```bash
# 1. 进入工作目录
cd /home/dev/app/xiaohongshu-mcp

# 2. 下载更新脚本
wget https://raw.githubusercontent.com/vmxmy/xiaohongshu-mcp/main/update-silent.sh
chmod +x update-silent.sh

# 3. 运行更新（会自动下载最新版本）
./update-silent.sh

# 4. 启动服务
pm2 start ecosystem.config.js
pm2 save
```

### 设置自动更新

```bash
# 编辑 crontab
crontab -e

# 添加定时任务（每天凌晨2点自动检查并更新）
0 2 * * * cd /home/dev/app/xiaohongshu-mcp && ./update-silent.sh >> /home/dev/app/xiaohongshu-mcp/update-cron.log 2>&1
```

## 脚本工作流程

### update.sh（交互式）
1. 检测系统架构（amd64/arm64）
2. 获取 GitHub 最新版本
3. 检查本地当前版本
4. 版本对比
5. 如果有新版本：
   - 备份当前版本
   - 下载新版本
   - 赋予执行权限
   - 验证新版本
   - 询问是否删除备份

### update-silent.sh（静默式）
1. 检测系统架构
2. 获取最新版本
3. 检查当前版本并记录日志
4. 版本对比
5. 如果有新版本：
   - 备份当前版本
   - 下载新版本
   - 赋予执行权限
   - 验证新版本
   - 自动删除备份
   - **自动重启 PM2 服务**
6. 所有操作记录到 `update.log`

## 日志查看

### update-silent.sh 日志

```bash
# 查看更新日志
tail -f update.log

# 示例输出:
# [2026-01-24 22:35:00] 当前版本: v2026.01.24.2135-64dc373
# [2026-01-24 22:35:01] 最新版本: v2026.01.24.2233-91ae7a1
# [2026-01-24 22:35:02] 开始下载 v2026.01.24.2233-91ae7a1 ...
# [2026-01-24 22:35:05] 下载成功: v2026.01.24.2233-91ae7a1
# [2026-01-24 22:35:05] 验证通过: v2026.01.24.2233-91ae7a1
# [2026-01-24 22:35:06] 重启 PM2 服务...
# [2026-01-24 22:35:08] 服务已重启
# [2026-01-24 22:35:08] 更新完成
```

### Cron 日志

```bash
# 查看 cron 执行日志
tail -f update-cron.log
```

## 故障恢复

如果更新失败，脚本会自动恢复备份：

```bash
# 手动恢复备份
mv xiaohongshu-mcp.backup xiaohongshu-mcp

# 重启服务
pm2 restart xiaohongshu-mcp
```

## 版本锁定

如果你想锁定在特定版本，可以修改 cron 任务：

```bash
# 禁用自动更新
crontab -e
# 注释掉或删除更新任务行

# 或者手动下载特定版本
gh release download v2026.01.24.2135-64dc373 \
  --repo vmxmy/xiaohongshu-mcp \
  --pattern "xiaohongshu-mcp-linux-amd64"
```

## 常见问题

### Q: 脚本提示 "无法获取最新版本"
**A**: 确保已使用 `gh auth login` 登录 GitHub CLI

### Q: 下载速度慢
**A**: 可以配置 GitHub CLI 使用代理：
```bash
export https_proxy=http://your-proxy:port
```

### Q: PM2 服务没有自动重启
**A**: 确保：
1. PM2 已安装并运行
2. 服务名称为 `xiaohongshu-mcp`（在 ecosystem.config.js 中配置）

### Q: 如何验证更新成功？
**A**: 运行以下命令：
```bash
./xiaohongshu-mcp --version
pm2 logs xiaohongshu-mcp --lines 20
```

## 安全建议

1. **定期检查日志**: `tail -100 update.log`
2. **保留最近备份**: 不要立即删除 `.backup` 文件
3. **测试新版本**: 更新后验证服务正常运行
4. **监控服务状态**: 使用 `pm2 monit`

## 卸载

```bash
# 删除更新脚本
rm update.sh update-silent.sh

# 删除日志文件
rm update.log update-cron.log

# 删除 cron 任务
crontab -e
# 删除相关行
```
