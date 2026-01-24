# 脚本归档说明

归档时间: 2026-01-25 04:29

## 归档内容

### 测试脚本
- `build.sh` - 构建脚本（已被 GitHub Actions 替代）
- `check_mcp_response.sh` - MCP 响应检查（调试用）
- `prepare_test.sh` - 发帖测试准备
- `quick_test.sh` - 快速测试
- `start_mcp_debug.sh` - MCP 调试启动
- `test_playwright_fix.sh` - Playwright 修复测试

### 登录相关（已合并到 login.sh）
- `start_login.sh` - 启动登录（有头浏览器）
- `local_login.sh` - 本地登录并生成同步命令
- `login_and_sync.sh` - 登录并同步
- `export_cookies.sh` - 导出 cookies base64

### Git 工具
- `cleanup-git-history.sh` - Git 历史清理

### 测试程序
- `cmd/test_navigation/` - 导航测试程序
- `cmd/config-gen/` - 配置生成工具

### 临时文件
- `login` - 旧的登录二进制文件
- `test_publish_browser.js` - 浏览器测试脚本
- `test_publish.json` - 测试数据
- `xhs-mcp` - 旧的主程序二进制文件

## 保留的核心脚本

项目根目录现在只保留两个核心脚本：

1. **`login.sh`** - 本地登录工具
   - 自动检测登录程序位置
   - 启动有头浏览器登录
   - 可选生成远程同步命令

2. **`deploy.sh`** - 远程部署脚本
   - 自动下载最新 release 版本
   - 配置 PM2 进程管理
   - 一键部署到 Linux 服务器

## 归档原因

- 项目已完成从 Rod 到 Playwright 的迁移
- CI/CD 流程已完善（GitHub Actions 自动构建）
- 核心功能稳定，不再需要大量测试脚本
- 简化项目结构，只保留必要的用户脚本
