# 配置文件化改造指南

## 概述

将所有硬编码的URL、CSS选择器、超时参数等提取到配置文件 `config.yaml` 中，使代码更易维护。当小红书页面更新时，只需修改配置文件，无需改动代码。

### Playwright 迁移注意

在 Playwright 迁移过程中，我们统一扩展了 `internal/infra/browser/engine.go` 中的 `browser.Page` 抽象，新增了导航等待、Eval、鼠标键盘、元素查询等方法，便于业务层不依赖具体引擎。后续若需要新增交互能力，请优先在该接口中补充，再由 Playwright 实现具体逻辑。

## 改造步骤

### 1. 已完成的工作

✅ 创建配置文件: `config.yaml`
✅ 创建配置加载器: `config/config.go`
✅ 修改 main.go 加载配置
✅ 创建示例文件: `xiaohongshu/publish_with_config.go`

### 2. 改造模式

#### 模式 1: URL 替换

**改造前:**
```go
const urlOfPublic = `https://creator.xiaohongshu.com/publish/publish?source=official`

pp.MustNavigate(urlOfPublic)
```

**改造后:**
```go
// 获取配置
cfg := config.Get()
publishURL := cfg.URLs.Creator.PublishImage  // 从配置读取

// 或使用辅助函数（带默认值）
publishURL := getPublishURL()

pp.MustNavigate(publishURL)
```

#### 模式 2: CSS选择器替换

**改造前:**
```go
uploadInput := page.MustElement(".upload-input")
titleElem := page.MustElement("div.d-input input")
```

**改造后:**
```go
cfg := config.Get()

// 直接使用配置
uploadInput := page.MustElement(cfg.Selectors.Publish.UploadInput)
titleElem := page.MustElement(cfg.Selectors.Publish.TitleInput)

// 或使用辅助函数（带默认值后备）
uploadSelector := getSelector(func(s *config.PublishSelectors) string {
    return s.UploadInput
}, ".upload-input")  // 后备默认值

uploadInput := page.MustElement(uploadSelector)
```

#### 模式 3: 超时时间替换

**改造前:**
```go
pp := page.Timeout(300 * time.Second)
deadline := time.Now().Add(15 * time.Second)
```

**改造后:**
```go
cfg := config.Get()

// 使用配置中的超时
timeout := cfg.Timeouts.GetNavigate()  // 300秒
pp := page.Timeout(timeout)

// TAB查找超时
searchTimeout := cfg.Timeouts.GetTabSearch()  // 15秒
deadline := time.Now().Add(searchTimeout)
```

#### 模式 4: 间隔时间替换

**改造前:**
```go
time.Sleep(200 * time.Millisecond)
time.Sleep(500 * time.Millisecond)
```

**改造后:**
```go
cfg := config.Get()

time.Sleep(cfg.Intervals.GetTabRetry())      // 200ms
time.Sleep(cfg.Intervals.GetTagSelectWait()) // 500ms
```

#### 模式 5: 限制值替换

**改造前:**
```go
if len(tags) >= 10 {
    tags = tags[:10]
}

if titleWidth > 40 {
    return errors.New("标题超长")
}
```

**改造后:**
```go
cfg := config.Get()

maxTags := cfg.Limits.MaxTags  // 10
if len(tags) >= maxTags {
    tags = tags[:maxTags]
}

maxWidth := cfg.Limits.MaxTitleWidth  // 40
if titleWidth > maxWidth {
    return errors.New("标题超长")
}
```

### 3. 需要改造的文件清单

#### 高优先级（核心功能）

- [x] `xiaohongshu/publish.go` - 发布功能（已创建示例）
- [ ] `xiaohongshu/publish_video.go` - 视频发布
- [ ] `xiaohongshu/login.go` - 登录功能
- [ ] `xiaohongshu/feeds.go` - Feed列表
- [ ] `xiaohongshu/feed_detail.go` - Feed详情
- [ ] `xiaohongshu/search.go` - 搜索功能
- [ ] `xiaohongshu/user_profile.go` - 用户主页
- [ ] `xiaohongshu/data.go` - 数据分析

#### 中优先级（交互功能）

- [ ] `xiaohongshu/action.go` - 互动操作
- [ ] `xiaohongshu/comment.go` - 评论相关
- [ ] `xiaohongshu/navigate.go` - 导航

### 4. 辅助函数说明

已在 `publish_with_config.go` 中提供了以下辅助函数：

```go
// getPublishURL 获取发布URL（带默认值）
func getPublishURL() string

// getSelector 获取选择器（带默认值）
func getSelector(selectorFunc func(*config.PublishSelectors) string, fallback string) string

// getTimeout 获取超时（带默认值）
func getTimeout(timeoutFunc func(*config.TimeoutsConfig) time.Duration, fallback time.Duration) time.Duration

