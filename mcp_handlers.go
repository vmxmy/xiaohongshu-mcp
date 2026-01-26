package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/cookies"
	apperrors "github.com/xpzouying/xiaohongshu-mcp/errors"
	domainpublish "github.com/xpzouying/xiaohongshu-mcp/internal/domain/publish"
	"github.com/xpzouying/xiaohongshu-mcp/xiaohongshu"
)

// MCP 工具处理函数

// parseBool 安全地解析布尔值，兼容 bool 和 string 类型
func parseBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		b, _ := strconv.ParseBool(val)
		return b
	default:
		return false
	}
}

// handleCheckLoginStatus 处理检查登录状态
func (s *AppServer) handleCheckLoginStatus(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: 检查登录状态")

	status, err := s.xiaohongshuService.CheckLoginStatus(ctx)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "检查登录状态失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	// 根据 IsLoggedIn 判断并返回友好的提示
	var resultText string
	if status.IsLoggedIn {
		resultText = fmt.Sprintf("✅ 已登录\n用户名: %s\n\n你可以使用其他功能了。", status.Username)
	} else {
		resultText = fmt.Sprintf("❌ 未登录\n\n请使用 get_login_qrcode 工具获取二维码进行登录。")
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: resultText,
		}},
	}
}

// handleGetLoginQrcode 处理获取登录二维码请求。
// 返回二维码图片的 Base64 编码和超时时间，供前端展示扫码登录。
func (s *AppServer) handleGetLoginQrcode(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: 获取登录扫码图片")

	result, err := s.xiaohongshuService.GetLoginQrcode(ctx)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "获取登录扫码图片失败: " + err.Error()}},
			IsError: true,
		}
	}

	if result.IsLoggedIn {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "你当前已处于登录状态"}},
		}
	}

	now := time.Now()
	deadline := func() string {
		d, err := time.ParseDuration(result.Timeout)
		if err != nil {
			return now.Format("2006-01-02 15:04:05")
		}
		return now.Add(d).Format("2006-01-02 15:04:05")
	}()

	text := "请用小红书 App 在 " + deadline + " 前扫码登录 👇"
	if strings.EqualFold(result.Stage, "security") {
		text = "请用小红书 App 在 " + deadline + " 前扫码完成安全认证 👇"
	}
	lines := []string{text}
	if result.Status != "" {
		lines = append(lines, "状态: "+result.Status)
	}
	if result.SessionID != "" {
		lines = append(lines, "会话: "+result.SessionID)
	}
	text = strings.Join(lines, "\n")

	// 已登录：文本 + 图片
	contents := []MCPContent{
		{Type: "text", Text: text},
		{
			Type:     "image",
			MimeType: "image/png",
			Data:     strings.TrimPrefix(result.Img, "data:image/png;base64,"),
		},
	}
	return &MCPToolResult{Content: contents}
}

func parseSyncCookiesPayload(args SyncCookiesArgs) ([]byte, error) {
	if strings.TrimSpace(args.CookiesBase64) != "" {
		data, err := base64.StdEncoding.DecodeString(args.CookiesBase64)
		if err != nil {
			return nil, fmt.Errorf("cookies_base64 解码失败: %w", err)
		}
		return data, nil
	}
	if strings.TrimSpace(args.CookiesJSON) != "" {
		return []byte(args.CookiesJSON), nil
	}
	return nil, errors.New("cookies_base64 或 cookies_json 至少提供一个")
}

func validateCookiesJSON(data []byte) error {
	var items []map[string]any
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("cookies JSON 无法解析: %w", err)
	}
	return nil
}

// handleSyncCookies 处理上传 cookies 请求。
func (s *AppServer) handleSyncCookies(ctx context.Context, args SyncCookiesArgs) *MCPToolResult {
	logrus.Info("MCP: 上传 cookies")

	payload, err := parseSyncCookiesPayload(args)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "cookies 输入无效: " + err.Error()}},
			IsError: true,
		}
	}
	if err := validateCookiesJSON(payload); err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "cookies JSON 校验失败: " + err.Error()}},
			IsError: true,
		}
	}

	path, size, err := s.xiaohongshuService.SyncCookies(ctx, payload)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "保存 cookies 失败: " + err.Error()}},
			IsError: true,
		}
	}
	logrus.WithFields(logrus.Fields{"path": path, "bytes": size}).Info("cookies 已写入")
	return &MCPToolResult{
		Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("cookies 已写入: %s (%d bytes)", path, size)}},
	}
}

