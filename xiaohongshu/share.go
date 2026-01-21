package xiaohongshu

import (
	"context"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"
)

// ShareAction 分享操作
type ShareAction struct {
	page *rod.Page
}

// NewShareAction 创建分享操作实例
func NewShareAction(page *rod.Page) *ShareAction {
	return &ShareAction{page: page}
}

// ShareFeed 分享笔记，获取分享链接
func (s *ShareAction) ShareFeed(ctx context.Context, feedID, xsecToken string) (string, error) {
	page := s.page.Context(ctx).Timeout(60 * time.Second)

	url := makeFeedDetailURL(feedID, xsecToken)
	logrus.Infof("打开 feed 详情页进行分享: %s", url)

	// 导航到详情页
	page.MustNavigate(url)
	page.MustWaitDOMStable()
	time.Sleep(2 * time.Second)

	// 检查页面是否可访问
	if err := checkPageAccessible(page); err != nil {
		return "", err
	}

	// 查找分享按钮
	shareBtn, err := s.findShareButton(page)
	if err != nil {
		return "", fmt.Errorf("未找到分享按钮: %w", err)
	}

	// 滚动到按钮位置
	logrus.Info("滚动到分享按钮...")
	if err := shareBtn.ScrollIntoView(); err != nil {
		logrus.Warnf("滚动失败: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// 点击分享按钮
	logrus.Info("点击分享按钮...")
	if err := shareBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		logrus.Warnf("点击失败: %v，尝试使用 JS 点击", err)

		// 备用方案：使用 JavaScript 点击
		_, err = page.Eval(`() => {
			const buttons = Array.from(document.querySelectorAll('button, [class*="share"]'));
			const shareBtn = buttons.find(btn =>
				btn.textContent.includes('分享') ||
				btn.className.includes('share')
			);
			if (shareBtn) {
				shareBtn.click();
				return true;
			}
			return false;
		}`)

		if err != nil {
			return "", fmt.Errorf("无法点击分享按钮: %w", err)
		}
	}

	time.Sleep(2 * time.Second)

	// 查找"复制链接"按钮
	copyLinkBtn, err := s.findCopyLinkButton(page)
	if err != nil {
		logrus.Warnf("未找到复制链接按钮: %v，尝试直接获取链接", err)
		// 直接返回当前页面URL作为分享链接
		return page.MustInfo().URL, nil
	}

	// 点击复制链接
	logrus.Info("点击复制链接按钮...")
	if err := copyLinkBtn.Click(proto.InputMouseButtonLeft, 1); err != nil {
		logrus.Warnf("点击复制链接失败: %v", err)
	}

	time.Sleep(1 * time.Second)

	// 尝试从剪贴板获取链接
	shareLink, err := s.getShareLinkFromClipboard(page)
	if err != nil {
		logrus.Warnf("从剪贴板获取链接失败: %v，使用当前URL", err)
		return page.MustInfo().URL, nil
	}

	logrus.Infof("成功获取分享链接: %s", shareLink)
	return shareLink, nil
}

// findShareButton 查找分享按钮
func (s *ShareAction) findShareButton(page *rod.Page) (*rod.Element, error) {
	// 尝试多个选择器
	selectors := []string{
		".share-wrapper",
		"[class*='share']",
		".interact-container .share",
		"button[aria-label*='分享']",
	}

	for _, sel := range selectors {
		elem, err := page.Timeout(3 * time.Second).Element(sel)
		if err == nil && elem != nil {
			logrus.Infof("找到分享按钮: %s", sel)
			return elem, nil
		}
	}

	// 尝试通过文本查找
	buttons, err := page.Elements("button")
	if err == nil {
		for _, btn := range buttons {
			text, _ := btn.Text()
			if text == "分享" {
				logrus.Info("通过文本找到分享按钮")
				return btn, nil
			}
		}
	}

	return nil, fmt.Errorf("所有选择器都失败")
}

// findCopyLinkButton 查找复制链接按钮
func (s *ShareAction) findCopyLinkButton(page *rod.Page) (*rod.Element, error) {
	// 尝试多个选择器
	selectors := []string{
		"button:has-text('复制链接')",
		"[class*='copy']",
		"button[aria-label*='复制']",
	}

	for _, sel := range selectors {
		elem, err := page.Timeout(3 * time.Second).Element(sel)
		if err == nil && elem != nil {
			logrus.Infof("找到复制链接按钮: %s", sel)
			return elem, nil
		}
	}

	// 尝试通过文本查找
	buttons, err := page.Elements("button")
	if err == nil {
		for _, btn := range buttons {
			text, _ := btn.Text()
			if text == "复制链接" || text == "复制" {
				logrus.Info("通过文本找到复制链接按钮")
				return btn, nil
			}
		}
	}

	return nil, fmt.Errorf("所有选择器都失败")
}

// getShareLinkFromClipboard 从剪贴板获取分享链接
func (s *ShareAction) getShareLinkFromClipboard(page *rod.Page) (string, error) {
	// 尝试使用 JavaScript 读取剪贴板（使用 Eval 而不是 MustEval 避免panic）
	result, err := page.Eval(`async () => {
		try {
			const text = await navigator.clipboard.readText();
			return text;
		} catch (err) {
			return "";
		}
	}`)

	if err != nil {
		return "", fmt.Errorf("读取剪贴板失败: %w", err)
	}

	text := result.Value.String()
	if text != "" {
		return text, nil
	}

	return "", fmt.Errorf("剪贴板为空")
}
