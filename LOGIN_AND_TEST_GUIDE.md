# 小红书 MCP 登录和测试流程

## 问题总结

当前发布失败的原因：**未登录小红书**
```
is_logged_in: false
```

发布页面需要登录后才能访问，所以出现了超时错误。

## 解决步骤

### 1. 启动有头浏览器登录

```bash
./start_login.sh
```

或者直接运行：
```bash
./bin/xiaohongshu-login
```

**操作流程：**
1. 浏览器窗口会自动打开
2. 显示小红书登录页面
3. 使用手机小红书 App 扫描二维码
4. 确认登录
5. Cookies 会自动保存到 `~/.xhs-mcp/cookies.yaml`
6. 浏览器自动关闭

### 2. 验证登录状态

```bash
curl -s http://localhost:18060/api/v1/login/status | jq .
```

应该看到：
```json
{
  "success": true,
  "data": {
    "is_logged_in": true,
    "username": "你的用户名"
  }
}
```

### 3. （可选）同步 Cookies 到远程

如果你有远程 VPS 部署，需要同步 cookies：

```bash
# 查看本地 cookies
cat ~/.xhs-mcp/cookies.yaml

# 使用 MCP 工具同步
# 调用 sync_cookies 工具
```

### 4. 测试 URL 图片发布

登录成功后，再次测试发布：

```bash
# 方法1: 通过 HTTP API
curl -X POST http://localhost:18060/api/v1/publish \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "URL图片发布测试",
    "content": "测试使用远程URL图片发布",
    "images": ["https://pub-c918abf638c7475fa46f2a62c795106f.r2.dev/images/20260123-144120-313.png"],
    "tags": ["测试", "URL图片"]
  }' | jq .
```

```bash
# 方法2: 通过 MCP Router
# 在 Claude Desktop 或 MCP Client 中调用 publish_content 工具
```

## URL 图片功能说明

✅ **当前发布功能已支持 URL 图片**

工作原理：
1. 用户提供图片 URL
2. `processImages()` 自动检测是否为 HTTP/HTTPS 链接
3. `downloader.ImageProcessor` 下载图片到本地临时目录
4. 返回本地文件路径
5. 使用本地路径上传到小红书

支持的图片来源：
- ✅ HTTP/HTTPS 图片链接（自动下载）
- ✅ 本地图片绝对路径
- ✅ 混合使用（部分 URL + 部分本地路径）

## 下一步

1. 运行 `./start_login.sh` 登录小红书
2. 扫码确认登录
3. 验证登录状态
4. 重新测试 URL 图片发布功能

---
**更新时间**: 2026-01-24 16:05
**状态**: 等待用户登录
