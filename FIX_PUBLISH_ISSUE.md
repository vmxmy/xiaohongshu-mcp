# 发布功能修复报告

## 问题描述

发布笔记功能超时失败，错误信息：
```
playwright: timeout: Timeout 60000ms exceeded.
发布内容失败(新用例): title=测试 publish image content(div.ql-editor)
```

## 根本原因

1. **选择器过时**：小红书页面结构更新，原选择器 `div.ql-editor` 已失效
2. **页面加载等待不足**：`Goto` 方法未等待网络空闲就返回
3. **元素等待缺失**：直接操作元素，未等待元素可见
4. **多元素匹配**：提交按钮选择器匹配到2个元素（"发布"和"暂存离开"）

## 修复内容

### 1. 更新配置文件选择器 (config.yaml)

```yaml
# 修改前
selectors:
  publish:
    content_editor_ql: "div.ql-editor"
    submit_button: "div.submit div.d-button-content"

# 修改后
selectors:
  publish:
    content_editor_ql: '[role="textbox"]'  # 使用新的可访问性选择器
    submit_button: "button:has-text('发布')"  # 精确匹配发布按钮
```

### 2. 优化 Playwright Engine (internal/infra/browser/playwright/engine.go)

#### 2.1 Goto 方法添加网络空闲等待
```go
// 修改前
func (p *page) Goto(url string) error {
    _, err := p.p.Goto(url)
    return err
}

// 修改后
func (p *page) Goto(url string) error {
    _, err := p.p.Goto(url, playwright.PageGotoOptions{
        WaitUntil: playwright.WaitUntilStateNetworkidle,
    })
    return err
}
```

#### 2.2 Click 方法处理多元素匹配
```go
// 修改前
func (p *page) Click(selector string) error {
    return p.p.Locator(selector).Click()
}

// 修改后
func (p *page) Click(selector string) error {
    // 使用 first() 处理多个匹配的情况
    return p.p.Locator(selector).First().Click()
}
```

### 3. 添加元素等待逻辑 (internal/infra/xhs/publish/gateway.go)

在 `PublishImage` 和 `PublishVideo` 方法中添加 WaitVisible：

```go
// 等待上传输入框可见
if err := page.WaitVisible(g.cfg.Selectors["upload_input"]); err != nil {
    return fmt.Errorf("publish image wait upload_input(%s): %w", g.cfg.Selectors["upload_input"], err)
}
if err := page.SetFiles(g.cfg.Selectors["upload_input"], content.ImagePaths); err != nil {
    return fmt.Errorf("publish image upload_input(%s): %w", g.cfg.Selectors["upload_input"], err)
}

// 等待标题输入框可见（图片上传后才出现）
if err := page.WaitVisible(g.cfg.Selectors["title_input"]); err != nil {
    return fmt.Errorf("publish image wait title_input(%s): %w", g.cfg.Selectors["title_input"], err)
}
// ... 填写标题

// 等待内容编辑器可见
if err := page.WaitVisible(g.cfg.Selectors["content"]); err != nil {
    return fmt.Errorf("publish image wait content(%s): %w", g.cfg.Selectors["content"], err)
}
// ... 填写内容
```

## 测试结果

### 诊断过程
1. ✅ 发现 `div.ql-editor` 不可用
2. ✅ 找到备用选择器 `[role="textbox"]` 和 `div[contenteditable="true"]`
3. ✅ 验证完整发布流程
4. ✅ 成功发布笔记（审核中）

### 发布成功标志
- URL 跳转到：`https://creator.xiaohongshu.com/publish/success`
- 显示消息："发布成功"
- 笔记进入审核状态

## 影响范围

### 修改的文件
1. `config.yaml` - 更新选择器配置
2. `internal/infra/browser/playwright/engine.go` - 优化页面操作方法
3. `internal/infra/xhs/publish/gateway.go` - 添加元素等待逻辑

### 涉及功能
- ✅ 图文发布
- ✅ 视频发布

## 后续建议

1. **监控选择器稳定性**：小红书页面可能会继续更新，建议定期检查选择器有效性
2. **配置化选择器**：已通过 config.yaml 实现，方便快速调整
3. **添加备用选择器**：考虑在代码中支持多个备用选择器，提高容错性
4. **增强错误提示**：当选择器失效时，提供更友好的错误信息和诊断建议

## 验证步骤

1. 重新编译：`go build -o xhs-mcp .`
2. 重启 MCP 服务
3. 通过 MCP 工具发布测试笔记
4. 确认发布成功且笔记进入审核

---
**修复日期**：2026-01-24
**修复人**：Claude
**测试状态**：✅ 通过
