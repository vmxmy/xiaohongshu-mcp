package xiaohongshu

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
)

// CommentFeedAction 表示 Feed 评论动作
type CommentFeedAction struct {
	page browser.Page
	// TODO: smartSelector 需要等待迁移到 browser 接口后再启用
	// smartSelector *selector.SmartSelector
}

// NewCommentFeedAction 创建 Feed 评论动作
func NewCommentFeedAction(page browser.Page) *CommentFeedAction {
	// TODO: 等待 selector 迁移后再启用
	// 尝试加载智能选择器
	// configPath := filepath.Join("configs", "selectors.yaml")
	// smartSelector, err := selector.NewSmartSelector(configPath, page)
	// if err != nil {
	// 	logrus.Warnf("加载智能选择器失败，使用传统方式: %v", err)
	// }

	return &CommentFeedAction{
		page: page,
		// smartSelector: smartSelector,
	}
}

// PostComment 发表评论到 Feed
func (f *CommentFeedAction) PostComment(ctx context.Context, feedID, xsecToken, content string) error {
	// 设置超时，不继承外部 context
	page := f.page.WithTimeout(60 * time.Second)

	url := makeFeedDetailURL_browser(feedID, xsecToken)
	logrus.Infof("打开 feed 详情页: %s", url)

	// 导航到详情页
	if err := page.Goto(url); err != nil {
		return fmt.Errorf("导航失败: %w", err)
	}
	if err := page.WaitDOMStable(5*time.Second, 0.1); err != nil {
		logrus.Warnf("等待 DOM 稳定失败: %v", err)
	}
	time.Sleep(1 * time.Second)

	// 检测页面是否可访问
	if err := checkPageAccessible_browser(page); err != nil {
		return err
	}

	// 使用智能选择器查找评论输入框
	var inputElem browser.Element
	var err error

	// TODO: 等待 selector 迁移后再启用
	// if f.smartSelector != nil {
	// 	logrus.Info("使用智能选择器查找评论输入框...")
	// 	inputElem, err = f.smartSelector.FindElement("comment_input")
	// 	if err != nil {
	// 		logrus.Warnf("智能选择器失败: %v，回退到传统方式", err)
	// 		inputElem, err = f.findInputFallback(page)
	// 	}
	// } else {
	inputElem, err = f.findInputFallback(page)
	// }

	if inputElem == nil {
		logrus.Warnf("Failed to find comment input box with all selectors")
		return fmt.Errorf("未找到评论输入框，该帖子可能不支持评论或网页端不可访问")
	}

	// 滚动到输入框位置，确保可见
	logrus.Info("滚动到评论输入框...")
	if err := inputElem.ScrollIntoView(); err != nil {
		logrus.Warnf("滚动到输入框失败: %v", err)
	}
	time.Sleep(1 * time.Second)

	// 等待元素可见
	logrus.Info("等待输入框可见...")
	if err := inputElem.WaitVisible(); err != nil {
		logrus.Warnf("等待输入框可见失败: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// 点击输入框以激活
	logrus.Info("点击输入框...")
	if err := inputElem.Click(); err != nil {
		logrus.Warnf("点击输入框失败: %v，尝试继续", err)
		// 不返回错误，尝试继续输入
	}
	time.Sleep(1 * time.Second)

	// 清空输入框（如果有内容）
	logrus.Info("清空输入框...")
	_, err = page.Eval(`() => {
		const elem = document.querySelector('#content-textarea');
		if (elem) {
			elem.textContent = '';
			elem.innerText = '';
		}
	}`)
	if err != nil {
		logrus.Warnf("清空输入框失败: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// 使用 JavaScript 直接设置内容（更可靠）
	logrus.Infof("输入评论内容: %s", content)
	_, err = page.Eval(fmt.Sprintf(`() => {
		const elem = document.querySelector('#content-textarea');
		if (elem) {
			elem.textContent = %s;
			elem.innerText = %s;
			elem.focus();
			// 触发 input 事件
			elem.dispatchEvent(new Event('input', { bubbles: true }));
			return true;
		}
		return false;
	}`, strconv.Quote(content), strconv.Quote(content)))

	if err != nil {
		logrus.Warnf("使用 JS 输入失败，尝试 Input 方法: %v", err)
		// 备用方案：使用 Input 方法
		if err := inputElem.Input(content); err != nil {
			logrus.Warnf("Input 方法也失败: %v", err)
			return fmt.Errorf("无法输入评论内容: %w", err)
		}
	}

	time.Sleep(1 * time.Second)

	// 直接使用传统方式查找提交按钮（更可靠）
	var submitButton browser.Element
	logrus.Info("查找提交按钮...")
	submitButton, err = f.findSubmitButtonFallback(page)

	if submitButton == nil {
		logrus.Warnf("Failed to find submit button with all selectors")
		return fmt.Errorf("未找到提交按钮")
	}

	// 滚动到按钮位置
	logrus.Info("滚动到提交按钮...")
	if err := submitButton.ScrollIntoView(); err != nil {
		logrus.Warnf("滚动到按钮失败: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// 等待按钮可见
	logrus.Info("等待提交按钮可见...")
	if err := submitButton.WaitVisible(); err != nil {
		logrus.Warnf("等待按钮可见失败: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// 点击提交按钮
	logrus.Info("点击提交按钮...")
	if err := submitButton.Click(); err != nil {
		logrus.Warnf("点击提交按钮失败: %v，尝试使用 JS 点击", err)

		// 备用方案：使用 JavaScript 点击
		_, err = page.Eval(`() => {
			const buttons = Array.from(document.querySelectorAll('button'));
			const submitBtn = buttons.find(btn =>
				btn.textContent.includes('发布') ||
				btn.textContent.includes('提交') ||
				btn.className.includes('submit')
			);
			if (submitBtn) {
				submitBtn.click();
				return true;
			}
			return false;
		}`)

		if err != nil {
			return fmt.Errorf("无法点击提交按钮: %w", err)
		}
	}

	time.Sleep(1 * time.Second)

	logrus.Infof("Comment posted successfully to feed: %s", feedID)
	return nil
}

// ReplyToComment 回复指定评论
func (f *CommentFeedAction) ReplyToComment(ctx context.Context, feedID, xsecToken, commentID, userID, content string) error {
	// 增加超时时间，因为需要滚动查找评论
	page := f.page.WithTimeout(5 * time.Minute)
	url := makeFeedDetailURL_browser(feedID, xsecToken)
	logrus.Infof("打开 feed 详情页进行回复: %s", url)

	// 导航到详情页
	if err := page.Goto(url); err != nil {
		return fmt.Errorf("导航失败: %w", err)
	}
	if err := page.WaitDOMStable(5*time.Second, 0.1); err != nil {
		logrus.Warnf("等待 DOM 稳定失败: %v", err)
	}
	time.Sleep(1 * time.Second)

	// 检测页面是否可访问
	if err := checkPageAccessible_browser(page); err != nil {
		return err
	}

	// 等待评论容器加载
	time.Sleep(2 * time.Second)

	// 使用 Go 实现的查找逻辑
	commentEl, err := findCommentElement(page, commentID, userID)
	if err != nil {
		return fmt.Errorf("无法找到评论: %w", err)
	}

	// 滚动到评论位置
	logrus.Info("滚动到评论位置...")
	if err := commentEl.ScrollIntoView(); err != nil {
		return fmt.Errorf("滚动到评论失败: %w", err)
	}
	time.Sleep(1 * time.Second)

	logrus.Info("准备点击回复按钮")

	// 查找并点击回复按钮
	replyBtn, err := commentEl.Element(".right .interactions .reply")
	if err != nil {
		return fmt.Errorf("无法找到回复按钮: %w", err)
	}

	if err := replyBtn.Click(); err != nil {
		return fmt.Errorf("点击回复按钮失败: %w", err)
	}

	time.Sleep(1 * time.Second)

	// 查找回复输入框
	inputEl, err := page.Element("div.input-box div.content-edit p.content-input")
	if err != nil {
		return fmt.Errorf("无法找到回复输入框: %w", err)
	}

	// 输入内容
	if err := inputEl.Input(content); err != nil {
		return fmt.Errorf("输入回复内容失败: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	// 查找并点击提交按钮
	submitBtn, err := page.Element("div.bottom button.submit")
	if err != nil {
		return fmt.Errorf("无法找到提交按钮: %w", err)
	}

	if err := submitBtn.Click(); err != nil {
		return fmt.Errorf("点击提交按钮失败: %w", err)
	}

	time.Sleep(2 * time.Second)
	logrus.Infof("回复评论成功")
	return nil
}

// findCommentElement 查找指定评论元素（参考 feed_detail.go 的滚动逻辑）
func findCommentElement(page browser.Page, commentID, userID string) (browser.Element, error) {
	logrus.Infof("开始查找评论 - commentID: %s, userID: %s", commentID, userID)

	const maxAttempts = 100
	const scrollInterval = 800 * time.Millisecond

	// 先滚动到评论区
	scrollToCommentsArea_browser(page)
	time.Sleep(1 * time.Second)

	var lastCommentCount = 0
	stagnantChecks := 0

	logrus.Infof("开始循环查找，最大尝试次数: %d", maxAttempts)

	for attempt := 0; attempt < maxAttempts; attempt++ {
		logrus.Infof("=== 查找尝试 %d/%d ===", attempt+1, maxAttempts)

		// === 1. 检查是否到达底部 ===
		if checkEndContainer_browser(page) {
			logrus.Info("已到达评论底部，未找到目标评论")
			break
		}

		// === 2. 获取当前评论数量 ===
		currentCount := getCommentCount_browser(page)
		logrus.Infof("当前评论数: %d", currentCount)

		if currentCount != lastCommentCount {
			logrus.Infof("✓ 评论数增加: %d -> %d", lastCommentCount, currentCount)
			lastCommentCount = currentCount
			stagnantChecks = 0
		} else {
			stagnantChecks++
			if stagnantChecks%5 == 0 {
				logrus.Infof("评论数停滞 %d 次", stagnantChecks)
			}
		}

		// === 3. 停滞检测 ===
		if stagnantChecks >= 10 {
			logrus.Info("评论数量停滞超过10次，可能已加载完所有评论")
			break
		}

		// === 4. 先滚动到最后一个评论（触发懒加载）===
		if currentCount > 0 {
			logrus.Infof("滚动到最后一个评论（共 %d 条）", currentCount)

			// 使用 Go 获取所有评论元素
			elements, err := page.WithTimeout(2 * time.Second).Elements(".parent-comment, .comment-item, .comment")
			if err == nil && len(elements) > 0 {
				// 滚动到最后一个评论
				lastComment := elements[len(elements)-1]
				err := lastComment.ScrollIntoView()
				if err != nil {
					logrus.Warnf("滚动到最后一个评论失败: %v", err)
				}
			} else {
				logrus.Warnf("未找到评论元素: %v", err)
			}
			time.Sleep(300 * time.Millisecond)
		}

		// === 5. 继续向下滚动 ===
		logrus.Infof("继续向下滚动...")
		_, err := page.Eval(`() => { window.scrollBy(0, window.innerHeight * 0.8); return true; }`)
		if err != nil {
			logrus.Warnf("滚动失败: %v", err)
		}
		time.Sleep(500 * time.Millisecond)

		// === 6. 滚动后立即查找（边滚动边查找）===
		// 优先通过 commentID 查找（使用 Timeout 避免长时间等待）
		if commentID != "" {
			selector := fmt.Sprintf("#comment-%s", commentID)
			logrus.Infof("尝试通过 commentID 查找: %s", selector)

			// 使用 Timeout 避免长时间等待
			el, err := page.WithTimeout(2 * time.Second).Element(selector)
			if err == nil && el != nil {
				logrus.Infof("✓ 通过 commentID 找到评论: %s (尝试 %d 次)", commentID, attempt+1)
				return el, nil
			}
			logrus.Infof("未找到 commentID (2秒超时)")
		}

		// 通过 userID 查找
		if userID != "" {
			logrus.Infof("尝试通过 userID 查找: %s", userID)

			// 使用 Timeout 避免长时间等待
			elements, err := page.WithTimeout(2 * time.Second).Elements(".comment-item, .comment, .parent-comment")
			if err == nil && len(elements) > 0 {
				logrus.Infof("找到 %d 个评论元素", len(elements))
				for i, el := range elements {
					// 快速检查，不等待
					userEl, err := el.Element(fmt.Sprintf(`[data-user-id="%s"]`, userID))
					if err == nil && userEl != nil {
						logrus.Infof("✓ 通过 userID 在第 %d 个元素中找到评论: %s (尝试 %d 次)", i+1, userID, attempt+1)
						return el, nil
					}
				}
				logrus.Infof("在 %d 个元素中未找到匹配的 userID", len(elements))
			} else {
				logrus.Infof("获取评论元素失败或超时: %v", err)
			}
		}

		logrus.Infof("本次尝试未找到目标评论，继续下一轮...")

		// === 7. 等待内容加载 ===
		time.Sleep(scrollInterval)
	}

	return nil, fmt.Errorf("未找到评论 (commentID: %s, userID: %s), 尝试次数: %d", commentID, userID, maxAttempts)
}

// findInputFallback 传统方式查找输入框
func (f *CommentFeedAction) findInputFallback(page browser.Page) (browser.Element, error) {
	selectors := []string{
		"#content-textarea",
		"p.content-input[contenteditable]",
		"[contenteditable='true']#content-textarea",
		"p[contenteditable='true']",
		"textarea",
	}

	for _, sel := range selectors {
		elem, err := page.WithTimeout(3 * time.Second).Element(sel)
		if err == nil && elem != nil {
			logrus.Infof("找到输入框（传统方式）: %s", sel)
			return elem, nil
		}
	}

	return nil, fmt.Errorf("所有传统选择器都失败")
}

// findSubmitButtonFallback 传统方式查找提交按钮
func (f *CommentFeedAction) findSubmitButtonFallback(page browser.Page) (browser.Element, error) {
	selectors := []string{
		"button.submit",
		"div.bottom button.submit",
		"button[class*='submit']",
		"div.bottom button",
	}

	for _, sel := range selectors {
		elem, err := page.WithTimeout(3 * time.Second).Element(sel)
		if err == nil && elem != nil {
			logrus.Infof("找到提交按钮（传统方式）: %s", sel)
			return elem, nil
		}
	}

	// 尝试文本匹配
	buttons, err := page.Elements("button")
	if err == nil {
		for _, btn := range buttons {
			text, _ := btn.Text()
			if text == "发布" || text == "提交" {
				logrus.Info("通过文本找到提交按钮（传统方式）")
				return btn, nil
			}
		}
	}

	return nil, fmt.Errorf("所有传统选择器都失败")
}

// makeFeedDetailURL_browser 构造 Feed 详情页 URL (browser 版本)
func makeFeedDetailURL_browser(feedID, xsecToken string) string {
	return fmt.Sprintf("https://www.xiaohongshu.com/explore/%s?xsec_token=%s&xsec_source=pc_feed", feedID, xsecToken)
}

// checkPageAccessible_browser 检查页面是否可访问 (browser 版本)
func checkPageAccessible_browser(page browser.Page) error {
	time.Sleep(500 * time.Millisecond)

	// 查找错误提示容器
	has, err := page.Has(".access-wrapper, .error-wrapper, .not-found-wrapper, .blocked-wrapper")
	if err != nil || !has {
		// 未找到错误容器，说明页面可访问
		return nil
	}

	// 获取文本内容
	text, err := page.Text(".access-wrapper, .error-wrapper, .not-found-wrapper, .blocked-wrapper")
	if err != nil {
		// 无法获取文本，假设页面可访问
		return nil
	}

	// 检查是否包含错误关键词
	if text == "" {
		return nil
	}

	logrus.Warnf("页面不可访问: %s", text)
	return fmt.Errorf("页面不可访问: %s", text)
}

// scrollToCommentsArea_browser 滚动到评论区 (browser 版本)
func scrollToCommentsArea_browser(page browser.Page) {
	logrus.Info("滚动到评论区...")

	// 先定位到评论区
	has, _ := page.Has(".comments-container")
	if has {
		_ = page.ScrollIntoView(".comments-container")
	}
	// 等待滚动完成
	time.Sleep(500 * time.Millisecond)

	// 触发一次小滚动，激活懒加载机制
	_ = page.ScrollBy(0, 100)
}

// getCommentCount_browser 获取当前评论数量 (browser 版本)
func getCommentCount_browser(page browser.Page) int {
	elements, err := page.WithTimeout(2 * time.Second).Elements(".parent-comment")
	if err != nil {
		return 0
	}
	return len(elements)
}

// checkEndContainer_browser 检查是否到达评论底部 (browser 版本)
func checkEndContainer_browser(page browser.Page) bool {
	// 查找结束容器
	has, err := page.WithTimeout(2 * time.Second).Has(".end-container")
	if err != nil || !has {
		return false
	}

	// 获取文本内容
	text, err := page.Text(".end-container")
	if err != nil {
		return false
	}

	// 检查文本是否表示已到底部
	return text == "没有更多了" || text == "已经到底了"
}
