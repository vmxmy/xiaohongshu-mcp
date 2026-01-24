package xiaohongshu

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/sirupsen/logrus"
)

// DataAction 数据获取操作
type DataAction struct {
	page *rod.Page
}

// NewDataAction 创建数据获取操作实例
func NewDataAction(page *rod.Page) *DataAction {
	return &DataAction{page: page}
}

// UserStats 用户统计数据
type UserStats struct {
	// 基础数据
	FollowerCount int `json:"follower_count"` // 粉丝数
	FollowCount   int `json:"follow_count"`   // 关注数
	LikedCount    int `json:"liked_count"`    // 获赞与收藏
	NoteCount     int `json:"note_count"`     // 笔记数
	CollectCount  int `json:"collect_count"`  // 收藏数

	// 创作者中心数据（近7日）
	ExposureCount       int     `json:"exposure_count,omitempty"`        // 曝光数
	ViewCount           int     `json:"view_count,omitempty"`            // 观看数
	CoverClickRate      float64 `json:"cover_click_rate,omitempty"`      // 封面点击率
	VideoCompleteRate   float64 `json:"video_complete_rate,omitempty"`   // 视频完播率
	LikeCount7d         int     `json:"like_count_7d,omitempty"`         // 点赞数（7日）
	CommentCount7d      int     `json:"comment_count_7d,omitempty"`      // 评论数（7日）
	CollectCount7d      int     `json:"collect_count_7d,omitempty"`      // 收藏数（7日）
	ShareCount7d        int     `json:"share_count_7d,omitempty"`        // 分享数（7日）
	NetFollowerGrowth   int     `json:"net_follower_growth,omitempty"`   // 净涨粉
	NewFollowerCount    int     `json:"new_follower_count,omitempty"`    // 新增关注
	UnfollowCount       int     `json:"unfollow_count,omitempty"`        // 取消关注
	ProfileVisitorCount int     `json:"profile_visitor_count,omitempty"` // 主页访客
}

// FanAnalytics 粉丝分析数据
type FanAnalytics struct {
	Overview     FanOverview     `json:"overview"`
	Demographics FanDemographics `json:"demographics"`
	ActiveFans   []ActiveFan     `json:"active_fans"`
}

// FanOverview 粉丝概览
type FanOverview struct {
	TotalFans int `json:"total_fans"` // 总粉丝数
	NewFans   int `json:"new_fans"`   // 新增粉丝数
	LostFans  int `json:"lost_fans"`  // 流失粉丝数
}

// FanDemographics 粉丝画像
type FanDemographics struct {
	Gender    map[string]int `json:"gender"`    // 性别分布 {"male": 59, "female": 41}
	Interests []string       `json:"interests"` // 兴趣分布
}

// ActiveFan 活跃粉丝
type ActiveFan struct {
	Nickname     string `json:"nickname"`     // 昵称
	Interactions int    `json:"interactions"` // 互动次数
}

// ContentAnalytics 内容分析数据
type ContentAnalytics struct {
	Notes []NoteMetrics `json:"notes"`
}

// NoteMetrics 笔记指标
type NoteMetrics struct {
	Title           string  `json:"title"`             // 标题
	PublishTime     string  `json:"publish_time"`      // 发布时间
	Exposure        int     `json:"exposure"`          // 曝光数
	Views           int     `json:"views"`             // 观看数
	ClickRate       float64 `json:"click_rate"`        // 点击率
	Likes           int     `json:"likes"`             // 点赞数
	Comments        int     `json:"comments"`          // 评论数
	Collects        int     `json:"collects"`          // 收藏数
	FollowerGrowth  int     `json:"follower_growth"`   // 涨粉数
	Shares          int     `json:"shares"`            // 分享数
	AvgViewDuration string  `json:"avg_view_duration"` // 人均观看时长
	FullScreen      int     `json:"full_screen"`       // 满屏
	Status          string  `json:"status"`            // 状态
}

// FollowerUser 粉丝/关注用户信息
type FollowerUser struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Desc     string `json:"desc"`
}