// handleDeleteCookies 处理删除 cookies 请求，用于登录重置
func (s *AppServer) handleDeleteCookies(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: 删除 cookies，重置登录状态")

	err := s.xiaohongshuService.DeleteCookies(ctx)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{Type: "text", Text: "删除 cookies 失败: " + err.Error()}},
			IsError: true,
		}
	}

	cookiePath := cookies.GetCookiesFilePath()
	resultText := fmt.Sprintf("Cookies 已成功删除，登录状态已重置。\n\n删除的文件路径: %s\n\n下次操作时，需要重新登录。", cookiePath)
	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: resultText,
		}},
	}
}

// handlePublishContent 处理发布内容
func (s *AppServer) handlePublishContent(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logrus.Info("MCP: 发布内容")
	logrus.Debugf("MCP: 原始参数 - %+v", args)

	// 解析参数
	title, _ := args["title"].(string)
	content, _ := args["content"].(string)
	imagePathsInterface, _ := args["images"].([]interface{})
	tagsInterface, _ := args["tags"].([]interface{})

	// 验证必需参数
	if title == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "发布失败: 标题不能为空",
			}},
			IsError: true,
		}
	}

	if content == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "发布失败: 内容不能为空",
			}},
			IsError: true,
		}
	}

	// 解析图片路径
	var imagePaths []string
	for _, path := range imagePathsInterface {
		if pathStr, ok := path.(string); ok {
			imagePaths = append(imagePaths, pathStr)
		}
	}

	// 验证图片
	if len(imagePaths) == 0 {
		logrus.Errorf("MCP: 图片参数错误 - 原始类型: %T, 值: %v", args["images"], args["images"])
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "发布失败: 至少需要1张图片。请确保 images 参数是字符串数组格式，如: [\"图片路径1\", \"图片路径2\"]",
			}},
			IsError: true,
		}
	}

	// 解析标签
	var tags []string
	for _, tag := range tagsInterface {
		if tagStr, ok := tag.(string); ok {
			tags = append(tags, tagStr)
		}
	}

	// 解析地点
	location, _ := args["location"].(string)

	// 解析标记标签
	markerTagsInterface, _ := args["marker_tags"].([]interface{})
	var markerTags []string
	for _, marker := range markerTagsInterface {
		if markerStr, ok := marker.(string); ok {
			markerTags = append(markerTags, markerStr)
		}
	}

	// 解析定时发布参数
	scheduleAt, _ := args["schedule_at"].(string)

	logrus.Infof("MCP: 发布内容 - 标题: %s, 图片数量: %d, 标签数量: %d, 地点: %s, 标记数量: %d, 定时: %s",
		title, len(imagePaths), len(tags), location, len(markerTags), scheduleAt)
	logrus.Debugf("MCP: 图片路径 - %v", imagePaths)
	logrus.Debugf("MCP: 标记标签 - %v", markerTags)

	// 构建发布请求
	req := &PublishRequest{
		Title:      title,
		Content:    content,
		Images:     imagePaths,
		Tags:       tags,
		Location:   location,
		MarkerTags: markerTags,
		ScheduleAt: scheduleAt,
	}

	// 执行发布
	result, err := s.xiaohongshuService.PublishContent(ctx, req)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "发布失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	resultText := fmt.Sprintf("内容发布成功: %+v", result)
	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: resultText,
		}},
	}
}

// handlePublishVideo 处理发布视频内容（仅本地单个视频文件）
func (s *AppServer) handlePublishVideo(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logrus.Info("MCP: 发布视频内容（本地）")

	title, _ := args["title"].(string)
	content, _ := args["content"].(string)
	videoPath, _ := args["video"].(string)
	tagsInterface, _ := args["tags"].([]interface{})

	var tags []string
	for _, tag := range tagsInterface {
		if tagStr, ok := tag.(string); ok {
			tags = append(tags, tagStr)
		}
	}

	if videoPath == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "发布失败: 缺少本地视频文件路径",
			}},
			IsError: true,
		}
	}

	// 解析定时发布参数
	scheduleAt, _ := args["schedule_at"].(string)

	logrus.Infof("MCP: 发布视频 - 标题: %s, 标签数量: %d, 定时: %s", title, len(tags), scheduleAt)

	// 构建发布请求
	req := &PublishVideoRequest{
		Title:      title,
		Content:    content,
		Video:      videoPath,
		Tags:       tags,
		ScheduleAt: scheduleAt,
	}

	// 执行发布
	result, err := s.xiaohongshuService.PublishVideo(ctx, req)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "发布失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	resultText := fmt.Sprintf("视频发布成功: %+v", result)
	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: resultText,
		}},
	}
}

