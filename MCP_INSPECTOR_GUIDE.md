# MCP Inspector 使用指南 - 图文发帖调试

## 修复内容

### 1. 增强参数验证（mcp_handlers.go）

**修复前的问题**:
```go
imagePathsInterface, _ := args["images"].([]interface{})
// 断言失败时返回 nil，导致 imagePaths 为空数组
```

**修复后**:
```go
// 1. 添加调试日志
logrus.Debugf("MCP: 原始参数 - %+v", args)

// 2. 验证必需参数
if title == "" {
    return &MCPToolResult{
        Content: []MCPContent{{
            Type: "text",
            Text: "发布失败: 标题不能为空",
        }},
        IsError: true,
    }
}

// 3. 验证图片参数
if len(imagePaths) == 0 {
    logrus.Errorf("MCP: 图片参数错误 - 原始类型: %T, 值: %v", args["images"], args["images"])
    return &MCPToolResult{
        Content: []MCPContent{{
            Type: "text",
            Text: "发布失败: 至少需要1张图片。请确保 images 参数是字符串数组格式",
        }},
        IsError: true,
    }
}
```

### 2. 增强 Schema 定义（mcp_server.go）

```go
type PublishContentArgs struct {
    Title      string   `json:"title" jsonschema:"required,title=内容标题,..."`
    Content    string   `json:"content" jsonschema:"required,title=正文内容,..."`
    Images     []string `json:"images" jsonschema:"required,minItems=1,title=图片路径列表,..."`
    Tags       []string `json:"tags,omitempty" jsonschema:"title=话题标签列表,..."`
    ScheduleAt string   `json:"schedule_at,omitempty" jsonschema:"title=定时发布时间,..."`
}
```

**改进**:
- ✅ 添加 `required` 标记（title, content, images 必填）
- ✅ 添加 `minItems=1`（至少1张图片）
- ✅ 添加清晰的 title 和 description
- ✅ MCP Inspector 会基于这些 schema 渲染表单并前端验证

---

## 使用 MCP Inspector 测试发帖

### 步骤 1: 启动 MCP 服务器（开启调试日志）

```bash
# 停止旧服务
pkill -f xhs-mcp

# 启动服务（开启调试日志，非 headless 模式）
LOG_LEVEL=debug ./xhs-mcp -headless=false 2>&1 | tee mcp_debug.log
```

**为什么要这样启动**:
- ✅ `LOG_LEVEL=debug` - 显示详细的参数解析日志
- ✅ `-headless=false` - 显示浏览器窗口，方便观察发布过程
- ✅ `tee mcp_debug.log` - 同时保存日志到文件

### 步骤 2: 启动 MCP Inspector

```bash
npx @modelcontextprotocol/inspector
```

浏览器会自动打开：`http://localhost:6274/`

### 步骤 3: 连接到 MCP 服务器

在 MCP Inspector 页面：

1. **Server URL**: 输入 `http://127.0.0.1:18060/mcp`
2. **Transport Type**: 选择 `Streamable HTTP`
3. 点击 **Connect**

**成功标志**:
```
✅ Connected to xiaohongshu-mcp v2.0.0
✅ Session ID: xxx
✅ Available tools: 20+ tools
```

### 步骤 4: 调用 publish_content 工具

#### 方法 A: 使用表单（推荐）

1. 在左侧工具列表找到 **publish_content**
2. 点击展开，会看到表单字段：
   - **title** (必填) - 输入: `MCP测试发帖`
   - **content** (必填) - 输入: `通过 MCP Inspector 测试发布`
   - **images** (必填) - 点击 "Add item"，输入图片路径
   - **tags** (可选) - 添加标签，如 `["测试", "MCP"]`

3. 填写图片路径示例：
   ```
   /Users/xumingyang/app/xiaohongshu-mcp/test_images/test.jpg
   ```

4. 点击 **Call Tool** 按钮

#### 方法 B: 使用 JSON（高级）

点击 "Show Raw JSON"，输入：

```json
{
  "title": "MCP测试发帖",
  "content": "通过 MCP Inspector 测试发布功能",
  "images": [
    "/Users/xumingyang/app/xiaohongshu-mcp/test_images/test.jpg"
  ],
  "tags": ["测试", "MCP"]
}
```

### 步骤 5: 查看结果

#### ✅ 成功的标志