// GetMyStats 获取当前用户的统计数据
func (d *DataAction) GetMyStats(ctx context.Context) (*UserStats, error) {
	page := d.page.Context(ctx).Timeout(60 * time.Second)

	// 导航到创作者中心页面（包含更详细的运营数据）
	logrus.Info("导航到创作者中心页面...")
	page.MustNavigate("https://creator.xiaohongshu.com/new/home?source=official")
	page.MustWaitDOMStable()
	time.Sleep(5 * time.Second) // 等待页面加载和数据渲染

	// 从创作者中心页面提取详细统计数据
	result := page.MustEval(`() => {
		const stats = {
			follower_count: 0,
			follow_count: 0,
			liked_count: 0,
			note_count: 0,
			collect_count: 0,
			exposure_count: 0,
			view_count: 0,
			cover_click_rate: 0,
			video_complete_rate: 0,
			like_count_7d: 0,
			comment_count_7d: 0,
			collect_count_7d: 0,
			share_count_7d: 0,
			net_follower_growth: 0,
			new_follower_count: 0,
			unfollow_count: 0,
			profile_visitor_count: 0
		};

		// 提取基础数据（关注数、粉丝数、获赞与收藏）
		const allText = document.body.innerText;

		// 匹配 "34关注数96粉丝数785获赞与收藏" 格式
		const basicMatch = allText.match(/(\d+)关注数(\d+)粉丝数(\d+)获赞与收藏/);
		if (basicMatch) {
			stats.follow_count = parseInt(basicMatch[1]);
			stats.follower_count = parseInt(basicMatch[2]);
			stats.liked_count = parseInt(basicMatch[3]);
		}

		// 提取笔记数据总览（近7日）
		// 匹配格式: "曝光数10.5万" "观看数1.1万" "点赞数375"
		const parseNumber = (text) => {
			if (!text) return 0;
			text = text.replace(/,/g, '');
			if (text.includes('万')) {
				return Math.round(parseFloat(text.replace('万', '')) * 10000);
			} else if (text.includes('亿')) {
				return Math.round(parseFloat(text.replace('亿', '')) * 100000000);
			}
			return parseInt(text) || 0;
		};

		const parsePercent = (text) => {
			if (!text) return 0;
			return parseFloat(text.replace('%', '')) || 0;
		};

		// 曝光数
		const exposureMatch = allText.match(/曝光数([\d.]+[万亿]?)/);
		if (exposureMatch) stats.exposure_count = parseNumber(exposureMatch[1]);

		// 观看数
		const viewMatch = allText.match(/观看数([\d.]+[万亿]?)/);
		if (viewMatch) stats.view_count = parseNumber(viewMatch[1]);

		// 封面点击率
		const clickRateMatch = allText.match(/封面点击率([\d.]+)%/);
		if (clickRateMatch) stats.cover_click_rate = parsePercent(clickRateMatch[1]);

		// 视频完播率
		const completeRateMatch = allText.match(/视频完播率([\d.]+)%/);
		if (completeRateMatch) stats.video_complete_rate = parsePercent(completeRateMatch[1]);

		// 点赞数
		const likeMatch = allText.match(/点赞数([\d.]+[万亿]?)/);
		if (likeMatch) stats.like_count_7d = parseNumber(likeMatch[1]);

		// 评论数
		const commentMatch = allText.match(/评论数([\d.]+[万亿]?)/);
		if (commentMatch) stats.comment_count_7d = parseNumber(commentMatch[1]);

		// 收藏数
		const collectMatch = allText.match(/收藏数([\d.]+[万亿]?)/);
		if (collectMatch) stats.collect_count_7d = parseNumber(collectMatch[1]);

		// 分享数
		const shareMatch = allText.match(/分享数([\d.]+[万亿]?)/);
		if (shareMatch) stats.share_count_7d = parseNumber(shareMatch[1]);

		// 净涨粉
		const netGrowthMatch = allText.match(/净涨粉([\d.]+[万亿]?)/);
		if (netGrowthMatch) stats.net_follower_growth = parseNumber(netGrowthMatch[1]);

		// 新增关注
		const newFollowerMatch = allText.match(/新增关注([\d.]+[万亿]?)/);
		if (newFollowerMatch) stats.new_follower_count = parseNumber(newFollowerMatch[1]);

		// 取消关注
		const unfollowMatch = allText.match(/取消关注([\d.]+[万亿]?)/);
		if (unfollowMatch) stats.unfollow_count = parseNumber(unfollowMatch[1]);

		// 主页访客
		const visitorMatch = allText.match(/主页访客([\d.]+[万亿]?)/);
		if (visitorMatch) stats.profile_visitor_count = parseNumber(visitorMatch[1]);

		return JSON.stringify(stats);
	}`).String()

	if result == "" {
		return nil, fmt.Errorf("无法获取统计数据")
	}

	var stats UserStats
	if err := json.Unmarshal([]byte(result), &stats); err != nil {
		return nil, fmt.Errorf("解析统计数据失败: %w", err)
	}

	logrus.Infof("获取统计数据成功: %+v", stats)
	return &stats, nil
}

