<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

# Repository Guidelines

## 项目结构与模块组织
- `main.go`、`routes.go`、`service.go`：MCP 服务入口与路由编排。
- `cmd/login/`：登录工具入口；`bin/` 为 `build.sh` 的产物。
- `xiaohongshu/`：核心业务动作（发布/搜索/Feeds），测试也在此目录。
- `browser/`：Rod 浏览器封装；`configs/`：浏览器与选择器配置（含 `configs/selectors.yaml`）。
- `docs/`：API 文档与指南；`docker/`：Docker 部署；`assets/`：README 图片资源。
- `pkg/`、`errors/`、`selector/`：通用能力与选择器相关逻辑；`examples/` 提供示例用法。

## 架构概览
- 运行时通过 Gin Web 框架承载 HTTP 服务，MCP SDK 负责协议适配。
- 浏览器自动化基于 Rod，登录与发布依赖本地 cookies 与选择器配置。
- `mcp_server.go` 与 `mcp_handlers.go` 负责 MCP 入口与工具注册。

## 构建、测试与开发命令
- `go run cmd/login/main.go`：运行登录工具，生成可用 cookies。
- `go run .` / `go run . -headless=false`：启动 MCP 服务（无头/有界面）。
- `./build.sh v1.2.3`：构建二进制到 `bin/`，默认版本可省略。
- `docker compose up -d`（在 `docker/` 内）：本地一键运行。
- `docker compose logs -f`：查看容器日志，排查启动问题。
- `go test ./...`：运行全部 Go 测试。
- `npx @modelcontextprotocol/inspector`：验证 MCP 连接与工具列表。

## 编码风格与命名约定
- Go 1.24 项目，改动后统一执行 `gofmt -w`。
- 缩进使用 Go 默认的制表符；文件名多为 `snake_case.go`。
- 导出符号用 `PascalCase`，包内使用 `camelCase`。
- 日志/错误信息语言保持与当前文件一致，避免混用。
- 优先保持函数单一职责，重复逻辑优先下沉到 `pkg/` 或 `xiaohongshu/` 内部辅助函数。

## 测试指南
- 测试位于 `xiaohongshu/*_test.go`，基于 `stretchr/testify`。
- 部分测试默认 `t.Skip`，需要登录 cookies、浏览器与真实素材；取消跳过前请确认环境。
- 新增测试使用 `TestXxx` 命名并保持隔离性。
- 定向执行可用 `go test ./xiaohongshu -run TestPublish` 等方式，建议 配合 `-v`。

## 提交与 PR 规范
- 参考历史提交：`feat:`、`fix:`、`docs:`、`publish:` 前缀 + 简短说明。
- PR 描述需包含变更摘要、测试结果、关联 Issue/PR；涉及文档/UI 时附截图。
- 禁止提交 `cookies.json`、`creator-cookies.json` 或构建产物。
- 若涉及选择器更新，说明验证方式（如 Inspector 截图或日志片段）。

## 安全与配置提示
- Cookies 通常保存在 `~/.xiaohongshu/cookies.json` 或 Docker 挂载的 `./data`；视为敏感信息。
- 修改 `configs/selectors.yaml` 后建议使用 `go run .` + Inspector 进行验证。
- `cookies/` 目录负责读写与加载逻辑，避免在调试时打印完整 cookies 内容。
- 变更登录或发布流程时，确认 `docs/API.md` 是否需要同步更新。