// handleSaveDraft 处理保存草稿
func (s *AppServer) handleSaveDraft(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logrus.Info("MCP: 保存草稿")

	// 解析参数
	title, _ := args["title"].(string)
	content, _ := args["content"].(string)
	imagePathsInterface, _ := args["images"].([]interface{})
	tagsInterface, _ := args["tags"].([]interface{})

	var imagePaths []string
	for _, path := range imagePathsInterface {
		if pathStr, ok := path.(string); ok {
			imagePaths = append(imagePaths, pathStr)
		}
	}

	var tags []string
	for _, tag := range tagsInterface {
		if tagStr, ok := tag.(string); ok {
			tags = append(tags, tagStr)
		}
	}

	logrus.Infof("MCP: 保存草稿 - 标题: %s, 图片数量: %d, 标签数量: %d", title, len(imagePaths), len(tags))

	// 调用保存草稿服务
	if s.publishUsecase == nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "保存草稿失败: 发布服务未初始化",
			}},
			IsError: true,
		}
	}

	publishContent := domainpublish.ImageContent{
		Title:      title,
		Content:    content,
		Tags:       tags,
		ImagePaths: imagePaths,
	}

	if err := s.publishUsecase.SaveImageDraft(ctx, publishContent); err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "保存草稿失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: "草稿保存成功",
		}},
	}
}

// handleSaveVideoDraft 处理保存视频草稿
func (s *AppServer) handleSaveVideoDraft(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logrus.Info("MCP: 保存视频草稿")

	title, _ := args["title"].(string)
	content, _ := args["content"].(string)
	videoPath, _ := args["video"].(string)
	tagsInterface, _ := args["tags"].([]interface{})

	var tags []string
	for _, tag := range tagsInterface {
		if tagStr, ok := tag.(string); ok {
			tags = append(tags, tagStr)
		}
	}

	if videoPath == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "保存草稿失败: 缺少本地视频文件路径",
			}},
			IsError: true,
		}
	}

	logrus.Infof("MCP: 保存视频草稿 - 标题: %s, 标签数量: %d", title, len(tags))

	// 调用保存视频草稿服务
	if s.publishUsecase == nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "保存草稿失败: 发布服务未初始化",
			}},
			IsError: true,
		}
	}

	publishContent := domainpublish.VideoContent{
		Title:     title,
		Content:   content,
		Tags:      tags,
		VideoPath: videoPath,
	}

	if err := s.publishUsecase.SaveVideoDraft(ctx, publishContent); err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "保存视频草稿失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: "视频草稿保存成功",
		}},
	}
}

// handleListFeeds 处理获取Feeds列表
func (s *AppServer) handleListFeeds(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: 获取Feeds列表")

	result, err := s.xiaohongshuService.ListFeeds(ctx)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "获取Feeds列表失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	// 格式化输出，转换为JSON字符串
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("获取Feeds列表成功，但序列化失败: %v", err),
			}},
			IsError: true,
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: string(jsonData),
		}},
	}
}

// handleSearchFeeds 处理搜索Feeds
func (s *AppServer) handleSearchFeeds(ctx context.Context, args SearchFeedsArgs) *MCPToolResult {
	logrus.Info("MCP: 搜索Feeds")

	if args.Keyword == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "搜索Feeds失败: 缺少关键词参数",
			}},
			IsError: true,
		}
	}

	logrus.Infof("MCP: 搜索Feeds - 关键词: %s", args.Keyword)

	// 将 MCP 的 FilterOption 转换为 xiaohongshu.FilterOption
	filter := xiaohongshu.FilterOption{
		SortBy:      args.Filters.SortBy,
		NoteType:    args.Filters.NoteType,
		PublishTime: args.Filters.PublishTime,
		SearchScope: args.Filters.SearchScope,
		Location:    args.Filters.Location,
	}

	result, err := s.xiaohongshuService.SearchFeeds(ctx, args.Keyword, filter)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "搜索Feeds失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	// 格式化输出，转换为JSON字符串
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("搜索Feeds成功，但序列化失败: %v", err),
			}},
			IsError: true,
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: string(jsonData),
		}},
	}
}

