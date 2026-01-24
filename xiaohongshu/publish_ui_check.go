package xiaohongshu

import (
	"log/slog"
	"time"

	"github.com/pkg/errors"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
)

// waitForPublishResultUI 通过UI元素等待并验证发布结果(备用方案)
func waitForPublishResultUI(page browser.Page) error {
	maxWaitTime := 30 * time.Second
	checkInterval := 500 * time.Millisecond
	start := time.Now()

	slog.Info("等待发布结果(UI检测)...")

	for time.Since(start) < maxWaitTime {
		// 1. 检查是否有错误提示
		if hasError, errorMsg := checkForErrorMessage(page); hasError {
			slog.Error("发布失败，检测到错误提示", "error", errorMsg)
			return errors.Errorf("发布失败: %s", errorMsg)
		}

		// 2. 检查是否有成功提示弹窗
		if hasSuccess, successMsg := checkForSuccessDialog(page); hasSuccess {
			slog.Info("发布成功", "message", successMsg)
			return nil
		}

		// 3. 检查提交按钮是否变为不可点击状态(正在提交中)
		if isSubmitting := checkSubmitButtonState(page); isSubmitting {
			slog.Info("正在提交中...")
		}

		time.Sleep(checkInterval)
	}

	// 超时后做最后一次检查
	if hasError, errorMsg := checkForErrorMessage(page); hasError {
		return errors.Errorf("发布失败(超时检查): %s", errorMsg)
	}

	// 如果超时但没有明确的错误，记录警告
	slog.Warn("发布结果验证超时，可能发布成功但未收到确认")
	return nil
}
