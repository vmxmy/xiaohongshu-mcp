package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/configs"
	"github.com/xpzouying/xiaohongshu-mcp/cookies"
	apppublish "github.com/xpzouying/xiaohongshu-mcp/internal/app/publish"
	domainpublish "github.com/xpzouying/xiaohongshu-mcp/internal/domain/publish"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
	"github.com/xpzouying/xiaohongshu-mcp/pkg/downloader"
	"github.com/xpzouying/xiaohongshu-mcp/xiaohongshu"
)

type loginProvider interface {
	GetQRCode(ctx context.Context) (loginQRResult, error)
}

// XiaohongshuService 小红书业务服务
type XiaohongshuService struct {
	publishUsecase *apppublish.Usecase
	loginManager   loginProvider
}

// NewXiaohongshuService 创建小红书服务实例
func NewXiaohongshuService() *XiaohongshuService {
	return NewXiaohongshuServiceWithUsecase(nil)
}

// NewXiaohongshuServiceWithUsecase 支持注入发布用例
func NewXiaohongshuServiceWithUsecase(publishUsecase *apppublish.Usecase) *XiaohongshuService {
	return &XiaohongshuService{
		publishUsecase: publishUsecase,
		loginManager:   NewLoginManager(newPlaywrightLoginSession, 4*time.Minute),
	}
}

// PublishRequest 发布请求
type PublishRequest struct {
	Title      string   `json:"title" binding:"required"`
	Content    string   `json:"content" binding:"required"`
	Images     []string `json:"images" binding:"required,min=1"`
	Tags       []string `json:"tags,omitempty"`
	ScheduleAt string   `json:"schedule_at,omitempty"` // 定时发布时间，ISO8601格式，为空则立即发布
}

// LoginStatusResponse 登录状态响应
type LoginStatusResponse struct {
	IsLoggedIn bool   `json:"is_logged_in"`
	Username   string `json:"username,omitempty"`
}

// LoginQrcodeResponse 登录扫码二维码
type LoginQrcodeResponse struct {
	Timeout    string `json:"timeout"`
	IsLoggedIn bool   `json:"is_logged_in"`
	Img        string `json:"img,omitempty"`
	Stage      string `json:"stage,omitempty"`
	Status     string `json:"status,omitempty"`
	SessionID  string `json:"session_id,omitempty"`
}

// PublishResponse 发布响应
type PublishResponse struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Images  int    `json:"images"`
	Status  string `json:"status"`
	PostID  string `json:"post_id,omitempty"`
}

// PublishVideoRequest 发布视频请求（仅支持本地单个视频文件）
type PublishVideoRequest struct {
	Title      string   `json:"title" binding:"required"`
	Content    string   `json:"content" binding:"required"`
	Video      string   `json:"video" binding:"required"`
	Tags       []string `json:"tags,omitempty"`
	ScheduleAt string   `json:"schedule_at,omitempty"` // 定时发布时间，ISO8601格式，为空则立即发布
}

// PublishVideoResponse 发布视频响应
type PublishVideoResponse struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Video   string `json:"video"`
	Status  string `json:"status"`
	PostID  string `json:"post_id,omitempty"`
}

// FeedsListResponse Feeds列表响应
type FeedsListResponse struct {
	Feeds []xiaohongshu.Feed `json:"feeds"`
	Count int                `json:"count"`
}

// UserProfileResponse 用户主页响应
type UserProfileResponse struct {
	UserBasicInfo xiaohongshu.UserBasicInfo      `json:"userBasicInfo"`
	Interactions  []xiaohongshu.UserInteractions `json:"interactions"`
	Feeds         []xiaohongshu.Feed             `json:"feeds"`
}

// DeleteCookies 删除 cookies 文件，用于登录重置
func (s *XiaohongshuService) DeleteCookies(ctx context.Context) error {
	cookiePath := cookies.GetCookiesFilePath()
	cookieLoader := cookies.NewLoadCookie(cookiePath)
	return cookieLoader.DeleteCookies()
}