// handleGetFeedDetail 处理获取Feed详情
func (s *AppServer) handleGetFeedDetail(ctx context.Context, args map[string]any) *MCPToolResult {
	logrus.Info("MCP: 获取Feed详情")

	// 解析参数
	feedID, ok := args["feed_id"].(string)
	if !ok || feedID == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "获取Feed详情失败: 缺少feed_id参数",
			}},
			IsError: true,
		}
	}

	xsecToken, ok := args["xsec_token"].(string)
	if !ok || xsecToken == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "获取Feed详情失败: 缺少xsec_token参数",
			}},
			IsError: true,
		}
	}

	loadAll := false
	if raw, ok := args["load_all_comments"]; ok {
		switch v := raw.(type) {
		case bool:
			loadAll = v
		case string:
			if parsed, err := strconv.ParseBool(v); err == nil {
				loadAll = parsed
			}
		case float64:
			loadAll = v != 0
		}
	}

	// 解析评论配置参数，如果未提供则使用默认值
	config := xiaohongshu.DefaultCommentLoadConfig()

	if raw, ok := args["click_more_replies"]; ok {
		switch v := raw.(type) {
		case bool:
			config.ClickMoreReplies = v
		case string:
			if parsed, err := strconv.ParseBool(v); err == nil {
				config.ClickMoreReplies = parsed
			}
		}
	}

	if raw, ok := args["max_replies_threshold"]; ok {
		switch v := raw.(type) {
		case float64:
			config.MaxRepliesThreshold = int(v)
		case string:
			if parsed, err := strconv.Atoi(v); err == nil {
				config.MaxRepliesThreshold = parsed
			}
		case int:
			config.MaxRepliesThreshold = v
		}
	}

	if raw, ok := args["max_comment_items"]; ok {
		switch v := raw.(type) {
		case float64:
			config.MaxCommentItems = int(v)
		case string:
			if parsed, err := strconv.Atoi(v); err == nil {
				config.MaxCommentItems = parsed
			}
		case int:
			config.MaxCommentItems = v
		}
	}

	if raw, ok := args["scroll_speed"].(string); ok && raw != "" {
		config.ScrollSpeed = raw
	}

	logrus.Infof("MCP: 获取Feed详情 - Feed ID: %s, loadAllComments=%v, config=%+v", feedID, loadAll, config)

	result, err := s.xiaohongshuService.GetFeedDetailWithConfig(ctx, feedID, xsecToken, loadAll, config)
	if err != nil {
		// 检查是否是笔记不可访问错误
		var notAccessibleErr *apperrors.ErrFeedNotAccessible
		if errors.As(err, &notAccessibleErr) {
			// 笔记不可访问，返回友好提示而不是错误
			return &MCPToolResult{
				Content: []MCPContent{{
					Type: "text",
					Text: fmt.Sprintf("⚠️ 笔记不可访问\n\nFeed ID: %s\n原因: %s\n\n可能的原因：\n- 笔记已被作者删除\n- 笔记因违规被平台删除\n- 笔记设置为私密，仅作者可见\n- 笔记暂时无法访问", feedID, notAccessibleErr.Reason),
				}},
				IsError: false, // 不标记为错误，因为这是预期的业务场景
			}
		}

		// 其他错误正常返回
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "获取Feed详情失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	// 格式化输出，转换为JSON字符串
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("获取Feed详情成功，但序列化失败: %v", err),
			}},
			IsError: true,
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: string(jsonData),
		}},
	}
}

// handleUserProfile 获取用户主页
func (s *AppServer) handleUserProfile(ctx context.Context, args map[string]any) *MCPToolResult {
	logrus.Info("MCP: 获取用户主页")

	// 解析参数
	userID, ok := args["user_id"].(string)
	if !ok || userID == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "获取用户主页失败: 缺少user_id参数",
			}},
			IsError: true,
		}
	}

	xsecToken, ok := args["xsec_token"].(string)
	if !ok || xsecToken == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "获取用户主页失败: 缺少xsec_token参数",
			}},
			IsError: true,
		}
	}

	logrus.Infof("MCP: 获取用户主页 - User ID: %s", userID)

	result, err := s.xiaohongshuService.UserProfile(ctx, userID, xsecToken)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "获取用户主页失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	// 格式化输出，转换为JSON字符串
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("获取用户主页，但序列化失败: %v", err),
			}},
			IsError: true,
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: string(jsonData),
		}},
	}
}

