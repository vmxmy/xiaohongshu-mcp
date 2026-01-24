# 保存草稿功能实现报告

## 功能概述

添加了两个新的 MCP 工具，允许用户将小红书内容保存为草稿而不立即发布：
- `save_draft` - 保存图文草稿
- `save_video_draft` - 保存视频草稿

## 实现内容

### 1. 配置更新 (config.yaml)

添加暂存按钮选择器：
```yaml
selectors:
  publish:
    save_draft_button: "button:has-text('暂存离开')"
```

### 2. Domain 层

无需修改，复用现有的 `ImageContent` 和 `VideoContent` 类型。

### 3. Port 层 (internal/app/ports/ports.go)

扩展 `PublishGateway` 接口：
```go
type PublishGateway interface {
    PublishImage(ctx context.Context, content publish.ImageContent) error
    PublishVideo(ctx context.Context, content publish.VideoContent) error
    SaveImageDraft(ctx context.Context, content publish.ImageContent) error  // 新增
    SaveVideoDraft(ctx context.Context, content publish.VideoContent) error  // 新增
}
```

### 4. Infrastructure 层 (internal/infra/xhs/publish/gateway.go)

实现草稿保存方法：
- `SaveImageDraft` - 与发布流程相同，但最后点击"暂存离开"按钮
- `SaveVideoDraft` - 视频草稿保存

### 5. Application 层 (internal/app/publish/usecase.go)

添加保存草稿用例：
```go
func (u Usecase) SaveImageDraft(ctx context.Context, content publish.ImageContent) error
func (u Usecase) SaveVideoDraft(ctx context.Context, content publish.VideoContent) error
```

### 6. MCP 接口层

#### 6.1 参数定义 (mcp_server.go)
```go
type SaveDraftArgs struct {
    Title   string   `json:"title"`
    Content string   `json:"content"`
    Images  []string `json:"images"`
    Tags    []string `json:"tags,omitempty"`
}

type SaveVideoDraftArgs struct {
    Title   string   `json:"title"`
    Content string   `json:"content"`
    Video   string   `json:"video"`
    Tags    []string `json:"tags,omitempty"`
}
```

#### 6.2 工具注册 (mcp_server.go)
注册两个新工具：
- 工具 11.5: `save_draft`
- 工具 11.6: `save_video_draft`

#### 6.3 处理器实现 (mcp_handlers.go)
- `handleSaveDraft` - 处理图文草稿保存
- `handleSaveVideoDraft` - 处理视频草稿保存

### 7. 服务器初始化 (app_server.go, main.go)

- 在 `AppServer` 中添加 `publishUsecase` 字段
- 创建 `NewAppServerWithPublish` 构造函数
- 在 main.go 中注入 publishUsecase

### 8. 配置加载 (wiring_bootstrap.go)

更新选择器配置结构，加载 `save_draft_button`。

## 使用方式

### 保存图文草稿

```json
{
  "tool": "save_draft",
  "arguments": {
    "title": "草稿标题",
    "content": "草稿内容",
    "images": ["/path/to/image1.jpg", "/path/to/image2.jpg"],
    "tags": ["标签1", "标签2"]
  }
}
```

### 保存视频草稿

```json
{
  "tool": "save_video_draft",
  "arguments": {
    "title": "视频草稿标题",
    "content": "视频草稿描述",
    "video": "/path/to/video.mp4",
    "tags": ["vlog", "生活"]
  }
}
```

## 与发布功能的对比

| 特性 | 发布功能 | 保存草稿功能 |
|------|---------|------------|
| 上传图片/视频 | ✅ | ✅ |
| 填写标题 | ✅ | ✅ |
| 填写内容 | ✅ | ✅ |
| 添加标签 | ✅ | ✅ |
| 定时发布 | ✅ | ❌ |
| 最后操作 | 点击"发布" | 点击"暂存离开" |
| 结果 | 进入审核 | 保存为草稿 |

## 测试建议

1. **图文草稿测试**：
   ```bash
   # 调用 MCP 工具 save_draft
   # 检查小红书创作者中心的草稿箱
   # 确认内容已保存
   ```

2. **视频草稿测试**：
   ```bash
   # 调用 MCP 工具 save_video_draft
   # 检查视频草稿是否正确保存
   ```

3. **边界测试**：
   - 空标题/内容
   - 无图片/视频
   - 无效文件路径

## 后续优化建议

1. **草稿列表管理**：
   - 添加获取草稿列表的工具
   - 添加编辑草稿的工具
   - 添加删除草稿的工具
   - 添加从草稿发布的工具

2. **错误处理增强**：
   - 更详细的错误提示
   - 重试机制

3. **用户体验**：
   - 保存草稿后返回草稿 ID
   - 支持部分字段更新

---
**实现日期**：2026-01-24
**实现人**：Claude
**编译状态**：✅ 通过
**测试状态**：⏳ 待测试
