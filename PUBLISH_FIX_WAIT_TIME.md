# 发布功能修复报告

## 问题诊断

**症状**：
- API 返回发布成功
- 小红书后台没有新帖子
- 没有报错

**根本原因**：
点击发布按钮后，代码立即返回成功，**没有等待发布流程完成**。

在 `gateway.go` 第71-74行：
```go
if err := page.Click(g.cfg.Selectors["submit"]); err != nil {
    return fmt.Errorf("publish image submit(%s): %w", g.cfg.Selectors["submit"], err)
}
return nil  // ❌ 立即返回，没有等待发布完成
```

## 修复内容

### 1. 图文发布流程（PublishImage）
```go
if err := page.Click(g.cfg.Selectors["submit"]); err != nil {
    return fmt.Errorf("publish image submit(%s): %w", g.cfg.Selectors["submit"], err)
}

// ✅ 新增：等待发布完成
time.Sleep(5 * time.Second)

return nil
```

### 2. 视频发布流程（PublishVideo）
同样添加 5 秒等待

### 3. 草稿保存流程（SaveImageDraft, SaveVideoDraft）
添加 3 秒等待（草稿保存通常更快）

### 4. 导入 time 包
```go
import (
    "context"
    "errors"
    "fmt"
    "time"  // ✅ 新增
    ...
)
```

## 重启服务步骤

1. **停止当前服务**：
   ```bash
   # 查找进程
   ps aux | grep xhs-mcp | grep -v grep

   # 杀掉进程
   kill <进程ID>
   ```

2. **启动新服务**：
   ```bash
   ./xhs-mcp --headless=false --port=:18060
   ```
   或者如果作为后台服务：
   ```bash
   nohup ./xhs-mcp --headless=false --port=:18060 > xhs-mcp.log 2>&1 &
   ```

3. **验证登录状态**：
   ```bash
   curl -s http://localhost:18060/api/v1/login/status | jq .
   ```

4. **重新测试发布**：
   ```bash
   ./test_url_final.sh
   ```

## 后续优化建议

当前使用固定的 5 秒等待，更好的方案是：

1. **监听 URL 变化**：
   - 发布成功后会跳转到 `/publish/success` 页面
   - 监听 URL 变化来判断发布完成

2. **检查成功消息**：
   - 等待页面出现"发布成功"提示
   - 使用选择器检测成功消息

3. **可配置等待时间**：
   - 在 `config.yaml` 中配置 `publish_wait_time`
   - 允许用户根据网络情况调整

示例改进代码：
```go
// 更好的方式（未实现）
func waitForPublishSuccess(page Page) error {
    // 方式1: 等待 URL 包含 /success
    // 方式2: 等待成功消息元素出现
    // 方式3: 设置最大等待时间 30 秒
}
```

---
**修复日期**：2026-01-24 16:15
**编译状态**：✅ 通过
**需要操作**：重启服务并重新测试