// getInterval 获取间隔（带默认值）
func getInterval(intervalFunc func(*config.IntervalsConfig) time.Duration, fallback time.Duration) time.Duration
```

这些函数的优点：
- ✅ 配置加载失败时自动使用默认值
- ✅ 向后兼容，不会因为配置问题导致程序崩溃
- ✅ 代码更简洁

### 5. 改造示例

#### 示例 1: 改造 uploadImages 函数

**改造前:**
```go
func uploadImages(page *rod.Page, imagesPaths []string) error {
    pp := page.Timeout(30 * time.Second)
    uploadInput := pp.MustElement(".upload-input")
    uploadInput.MustSetFiles(validPaths...)
    return waitForUploadComplete(pp, len(validPaths))
}
```

**改造后:**
```go
func uploadImages(page *rod.Page, imagesPaths []string, cfg *config.Config) error {
    // 使用配置的超时
    timeout := 30 * time.Second
    if cfg != nil {
        timeout = cfg.Timeouts.GetImageUpload()
    }
    pp := page.Timeout(timeout)

    // 使用配置的选择器
    selector := ".upload-input"
    if cfg != nil {
        selector = cfg.Selectors.Publish.UploadInput
    }
    uploadInput := pp.MustElement(selector)

    uploadInput.MustSetFiles(validPaths...)
    return waitForUploadComplete(pp, len(validPaths), cfg)
}
```

#### 示例 2: 改造 mustClickPublishTab 函数

**改造前:**
```go
func mustClickPublishTab(page *rod.Page, tabname string) error {
    page.MustElement(`div.upload-content`).MustWaitVisible()

    deadline := time.Now().Add(15 * time.Second)
    for time.Now().Before(deadline) {
        tab, blocked, err := getTabElement(page, tabname)
        // ...
        time.Sleep(200 * time.Millisecond)
    }
    return errors.Errorf("没有找到发布 TAB - %s", tabname)
}
```

**改造后:**
```go
func mustClickPublishTab(page *rod.Page, tabname string, cfg *config.Config) error {
    // 使用配置的选择器
    uploadContentSelector := "div.upload-content"
    if cfg != nil {
        uploadContentSelector = cfg.Selectors.Publish.UploadContent
    }
    page.MustElement(uploadContentSelector).MustWaitVisible()

    // 使用配置的超时
    timeout := 15 * time.Second
    if cfg != nil {
        timeout = cfg.Timeouts.GetTabSearch()
    }
    deadline := time.Now().Add(timeout)

    // 使用配置的重试间隔
    retryInterval := 200 * time.Millisecond
    if cfg != nil {
        retryInterval = cfg.Intervals.GetTabRetry()
    }

    for time.Now().Before(deadline) {
        tab, blocked, err := getTabElement(page, tabname, cfg)
        // ...
        time.Sleep(retryInterval)
    }
    return errors.Errorf("没有找到发布 TAB - %s", tabname)
}
```

### 6. 测试配置文件化

#### 测试步骤

1. **编译程序:**
```bash
go build -o build/xiaohongshu-mcp-configurable .
```

2. **不指定配置文件运行（自动查找）:**
```bash
./build/xiaohongshu-mcp-configurable
# 会自动查找 config.yaml
```

3. **指定配置文件运行:**
```bash
./build/xiaohongshu-mcp-configurable --config ./custom_config.yaml
```

4. **测试配置修改:**

修改 `config.yaml` 中的某个选择器：
```yaml
selectors:
  publish:
    upload_input: ".new-upload-input"  # 修改为新的选择器
```

重启程序，新配置立即生效，无需重新编译。

### 7. 配置文件版本管理

建议：
```bash
# 将配置文件加入版本控制
git add config.yaml

# 创建示例配置文件
cp config.yaml config.example.yaml

# 忽略本地自定义配置
echo "config.local.yaml" >> .gitignore
```

用户可以：
1. 复制 `config.example.yaml` 为 `config.yaml`
2. 或创建 `config.local.yaml` 进行本地自定义

### 8. 配置热更新（可选）

如果需要在不重启的情况下更新配置，可以添加：

```go
// 在 config/config.go 中添加
func Reload() error {
    return LoadDefault()
}

// 在某个地方监听配置文件变化
// 使用 fsnotify 库
```

### 9. 配置验证

添加配置验证函数：

```go
func (c *Config) Validate() error {
    if c.URLs.Creator.PublishImage == "" {
        return errors.New("publish_image URL不能为空")
    }
    if c.Limits.MaxTags <= 0 {
        return errors.New("max_tags必须大于0")
    }
    // ... 更多验证
    return nil
}
```

### 10. 迁移进度跟踪

使用以下命令检查还有多少硬编码需要迁移：

```bash
# 检查硬编码URL
grep -rn "https://.*xiaohongshu.com" xiaohongshu/ --include="*.go" | wc -l

# 检查硬编码选择器（简单检测）
grep -rn 'MustElement("' xiaohongshu/ --include="*.go" | wc -l

# 检查硬编码超时
grep -rn "[0-9]* \* time\.(Second\|Millisecond)" xiaohongshu/ --include="*.go" | wc -l
```

## 总结

配置文件化改造的优势：

1. ✅ **易维护**: 页面更新时只需修改 YAML 文件
2. ✅ **可测试**: 可以快速切换不同配置进行测试
3. ✅ **可扩展**: 轻松添加新的配置项
4. ✅ **向后兼容**: 配置加载失败时使用默认值
5. ✅ **无需重编译**: 修改配置后直接生效

下一步建议：
1. 逐步迁移各个模块使用配置
2. 完善配置验证
3. 添加配置文档
4. 考虑添加配置热更新