// SyncCookies 写入 cookies 文件，供无头模式加载。
func (s *XiaohongshuService) SyncCookies(ctx context.Context, data []byte) (string, int64, error) {
	cookiePath := cookies.GetCookiesFilePath()
	cookieLoader := cookies.NewLoadCookie(cookiePath)
	if err := cookieLoader.SaveCookies(data); err != nil {
		return "", 0, err
	}
	return cookiePath, int64(len(data)), nil
}

// CheckLoginStatus 检查登录状态
func (s *XiaohongshuService) CheckLoginStatus(ctx context.Context) (*LoginStatusResponse, error) {
	var isLoggedIn bool
	var err error

	err = withBrowserPage(func(page browser.Page) error {
		loginAction := xiaohongshu.NewLogin(page)
		isLoggedIn, err = loginAction.CheckLoginStatus(ctx)
		return err
	})

	if err != nil {
		return nil, err
	}

	response := &LoginStatusResponse{
		IsLoggedIn: isLoggedIn,
		Username:   configs.Username,
	}

	return response, nil
}

// GetLoginQrcode 获取登录的扫码二维码
func (s *XiaohongshuService) GetLoginQrcode(ctx context.Context) (*LoginQrcodeResponse, error) {
	if s.loginManager == nil {
		return s.getLoginQrcodeLegacy(ctx)
	}

	result, err := s.loginManager.GetQRCode(ctx)
	if err != nil {
		return nil, err
	}

	img := result.Img
	if img != "" && !strings.HasPrefix(img, "data:image/png;base64,") {
		img = "data:image/png;base64," + img
	}

	return &LoginQrcodeResponse{
		Timeout:    result.Timeout,
		Img:        img,
		IsLoggedIn: result.IsLoggedIn,
		Stage:      result.Stage,
		Status:     result.Status,
		SessionID:  result.SessionID,
	}, nil
}

func (s *XiaohongshuService) getLoginQrcodeLegacy(ctx context.Context) (*LoginQrcodeResponse, error) {
	var img string
	var loggedIn bool
	var err error

	// 注意: 这个函数需要保持页面打开来等待登录，但 withBrowserPage 会在函数返回后关闭页面
	// 这里保持原有的逻辑，但使用新的浏览器引擎
	engine := newBrowserEngine()
	if err := engine.Start(); err != nil {
		return nil, err
	}

	page, err := engine.NewPage()
	if err != nil {
		engine.Close()
		return nil, err
	}

	deferFunc := func() {
		_ = page.Close()
		engine.Close()
	}

	loginAction := xiaohongshu.NewLogin(page)

	img, loggedIn, err = loginAction.FetchQrcodeImage(ctx)
	if err != nil || loggedIn {
		defer deferFunc()
	}
	if err != nil {
		return nil, err
	}

	timeout := 4 * time.Minute

	if !loggedIn {
		go func() {
			ctxTimeout, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			defer deferFunc()

			// Playwright 引擎会自动保存 cookies，无需手动调用 saveCookies
			_ = loginAction.WaitForLogin(ctxTimeout)
		}()
	}

	return &LoginQrcodeResponse{
		Timeout: func() string {
			if loggedIn {
				return "0s"
			}
			return timeout.String()
		}(),
		Img:        img,
		IsLoggedIn: loggedIn,
	}, nil
}