// GetMyFeeds 获取自己发布的笔记列表
func (d *DataAction) GetMyFeeds(ctx context.Context, limit int) ([]Feed, error) {
	page := d.page.Context(ctx).Timeout(5 * time.Minute)

	// 通过侧边栏导航到个人主页
	logrus.Info("通过侧边栏导航到个人主页获取笔记...")
	navigate := NewNavigate(page)
	if err := navigate.ToProfilePage(ctx); err != nil {
		return nil, fmt.Errorf("导航到个人主页失败: %w", err)
	}
	page.MustWaitDOMStable()
	time.Sleep(3 * time.Second)

	// 使用JavaScript提取笔记列表
	feeds := d.extractFeedsFromPage(page, limit)

	logrus.Infof("获取笔记列表成功，共 %d 条", len(feeds))
	return feeds, nil
}

// extractFeedsFromPage 从页面提取笔记列表
func (d *DataAction) extractFeedsFromPage(page *rod.Page, limit int) []Feed {
	var feeds []Feed
	lastCount := 0
	stagnantChecks := 0
	maxAttempts := 20

	for attempt := 0; attempt < maxAttempts && len(feeds) < limit; attempt++ {
		// 使用JavaScript提取笔记
		result := page.MustEval(`(limit) => {
			const notes = [];
			const seen = new Set();

			// 查找所有笔记链接 (包含xsec_token的链接)
			document.querySelectorAll('a[href*="xsec_token"]').forEach(link => {
				const href = link.getAttribute('href');
				// 匹配用户主页下的笔记链接
				const match = href.match(/\/user\/profile\/(\w+)\/(\w+)\?/);
				if (match && !seen.has(match[2])) {
					seen.add(match[2]);
					const [_, userId, noteId] = match;

					// 提取标题 - 查找链接内的文本
					let title = '';
					const titleEl = link.querySelector('span, div');
					if (titleEl) {
						title = titleEl.textContent.trim();
					}

					// 提取封面图
					let cover = '';
					const img = link.querySelector('img');
					if (img) {
						cover = img.src || '';
					}

					// 提取xsec_token
					const tokenMatch = href.match(/xsec_token=([^&]+)/);
					const xsecToken = tokenMatch ? tokenMatch[1] : '';

					if (notes.length < limit) {
						notes.push({
							id: noteId,
							user_id: userId,
							title: title,
							cover: cover,
							xsec_token: xsecToken
						});
					}
				}
			});

			return JSON.stringify(notes);
		}`, limit).String()

		var extractedNotes []struct {
			ID        string `json:"id"`
			UserID    string `json:"user_id"`
			Title     string `json:"title"`
			Cover     string `json:"cover"`
			XsecToken string `json:"xsec_token"`
		}

		if err := json.Unmarshal([]byte(result), &extractedNotes); err != nil {
			logrus.WithError(err).Error("解析笔记数据失败")
			break
		}

		// 转换为Feed结构
		feeds = make([]Feed, 0, len(extractedNotes))
		for _, note := range extractedNotes {
			feed := Feed{
				ID:        note.ID,
				XsecToken: note.XsecToken,
			}
			feed.NoteCard.DisplayTitle = note.Title
			feed.NoteCard.Cover.URLDefault = note.Cover
			feed.NoteCard.User.UserID = note.UserID
			feeds = append(feeds, feed)
		}

		currentCount := len(feeds)
		if currentCount != lastCount {
			logrus.Infof("加载笔记: %d -> %d", lastCount, currentCount)
			lastCount = currentCount
			stagnantChecks = 0
		} else {
			stagnantChecks++
			if stagnantChecks >= 3 {
				logrus.Info("笔记列表停滞，停止加载")
				break
			}
		}

		if len(feeds) >= limit {
			break
		}

		// 滚动到底部加载更多
		page.MustEval(`() => { window.scrollBy(0, window.innerHeight); }`)
		time.Sleep(1 * time.Second)
	}

	// 限制返回数量
	if len(feeds) > limit {
		feeds = feeds[:limit]
	}

	return feeds
}

