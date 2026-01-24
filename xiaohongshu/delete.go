package xiaohongshu

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
)

// DeleteAction 删除操作
type DeleteAction struct {
	page browser.Page
}

// NewDeleteAction 创建删除操作实例
func NewDeleteAction(page browser.Page) *DeleteAction {
	return &DeleteAction{page: page}
}

// DeleteFeed 删除自己的笔记
func (d *DeleteAction) DeleteFeed(ctx context.Context, feedID, xsecToken string) error {
	page := d.page.WithContext(ctx).WithTimeout(60 * time.Second)

	url := makeFeedDetailURL(feedID, xsecToken)
	logrus.Infof("打开 feed 详情页进行删除: %s", url)

	// 导航到详情页
	if err := page.Goto(url); err != nil {
		return fmt.Errorf("导航到详情页失败: %w", err)
	}
	if err := page.WaitDOMStable(time.Second, 0.1); err != nil {
		return fmt.Errorf("等待DOM稳定失败: %w", err)
	}
	time.Sleep(2 * time.Second)

	// 检查页面是否可访问
	if err := checkPageAccessible(page); err != nil {
		return err
	}

	// 查找更多按钮（三个点）
	moreBtn, err := d.findMoreButton(page)
	if err != nil {
		return fmt.Errorf("未找到更多按钮: %w", err)
	}

	// 点击更多按钮
	logrus.Info("点击更多按钮...")
	if err := moreBtn.Click(); err != nil {
		return fmt.Errorf("点击更多按钮失败: %w", err)
	}

	time.Sleep(1 * time.Second)

	// 查找删除按钮
	deleteBtn, err := d.findDeleteButton(page)
	if err != nil {
		return fmt.Errorf("未找到删除按钮: %w", err)
	}

	// 点击删除按钮
	logrus.Info("点击删除按钮...")
	if err := deleteBtn.Click(); err != nil {
		return fmt.Errorf("点击删除按钮失败: %w", err)
	}

	time.Sleep(1 * time.Second)

	// 查找确认删除按钮
	confirmBtn, err := d.findConfirmButton(page)
	if err != nil {
		return fmt.Errorf("未找到确认按钮: %w", err)
	}

	// 点击确认删除
	logrus.Info("点击确认删除...")
	if err := confirmBtn.Click(); err != nil {
		return fmt.Errorf("点击确认按钮失败: %w", err)
	}

	time.Sleep(2 * time.Second)

	logrus.Infof("笔记删除成功: %s", feedID)
	return nil
}

// DeleteComment 删除自己的评论
func (d *DeleteAction) DeleteComment(ctx context.Context, feedID, xsecToken, commentID, userID string) error {
	page := d.page.WithContext(ctx).WithTimeout(5 * time.Minute)

	url := makeFeedDetailURL(feedID, xsecToken)
	logrus.Infof("打开 feed 详情页进行删除评论: %s", url)

	// 导航到详情页
	if err := page.Goto(url); err != nil {
		return fmt.Errorf("导航到详情页失败: %w", err)
	}
	if err := page.WaitDOMStable(time.Second, 0.1); err != nil {
		return fmt.Errorf("等待DOM稳定失败: %w", err)
	}
	time.Sleep(2 * time.Second)

	// 检查页面是否可访问
	if err := checkPageAccessible(page); err != nil {
		return err
	}

	// 等待评论容器加载
	time.Sleep(2 * time.Second)

	// 查找评论元素
	commentEl, err := findCommentElement(page, commentID, userID)
	if err != nil {
		return fmt.Errorf("无法找到评论: %w", err)
	}

	// 滚动到评论位置
	logrus.Info("滚动到评论位置...")
	if err := commentEl.ScrollIntoView(); err != nil {
		return fmt.Errorf("滚动到评论位置失败: %w", err)
	}
	time.Sleep(1 * time.Second)

	// 查找评论的更多按钮
	moreBtn, err := d.findCommentMoreButton(commentEl)
	if err != nil {
		return fmt.Errorf("未找到评论更多按钮: %w", err)
	}

	// 点击更多按钮
	logrus.Info("点击评论更多按钮...")
	if err := moreBtn.Click(); err != nil {
		return fmt.Errorf("点击更多按钮失败: %w", err)
	}

	time.Sleep(1 * time.Second)

	// 查找删除按钮
	deleteBtn, err := d.findDeleteButton(page)
	if err != nil {
		return fmt.Errorf("未找到删除按钮: %w", err)
	}

	// 点击删除按钮
	logrus.Info("点击删除按钮...")
	if err := deleteBtn.Click(); err != nil {
		return fmt.Errorf("点击删除按钮失败: %w", err)
	}

	time.Sleep(1 * time.Second)

	// 查找确认删除按钮
	confirmBtn, err := d.findConfirmButton(page)
	if err != nil {
		logrus.Warnf("未找到确认按钮，可能已直接删除: %v", err)
		return nil
	}

	// 点击确认删除
	logrus.Info("点击确认删除...")
	if err := confirmBtn.Click(); err != nil {
		return fmt.Errorf("点击确认按钮失败: %w", err)
	}

	time.Sleep(2 * time.Second)

	logrus.Infof("评论删除成功")
	return nil
}

