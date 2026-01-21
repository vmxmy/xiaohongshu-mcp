# 智能选择器系统

## 概述

智能选择器系统是一个自适应的页面元素定位引擎，能够在网站页面结构变化时自动适应，大大提高系统的健壮性和可维护性。

## 核心特性

### 1. 配置化管理
- 所有选择器存储在 `configs/selectors.yaml` 配置文件中
- 支持热更新，无需重启服务
- 易于维护和更新

### 2. 多级备用策略
- **主选择器（Primary）**: 当前版本的最优选择器
- **备用选择器（Fallback）**: 旧版本或通用选择器
- **文本匹配（TextMatch）**: 通过文本内容查找元素

### 3. 智能验证
- 可见性验证
- 可点击性验证
- 可编辑性验证
- 父元素验证
- 文本内容验证
- 属性验证

### 4. 统计与监控
- 记录每个选择器的成功率
- 记录平均响应时间
- 自动优化选择器顺序
- 生成统计报告

### 5. 自动降级
- 主选择器失败时自动尝试备用选择器
- 所有选择器失败时尝试文本匹配
- 智能选择器失败时回退到传统方式

## 配置文件结构

```yaml
# configs/selectors.yaml

like_button:
  name: "点赞按钮"

  # 主选择器（按优先级排序）
  primary:
    - selector: ".like-wrapper .like-lottie"
      version: "2024-12"
      confidence: 0.95
      description: "当前版本选择器"

  # 备用选择器
  fallback:
    - selector: ".interact-container .left .like-lottie"
      version: "2024-06"
      confidence: 0.80
      description: "旧版本选择器"

    - selector: "[class*='like']"
      version: "universal"
      confidence: 0.60
      description: "模糊匹配"

  # 验证规则
  validation:
    must_be_visible: true
    must_be_clickable: true
    parent_contains: ["interact", "like"]
```

## 使用方法

### 基本用法

```go
// 创建智能选择器
smartSelector, err := selector.NewSmartSelector("configs/selectors.yaml", page)
if err != nil {
    log.Fatal(err)
}

// 查找元素
likeButton, err := smartSelector.FindElement("like_button")
if err != nil {
    log.Fatal(err)
}

// 使用元素
likeButton.MustClick()
```

### 在现有代码中集成

```go
// 点赞功能
action := NewLikeActionV2(page)
err := action.Like(ctx, feedID, xsecToken)

// 评论功能
commentAction := NewCommentFeedAction(page)
err := commentAction.PostComment(ctx, feedID, xsecToken, content)
```

## 热更新配置

### 修改配置文件

1. 编辑 `configs/selectors.yaml`
2. 修改或添加选择器
3. 保存文件

系统会在 5 秒内自动检测并加载新配置，无需重启服务。

### 示例：添加新选择器

```yaml
like_button:
  primary:
    # 添加新的选择器到列表顶部
    - selector: ".new-like-button"
      version: "2025-01"
      confidence: 0.98
      description: "2025年新版本"

    - selector: ".like-wrapper .like-lottie"
      version: "2024-12"
      confidence: 0.95
      description: "当前版本选择器"
```

## 统计信息

系统会自动记录选择器使用统计，保存在 `configs/selector_stats.json`：

```json
{
  "like_button": {
    "element_name": "like_button",
    "total_attempts": 100,
    "success_count": 98,
    "selector_stats": {
      ".like-wrapper .like-lottie": {
        "selector": ".like-wrapper .like-lottie",
        "success_count": 95,
        "fail_count": 2,
        "avg_time_ms": 45.3,
        "last_used": "2026-01-21T04:00:00Z"
      }
    },
    "last_successful": ".like-wrapper .like-lottie"
  }
}
```

## 添加新元素

### 1. 在配置文件中定义

```yaml
new_element:
  name: "新元素"
  primary:
    - selector: ".new-element-class"
      version: "2025-01"
      confidence: 0.95
      description: "主选择器"

  fallback:
    - selector: "[class*='new-element']"
      version: "universal"
      confidence: 0.70
      description: "备用选择器"

  validation:
    must_be_visible: true
```

### 2. 在代码中使用

```go
element, err := smartSelector.FindElement("new_element")
if err != nil {
    return err
}

element.MustClick()
```

## 最佳实践

### 1. 选择器优先级

1. **ID 选择器**: 最稳定，优先使用
   ```yaml
   selector: "#element-id"
   ```

2. **语义化属性**: ARIA 标签、data 属性
   ```yaml
   selector: "[aria-label='点赞']"
   ```

3. **Class 选择器**: 具体的 class 名称
   ```yaml
   selector: ".like-button"
   ```

4. **模糊匹配**: 包含特定关键词
   ```yaml
   selector: "[class*='like']"
   ```

5. **文本匹配**: 最后的备用方案
   ```yaml
   text_match: ["点赞", "Like"]
   ```

### 2. 验证规则

根据元素类型设置合适的验证规则：

```yaml
# 按钮
validation:
  must_be_visible: true
  must_be_clickable: true

# 输入框
validation:
  must_be_visible: true
  must_be_editable: true
  attributes:
    contenteditable: "true"

# 文本元素
validation:
  must_be_visible: true
  text_contains: ["关键词"]
```

### 3. 置信度设置

- **0.90-1.00**: 非常可靠的选择器（ID、特定class）
- **0.70-0.89**: 较可靠的选择器（组合选择器）
- **0.50-0.69**: 通用选择器（模糊匹配）
- **< 0.50**: 不推荐使用

## 故障排查

### 问题：所有选择器都失败

1. 检查配置文件是否正确加载
2. 查看日志中的详细错误信息
3. 使用浏览器开发者工具检查实际的页面结构
4. 更新配置文件中的选择器

### 问题：热更新不生效

1. 检查配置文件语法是否正确（YAML格式）
2. 查看日志中是否有加载错误
3. 确认文件修改时间已更新

### 问题：验证失败

1. 检查元素是否真的可见/可点击
2. 调整验证规则
3. 添加等待时间

## 未来扩展

### 计划中的功能

1. **AI 辅助识别**: 使用 Claude Vision API 自动识别元素
2. **自动学习**: 根据成功率自动调整选择器顺序
3. **可视化监控**: Web 界面查看统计信息
4. **告警系统**: 选择器失败率过高时自动告警
5. **A/B 测试**: 测试不同选择器的效果

## 贡献指南

欢迎贡献新的选择器配置或改进建议！

1. Fork 项目
2. 创建特性分支
3. 提交更改
4. 发起 Pull Request

## 许可证

MIT License