// handleLikeFeed 处理点赞/取消点赞
func (s *AppServer) handleLikeFeed(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	feedID, ok := args["feed_id"].(string)
	if !ok || feedID == "" {
		return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: "操作失败: 缺少feed_id参数"}}, IsError: true}
	}
	xsecToken, ok := args["xsec_token"].(string)
	if !ok || xsecToken == "" {
		return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: "操作失败: 缺少xsec_token参数"}}, IsError: true}
	}
	unlike := parseBool(args["unlike"])

	var res *ActionResult
	var err error

	if unlike {
		res, err = s.xiaohongshuService.UnlikeFeed(ctx, feedID, xsecToken)
	} else {
		res, err = s.xiaohongshuService.LikeFeed(ctx, feedID, xsecToken)
	}

	if err != nil {
		action := "点赞"
		if unlike {
			action = "取消点赞"
		}
		return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: action + "失败: " + err.Error()}}, IsError: true}
	}

	action := "点赞"
	if unlike {
		action = "取消点赞"
	}
	return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("%s成功 - Feed ID: %s", action, res.FeedID)}}}
}

// handleFavoriteFeed 处理收藏/取消收藏
func (s *AppServer) handleFavoriteFeed(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	feedID, ok := args["feed_id"].(string)
	if !ok || feedID == "" {
		return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: "操作失败: 缺少feed_id参数"}}, IsError: true}
	}
	xsecToken, ok := args["xsec_token"].(string)
	if !ok || xsecToken == "" {
		return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: "操作失败: 缺少xsec_token参数"}}, IsError: true}
	}
	unfavorite := parseBool(args["unfavorite"])

	var res *ActionResult
	var err error

	if unfavorite {
		res, err = s.xiaohongshuService.UnfavoriteFeed(ctx, feedID, xsecToken)
	} else {
		res, err = s.xiaohongshuService.FavoriteFeed(ctx, feedID, xsecToken)
	}

	if err != nil {
		action := "收藏"
		if unfavorite {
			action = "取消收藏"
		}
		return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: action + "失败: " + err.Error()}}, IsError: true}
	}

	action := "收藏"
	if unfavorite {
		action = "取消收藏"
	}
	return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("%s成功 - Feed ID: %s", action, res.FeedID)}}}
}

// handlePostComment 处理发表评论到Feed
func (s *AppServer) handlePostComment(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logrus.Info("MCP: 发表评论到Feed")

	// 解析参数
	feedID, ok := args["feed_id"].(string)
	if !ok || feedID == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "发表评论失败: 缺少feed_id参数",
			}},
			IsError: true,
		}
	}

	xsecToken, ok := args["xsec_token"].(string)
	if !ok || xsecToken == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "发表评论失败: 缺少xsec_token参数",
			}},
			IsError: true,
		}
	}

	content, ok := args["content"].(string)
	if !ok || content == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "发表评论失败: 缺少content参数",
			}},
			IsError: true,
		}
	}

	logrus.Infof("MCP: 发表评论 - Feed ID: %s, 内容长度: %d", feedID, len(content))

	// 发表评论
	result, err := s.xiaohongshuService.PostCommentToFeed(ctx, feedID, xsecToken, content)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "发表评论失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	// 返回成功结果，只包含feed_id
	resultText := fmt.Sprintf("评论发表成功 - Feed ID: %s", result.FeedID)
	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: resultText,
		}},
	}
}

// handleReplyComment 处理回复评论
func (s *AppServer) handleReplyComment(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logrus.Info("MCP: 回复评论")

	// 解析参数
	feedID, ok := args["feed_id"].(string)
	if !ok || feedID == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "回复评论失败: 缺少feed_id参数",
			}},
			IsError: true,
		}
	}

	xsecToken, ok := args["xsec_token"].(string)
	if !ok || xsecToken == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "回复评论失败: 缺少xsec_token参数",
			}},
			IsError: true,
		}
	}

	commentID, _ := args["comment_id"].(string)
	userID, _ := args["user_id"].(string)
	if commentID == "" && userID == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "回复评论失败: 缺少comment_id或user_id参数",
			}},
			IsError: true,
		}
	}

	content, ok := args["content"].(string)
	if !ok || content == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "回复评论失败: 缺少content参数",
			}},
			IsError: true,
		}
	}

	logrus.Infof("MCP: 回复评论 - Feed ID: %s, Comment ID: %s, User ID: %s, 内容长度: %d", feedID, commentID, userID, len(content))

	// 回复评论
	result, err := s.xiaohongshuService.ReplyCommentToFeed(ctx, feedID, xsecToken, commentID, userID, content)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "回复评论失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	// 返回成功结果
	responseText := fmt.Sprintf("评论回复成功 - Feed ID: %s, Comment ID: %s, User ID: %s", result.FeedID, result.TargetCommentID, result.TargetUserID)
	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: responseText,
		}},
	}
}

