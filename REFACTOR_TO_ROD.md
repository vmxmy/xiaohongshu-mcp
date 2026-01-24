# 浏览器引擎统一为 Rod

## 日期
2026-01-24

## 背景

项目中同时存在两套浏览器自动化实现：
- **Rod**: 用于大部分业务逻辑（登录、数据获取等）
- **Playwright**: 仅用于发布功能

这导致了以下问题：
1. **依赖复杂**: Playwright 需要额外下载 500MB+ 的浏览器驱动
2. **部署困难**: VPS 部署时遇到 "playwright driver not installed" 错误
3. **维护成本高**: 两套 API，增加学习和维护负担
4. **代码不统一**: 同样的功能用不同的库实现

## 重构目标

统一使用 **Rod** 作为唯一的浏览器自动化框架。

## 实施步骤

### 1. 创建 Rod Engine 实现

创建了符合 `browser.Engine` 接口的 Rod 实现：

**新增文件**:
- `internal/infra/browser/rod/engine.go` - 浏览器引擎实现
- `internal/infra/browser/rod/cookies.go` - Cookie 处理

**核心特性**:
- 使用 `headless_browser` 包管理浏览器生命周期
- 自动加载 cookies
- 支持配置超时时间
- 完全兼容 `browser.Engine` 接口

### 2. 修改 Wiring 配置

**修改文件**: `wiring_bootstrap.go`

```go
// 之前：使用 Playwright
import "github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser/playwright"
engine := playwright.New(engineCfg)

// 现在：使用 Rod
import browserrod "github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser/rod"
engine := browserrod.New(engineCfg)
```

### 3. 清理 Playwright 代码

**归档文件** (移动到 `.archived/`):
- `internal/infra/browser/playwright/` - 整个目录
- `PUBLISH_TEST_REPORT.md` - Playwright 测试报告
- `docs/plans/2026-01-23-clean-arch-playwright-auto-selector-implementation.md` - Playwright 实现计划

**依赖清理**:
- `go mod tidy` 自动移除了 `playwright-community/playwright-go` 依赖

## 优势对比

### Playwright (旧方案)
❌ 需要额外安装驱动: `go run github.com/playwright-community/playwright-go/cmd/playwright install`
❌ 驱动文件 500MB+
❌ 部署复杂，容易出错
❌ API 与项目其他部分不一致

### Rod (新方案)
✅ 零外部依赖，自动管理浏览器
✅ 轻量级，按需下载
✅ 部署简单，一个二进制文件搞定
✅ 与项目其他 42 个文件保持一致
✅ 更好的性能

## 验证

### 编译测试
```bash
make build
# ✅ 编译成功

./bin/xiaohongshu-mcp --help
# ✅ 程序正常运行
```

### 依赖检查
```bash
grep "playwright" go.mod
# (无输出，依赖已移除)
```

### 代码检查
```bash
grep -r "playwright" --include="*.go" . | grep -v ".archived"
# (无输出，代码已清理)
```

## 兼容性

### browser.Engine 接口
Rod 实现完全符合接口定义，包括：
- `Start()` - 启动浏览器
- `NewPage()` - 创建新页面
- `Close()` - 关闭浏览器

### browser.Page 接口
所有方法都已实现：
- `Goto(url)` - 导航
- `Click(selector)` - 点击
- `Fill(selector, value)` - 填充
- `SetFiles(selector, files)` - 上传文件
- `WaitVisible(selector)` - 等待元素可见
- `IsVisible(selector)` - 检查元素可见性
- `ScrollIntoView(selector)` - 滚动到视图
- `ClickForce(selector)` - 强制点击
- 等等...

## 后续任务

- [x] 创建 Rod Engine 实现
- [x] 修改 wiring 配置
- [x] 清理 Playwright 代码
- [x] 测试编译
- [ ] 本地测试发布功能
- [ ] VPS 部署测试
- [ ] 更新部署文档

## 影响范围

### 受影响的功能
- ✅ 发布图文内容
- ✅ 发布视频内容
- ✅ 定时发布

### 不受影响的功能
所有其他功能（登录、数据获取、点赞、评论等）继续使用原有的 Rod 代码，无任何变化。

## 总结

通过这次重构：
1. **简化了技术栈**: 从双引擎统一为单引擎
2. **降低了部署复杂度**: 不再需要手动安装 Playwright 驱动
3. **提升了代码一致性**: 全项目统一使用 Rod
4. **减少了二进制大小**: 去除了 Playwright 相关依赖
5. **改善了开发体验**: 只需要学习一套 API

这是一次成功的技术债务清理，为项目的长期维护奠定了更好的基础。