**MCP Inspector 显示**:
```
✅ Tool call successful
返回内容:
发布成功！
标题：MCP测试发帖
内容长度：17 字符
图片数量：1 张
状态：发布完成
```

**服务器日志显示**:
```
INFO[0001] MCP: 发布内容
DEBUG[0001] MCP: 原始参数 - map[content:通过... images:[...] title:MCP测试发帖]
INFO[0001] MCP: 发布内容 - 标题: MCP测试发帖, 图片数量: 1, 标签数量: 2, 定时:
DEBUG[0001] MCP: 图片路径 - [/Users/xumingyang/app/xiaohongshu-mcp/test_images/test.jpg]
INFO[0011] POST /mcp  200
```

**小红书创作中心**:
- 访问: https://creator.xiaohongshu.com
- 查看最新笔记
- 状态: 审核中 或 已发布

#### ❌ 失败的情况

**情况 1: 参数格式错误**

MCP Inspector 显示:
```
❌ Tool call failed
发布失败: 至少需要1张图片。请确保 images 参数是字符串数组格式
```

**原因**: images 字段输入格式不对

**修复**:
- ✅ 使用 "Add item" 按钮添加图片
- ✅ 确保每个图片是单独的字符串
- ❌ 不要输入 JSON 字符串如 `"[\"/path\"]"`

**情况 2: 图片路径不存在**

服务器日志显��:
```
ERROR[0001] 发布内容失败 - 图片下载失败: stat /path/to/image.jpg: no such file or directory
```

**修复**: 使用绝对路径，确保文件存在

**情况 3: 未登录**

服务器日志显示:
```
ERROR[0001] 发布内容失败 - playwright: timeout: Timeout 60000ms exceeded.
```

**修复**:
```bash
# 先登录
./start_login.sh

# 或使用 sync_cookies 工具
```

---

## 调试技巧

### 1. 实时查看日志

```bash
# 在另一个终端窗口
tail -f mcp_debug.log | grep -E "MCP:|ERROR|发布"
```

### 2. 对比 HTTP API 和 MCP 工具

**HTTP API（已验证成功）**:
```bash
./test_publish_complete.sh
```

**MCP 工具（正在调试）**:
- 使用 MCP Inspector 调用
- 对比两者的服务器日志
- 找出差异点

### 3. 检查参数传递

在 MCP Inspector 中:
1. 点击 **Show Raw Request/Response**
2. 查看实际发送的 JSON
3. 确认 `arguments.images` 是数组格式

### 4. 查看完整错误

服务器日志中搜索:
```bash
grep -A 10 "MCP: 发布内容" mcp_debug.log
grep "ERROR" mcp_debug.log
```

---

## 常见问题

### Q1: MCP Inspector 连接失败

**错误**: `Failed to connect to server`

**解决**:
1. 确认 xhs-mcp 服务已启动: `ps aux | grep xhs-mcp`
2. 确认端口正确: `http://127.0.0.1:18060/mcp`（不是 18060/api/v1）
3. 检查防火墙设置

### Q2: 工具列表为空

**错误**: 连接成功但看不到工具

**解决**:
1. 检查服务器日志是否有错误
2. 重启 MCP 服务器
3. 刷新 MCP Inspector 页面

### Q3: 图片上传失败

**错误**: `no valid images found`

**原因**: images 参数格式错误

**正确格式**:
```json
{
  "images": [
    "/Users/xumingyang/app/xiaohongshu-mcp/test_images/test.jpg"
  ]
}
```

**错误格式**:
```json
{
  "images": "/Users/xumingyang/app/xiaohongshu-mcp/test_images/test.jpg"  // ❌ 不是数组
}
```

---

## 下一步

1. ✅ 重启 MCP 服务器（使用新编译的版本）
2. ✅ 使用 MCP Inspector 测试 publish_content
3. ✅ 查看详细日志确认参数正确传递
4. ✅ 在小红书创作中心验证笔记发布成功

## 测试清单

- [ ] MCP Inspector 成功连接
- [ ] 工具列表显示 publish_content
- [ ] 表单字段验证正确（title, content, images 必填）
- [ ] 发送测试请求
- [ ] 服务器日志显示正确的参数
- [ ] 浏览器窗口显示发布过程
- [ ] MCP Inspector 返回成功消息
- [ ] 小红书创作中心显示新笔记

---

**总结**: 通过增强参数验证和 schema 定义，现在 MCP 工具应该能够正确处理参数并成功发布笔记了！
