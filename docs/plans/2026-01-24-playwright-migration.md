# Playwright Migration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将小红书 MCP 的浏览器层彻底迁移到 Playwright，实现统一的 `browser.Page` 抽象并更新所有业务动作与配置。

**Architecture:** 通过扩展 `internal/infra/browser` 抽象来覆盖既有 Rod 能力，在 `internal/infra/browser/playwright` 中提供完整实现，并让 `service.go`、登录流程与 `xiaohongshu/` 目录全部基于该接口；配置文件 `config.yaml` 更新为 Playwright 选择器语法。

**Tech Stack:** Go 1.24、Playwright-Go、现有 MCP 服务架构、YAML 配置。

### Task 1: 勘察 Rod 用法并扩展浏览器抽象

**Files:**
- Modify: `internal/infra/browser/engine.go`
- Modify: `docs/`（若需记录接口变更，可补充在 `CONFIG_MIGRATION_GUIDE.md`）

**Step 1: 列出 Rod API 用法**

```bash
rg -n "rod" -g'*.go' xiaohongshu selector login_session_rod.go > /tmp/rod-usages.txt
```

确认涉及导航、DOM 查询、鼠标键盘、事件监听、网络拦截、截图等场景。

**Step 2: 设计接口**
- 在 `internal/infra/browser/engine.go` 内补充 `Page` 方法：`WaitLoad`, `Eval`, `Hover`, `Press`, `Type`, `SelectOption`, `Focus`, `WaitForSelector(selector string, timeout time.Duration)`, `Elements(selector string) ([]Element, error)`, `Element(selector string) (Element, error)`, `MouseMove(x,y float64) error`, `MouseClick(Button) error`, `KeyboardType(text string) error`, `Screenshot(path string) error`, `Context(ctx context.Context) Page`（如需要上下文绑定）。
- 定义 `Element`/`Keyboard`/`Mouse` 子接口涵盖 Rod 常用方法（点击、输入、属性、滚动、等待可见、文本读取、Focus、Remove 等）。

**Step 3: 文档同步**
- 在 `CONFIG_MIGRATION_GUIDE.md` 或单独小节记录接口新增说明，确保团队了解 Playwright 适配层职责。

### Task 2: Playwright 引擎与元素实现

**Files:**
- Modify: `internal/infra/browser/playwright/engine.go`
- Add: `internal/infra/browser/playwright/page.go`, `element.go`, `mouse.go`, `keyboard.go`（如需拆分）
- Modify/Create: `internal/infra/browser/playwright/cookies.go`

**Step 1: Page 实现**
- 在新文件中实现 `type page struct { p playwright.Page; ctx playwright.BrowserContext }` 对应的所有接口方法。
- 使用 Playwright `Locator` 与 `FrameLocator` 实现 `Element`/`Elements`，必要时封装 `type element struct { locator playwright.Locator }`。

**Step 2: 输入控制**
- 给出鼠标/键盘封装，映射 Rod 的 `Mouse.MoveTo/Click`、`Keyboard.Type/Press`。
- 处理 `Press` 组合键（分解为 Playwright `Keyboard.Press`）。

**Step 3: Eval/Wait**
- `Eval` 使用 `Page.Evaluate` 并返回 `any`；`WaitLoad` 使用 `WaitForLoadState(NetworkIdle)`。
- `WaitForSelector` 支持自定义 timeout，内部调用 `Locator.WaitFor`。

**Step 4: Cookies/超时**
- 维持 `applyTimeouts`，并支持 action/navigation timeout；确保 `loadCookies` 兼容 Rod/Playwright 格式。

**Step 5: gofmt / tests**

```bash
gofmt -w internal/infra/browser/playwright
```

### Task 3: 浏览器工厂与全局依赖替换

**Files:**
- Delete/Replace: `browser/browser.go`
- Modify: `service.go`
- Modify: `cmd/login/main.go`
- Modify: `login_session_rod.go`（改为 Playwright 版，如 `login_session_pw.go`）
- Modify: `internal/interfaces/wiring/wiring.go`
- Modify: `configs/browser.go`（若需要新的配置参数）

**Step 1: 新建工厂**
- 创建 `internal/infra/browser/factory.go`（或在 `playwright` 包中导出 `func NewEngineFromConfig(cfg playwright.Config) browser.Engine`）。
- 读取 `configs.IsHeadless()`、cookies 路径，构造 `playwright.Config`。

**Step 2: 替换 `newBrowser()`**
- 在 `service.go` 中创建 `func newBrowserEngine() browser.Engine` 返回 Playwright。
- 更新所有 `page := b.NewPage()` 调用为：

```go
engine := newBrowserEngine()
if err := engine.Start(); err != nil { ... }
page, err := engine.NewPage()
```

或封装 `withBrowserPage(ctx, func(page browser.Page) error)` 来负责 Start/NewPage/Close。

**Step 3: 登录 CLI**
- `cmd/login/main.go` 改为使用新的 `withBrowserPage`，并移除 Rod 导入。

**Step 4: 登录 Session**
- 将 `login_session_rod.go` 重写为 Playwright 版本：实现 `qrPage` 接口使用 `browser.Page`/`Element` 方法；替换 `rodPageAdapter`。

