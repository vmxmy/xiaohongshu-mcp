package xiaohongshu

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	myerrors "github.com/xpzouying/xiaohongshu-mcp/errors"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
)

// ActionResult 通用动作响应（点赞/收藏等）
type ActionResult struct {
	FeedID  string `json:"feed_id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// 选择器常量
const (
	SelectorLikeButton    = ".like-wrapper .like-lottie"
	SelectorCollectButton = ".collect-wrapper .collect-icon"
)

// interactActionType 交互动作类型
type interactActionType string

const (
	actionLike       interactActionType = "点赞"
	actionFavorite   interactActionType = "收藏"
	actionUnlike     interactActionType = "取消点赞"
	actionUnfavorite interactActionType = "取消收藏"
)

type interactAction struct {
	page browser.Page
}

func newInteractAction(page browser.Page) *interactAction {
	return &interactAction{page: page}
}

func (a *interactAction) preparePage(ctx context.Context, actionType interactActionType, feedID, xsecToken string) browser.Page {
	page := a.page.WithContext(ctx).WithTimeout(60 * time.Second)
	url := makeFeedDetailURL(feedID, xsecToken)
	logrus.Infof("Opening feed detail page for %s: %s", actionType, url)

	if err := page.Goto(url); err != nil {
		logrus.Warnf("failed to navigate to %s: %v", url, err)
	}
	if err := page.WaitDOMStable(5*time.Second, 0.95); err != nil {
		logrus.Warnf("WaitDOMStable failed: %v", err)
	}
	time.Sleep(2 * time.Second)

	// 等待 __INITIAL_STATE__ 就绪
	a.waitForInitialState(page)

	return page
}

// waitForInitialState 等待页面 __INITIAL_STATE__ 数据就绪
func (a *interactAction) waitForInitialState(page browser.Page) {
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		result, err := page.Eval(`() => {
			return !!(window.__INITIAL_STATE__ &&
				window.__INITIAL_STATE__.note &&
				window.__INITIAL_STATE__.note.noteDetailMap &&
				Object.keys(window.__INITIAL_STATE__.note.noteDetailMap).length > 0);
		}`)
		if err != nil {
			logrus.Warnf("Eval error when waiting for __INITIAL_STATE__: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if boolResult, ok := result.(bool); ok && boolResult {
			logrus.Info("__INITIAL_STATE__ 数据就绪")
			return
		}

		logrus.Infof("等待 __INITIAL_STATE__ 就绪... (%d/%d)", i+1, maxRetries)
		time.Sleep(1 * time.Second)
	}
	logrus.Warn("__INITIAL_STATE__ 等待超时，继续尝试操作")
}

func (a *interactAction) performClick(page browser.Page, selector string) {
	if err := page.Click(selector); err != nil {
		logrus.Warnf("click selector %s failed: %v", selector, err)
	}
}

// LikeAction 负责处理点赞相关交互
type LikeAction struct {
	*interactAction
}

func NewLikeAction(page browser.Page) *LikeAction {
	return &LikeAction{interactAction: newInteractAction(page)}
}

// Like 点赞指定笔记，如果已点赞则直接返回
func (a *LikeAction) Like(ctx context.Context, feedID, xsecToken string) error {
	return a.perform(ctx, feedID, xsecToken, true)
}

// Unlike 取消点赞指定笔记，如果未点赞则直接返回
func (a *LikeAction) Unlike(ctx context.Context, feedID, xsecToken string) error {
	return a.perform(ctx, feedID, xsecToken, false)
}

func (a *LikeAction) perform(ctx context.Context, feedID, xsecToken string, targetLiked bool) error {
	actionType := actionLike
	if !targetLiked {
		actionType = actionUnlike
	}

	page := a.preparePage(ctx, actionType, feedID, xsecToken)

	liked, _, err := a.getInteractState(page, feedID)
	if err != nil {
		logrus.Warnf("failed to read interact state: %v (continue to try clicking)", err)
		return a.toggleLike(page, feedID, targetLiked, actionType)
	}

	if targetLiked && liked {
		logrus.Infof("feed %s already liked, skip clicking", feedID)
		return nil
	}
	if !targetLiked && !liked {
		logrus.Infof("feed %s not liked yet, skip clicking", feedID)
		return nil
	}

	return a.toggleLike(page, feedID, targetLiked, actionType)
}

func (a *LikeAction) toggleLike(page browser.Page, feedID string, targetLiked bool, actionType interactActionType) error {
	a.performClick(page, SelectorLikeButton)
	time.Sleep(3 * time.Second)

	liked, _, err := a.getInteractState(page, feedID)
	if err != nil {
		logrus.Warnf("验证%s状态失败: %v", actionType, err)
		return nil
	}
	if liked == targetLiked {
		logrus.Infof("feed %s %s成功", feedID, actionType)
		return nil
	}

	logrus.Warnf("feed %s %s可能未成功，状态未变化，尝试再次点击", feedID, actionType)
	a.performClick(page, SelectorLikeButton)
	time.Sleep(2 * time.Second)

	liked, _, err = a.getInteractState(page, feedID)
	if err != nil {
		logrus.Warnf("第二次验证%s状态失败: %v", actionType, err)
		return nil
	}
	if liked == targetLiked {
		logrus.Infof("feed %s 第二次点击%s成功", feedID, actionType)
		return nil
	}

	return nil
}

// FavoriteAction 负责处理收藏相关交互
type FavoriteAction struct {
	*interactAction
}

func NewFavoriteAction(page browser.Page) *FavoriteAction {
	return &FavoriteAction{interactAction: newInteractAction(page)}
}

// Favorite 收藏指定笔记，如果已收藏则直接返回
func (a *FavoriteAction) Favorite(ctx context.Context, feedID, xsecToken string) error {
	return a.perform(ctx, feedID, xsecToken, true)
}

// Unfavorite 取消收藏指定笔记，如果未收藏则直接返回
func (a *FavoriteAction) Unfavorite(ctx context.Context, feedID, xsecToken string) error {
	return a.perform(ctx, feedID, xsecToken, false)
}

func (a *FavoriteAction) perform(ctx context.Context, feedID, xsecToken string, targetCollected bool) error {
	actionType := actionFavorite
	if !targetCollected {
		actionType = actionUnfavorite
	}

	page := a.preparePage(ctx, actionType, feedID, xsecToken)

	_, collected, err := a.getInteractState(page, feedID)
	if err != nil {
		logrus.Warnf("failed to read interact state: %v (continue to try clicking)", err)
		return a.toggleFavorite(page, feedID, targetCollected, actionType)
	}

	if targetCollected && collected {
		logrus.Infof("feed %s already favorited, skip clicking", feedID)
		return nil
	}
	if !targetCollected && !collected {
		logrus.Infof("feed %s not favorited yet, skip clicking", feedID)
		return nil
	}

	return a.toggleFavorite(page, feedID, targetCollected, actionType)
}

func (a *FavoriteAction) toggleFavorite(page browser.Page, feedID string, targetCollected bool, actionType interactActionType) error {
	a.performClick(page, SelectorCollectButton)
	time.Sleep(3 * time.Second)

	_, collected, err := a.getInteractState(page, feedID)
	if err != nil {
		logrus.Warnf("验证%s状态失败: %v", actionType, err)
		return nil
	}
	if collected == targetCollected {
		logrus.Infof("feed %s %s成功", feedID, actionType)
		return nil
	}

	logrus.Warnf("feed %s %s可能未成功，状态未变化，尝试再次点击", feedID, actionType)
	a.performClick(page, SelectorCollectButton)
	time.Sleep(2 * time.Second)

	_, collected, err = a.getInteractState(page, feedID)
	if err != nil {
		logrus.Warnf("第二次验证%s状态失败: %v", actionType, err)
		return nil
	}
	if collected == targetCollected {
		logrus.Infof("feed %s 第二次点击%s成功", feedID, actionType)
		return nil
	}

	return nil
}

// getInteractState 从页面读取笔记的点赞/收藏状态（优先用DOM，fallback到__INITIAL_STATE__）
func (a *interactAction) getInteractState(page browser.Page, feedID string) (liked bool, collected bool, err error) {
	// 优先使用 DOM 判断状态（更可靠）
	result, evalErr := page.Eval(`() => {
		const likeBtn = document.querySelector('.like-wrapper .like-lottie, .like-wrapper');
		const collectBtn = document.querySelector('.collect-wrapper .collect-icon, .collect-wrapper');

		let liked = false;
		let collected = false;

		// 通过类名或属性判断点赞状态
		if (likeBtn) {
			liked = likeBtn.classList.contains('active') ||
				likeBtn.classList.contains('liked') ||
				likeBtn.closest('.like-wrapper')?.classList.contains('active') ||
				likeBtn.getAttribute('data-active') === 'true';
		}

		// 通过类名或属性判断收藏状态
		if (collectBtn) {
			collected = collectBtn.classList.contains('active') ||
				collectBtn.classList.contains('collected') ||
				collectBtn.closest('.collect-wrapper')?.classList.contains('active') ||
				collectBtn.getAttribute('data-active') === 'true';
		}

		// 如果DOM判断失败，尝试从 __INITIAL_STATE__ 读取
		if (!liked && !collected) {
			if (window.__INITIAL_STATE__ &&
				window.__INITIAL_STATE__.note &&
				window.__INITIAL_STATE__.note.noteDetailMap) {
				const noteDetailMap = window.__INITIAL_STATE__.note.noteDetailMap;
				const keys = Object.keys(noteDetailMap);
				if (keys.length > 0) {
					const detail = noteDetailMap[keys[0]];
					if (detail && detail.note && detail.note.interactInfo) {
						liked = detail.note.interactInfo.liked || false;
						collected = detail.note.interactInfo.collected || false;
					}
				}
			}
		}

		return JSON.stringify({liked: liked, collected: collected});
	}`)
	if evalErr != nil {
		return false, false, errors.Wrap(evalErr, "eval interactState failed")
	}

	resultStr, ok := result.(string)
	if !ok || resultStr == "" {
		return false, false, myerrors.ErrNoFeedDetail
	}

	// 解析结果
	var interactInfo struct {
		Liked     bool `json:"liked"`
		Collected bool `json:"collected"`
	}
	if err := json.Unmarshal([]byte(resultStr), &interactInfo); err != nil {
		return false, false, errors.Wrap(err, "unmarshal interactInfo failed")
	}

	return interactInfo.Liked, interactInfo.Collected, nil
}
