# 小红书 MCP 工具 - 配置文件化版本使用说明

## 🎉 改造完成

已成功将小红书MCP工具进行配置文件化改造，现在所有关键参数都可以通过 `config.yaml` 进行配置，无需修改代码。

## 📦 文件说明

### 核心文件
- `config.yaml` - **配置文件** (最重要！页面更新时只需修改这个文件)
- `config/config.go` - 配置加载器
- `build/xiaohongshu-mcp-config` - 支持配置文件的可执行程序
- `build/xiaohongshu-mcp-fixed` - URL修复版本

### 文档文件
- `CONFIG_SUMMARY.md` - 配置化改造总结报告
- `CONFIG_MIGRATION_GUIDE.md` - 代码改造迁移指南
- `FIX_REPORT.md` - 发布功能问题修复报告
- `DEBUG_PUBLISH.md` - 发布功能调试指南
- `MANUAL_TEST_GUIDE.md` - 手动测试指南

## 🚀 快速开始

### 1. 使用修复版本（推荐）

```bash
# 直接使用URL修复版本（已内置修复）
./build/xiaohongshu-mcp-fixed

# 或使用支持配置文件的版本
./build/xiaohongshu-mcp-config
```

### 2. 自定义配置

如果小红书页面再次更新，修改 `config.yaml`:

```yaml
# 例如：发布URL变化了
urls:
  creator:
    publish_image: "https://creator.xiaohongshu.com/new/publish?mode=image"

# 例如：选择器变化了
selectors:
  publish:
    upload_input: ".new-upload-selector"
    submit_button: "button.new-submit-class"

# 例如：需要增加超时
timeouts:
  image_upload: 120  # 从60秒改为120秒
```

然后重启程序即可，**无需重新编译**。

### 3. 指定配置文件路径

```bash
./build/xiaohongshu-mcp-config --config /path/to/custom_config.yaml
```

## 🔧 主要改进

### 1. URL修复 ✅

**问题:**
- 小红书更新后，原URL默认打开视频上传页
- 导致发布图文时找不到正确的页面元素

**解决:**
```yaml
# config.yaml 中已更新
urls:
  creator:
    publish_image: "https://creator.xiaohongshu.com/publish/publish?source=official&target=image"
    # 注意：添加了 &target=image 参数
```

### 2. 配置文件化 ✅

**优势:**
- ✅ 页面更新时只需改YAML
- ✅ 无需懂Go语言
- ✅ 修改后立即生效
- ✅ 易于版本控制

**配置内容:**
- 所有URL（20+个）
- 所有CSS选择器（30+个）
- 所有超时参数（7个）
- 所有间隔参数（6个）
- 所有限制值（6个）
- API拦截配置
- 搜索过滤器配置

## 📋 配置文件示例

### 修改URL

```yaml
urls:
  creator:
    publish_image: "新的URL"
    publish_video: "新的URL"
    fan_analytics: "新的URL"
```

### 修改选择器

```yaml
selectors:
  publish:
    upload_input: ".new-class"
    title_input: "div.new-input input"
    submit_button: "button.publish-btn"
```

### 修改超时

```yaml
timeouts:
  navigate: 300      # 导航超时(秒)
  image_upload: 60   # 图片上传超时(秒)
  api_response: 30   # API响应超时(秒)
```

### 修改限制

```yaml
limits:
  max_tags: 10         # 最多10个标签
  max_title_width: 40  # 标题最多40字符
  max_images: 9        # 最多9张图片
```

## 🐛 调试

### 查看配置加载状态

启动程序时会显示：
```
INFO 配置文件加载成功
```

或

```
WARN 加载配置文件失败（将使用默认值）: ...
```

### 验证配置��件语法

```bash
# 检查YAML语法
cat config.yaml | python3 -c "import sys, yaml; yaml.safe_load(sys.stdin)" && echo "✅ 配置文件格式正确"
```

### 测试配置修改

1. 修改 `config.yaml` 中的某个值
2. 重启程序
3. 观察日志确认新配置生效

## 📝 常见场景

### 场景1: 小红书更新了发布按钮样式

**症状:** 点击发布后没反应

**解决:**
```yaml
# 找到新的发布按钮选择器，修改 config.yaml
selectors:
  publish:
    submit_button: "新的选择器"  # 修改这里
```

### 场景2: 图片上传变慢了

**症状:** 上传超时

**解决:**
```yaml
# 增加超时时间
timeouts:
  image_upload: 120  # 从60秒改为120秒
```

### 场景3: 发布页面URL改变了

**症状:** 打开的页面不对

**解决:**
```yaml
# 更新URL
urls:
  creator:
    publish_image: "新的发布页面URL"
```

## 🔄 版本对比

| 版本 | 特性 | 适用场景 |
|------|------|----------|
| `xiaohongshu-mcp-fixed` | URL已修复，硬编码 | 立即使用，简单可靠 |
| `xiaohongshu-mcp-config` | 支持配置文件 | 需要频繁调整参数 |

## 📚 进一步学习

- 阅读 `CONFIG_SUMMARY.md` 了解改造详情
- 阅读 `FIX_REPORT.md` 了解问题根因
- 阅读 `CONFIG_MIGRATION_GUIDE.md` 了解如何继续改造其他模块

## ⚠️ 注意事项

1. **配置文件位置**:
   - 程序会自动查找当前目录的 `config.yaml`
   - 或可执行文件所在目录的 `config.yaml`
   - 或使用 `--config` 参数指定路径

2. **配置加载失败**:
   - 不会导致程序崩溃
   - 会使用代码中的默认值
   - 会记录警告日志

3. **配置验证**:
   - 目前没有强制验证
   - 错误的配置可能导致功能异常
   - 建议测试后再使用

## 🎯 下一步

### 立即可做
1. ✅ 测试修复版本是否能正常发布
2. ✅ 熟悉配置文件的结构
3. ✅ 备份 `config.yaml` 文件

### 未来计划
- ⬜ 逐步改造其他模块使用配置
- ⬜ 添加配置验证
- ⬜ 添加配置热更新
- ⬜ 添加环境变量覆盖支持

## 💡 提示

**当小红书下次更新页面时:**
1. 不要慌张，不需要等开发者更新代码
2. 打开浏览器开发者工具，找到新的选择器或URL
3. 修改 `config.yaml` 文件
4. 重启程序即可

**这就是配置文件化的最大价值！** 🎉
