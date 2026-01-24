# 使用 MCP Inspector 调试内容分析

## 步骤

### 1. 启动服务（有头模式）
```bash
pkill -f xhs-mcp
LOG_LEVEL=debug ./xhs-mcp -headless=false -port=:18060
```

### 2. 启动 MCP Inspector（新终端）
```bash
npx @modelcontextprotocol/inspector
```

浏览器会自动打开 http://localhost:6274/

### 3. 连接服务器
- Server URL: `http://127.0.0.1:18060/mcp`
- Transport Type: `Streamable HTTP`
- 点击 **Connect**

### 4. 调用 get_content_analytics 工具
- 工具列表中找到: `get_content_analytics`
- 参数设置: `limit = 100`
- 点击 **Call Tool**

### 5. 观察浏览器窗口
**这时候才会打开小红书页面！**

- 会自动访问内容分析页面
- **立即按 F12 打开开发者工具**
- **切换到 Console 标签页**

### 6. 查看 Console 输出

你会看到：
```
=== Start extracting notes ===
Total rows found: XX
Row 0 preview: ...
Row 1 title matched: ...
...
Successfully extracted notes: 7
```

### 7. 把 Console 输出告诉我

特别注意：
- Total rows found: 多少行？
- 有多少行 "title match failed"？
- 有多少行 "insufficient numbers"？
- 为什么只提取了 7 条？

---

## 为什么用 Inspector？

- ✅ 自动处理 MCP 初始化
- ✅ 可视化界面，容易操作
- ✅ 只有在点击 Call Tool 时才会打开浏览器
- ✅ 可以看到完整的响应数据