// PublishContent 发布内容
func (s *XiaohongshuService) PublishContent(ctx context.Context, req *PublishRequest) (*PublishResponse, error) {
	// 验证标题长度
	// 小红书限制：最大40个单位长度
	// 中文/日文/韩文占2个单位，英文/数字占1个单位
	if titleWidth := runewidth.StringWidth(req.Title); titleWidth > 40 {
		return nil, fmt.Errorf("标题长度超过限制")
	}

	// 处理图片：下载URL图片或使用本地路径
	imagePaths, err := s.processImages(req.Images)
	if err != nil {
		return nil, err
	}

	// 解析定时发布时间
	var scheduleTime *time.Time
	if req.ScheduleAt != "" {
		t, err := time.Parse(time.RFC3339, req.ScheduleAt)
		if err != nil {
			return nil, fmt.Errorf("定时发布时间格式错误，请使用 ISO8601 格式: %v", err)
		}

		// 校验定时发布时间范围：1小时至14天
		now := time.Now()
		minTime := now.Add(1 * time.Hour)
		maxTime := now.Add(14 * 24 * time.Hour)

		if t.Before(minTime) {
			return nil, fmt.Errorf("定时发布时间必须至少在1小时后，当前设置: %s，最早可选: %s",
				t.Format("2006-01-02 15:04"), minTime.Format("2006-01-02 15:04"))
		}
		if t.After(maxTime) {
			return nil, fmt.Errorf("定时发布时间不能超过14天，当前设置: %s，最晚可选: %s",
				t.Format("2006-01-02 15:04"), maxTime.Format("2006-01-02 15:04"))
		}

		scheduleTime = &t
		logrus.Infof("设置定时发布时间: %s", t.Format("2006-01-02 15:04"))
	}

	// 构建发布内容
	content := xiaohongshu.PublishImageContent{
		Title:        req.Title,
		Content:      req.Content,
		Tags:         req.Tags,
		ImagePaths:   imagePaths,
		ScheduleTime: scheduleTime,
	}

	// 执行发布（优先使用新用例）
	if s.publishUsecase != nil {
		if err := s.publishUsecase.PublishImage(ctx, domainpublish.ImageContent{
			Title:        content.Title,
			Content:      content.Content,
			Tags:         content.Tags,
			ImagePaths:   content.ImagePaths,
			ScheduleTime: content.ScheduleTime,
		}); err != nil {
			logrus.Errorf("发布内容失败(新用例): title=%s %v", content.Title, err)
			return nil, err
		}
	} else {
		if err := s.publishContent(ctx, content); err != nil {
			logrus.Errorf("发布内容失败: title=%s %v", content.Title, err)
			return nil, err
		}
	}

	response := &PublishResponse{
		Title:   req.Title,
		Content: req.Content,
		Images:  len(imagePaths),
		Status:  "发布完成",
	}

	return response, nil
}

// processImages 处理图片列表，支持URL下载和本地路径
func (s *XiaohongshuService) processImages(images []string) ([]string, error) {
	processor := downloader.NewImageProcessor()
	return processor.ProcessImages(images)
}

// publishContent 执行内容发布
func (s *XiaohongshuService) publishContent(ctx context.Context, content xiaohongshu.PublishImageContent) error {
	return withBrowserPage(func(page browser.Page) error {
		action, err := xiaohongshu.NewPublishImageAction(page)
		if err != nil {
			return err
		}

		// 执行发布
		return action.Publish(ctx, content)
	})
}

// PublishVideo 发布视频（本地文件）
func (s *XiaohongshuService) PublishVideo(ctx context.Context, req *PublishVideoRequest) (*PublishVideoResponse, error) {
	// 标题长度校验
	if titleWidth := runewidth.StringWidth(req.Title); titleWidth > 40 {
		return nil, fmt.Errorf("标题长度超过限制")
	}

	// 本地视频文件校验
	if req.Video == "" {
		return nil, fmt.Errorf("必须提供本地视频文件")
	}
	if _, err := os.Stat(req.Video); err != nil {
		return nil, fmt.Errorf("视频文件不存在或不可访问: %v", err)
	}

	// 解析定时发布时间
	var scheduleTime *time.Time
	if req.ScheduleAt != "" {
		t, err := time.Parse(time.RFC3339, req.ScheduleAt)
		if err != nil {
			return nil, fmt.Errorf("定时发布时间格式错误，请使用 ISO8601 格式: %v", err)
		}

		// 校验定时发布时间范围：1小时至14天
		now := time.Now()
		minTime := now.Add(1 * time.Hour)
		maxTime := now.Add(14 * 24 * time.Hour)

		if t.Before(minTime) {
			return nil, fmt.Errorf("定时发布时间必须至少在1小时后，当前设置: %s，最早可选: %s",
				t.Format("2006-01-02 15:04"), minTime.Format("2006-01-02 15:04"))
		}
		if t.After(maxTime) {
			return nil, fmt.Errorf("定时发布时间不能超过14天，当前设置: %s，最晚可选: %s",
				t.Format("2006-01-02 15:04"), maxTime.Format("2006-01-02 15:04"))
		}

		scheduleTime = &t
		logrus.Infof("设置定时发布时间: %s", t.Format("2006-01-02 15:04"))
	}

	// 构建发布内容
	content := xiaohongshu.PublishVideoContent{
		Title:        req.Title,
		Content:      req.Content,
		Tags:         req.Tags,
		VideoPath:    req.Video,
		ScheduleTime: scheduleTime,
	}

	// 执行发布
	if err := s.publishVideo(ctx, content); err != nil {
		return nil, err
	}

	resp := &PublishVideoResponse{
		Title:   req.Title,
		Content: req.Content,
		Video:   req.Video,
		Status:  "发布完成",
	}
	return resp, nil
}

