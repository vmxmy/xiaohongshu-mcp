package xiaohongshu

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/xpzouying/xiaohongshu-mcp/errors"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
)

type FeedsListAction struct {
	page browser.Page
}

func NewFeedsListAction(page browser.Page) *FeedsListAction {
	pp := page.WithTimeout(60 * time.Second)

	if err := pp.Goto("https://www.xiaohongshu.com"); err != nil {
		panic(fmt.Sprintf("导航失败: %v", err))
	}
	if err := pp.WaitDOMStable(time.Second, 0.1); err != nil {
		panic(fmt.Sprintf("等待 DOM 稳定失败: %v", err))
	}

	return &FeedsListAction{page: pp}
}

// GetFeedsList 获取页面的 Feed 列表数据
func (f *FeedsListAction) GetFeedsList(ctx context.Context) ([]Feed, error) {
	page := f.page.WithContext(ctx)

	time.Sleep(1 * time.Second)

	resultRaw, err := page.Eval(`() => {
		if (window.__INITIAL_STATE__ &&
		    window.__INITIAL_STATE__.feed &&
		    window.__INITIAL_STATE__.feed.feeds) {
			const feeds = window.__INITIAL_STATE__.feed.feeds;
			const feedsData = feeds.value !== undefined ? feeds.value : feeds._value;
			if (feedsData) {
				return JSON.stringify(feedsData);
			}
		}
		return "";
	}`)
	if err != nil {
		return nil, fmt.Errorf("failed to eval feeds: %w", err)
	}

	result, ok := resultRaw.(string)
	if !ok || result == "" {
		return nil, errors.ErrNoFeeds
	}

	var feeds []Feed
	if err := json.Unmarshal([]byte(result), &feeds); err != nil {
		return nil, fmt.Errorf("failed to unmarshal feeds: %w", err)
	}

	return feeds, nil
}