// GetFanAnalytics 获取粉丝分析数据
func (d *DataAction) GetFanAnalytics(ctx context.Context, period string) (*FanAnalytics, error) {
	page := d.page.Context(ctx).Timeout(5 * time.Minute)

	// 导航到粉丝数据页面
	logrus.Info("导航到粉丝数据页面...")
	url := "https://creator.xiaohongshu.com/statistics/fans-data?source=official"
	page.MustNavigate(url)
	page.MustWaitDOMStable()
	time.Sleep(5 * time.Second)

	// 提取粉丝分析数据
	result := page.MustEval(`() => {
		const data = {
			overview: {total_fans: 0, new_fans: 0, lost_fans: 0},
			demographics: {gender: {}, interests: []},
			active_fans: []
		};

		const text = document.body.innerText;

		// 提取粉丝概览数据
		const totalMatch = text.match(/总粉丝数\s*(\d+)/);
		if (totalMatch) data.overview.total_fans = parseInt(totalMatch[1]);

		const newMatch = text.match(/新增粉丝数\s*(\d+)/);
		if (newMatch) data.overview.new_fans = parseInt(newMatch[1]);

		const lostMatch = text.match(/流失粉丝数\s*(\d+)/);
		if (lostMatch) data.overview.lost_fans = parseInt(lostMatch[1]);

		// 提取性别分布
		const maleMatch = text.match(/男性\s*(\d+)%/);
		const femaleMatch = text.match(/女性\s*(\d+)%/);
		if (maleMatch) data.demographics.gender.male = parseInt(maleMatch[1]);
		if (femaleMatch) data.demographics.gender.female = parseInt(femaleMatch[1]);

		// 提取兴趣分布
		const interestKeywords = ['美食', '生活记录', '社科', '娱乐', '家居家装', '影视', '科技数码', '职场'];
		interestKeywords.forEach(keyword => {
			if (text.includes(keyword)) {
				data.demographics.interests.push(keyword);
			}
		});

		// 提取活跃粉丝列表
		const fanItems = document.querySelectorAll('li, .fan-item, [class*="fan"]');
		fanItems.forEach(item => {
			const itemText = item.textContent;
			const match = itemText.match(/(.+?)\s*互动\s*(\d+)\s*次/);
			if (match && data.active_fans.length < 10) {
				data.active_fans.push({
					nickname: match[1].trim(),
					interactions: parseInt(match[2])
				});
			}
		});

		return JSON.stringify(data);
	}`).String()

	var analytics FanAnalytics
	if err := json.Unmarshal([]byte(result), &analytics); err != nil {
		return nil, fmt.Errorf("解析粉丝分析数据失败: %w", err)
	}

	logrus.Infof("获取粉丝分析数据成功")
	return &analytics, nil
}

