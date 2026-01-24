# 小红书 MCP 工具 - 快速开始

## 🚀 快速启动

### 1. 编译项目
```bash
go build -o xhs-mcp .
```

### 2. 启动服务
```bash
# 生产环境（无头模式，推荐）
./xhs-mcp -port=:18060

# 调试模式（显示浏览器）
./start_mcp_debug.sh
```

### 3. 验证服务
```bash
curl http://localhost:18060/health
```

---

## 📝 使用方式

### 方式 1: HTTP API

```bash
curl -X POST http://localhost:18060/api/v1/publish \
  -H "Content-Type: application/json" \
  -d '{
    "title": "我的笔记",
    "content": "这是内容",
    "images": ["/path/to/image.jpg"],
    "tags": ["标签1", "标签2"]
  }'
```

### 方式 2: MCP Inspector（调试）

1. 启动 Inspector:
```bash
npx @modelcontextprotocol/inspector
```

2. 连接到: `http://127.0.0.1:18060/mcp`

3. 使用 `publish_content` 工具

### 方式 3: Claude Desktop（生产）

配置 `~/Library/Application Support/Claude/claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "xiaohongshu": {
      "command": "/path/to/xhs-mcp",
      "args": []
    }
  }
}
```

然后在 Claude Desktop 中直接对话发布。

---

## 🔧 常用命令

### 登录小红书
```bash
./start_login.sh
```

### 快速测试
```bash
./quick_test.sh
```

### 查看日志
```bash
# 启动时输出到文件
./xhs-mcp 2>&1 | tee mcp.log

# 实时查看
tail -f mcp.log
```

---

## 📚 完整文档

- **HEADLESS_MODE_FIX.md** - 无头模式修复说明
- **MCP_USAGE.md** - MCP工具详细使用指南
- **config.yaml** - 配置文件说明

---

## ✅ 功能支持

| 功能 | HTTP API | MCP 工具 | 状态 |
|------|---------|---------|------|
| 图文发布 | ✅ | ✅ | 已验证 |
| 视频发布 | ✅ | ✅ | 已修复 |
| 图文草稿 | ✅ | ✅ | 已修复 |
| 视频草稿 | ✅ | ✅ | 已修复 |
| 定时发布 | ✅ | ✅ | 支持 |
| 标签添加 | ✅ | ✅ | 支持 |
| 无头模式 | ✅ | ✅ | 已修复 |

---

## 🐛 故障排查

### 发布失败

1. 检查是否已登录:
```bash
curl http://localhost:18060/api/v1/login/status
```

2. 如果未登录，运行:
```bash
./start_login.sh
```

### 服务无法启动

检查端口是否被占用:
```bash
lsof -i :18060
```

如果被占用，停止旧进程:
```bash
pkill -f xhs-mcp
```

### 无头模式失败

1. 确认视口设置正确（已自动配置为1920x1080）
2. 查看日志中"按钮可见性"状态
3. 如仍有问题，使用有头模式调试:
```bash
./xhs-mcp -headless=false
```

---

## 💡 最佳实践

### 生产部署
- ✅ 使用无头模式（默认）
- ✅ 配置日志轮转
- ✅ 使用 systemd/supervisor 管理进程
- ✅ 定期备份 cookies.json

### 开发调试
- ✅ 使用有头模式观察浏览器
- ✅ 启用 DEBUG 日志: `LOG_LEVEL=debug ./xhs-mcp`
- ✅ 使用 MCP Inspector 测试工具

---

**项目地址**: https://github.com/xpzouying/xiaohongshu-mcp