// findMoreButton 查找更多按钮（三个点）
func (d *DeleteAction) findMoreButton(page browser.Page) (browser.Element, error) {
	selectors := []string{
		".more-button",
		"[class*='more']",
		"button[aria-label*='更多']",
		".operate-button",
	}

	for _, sel := range selectors {
		elem, err := page.WithTimeout(3 * time.Second).Element(sel)
		if err == nil && elem != nil {
			logrus.Infof("找到更多按钮: %s", sel)
			return elem, nil
		}
	}

	return nil, fmt.Errorf("所有选择器都失败")
}

// findCommentMoreButton 查找评论的更多按钮
func (d *DeleteAction) findCommentMoreButton(commentEl browser.Element) (browser.Element, error) {
	selectors := []string{
		".more",
		"[class*='more']",
		".operate",
	}

	for _, sel := range selectors {
		elem, err := commentEl.Element(sel)
		if err == nil && elem != nil {
			logrus.Infof("找到评论更多按钮: %s", sel)
			return elem, nil
		}
	}

	return nil, fmt.Errorf("所有选择器都失败")
}

// findDeleteButton 查找删除按钮
func (d *DeleteAction) findDeleteButton(page browser.Page) (browser.Element, error) {
	selectors := []string{
		"button:has-text('删除')",
		"[class*='delete']",
	}

	for _, sel := range selectors {
		elem, err := page.WithTimeout(3 * time.Second).Element(sel)
		if err == nil && elem != nil {
			logrus.Infof("找到删除按钮: %s", sel)
			return elem, nil
		}
	}

	// 尝试通过文本查找
	buttons, err := page.Elements("button")
	if err == nil {
		for _, btn := range buttons {
			text, _ := btn.Text()
			if text == "删除" {
				logrus.Info("通过文本找到删除按钮")
				return btn, nil
			}
		}
	}

	return nil, fmt.Errorf("所有选择器都失败")
}

// findConfirmButton 查找确认按钮
func (d *DeleteAction) findConfirmButton(page browser.Page) (browser.Element, error) {
	selectors := []string{
		"button:has-text('确认')",
		"button:has-text('确定')",
		"[class*='confirm']",
	}

	for _, sel := range selectors {
		elem, err := page.WithTimeout(3 * time.Second).Element(sel)
		if err == nil && elem != nil {
			logrus.Infof("找到确认按钮: %s", sel)
			return elem, nil
		}
	}

	// 尝试通过文本查找
	buttons, err := page.Elements("button")
	if err == nil {
		for _, btn := range buttons {
			text, _ := btn.Text()
			if text == "确认" || text == "确定" {
				logrus.Info("通过文本找到确认按钮")
				return btn, nil
			}
		}
	}

	return nil, fmt.Errorf("所有选择器都失败")
}