// GetContentAnalytics 获取内容分析数据
func (d *DataAction) GetContentAnalytics(ctx context.Context, limit int) (*ContentAnalytics, error) {
	page := d.page.Context(ctx).Timeout(5 * time.Minute)

	// 导航到数据分析页面
	logrus.Info("导航到数据分析页面...")
	url := "https://creator.xiaohongshu.com/statistics/data-analysis?source=official"

	// 设置请求拦截，捕获API响应
	capturedData := make(map[string]map[string]interface{}) // 用map去重，key为note id
	var captureMutex sync.Mutex

	router := page.HijackRequests()
	router.MustAdd("*/api/galaxy/creator/datacenter/note/analyze/list*", func(ctx *rod.Hijack) {
		// 记录请求URL（调试用）
		logrus.Debugf("拦截到请求: %s", ctx.Request.URL().String())

		// 让请求正常发出
		ctx.MustLoadResponse()

		// 记录响应状态
		statusCode := ctx.Response.Payload().ResponseCode
		logrus.Debugf("响应状态: %d", statusCode)

		// 如果响应成功，解析数据
		if statusCode == 200 {
			body := ctx.Response.Body()

			var apiResp struct {
				Success bool `json:"success"`
				Data    struct {
					NoteInfos []map[string]interface{} `json:"note_infos"`
					Total     int                      `json:"total"`
				} `json:"data"`
			}

			if err := json.Unmarshal([]byte(body), &apiResp); err == nil && apiResp.Success {
				captureMutex.Lock()
				for _, note := range apiResp.Data.NoteInfos {
					// 使用note id去重
					if id, ok := note["id"].(string); ok {
						capturedData[id] = note
					}
				}
				logrus.Infof("已捕获 %d 条笔记数据（去重后共 %d 条，总共 %d 条）", len(apiResp.Data.NoteInfos), len(capturedData), apiResp.Data.Total)
				captureMutex.Unlock()
			}
		} else {
			logrus.Warnf("API请求失败，状态码: %d", statusCode)
		}
	})
	go router.Run()
	defer router.MustStop()

	page.MustNavigate(url)
	page.MustWaitDOMStable()

	// 等待初始API请求完成
	time.Sleep(3 * time.Second)

	logrus.Infof("初始加载完成，已捕获 %d 条笔记", len(capturedData))

	// 如果需要更多数据，触发翻页
	for len(capturedData) < limit {
		// 计算还需要多少条
		needed := limit - len(capturedData)
		if needed <= 0 {
			break
		}

		logrus.Infof("需要更多数据（当前 %d 条，目标 %d 条），尝试点击下一页", len(capturedData), limit)

		// 查找分页容器中的所有按钮
		pageButtons := page.MustElements(".pagination .d-pagination-page")
		if len(pageButtons) == 0 {
			logrus.Info("没有找到分页按钮")
			break
		}

		// 最后一个按钮是"下一页"
		nextButton := pageButtons[len(pageButtons)-1]

		// 检查是否被禁用
		classes := nextButton.MustProperty("className").String()
		if contains(classes, "disabled") {
			logrus.Info("下一页按钮已禁用，已到达数据末尾")
			break
		}

		// 记录点击前的数据量
		beforeCount := len(capturedData)

		// 点击下一页
		logrus.Info("点击下一页按钮...")
		nextButton.MustClick()

		// 等待API请求完成
		time.Sleep(2 * time.Second)

		// 如果捕获的数据没有增加，说明已经到底了
		if len(capturedData) == beforeCount {
			logrus.Info("数据未增加，停止翻页")
			break
		}
	}

	logrus.Infof("数据获取完成，共 %d 条笔记", len(capturedData))

	// 转换为我们的数据格式
	var notes []NoteMetrics
	for _, noteData := range capturedData {
		if len(notes) >= limit {
			break
		}
		note := convertNoteData(noteData)
		notes = append(notes, note)
	}

	logrus.Infof("获取内容分析数据成功，共 %d 条笔记", len(notes))
	return &ContentAnalytics{Notes: notes}, nil
}

// convertNoteData 将API返回的笔记数据转换为NoteMetrics格式
func convertNoteData(data map[string]interface{}) NoteMetrics {
	// 安全地获取各个字段
	getInt := func(key string) int {
		if v, ok := data[key].(float64); ok {
			return int(v)
		}
		return 0
	}

	getFloat := func(key string) float64 {
		if v, ok := data[key].(float64); ok {
			return v
		}
		return 0
	}

	getString := func(key string) string {
		if v, ok := data[key].(string); ok {
			return v
		}
		return ""
	}

	// 转换发布时间
	var publishTime string
	if postTime, ok := data["post_time"].(float64); ok {
		t := time.Unix(int64(postTime/1000), 0)
		publishTime = t.Format("2006-01-02 15:04")
	}

	// 转换人均观看时长
	avgDuration := ""
	if viewTimeAvg := getInt("view_time_avg"); viewTimeAvg > 0 {
		avgDuration = fmt.Sprintf("%ds", viewTimeAvg)
	}

	// 转换点击率为百分比
	clickRate := getFloat("coverClickRate") * 100

	return NoteMetrics{
		Title:           getString("title"),
		PublishTime:     publishTime,
		Exposure:        getInt("imp_count"),
		Views:           getInt("read_count"),
		ClickRate:       clickRate,
		Likes:           getInt("like_count"),
		Comments:        getInt("comment_count"),
		Collects:        getInt("fav_count"),
		FollowerGrowth:  getInt("increase_fans_count"),
		Shares:          getInt("share_count"),
		AvgViewDuration: avgDuration,
		FullScreen:      getInt("danmaku_count"),
		Status:          "normal", // 默认正常状态
	}
}

// truncate 截断字符串到指定长度
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