// publishVideo 执行视频发布
func (s *XiaohongshuService) publishVideo(ctx context.Context, content xiaohongshu.PublishVideoContent) error {
	return withBrowserPage(func(page browser.Page) error {
		action, err := xiaohongshu.NewPublishVideoAction(page)
		if err != nil {
			return err
		}

		return action.PublishVideo(ctx, content)
	})
}

// ListFeeds 获取Feeds列表
func (s *XiaohongshuService) ListFeeds(ctx context.Context) (*FeedsListResponse, error) {
	var feeds []xiaohongshu.Feed
	var err error

	err = withBrowserPage(func(page browser.Page) error {
		// 创建 Feeds 列表 action
		action := xiaohongshu.NewFeedsListAction(page)

		// 获取 Feeds 列表
		feeds, err = action.GetFeedsList(ctx)
		return err
	})

	if err != nil {
		logrus.Errorf("获取 Feeds 列表失败: %v", err)
		return nil, err
	}

	response := &FeedsListResponse{
		Feeds: feeds,
		Count: len(feeds),
	}

	return response, nil
}

func (s *XiaohongshuService) SearchFeeds(ctx context.Context, keyword string, filters ...xiaohongshu.FilterOption) (*FeedsListResponse, error) {
	var feeds []xiaohongshu.Feed
	var err error

	err = withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewSearchAction(page)
		feeds, err = action.Search(ctx, keyword, filters...)
		return err
	})

	if err != nil {
		return nil, err
	}

	response := &FeedsListResponse{
		Feeds: feeds,
		Count: len(feeds),
	}

	return response, nil
}

// GetFeedDetail 获取Feed详情
func (s *XiaohongshuService) GetFeedDetail(ctx context.Context, feedID, xsecToken string, loadAllComments bool) (*FeedDetailResponse, error) {
	return s.GetFeedDetailWithConfig(ctx, feedID, xsecToken, loadAllComments, xiaohongshu.DefaultCommentLoadConfig())
}

// GetFeedDetailWithConfig 使用配置获取Feed详情
func (s *XiaohongshuService) GetFeedDetailWithConfig(ctx context.Context, feedID, xsecToken string, loadAllComments bool, config xiaohongshu.CommentLoadConfig) (*FeedDetailResponse, error) {
	var result *xiaohongshu.FeedDetailResponse
	var err error

	err = withBrowserPage(func(page browser.Page) error {
		// 创建 Feed 详情 action
		action := xiaohongshu.NewFeedDetailAction(page)

		// 获取 Feed 详情
		result, err = action.GetFeedDetailWithConfig(ctx, feedID, xsecToken, loadAllComments, config)
		return err
	})

	if err != nil {
		return nil, err
	}

	response := &FeedDetailResponse{
		FeedID: feedID,
		Data:   result,
	}

	return response, nil
}

// UserProfile 获取用户信息
func (s *XiaohongshuService) UserProfile(ctx context.Context, userID, xsecToken string) (*UserProfileResponse, error) {
	var result *xiaohongshu.UserProfileResponse
	var err error

	err = withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewUserProfileAction(page)
		result, err = action.UserProfile(ctx, userID, xsecToken)
		return err
	})

	if err != nil {
		return nil, err
	}

	response := &UserProfileResponse{
		UserBasicInfo: result.UserBasicInfo,
		Interactions:  result.Interactions,
		Feeds:         result.Feeds,
	}

	return response, nil
}

// PostCommentToFeed 发表评论到Feed
func (s *XiaohongshuService) PostCommentToFeed(ctx context.Context, feedID, xsecToken, content string) (*PostCommentResponse, error) {
	err := withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewCommentFeedAction(page)
		return action.PostComment(ctx, feedID, xsecToken, content)
	})

	if err != nil {
		return nil, err
	}

	return &PostCommentResponse{FeedID: feedID, Success: true, Message: "评论发表成功"}, nil
}