**Step 5: Wire publish usecase**
- `internal/interfaces/wiring/wiring.go` 使用新的 engine 构造函数，并确保传入 selectors。

### Task 4: 业务动作迁移（一）——登录 & 发布

**Files:**
- Modify: `xiaohongshu/login.go`
- Modify: `xiaohongshu/publish.go`
- Modify: `xiaohongshu/publish_video.go`
- Modify: `xiaohongshu/publish_ui_check.go`
- Modify: `xiaohongshu/publish_test.go`

**Step 1: 类型替换**
- 将结构体中的 `*rod.Page` 字段改为 `browser.Page`。
- 替换 `rod` import 为 `internal/infra/browser`。

**Step 2: Rod API 映射**
- 使用接口方法实现导航/等待/输入，例如：

```go
if err := page.Goto(url); err != nil { ... }
if err := page.WaitForSelector(selectors.Publish.UploadInput, timeout); err != nil { ... }
```

- 遇到 `MustEval`/`Eval` 转换为 `page.Eval`。
- 将 `page.Mouse.MustMoveTo...` 重写为 `page.Mouse().MoveTo(x,y)`。

**Step 3: 多元素处理**
- 将 `page.Elements(selector)` 返回的 `[]browser.Element` 用于 `getTabElement` 等函数；`Element` 提供 `Click`, `Attribute`, `IsVisible`, `ScrollIntoView`, `InnerText`, `SetFiles` 等方法。

**Step 4: Tag 输入/内容编辑器**
- 需要 `Element.Type`, `Focus`, `Press`；根据 Playwright API 适配。

**Step 5: Tests**
- 更新 `publish_test.go` 依赖的 mocks（可实现 `browser.Page` fake）。

### Task 5: 业务动作迁移（二）——数据、互动、搜索等

**Files:**
- Modify: `xiaohongshu/data.go`
- Modify: `xiaohongshu/search.go`
- Modify: `xiaohongshu/feeds.go`
- Modify: `xiaohongshu/feed_detail.go`
- Modify: `xiaohongshu/comment_feed.go`
- Modify: `xiaohongshu/comment_like.go`
- Modify: `xiaohongshu/follow.go`
- Modify: `xiaohongshu/like_favorite.go`
- Modify: `xiaohongshu/share.go`
- Modify: `xiaohongshu/delete.go`
- Modify: `xiaohongshu/navigate.go`
- Modify: `xiaohongshu/user_profile.go`

**Step 1: 结构体字段替换**
- 与 Task4 相同，全面改为 `browser.Page`。

**Step 2: 自定义行为迁移**
- Rod `WaitLoad`, `WaitDOMStable`, `Eval`, `Scroll`, `HasR`, `ElementR`, `InputTime`, `Hover`, `MouseWheel`, `Keyboard.Press` 等都基于接口方法重写。
- 若缺少 API（如正则匹配），通过 `Eval` + JS 查找元素。

**Step 3: 交互工具**
- 重写 `humanScroll`, `clickElementWithHumanBehavior`, `smartScroll` 等函数，使用 `page.ScrollIntoView`、`page.MouseMove`、`page.MouseClick`。

**Step 4: selector/smart_selector`**
- 更新 `SmartSelector` 依赖的 `page`/`element` 类型，支持 Playwright 查询；增加必要的 `Element.Text`, `Element.Remove`、`BoundingBox` 方法。

**Step 5: Tests**
- 调整 `feeds_test.go`, `search_test.go`, 其他 `_test.go` 使用新的接口 mocks。

### Task 6: 配置与选择器 Playwright 化

**Files:**
- Modify: `config.yaml`
- Modify: `README_CONFIG.md`, `CONFIG_MIGRATION_GUIDE.md`（若需）
- Modify: `config/config.go` (若 selector 字段新增)

**Step 1: 审核并更新选择器**
- 统一使用 Playwright 语法（`:has-text()`, `nth`, `>>` 等）。
- 校准 `upload_input`, `submit_button`, `save_draft_button` 等，保证 `internal/infra/xhs/publish` 与 `xiaohongshu` 逻辑一致。

**Step 2: 配置校验**

```bash
yq . config.yaml > /dev/null
```

**Step 3: 文档同步**
- 更新 README/指南，说明 Playwright 选择器策略及 `:has-text()` 用法。

### Task 7: 清理 Rod 依赖与回归测试

**Files:**
- Modify: `go.mod`, `go.sum`
- Delete: `login_session_rod.go`, `login_session_rod_test.go`（或替换为新版本）
- Global: 移除 `github.com/go-rod/rod`, `github.com/xpzouying/headless_browser`

**Step 1: go.mod 清理**

```bash
go mod tidy
```

确保 Rod/headless_browser 被移除。

**Step 2: 全量格式化**

```bash
gofmt -w ./xiaohongshu ./internal ./login_session*.go ./service.go ./cmd/login
```

**Step 3: 测试与验证**

```bash
go test ./...
go run cmd/login/main.go -bin=<chromium?>  # 如需人工验证
```

记录潜在的 Playwright 环境依赖（需先 `playwright install`）。

**Step 4: Git 检查**

```bash
git status -sb
```

确认仅包含计划内文件。
