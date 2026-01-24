package publish

import (
	"fmt"
	"strings"
	"time"

	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
)

// waitForPublishResult 验证发布结果
// 检查页面 URL 是否跳转到成功页面，或检查是否有错误提示
func (g *Gateway) waitForPublishResult(page browser.Page) error {
	// 最多等待 30 秒
	maxWaitTime := 30 * time.Second
	checkInterval := 500 * time.Millisecond
	deadline := time.Now().Add(maxWaitTime)

	for time.Now().Before(deadline) {
		// 1. 检查是否跳转到成功页面
		currentURL := page.URL()
		if currentURL != "" && (
		// 小红书发布成功后会跳转到成功页面或内容管理页面
		containsAny(currentURL, []string{
			"/publish/success",
			"/creator/publish/publish/complete",
			"/content",
		})) {
			return nil
		}

		// 2. 检查是否有错误提示（任意一个错误选择器）
		if errSelectors, ok := g.cfg.Selectors["error_message"]; ok {
			// error_message 可能是一个选择器字符串，检查是否可见
			isVisible, err := page.IsVisible(errSelectors)
			if err == nil && isVisible {
				// 尝试获取错误文本
				errText, _ := page.Text(errSelectors)
				if errText != "" {
					return fmt.Errorf("发布失败: %s", errText)
				}
				return fmt.Errorf("发布失败: 检测到错误提示")
			}
		}

		// 3. 检查是否有成功提示
		if successSelectors, ok := g.cfg.Selectors["success_message"]; ok {
			isVisible, err := page.IsVisible(successSelectors)
			if err == nil && isVisible {
				return nil
			}
		}

		time.Sleep(checkInterval)
	}

	// 超时，但不一定失败（小红书可能在处理中）
	// 检查最终 URL，如果还在发布页面，可能有问题
	finalURL := page.URL()
	if containsAny(finalURL, []string{"/publish/publish"}) {
		return fmt.Errorf("发布超时: 页面未跳转，可能发布失败或仍在处理中")
	}

	// URL 已改变但未匹配到已知的成功模式，谨慎认为成功
	return nil
}

// containsAny 检查字符串是否包含任意一个子串
func containsAny(s string, substrs []string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
