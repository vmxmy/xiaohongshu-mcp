# 小红书 MCP 工具

[![All Contributors](https://img.shields.io/badge/all_contributors-21-orange.svg?style=flat-square)](#contributors-)
[![Docker Pulls](https://img.shields.io/docker/pulls/xpzouying/xiaohongshu-mcp?style=flat-square&logo=docker)](https://hub.docker.com/r/xpzouying/xiaohongshu-mcp)
[![GitHub Release](https://img.shields.io/github/v/release/vmxmy/xiaohongshu-mcp?style=flat-square)](https://github.com/vmxmy/xiaohongshu-mcp/releases)

通过 MCP（Model Context Protocol）协议操作小红书，支持发布内容、获取数据分析、管理草稿等功能。

> **注意**: 本项目基于浏览器自动化实现，已完成从 Rod 到 Playwright 的迁移，稳定性和性能显著提升。

---

## 📑 目录

- [功能特性](#功能特性)
- [快速开始](#快速开始)
  - [方式1: 下载 Release 版本](#方式1-下载-release-版本推荐)
  - [方式2: 本地编译](#方式2-本地编译)
  - [方式3: 远程服务器部署](#方式3-远程服务器部署)
- [核心脚本](#核心脚本)
- [MCP 工具列表](#mcp-工具列表)
- [使用示例](#使用示例)
- [配置说明](#配置说明)
- [常见问题](#常见问题)
- [开发指南](#开发指南)
- [贡献者](#contributors-)
- [致谢](#致谢)

---

## 功能特性

### 📝 内容发布
- ✅ **图文发布**: 支持标题、正文、图片（最多9张）、话题标签
- ✅ **视频发布**: 支持本地视频文件上传
- ✅ **草稿保存**: 保存图文/视频草稿，稍后发布
- ✅ **定时发布**: 支持1小时至14天内的定时发布
- ✅ **图片支持**: HTTP/HTTPS链接或本地文件路径

### 📊 数据分析
- ✅ **内容分析**: 获取笔记的曝光、观看、点赞、评论等数据（支持翻页和排序）
- ✅ **粉丝分析**: 获取粉丝概览、画像分布、活跃粉丝列表

### 🔐 账号管理
- ✅ **二维码登录**: 有头浏览器扫码登录
- ✅ **登录状态检查**: 检查当前登录状态
- ✅ **Cookies 同步**: 本地登录后可同步到远程服务器

### 🔄 互动功能
- ✅ **笔记详情**: 获取笔记内容、评论列表
- ✅ **点赞/收藏**: 对笔记进行点赞或收藏
- ✅ **发布评论**: 评论笔记或回复评论
- ✅ **关注用户**: 关注或取关指定用户
- ✅ **搜索内容**: 搜索小红书笔记并支持筛选

### 🎯 其他功能
- ✅ **分享链接**: 获取笔记分享链接
- ✅ **删除内容**: 删除自己的笔记或评论
- ✅ **用户主页**: 获取用户信息和笔记列表
- ✅ **Feed 列表**: 获取首页推荐内容

---

## 快速开始

### 方式1: 下载 Release 版本（推荐）

适合不熟悉 Go 编译的用户。

#### 1. 下载二进制文件

访问 [Releases 页面](https://github.com/vmxmy/xiaohongshu-mcp/releases)，下载对应平台的文件：

- **macOS Apple Silicon**: `xiaohongshu-mcp-darwin-arm64`
- **macOS Intel**: `xiaohongshu-mcp-darwin-amd64`
- **Windows x64**: `xiaohongshu-mcp-windows-amd64.exe`
- **Linux x64**: `xiaohongshu-mcp-linux-amd64`

#### 2. 添加执行权限（Linux/macOS）

```bash
chmod +x xiaohongshu-mcp-darwin-arm64
```

#### 3. 启动服务

```bash
# macOS/Linux
./xiaohongshu-mcp-darwin-arm64

# Windows
xiaohongshu-mcp-windows-amd64.exe
```

#### 4. 验证服务

```bash
curl http://localhost:18060/health
```

---

### 方式2: 本地编译

适合开发者或需要定制的用户。

#### 前置要求
- Go 1.24+
- Git

#### 编译步骤

```bash
# 1. 克隆仓库
git clone https://github.com/vmxmy/xiaohongshu-mcp.git
cd xiaohongshu-mcp

# 2. 编译主程序
go build -o xiaohongshu-mcp .

# 3. 编译登录工具
go build -o xiaohongshu-login ./cmd/login

# 4. 启动服务
./xiaohongshu-mcp
```

---

### 方式3: 远程服务器部署

使用一键部署脚本，适合在 Linux 服务器上运行。

```bash
# 下载并执行部署脚本
curl -fsSL https://raw.githubusercontent.com/vmxmy/xiaohongshu-mcp/main/deploy.sh | bash
```

部署脚本会自动完成：
- 安装系统依赖
- 安装 PM2 进程管理器
- 下载最新 release 版本
- 配置并启动服务

**部署后管理命令**:

```bash
# 查看状态
pm2 status

# 查看日志
pm2 logs xiaohongshu-mcp

# 重启服务
pm2 restart xiaohongshu-mcp

# 停止服务
pm2 stop xiaohongshu-mcp
```

---

## 核心脚本

项目提供两个核心脚本，简化常见操作。

### 1. login.sh - 本地登录工具

在本地运行登录程序，扫码登录小红书。

```bash
./login.sh
```

**功能**:
- 自动检测登录程序位置
- 启动有头浏览器显示二维码
- 登录成功后保存 cookies
- 可选生成远程同步命令

**使用流程**:
1. 运行脚本
2. 浏览器自动打开
3. 扫描二维码登录
4. 选择是否同步到远程服务器

### 2. deploy.sh - 远程一键部署

在 Linux 服务器上一键部署 MCP 服务。

```bash
# 方式1: 远程直接执行
curl -fsSL https://raw.githubusercontent.com/vmxmy/xiaohongshu-mcp/main/deploy.sh | bash

# 方式2: 下载后执行
wget https://raw.githubusercontent.com/vmxmy/xiaohongshu-mcp/main/deploy.sh
chmod +x deploy.sh
./deploy.sh
```

**部署配置**:
- 部署目录: `/opt/xiaohongshu-mcp`
- 默认端口: `18060`
- 运行模式: 无头浏览器
- 进程管理: PM2

---

## MCP 工具列表

通过 MCP 协议可调用以下工具。完整的 API 文档参见 [docs/API.md](docs/API.md)。

### 登录相关
| 工具名称 | 功能 | 参数 |
|---------|------|------|
| `get_login_qrcode` | 获取登录二维码 | - |
| `check_login_status` | 检查登录状态 | - |
| `sync_cookies` | 同步 cookies | `cookies_base64` 或 `cookies_json` |

### 内容发布
| 工具名称 | 功能 | 参数 |
|---------|------|------|
| `publish_content` | 发布图文内容 | `title`, `content`, `images`, `tags`, `schedule_at` |
| `publish_with_video` | 发布视频内容 | `title`, `content`, `video`, `tags`, `schedule_at` |
| `save_draft` | 保存图文草稿 | `title`, `content`, `images`, `tags` |
| `save_video_draft` | 保存视频草稿 | `title`, `content`, `video`, `tags` |

### 数据分析
| 工具名称 | 功能 | 参数 |
|---------|------|------|
| `get_content_analytics` | 获取内容分析 | `limit`, `sort_by`, `sort_order` |
| `get_fan_analytics` | 获取粉丝分析 | `period` (7d/30d) |

### 笔记操作
| 工具名称 | 功能 | 参数 |
|---------|------|------|
| `get_feed_detail` | 获取笔记详情 | `feed_id`, `xsec_token`, `load_all_comments` |
| `like_feed` | 点赞笔记 | `feed_id`, `xsec_token`, `unlike` |
| `favorite_feed` | 收藏笔记 | `feed_id`, `xsec_token`, `unfavorite` |
| `share_feed` | 分享笔记 | `feed_id`, `xsec_token` |
| `delete_feed` | 删除笔记 | `feed_id`, `xsec_token` |

### 评论操作
| 工具名称 | 功能 | 参数 |
|---------|------|------|
| `post_comment_to_feed` | 发布评论 | `feed_id`, `xsec_token`, `content` |
| `reply_comment_in_feed` | 回复评论 | `feed_id`, `comment_id`, `user_id`, `xsec_token`, `content` |
| `like_comment` | 点赞评论 | `feed_id`, `comment_id`, `user_id`, `xsec_token`, `unlike` |
| `delete_comment` | 删除评论 | `feed_id`, `comment_id`, `user_id`, `xsec_token` |

### 用户操作
| 工具名称 | 功能 | 参数 |
|---------|------|------|
| `follow_user` | 关注用户 | `user_id`, `xsec_token`, `unfollow` |
| `user_profile` | 获取用户主页 | `user_id`, `xsec_token` |

### 搜索和浏览
| 工具名称 | 功能 | 参数 |
|---------|------|------|
| `search_feeds` | 搜索笔记 | `keyword`, `filters` |
| `list_feeds` | 获取首页 Feed | - |

### Cookies 管理
| 工具名称 | 功能 | 参数 |
|---------|------|------|
| `delete_cookies` | 删除 cookies | - |

---

## 使用示例

### 示例1: 发布图文内容

通过 MCP Inspector 或 Claude Desktop 调用 `publish_content` 工具：

```json
{
  "title": "我的第一篇笔记",
  "content": "这是内容描述",
  "images": [
    "/Users/username/Pictures/image1.jpg",
    "https://example.com/image2.jpg"
  ],
  "tags": ["生活", "分享"]
}
```

### 示例2: 获取内容分析数据（带排序）

调用 `get_content_analytics` 工具：

```json
{
  "limit": 20,
  "sort_by": "views",
  "sort_order": "desc"
}
```

可排序字段: `exposure`（曝光）, `views`（观看）, `click_rate`（点击率）, `likes`（点赞）, `comments`（评论）, `collects`（收藏）, `follower_growth`（涨粉）, `shares`（分享）, `duration`（观看时长）, `barrage`（弹幕）

### 示例3: 本地登录后同步到服务器

```bash
# 1. 本地运行登录脚本
./login.sh

# 2. 选择 Y 生成同步命令

# 3. 复制显示的 curl 命令到远程服务器执行
curl -X POST http://localhost:18060/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"sync_cookies","arguments":{"cookies_base64":"..."}}}'
```

---

## 配置说明

### 命令行参数

```bash
./xiaohongshu-mcp [选项]
```

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `-headless` | 是否使用无头模式 | `true` |
| `-port` | 监听端口 | `:18060` |
| `-bin` | 浏览器二进制文件路径 | 自动检测 |
| `-config` | 配置文件路径 | 自动查找 config.yaml |

### 配置文件

项目支持 YAML 配置文件（`config.yaml`），可自定义：
- 小红书创作中心 URL
- 页面元素选择器
- 超时时间和等待间隔
- API 拦截规则

详见项目根目录的 `config.yaml` 文件。

### 环境变量

可通过环境变量覆盖部分配置：

```bash
export LOG_LEVEL=debug
export PORT=8080
./xiaohongshu-mcp
```

---

## 常见问题

### Q: 首次运行提示下载浏览器？

A: Playwright 首次运行会自动下载 Chromium 浏览器（约150MB），请确保网络连接正常。后续运行无需重复下载。

### Q: 如何在 Claude Desktop 中使用？

A: 编辑 Claude Desktop 配置文件：

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "xiaohongshu": {
      "command": "/path/to/xiaohongshu-mcp",
      "args": []
    }
  }
}
```

然后重启 Claude Desktop。

### Q: 登录后 cookies 保存在哪里？

A: 默认保存在 `~/.xiaohongshu/cookies.json`。

### Q: 为什么图片上传失败？

A:
1. 检查图片路径是否正确（建议使用绝对路径）
2. 确保图片格式支持（JPG, PNG, WEBP）
3. 图片大小不超过 20MB
4. 最多上传 9 张图片

### Q: 如何查看 MCP 工具列表？

A:
1. 使用 MCP Inspector: `npx @modelcontextprotocol/inspector`
2. 连接到 `http://127.0.0.1:18060/mcp`
3. 在 Inspector 中查看所有可用工具

### Q: 排序功能如何使用？

A: 调用 `get_content_analytics` 时传入 `sort_by` 和 `sort_order` 参数。排序在前端实现，不会触发新的 API 请求。

### Q: 如何更新到最新版本？

A:
- **Release 用户**: 重新下载最新版本替换旧文件
- **编译用户**: `git pull && go build`
- **服务器部署**: 修改 `deploy.sh` 中的 `VERSION` 变量后重新运行

---

## 开发指南

### 项目结构

```
xiaohongshu-mcp/
├── cmd/
│   └── login/           # 登录工具源码
├── docs/                # API 和开发文档
├── examples/            # 集成示例
├── internal/
│   └── infra/
│       └── browser/     # 浏览器引擎（Playwright）
├── xiaohongshu/         # 核心业务逻辑
├── login.sh             # 本地登录脚本
├── deploy.sh            # 远程部署脚本
├── config.yaml          # 配置文件
└── main.go              # 主程序入口
```

### 技术栈

- **语言**: Go 1.24+
- **浏览器引擎**: Playwright (已从 Rod 迁移)
- **协议**: MCP (Model Context Protocol)
- **进程管理**: PM2 (生产环境)

### 本地开发

```bash
# 1. 克隆仓库
git clone https://github.com/vmxmy/xiaohongshu-mcp.git
cd xiaohongshu-mcp

# 2. 安装依赖
go mod download

# 3. 运行测试
go test ./...

# 4. 启动开发服务（有头模式）
go run . -headless=false
```

### 代码格式化

```bash
# 格式化所有 Go 文件
gofmt -w .
```

### 构建 Release

```bash
# 本地构建
go build -o xiaohongshu-mcp .

# 交叉编译（示例：Linux x64）
GOOS=linux GOARCH=amd64 go build -o xiaohongshu-mcp-linux-amd64 .
```

---

## 集成示例

项目提供多个平台的集成示例：

- [Claude Code 集成](examples/claude-code/claude-code-kimi-k2.md)
- [n8n 集成](examples/n8n/README.md)
- [AnythingLLM 集成](examples/anythingLLM/readme.md)
- [CherryStudio 集成](examples/cherrystudio/README.md)

更多示例详见 [examples/](examples/) 目录。

---

## 更新日志

### v2026.01.25 (最新)
- ✨ 新增内容分析排序功能（支持10个字段）
- ✨ 新增内容分析翻页功能（最多50页）
- ♻️ 完成从 Rod 到 Playwright 的浏览器引擎迁移
- 🐛 修复循环排序图标逻辑
- 📝 简化项目结构，清理历史文档
- 🚀 优化 GitHub Actions 自动构建

### v2026.01.24
- 🐛 修复 Rod Engine 的 Fill 方法支持 contenteditable 元素
- 📝 添加 PM2 部署配置和快速开始指南

完整更新日志: [RELEASE_NOTES.md](RELEASE_NOTES.md)

---

## Contributors ✨

感谢所有为本项目做出贡献的开发者！

<!-- ALL-CONTRIBUTORS-LIST:START - Do not remove or modify this section -->
<!-- 贡献者列表自动生成 -->
<!-- ALL-CONTRIBUTORS-LIST:END -->

---

## 致谢

本项目 Fork 自 [xpzouying/xiaohongshu-mcp](https://github.com/xpzouying/xiaohongshu-mcp)，感谢原作者的开创性工作。

**主要改进**:
- ✅ 完成浏览器引擎从 Rod 迁移到 Playwright
- ✅ 新增内容分析排序和翻页功能
- ✅ 优化项目结构和文档
- ✅ 改进部署脚本和自动化流程
- ✅ 清理历史代码和测试文件

如有问题或建议，欢迎提 [Issue](https://github.com/vmxmy/xiaohongshu-mcp/issues) 或 [Pull Request](https://github.com/vmxmy/xiaohongshu-mcp/pulls)。

---

## 许可证

MIT License

---

## 相关链接

- **原项目**: https://github.com/xpzouying/xiaohongshu-mcp
- **MCP 协议**: https://modelcontextprotocol.io
- **Playwright**: https://playwright.dev
- **Claude Desktop**: https://claude.ai/download