// LikeFeed 点赞笔记
func (s *XiaohongshuService) LikeFeed(ctx context.Context, feedID, xsecToken string) (*ActionResult, error) {
	err := withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewLikeAction(page)
		return action.Like(ctx, feedID, xsecToken)
	})

	if err != nil {
		return nil, err
	}

	return &ActionResult{FeedID: feedID, Success: true, Message: "点赞成功或已点赞"}, nil
}

// UnlikeFeed 取消点赞笔记
func (s *XiaohongshuService) UnlikeFeed(ctx context.Context, feedID, xsecToken string) (*ActionResult, error) {
	err := withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewLikeAction(page)
		return action.Unlike(ctx, feedID, xsecToken)
	})

	if err != nil {
		return nil, err
	}

	return &ActionResult{FeedID: feedID, Success: true, Message: "取消点赞成功或未点赞"}, nil
}

// FavoriteFeed 收藏笔记
func (s *XiaohongshuService) FavoriteFeed(ctx context.Context, feedID, xsecToken string) (*ActionResult, error) {
	err := withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewFavoriteAction(page)
		return action.Favorite(ctx, feedID, xsecToken)
	})

	if err != nil {
		return nil, err
	}

	return &ActionResult{FeedID: feedID, Success: true, Message: "收藏成功或已收藏"}, nil
}

// UnfavoriteFeed 取消收藏笔记
func (s *XiaohongshuService) UnfavoriteFeed(ctx context.Context, feedID, xsecToken string) (*ActionResult, error) {
	err := withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewFavoriteAction(page)
		return action.Unfavorite(ctx, feedID, xsecToken)
	})

	if err != nil {
		return nil, err
	}

	return &ActionResult{FeedID: feedID, Success: true, Message: "取消收藏成功或未收藏"}, nil
}

// ReplyCommentToFeed 回复指定评论
func (s *XiaohongshuService) ReplyCommentToFeed(ctx context.Context, feedID, xsecToken, commentID, userID, content string) (*ReplyCommentResponse, error) {
	err := withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewCommentFeedAction(page)
		return action.ReplyToComment(ctx, feedID, xsecToken, commentID, userID, content)
	})

	if err != nil {
		return nil, err
	}

	return &ReplyCommentResponse{
		FeedID:          feedID,
		TargetCommentID: commentID,
		TargetUserID:    userID,
		Success:         true,
		Message:         "评论回复成功",
	}, nil
}

// 注意：newBrowserEngine, withBrowserPage 已移至 browser_factory.go

// GetMyProfile 获取当前登录用户的个人信息
func (s *XiaohongshuService) GetMyProfile(ctx context.Context) (*UserProfileResponse, error) {
	var result *xiaohongshu.UserProfileResponse
	var err error

	err = withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewUserProfileAction(page)
		result, err = action.GetMyProfileViaSidebar(ctx)
		return err
	})

	if err != nil {
		return nil, err
	}

	response := &UserProfileResponse{
		UserBasicInfo: result.UserBasicInfo,
		Interactions:  result.Interactions,
		Feeds:         result.Feeds,
	}

	return response, nil
}

// FollowUser 关注用户
func (s *XiaohongshuService) FollowUser(ctx context.Context, userID, xsecToken string) (*ActionResult, error) {
	var err error

	err = withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewFollowAction(page)
		return action.Follow(ctx, userID, xsecToken)
	})

	if err != nil {
		return nil, err
	}

	return &ActionResult{
		FeedID:  userID,
		Success: true,
		Message: "关注成功",
	}, nil
}

// UnfollowUser 取关用户
func (s *XiaohongshuService) UnfollowUser(ctx context.Context, userID, xsecToken string) (*ActionResult, error) {
	var err error

	err = withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewFollowAction(page)
		return action.Unfollow(ctx, userID, xsecToken)
	})

	if err != nil {
		return nil, err
	}

	return &ActionResult{
		FeedID:  userID,
		Success: true,
		Message: "取关成功",
	}, nil
}