// handleFollowUser 处理关注/取关用户
func (s *AppServer) handleFollowUser(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logrus.Info("MCP: 关注/取关用户")

	// 解析参数
	userID, ok := args["user_id"].(string)
	if !ok || userID == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "操作失败: 缺少user_id参数",
			}},
			IsError: true,
		}
	}

	xsecToken, ok := args["xsec_token"].(string)
	if !ok || xsecToken == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "操作失败: 缺少xsec_token参数",
			}},
			IsError: true,
		}
	}

	unfollow := parseBool(args["unfollow"])

	var res *ActionResult
	var err error

	if unfollow {
		res, err = s.xiaohongshuService.UnfollowUser(ctx, userID, xsecToken)
	} else {
		res, err = s.xiaohongshuService.FollowUser(ctx, userID, xsecToken)
	}

	if err != nil {
		action := "关注"
		if unfollow {
			action = "取关"
		}
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: action + "失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	action := "关注"
	if unfollow {
		action = "取关"
	}
	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: fmt.Sprintf("%s成功 - User ID: %s", action, res.FeedID),
		}},
	}
}

// handleLikeComment 处理评论点赞/取消点赞
func (s *AppServer) handleLikeComment(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logrus.Info("MCP: 评论点赞/取消点赞")

	// 解析参数
	feedID, ok := args["feed_id"].(string)
	if !ok || feedID == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "操作失败: 缺少feed_id参数",
			}},
			IsError: true,
		}
	}

	xsecToken, ok := args["xsec_token"].(string)
	if !ok || xsecToken == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "操作失败: 缺少xsec_token参数",
			}},
			IsError: true,
		}
	}

	commentID, _ := args["comment_id"].(string)
	userID, _ := args["user_id"].(string)
	if commentID == "" && userID == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "操作失败: 缺少comment_id或user_id参数",
			}},
			IsError: true,
		}
	}

	unlike := parseBool(args["unlike"])

	var res *ActionResult
	var err error

	if unlike {
		res, err = s.xiaohongshuService.UnlikeComment(ctx, feedID, xsecToken, commentID, userID)
	} else {
		res, err = s.xiaohongshuService.LikeComment(ctx, feedID, xsecToken, commentID, userID)
	}

	if err != nil {
		action := "点赞评论"
		if unlike {
			action = "取消点赞评论"
		}
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: action + "失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	action := "点赞评论"
	if unlike {
		action = "取消点赞评论"
	}
	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: fmt.Sprintf("%s成功 - Feed ID: %s", action, res.FeedID),
		}},
	}
}

// handleShareFeed 处理分享笔记
func (s *AppServer) handleShareFeed(ctx context.Context, feedID, xsecToken string) *MCPToolResult {
	logrus.Info("MCP: 分享笔记")

	if feedID == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "操作失败: 缺少feed_id参数",
			}},
			IsError: true,
		}
	}

	if xsecToken == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "操作失败: 缺少xsec_token参数",
			}},
			IsError: true,
		}
	}

	shareLink, err := s.xiaohongshuService.ShareFeed(ctx, feedID, xsecToken)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "分享失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: fmt.Sprintf("分享成功 - 分享链接: %s", shareLink),
		}},
	}
}

// handleDeleteFeed 处理删除笔记
func (s *AppServer) handleDeleteFeed(ctx context.Context, feedID, xsecToken string) *MCPToolResult {
	logrus.Info("MCP: 删除笔记")

	if feedID == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "操作失败: 缺少feed_id参数",
			}},
			IsError: true,
		}
	}

	if xsecToken == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "操作失败: 缺少xsec_token参数",
			}},
			IsError: true,
		}
	}

	res, err := s.xiaohongshuService.DeleteFeed(ctx, feedID, xsecToken)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "删除失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: fmt.Sprintf("删除成功 - Feed ID: %s", res.FeedID),
		}},
	}
}

