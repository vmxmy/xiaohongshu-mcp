# 脚本使用指南

项目现在只保留两个核心用户脚本，简洁高效。

## 📦 核心脚本

### 1. login.sh - 本地登录工具

**用途**: 在本地运行登录工具，扫码登录小红书并保存 cookies

**使用方法**:

```bash
# 直接运行
./login.sh

# 或者
bash login.sh
```

**功能**:
- ✅ 自动检测登录程序位置（支持 `./xiaohongshu-login`, `./bin/xiaohongshu-login`, `./cmd/login/login`）
- ✅ 启动有头浏览器，显示二维码登录界面
- ✅ 登录成功后自动保存 cookies
- ✅ 可选生成远程同步命令（直接复制粘贴到服务器）

**交互流程**:
```
1. 运行脚本
2. 等待浏览器打开
3. 扫描二维码登录
4. 登录成功
5. 询问是否需要同步到远程
   - 选择 Y: 显示 curl 命令供复制
   - 选择 N: 直接完成
```

**前置条件**:
- 本地已编译登录工具: `go build -o xiaohongshu-login ./cmd/login`

---

### 2. deploy.sh - 远程一键部署

**用途**: 在 Linux 服务器上一键部署小红书 MCP 服务

**使用方法**:

```bash
# 方式1: 远程直接执行（推荐）
curl -fsSL https://raw.githubusercontent.com/vmxmy/xiaohongshu-mcp/main/deploy.sh | bash

# 方式2: 下载后执行
wget https://raw.githubusercontent.com/vmxmy/xiaohongshu-mcp/main/deploy.sh
chmod +x deploy.sh
./deploy.sh
```

**功能**:
- ✅ 自动安装系统依赖（curl, wget）
- ✅ 安装 PM2 进程管理器
- ✅ 下载最新 release 版本的 Linux 二进制文件
- ✅ 创建 PM2 配置并启动服务
- ✅ 配置开机自启

**部署配置**:
- 部署目录: `/opt/xiaohongshu-mcp`
- 默认端口: `18060`
- 运行模式: 无头浏览器（headless）
- 进程管理: PM2

**部署后管理**:
```bash
# 查看服务状态
pm2 status

# 查看日志
pm2 logs xiaohongshu-mcp

# 重启服务
pm2 restart xiaohongshu-mcp

# 停止服务
pm2 stop xiaohongshu-mcp

# 测试 API
curl http://localhost:18060/health
```

**前置条件**:
- Linux 服务器（x64 或 ARM64）
- 已安装 Node.js 和 npm（用于 PM2）

---

## 🔄 完整工作流

### 场景1: 本地开发测试

```bash
# 1. 编译登录工具
go build -o xiaohongshu-login ./cmd/login

# 2. 运行登录
./login.sh

# 3. 编译并启动 MCP 服务
go build -o xiaohongshu-mcp .
./xiaohongshu-mcp
```

### 场景2: 远程服务器部署

```bash
# 在远程服务器执行
curl -fsSL https://raw.githubusercontent.com/vmxmy/xiaohongshu-mcp/main/deploy.sh | bash

# 在本地获取登录 cookies
./login.sh
# 选择 Y 生成同步命令

# 将生成的 curl 命令复制到远程服务器执行
curl -X POST http://localhost:18060/mcp ...
```

---

## 📂 项目结构

```
xiaohongshu-mcp/
├── login.sh              # 本地登录脚本
├── deploy.sh             # 远程部署脚本
├── cmd/
│   └── login/            # 登录工具源码
├── .archived/            # 归档的测试脚本和临时文件
└── ...                   # 其他源码文件
```

---

## 🗑️ 已归档内容

所有测试脚本、过渡脚本和临时文件已移至 `.archived/scripts_20260125_042923/`

归档内容包括:
- 构建脚本（已被 GitHub Actions 替代）
- 测试脚本（功能已稳定）
- 旧的登录脚本（已合并到 login.sh）
- 配置生成工具（已不再需要）
- 临时测试程序

详见: `.archived/scripts_20260125_042923/README.md`

---

## 💡 常见问题

**Q: 找不到登录程序怎么办？**
```bash
go build -o xiaohongshu-login ./cmd/login
```

**Q: 如何更新 deploy.sh 中的版本？**

编辑 `deploy.sh` 第8行的 `VERSION` 变量即可。

**Q: 远程服务器需要什么环境？**

- Node.js + npm（用于 PM2）
- 足够的磁盘空间（Playwright 首次运行会下载约 150MB 浏览器）

**Q: 如何查看归档的脚本？**

所有归档内容保存在 `.archived/` 目录，可随时查看或恢复。