// LikeComment 点赞评论
func (s *XiaohongshuService) LikeComment(ctx context.Context, feedID, xsecToken, commentID, userID string) (*ActionResult, error) {
	var err error

	err = withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewCommentLikeAction(page)
		return action.LikeComment(ctx, feedID, xsecToken, commentID, userID)
	})

	if err != nil {
		return nil, err
	}

	return &ActionResult{
		FeedID:  feedID,
		Success: true,
		Message: "评论点赞成功",
	}, nil
}

// UnlikeComment 取消点赞评论
func (s *XiaohongshuService) UnlikeComment(ctx context.Context, feedID, xsecToken, commentID, userID string) (*ActionResult, error) {
	var err error

	err = withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewCommentLikeAction(page)
		return action.UnlikeComment(ctx, feedID, xsecToken, commentID, userID)
	})

	if err != nil {
		return nil, err
	}

	return &ActionResult{
		FeedID:  feedID,
		Success: true,
		Message: "取消评论点赞成功",
	}, nil
}

// ShareFeed 分享笔记，获取分享链接
func (s *XiaohongshuService) ShareFeed(ctx context.Context, feedID, xsecToken string) (string, error) {
	var shareLink string
	var err error

	err = withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewShareAction(page)
		shareLink, err = action.ShareFeed(ctx, feedID, xsecToken)
		return err
	})

	if err != nil {
		return "", err
	}

	return shareLink, nil
}

// DeleteFeed 删除自己的笔记
func (s *XiaohongshuService) DeleteFeed(ctx context.Context, feedID, xsecToken string) (*ActionResult, error) {
	var err error

	err = withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewDeleteAction(page)
		return action.DeleteFeed(ctx, feedID, xsecToken)
	})

	if err != nil {
		return nil, err
	}

	return &ActionResult{
		FeedID:  feedID,
		Success: true,
		Message: "笔记删除成功",
	}, nil
}

// DeleteComment 删除自己的评论
func (s *XiaohongshuService) DeleteComment(ctx context.Context, feedID, xsecToken, commentID, userID string) (*ActionResult, error) {
	var err error

	err = withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewDeleteAction(page)
		return action.DeleteComment(ctx, feedID, xsecToken, commentID, userID)
	})

	if err != nil {
		return nil, err
	}

	return &ActionResult{
		FeedID:  feedID,
		Success: true,
		Message: "评论删除成功",
	}, nil
}

// GetMyStats 获取当前用户的统计数据
func (s *XiaohongshuService) GetMyStats(ctx context.Context) (*xiaohongshu.UserStats, error) {
	var stats *xiaohongshu.UserStats
	var err error

	err = withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewDataAction(page)
		stats, err = action.GetMyStats(ctx)
		return err
	})

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetMyFeeds 获取自己发布的笔记列表
func (s *XiaohongshuService) GetMyFeeds(ctx context.Context, limit int) ([]xiaohongshu.Feed, error) {
	var feeds []xiaohongshu.Feed
	var err error

	err = withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewDataAction(page)
		feeds, err = action.GetMyFeeds(ctx, limit)
		return err
	})

	if err != nil {
		return nil, err
	}

	return feeds, nil
}

// GetFanAnalytics 获取粉丝分析数据
func (s *XiaohongshuService) GetFanAnalytics(ctx context.Context, period string) (*xiaohongshu.FanAnalytics, error) {
	var analytics *xiaohongshu.FanAnalytics
	var err error

	err = withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewDataAction(page)
		analytics, err = action.GetFanAnalytics(ctx, period)
		return err
	})

	if err != nil {
		return nil, err
	}

	return analytics, nil
}

// GetContentAnalytics 获取内容分析数据
func (s *XiaohongshuService) GetContentAnalytics(ctx context.Context, limit int) (*xiaohongshu.ContentAnalytics, error) {
	var analytics *xiaohongshu.ContentAnalytics
	var err error

	err = withBrowserPage(func(page browser.Page) error {
		action := xiaohongshu.NewDataAction(page)
		analytics, err = action.GetContentAnalytics(ctx, limit)
		return err
	})

	if err != nil {
		return nil, err
	}

	return analytics, nil
}