// handleDeleteComment 处理删除评论
func (s *AppServer) handleDeleteComment(ctx context.Context, args map[string]interface{}) *MCPToolResult {
	logrus.Info("MCP: 删除评论")

	feedID, ok := args["feed_id"].(string)
	if !ok || feedID == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "操作失败: 缺少feed_id参数",
			}},
			IsError: true,
		}
	}

	xsecToken, ok := args["xsec_token"].(string)
	if !ok || xsecToken == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "操作失败: 缺少xsec_token参数",
			}},
			IsError: true,
		}
	}

	commentID, _ := args["comment_id"].(string)
	userID, _ := args["user_id"].(string)
	if commentID == "" && userID == "" {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "操作失败: 缺少comment_id或user_id参数",
			}},
			IsError: true,
		}
	}

	res, err := s.xiaohongshuService.DeleteComment(ctx, feedID, xsecToken, commentID, userID)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "删除评论失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: fmt.Sprintf("删除评论成功 - Feed ID: %s", res.FeedID),
		}},
	}
}

// handleGetMyStats 处理获取个人统计数据
func (s *AppServer) handleGetMyStats(ctx context.Context) *MCPToolResult {
	logrus.Info("MCP: 获取个人统计数据")

	stats, err := s.xiaohongshuService.GetMyStats(ctx)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "获取统计数据失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	// 格式化输出
	jsonData, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("获取统计数据成功，但序列化失败: %v", err),
			}},
			IsError: true,
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: string(jsonData),
		}},
	}
}

// handleGetMyFeeds 处理获取自己的笔记列表
func (s *AppServer) handleGetMyFeeds(ctx context.Context, limit int) *MCPToolResult {
	logrus.Infof("MCP: 获取自己的笔记列表，限制: %d", limit)

	feeds, err := s.xiaohongshuService.GetMyFeeds(ctx, limit)
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: "获取笔记列表失败: " + err.Error(),
			}},
			IsError: true,
		}
	}

	// 格式化输出
	jsonData, err := json.MarshalIndent(feeds, "", "  ")
	if err != nil {
		return &MCPToolResult{
			Content: []MCPContent{{
				Type: "text",
				Text: fmt.Sprintf("获取笔记列表成功，但序列化失败: %v", err),
			}},
			IsError: true,
		}
	}

	return &MCPToolResult{
		Content: []MCPContent{{
			Type: "text",
			Text: string(jsonData),
		}},
	}
}

// handleGetFanAnalytics 处理获取粉丝分析请求
func (s *AppServer) handleGetFanAnalytics(ctx context.Context, period string) *MCPToolResult {
	analytics, err := s.xiaohongshuService.GetFanAnalytics(ctx, period)
	if err != nil {
		return &MCPToolResult{
			IsError: true,
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("获取粉丝分析数据失败: %v", err)}},
		}
	}

	data, err := json.Marshal(analytics)
	if err != nil {
		return &MCPToolResult{
			IsError: true,
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("序列化数据失败: %v", err)}},
		}
	}

	return &MCPToolResult{
		IsError: false,
		Content: []MCPContent{{Type: "text", Text: string(data)}},
	}
}

// handleGetContentAnalytics 处理获取内容分析请求
func (s *AppServer) handleGetContentAnalytics(ctx context.Context, limit int, sortBy, sortOrder string) *MCPToolResult {
	// 转换排序参数
	var sortField xiaohongshu.SortField
	var order xiaohongshu.SortOrder

	if sortBy != "" {
		sortField = xiaohongshu.SortField(sortBy)
	}
	if sortOrder != "" {
		order = xiaohongshu.SortOrder(sortOrder)
	} else {
		order = xiaohongshu.SortDesc // 默认降序
	}

	analytics, err := s.xiaohongshuService.GetContentAnalytics(ctx, limit, sortField, order)
	if err != nil {
		return &MCPToolResult{
			IsError: true,
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("获取内容分析数据失败: %v", err)}},
		}
	}

	data, err := json.Marshal(analytics)
	if err != nil {
		return &MCPToolResult{
			IsError: true,
			Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("序列化数据失败: %v", err)}},
		}
	}

	return &MCPToolResult{
		IsError: false,
		Content: []MCPContent{{Type: "text", Text: string(data)}},
	}
}
