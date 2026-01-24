package xiaohongshu

import (
	"context"
	"time"

	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
)

type NavigateAction struct {
	page browser.Page
}

func NewNavigate(page browser.Page) *NavigateAction {
	return &NavigateAction{page: page}
}

func (n *NavigateAction) ToExplorePage(ctx context.Context) error {
	page := n.page.WithContext(ctx)

	// 导航到探索页面
	if err := page.Goto("https://www.xiaohongshu.com/explore"); err != nil {
		return err
	}

	// 等待页面加载完成
	if err := page.WaitLoad(); err != nil {
		return err
	}

	// 等待主要元素出现
	_, err := page.Element(`div#app`)
	if err != nil {
		return err
	}

	return nil
}

func (n *NavigateAction) ToProfilePage(ctx context.Context) error {
	page := n.page.WithContext(ctx)

	// 首先导航到探索页面
	if err := n.ToExplorePage(ctx); err != nil {
		return err
	}

	// 等待 DOM 稳定
	if err := page.WaitDOMStable(time.Second, 0.1); err != nil {
		return err
	}

	// 查找并点击侧边栏中的"我"频道链接
	profileLink, err := page.Element(`div.main-container li.user.side-bar-component a.link-wrapper span.channel`)
	if err != nil {
		return err
	}

	if err := profileLink.Click(); err != nil {
		return err
	}

	// 等待导航完成
	if err := page.WaitLoad(); err != nil {
		return err
	}

	return nil
}
