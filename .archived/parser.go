package xiaohongshu

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

// HTMLParser 专业的HTML解析器，使用goquery进行DOM解析
type HTMLParser struct {
	doc *goquery.Document
}

// NewHTMLParser 创建HTML解析器
func NewHTMLParser(html string) (*HTMLParser, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}
	return &HTMLParser{doc: doc}, nil
}

// ParseUserList 解析用户列表（粉丝/关注）
// 支持多种可能的DOM结构
func (p *HTMLParser) ParseUserList() []FollowerUser {
	var users []FollowerUser

	// 先输出调试信息
	text := p.doc.Text()
	logrus.Debugf("页面文本长度: %d", len(text))
	logrus.Debugf("页面文本前200字符: %s", text[:min(len(text), 200)])

	// 输出所有可能的class名称
	p.doc.Find("div, li, article").Each(func(i int, s *goquery.Selection) {
		if i < 10 { // 只输出前10个
			if class, exists := s.Attr("class"); exists && class != "" {
				logrus.Debugf("元素 %d class: %s", i, class)
			}
		}
	})

	// 尝试多种选择器策略
	selectors := []string{
		".user-card",           // 标准用户卡片
		"[class*='user-item']", // 包含user-item的类名
		"[class*='UserCard']",  // React组件命名
		"[class*='follower']",  // 粉丝相关
		"li[class*='user']",    // 列表项
	}

	for _, selector := range selectors {
		p.doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			user := p.extractUser(s)
			if user.Nickname != "" || user.UserID != "" {
				users = append(users, user)
			}
		})

		if len(users) > 0 {
			logrus.Infof("使用选择器 '%s' 找到 %d 个用户", selector, len(users))
			break
		}
	}

	// 如果所有选择器都失败，尝试通过文本模式匹配
	if len(users) == 0 {
		logrus.Warn("所有选择器失败，尝试文本模式匹配")
		users = p.parseUserListByText()
	}

	return users
}

// extractUser 从DOM节点提取用户信息
func (p *HTMLParser) extractUser(s *goquery.Selection) FollowerUser {
	user := FollowerUser{}

	// 提取用户ID
	if id, exists := s.Attr("data-user-id"); exists {
		user.UserID = id
	} else if id, exists := s.Attr("data-id"); exists {
		user.UserID = id
	}

	// 提取昵称 - 尝试多种选择器
	nicknameSelectors := []string{
		".nickname",
		"[class*='name']",
		"[class*='Name']",
		".user-name",
		"h3", "h4", // 标题标签
	}
	for _, sel := range nicknameSelectors {
		if text := s.Find(sel).First().Text(); text != "" {
			user.Nickname = strings.TrimSpace(text)
			break
		}
	}

	// 提取头像
	if avatar, exists := s.Find("img").First().Attr("src"); exists {
		user.Avatar = avatar
	}

	// 提取描述
	descSelectors := []string{
		".desc",
		"[class*='desc']",
		"[class*='Desc']",
		".bio",
		"p",
	}
	for _, sel := range descSelectors {
		if text := s.Find(sel).First().Text(); text != "" {
			user.Desc = strings.TrimSpace(text)
			break
		}
	}

	return user
}

// parseUserListByText 通过文本模式匹配解析用户列表（备用方案）
func (p *HTMLParser) parseUserListByText() []FollowerUser {
	// 这是一个备用方案，当DOM结构无法识别时使用
	// 可以根据实际页面文本格式进行正则匹配
	return []FollowerUser{}
}

// ParseFeedList 解析笔记列表
func (p *HTMLParser) ParseFeedList() []Feed {
	var feeds []Feed

	selectors := []string{
		".note-card",
		"[class*='note-item']",
		"[class*='NoteCard']",
		"[class*='feed-item']",
		"article",
	}

	for _, selector := range selectors {
		p.doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			feed := p.extractFeed(s)
			if feed.ID != "" || feed.NoteCard.DisplayTitle != "" {
				feeds = append(feeds, feed)
			}
		})

		if len(feeds) > 0 {
			logrus.Infof("使用选择器 '%s' 找到 %d 条笔记", selector, len(feeds))
			break
		}
	}

	return feeds
}

// extractFeed 从DOM节点提取笔记信息
func (p *HTMLParser) extractFeed(s *goquery.Selection) Feed {
	feed := Feed{}

	// 提取笔记ID
	if id, exists := s.Attr("data-note-id"); exists {
		feed.ID = id
	} else if id, exists := s.Attr("data-id"); exists {
		feed.ID = id
	}

	// 提取标题
	titleSelectors := []string{
		".title",
		"[class*='title']",
		"[class*='Title']",
		"h3", "h4",
	}
	for _, sel := range titleSelectors {
		if text := s.Find(sel).First().Text(); text != "" {
			feed.NoteCard.DisplayTitle = strings.TrimSpace(text)
			break
		}
	}

	// 提取封面
	if cover, exists := s.Find("img").First().Attr("src"); exists {
		feed.NoteCard.Cover.URLDefault = cover
	}

	// 提取链接并从中提取ID（如果没有data-note-id）
	if link, exists := s.Find("a").First().Attr("href"); exists {
		if feed.ID == "" && strings.Contains(link, "/") {
			parts := strings.Split(link, "/")
			if len(parts) > 0 {
				feed.ID = parts[len(parts)-1]
			}
		}
	}

	return feed
}

// ParseCreatorStats 解析创作者统计数据
func (p *HTMLParser) ParseCreatorStats() *UserStats {
	stats := &UserStats{}
	text := p.doc.Text()

	// 使用正则表达式提取数字
	// 这里可以添加更复杂的解析逻辑
	logrus.Debug("解析创作者统计数据:", text[:min(len(text), 200)])

	return stats
}

// ParseFanAnalytics 解析粉丝分析数据
func (p *HTMLParser) ParseFanAnalytics() *FanAnalytics {
	analytics := &FanAnalytics{
		Overview:     FanOverview{},
		Demographics: FanDemographics{Gender: make(map[string]int), Interests: []string{}},
		ActiveFans:   []ActiveFan{},
	}

	text := p.doc.Text()

	// 提取活跃粉丝 - 使用更精确的选择器
	// 查找包含"互动"和数字的元素
	p.doc.Find("li, div, tr").Each(func(i int, s *goquery.Selection) {
		itemText := s.Text()
		// 匹配格式: "昵称 互动 X 次"
		if strings.Contains(itemText, "互动") && strings.Contains(itemText, "次") {
			// 提取昵称和互动次数
			parts := strings.Fields(itemText)
			if len(parts) >= 3 {
				nickname := ""
				interactions := 0

				// 查找昵称（第一个非数字字段）
				for _, part := range parts {
					if !strings.Contains(part, "互动") && !strings.Contains(part, "次") {
						// 尝试解析为数字
						if num := parseInt(part); num > 0 {
							interactions = num
						} else if nickname == "" {
							nickname = part
						}
					}
				}

				if nickname != "" && interactions > 0 {
					// 过滤掉包含页面标题的噪音数据
					if !strings.Contains(nickname, "粉丝数据") && !strings.Contains(nickname, "我的活跃粉丝") {
						analytics.ActiveFans = append(analytics.ActiveFans, ActiveFan{
							Nickname:     nickname,
							Interactions: interactions,
						})
					}
				}
			}
		}
	})

	logrus.Infof("解析到 %d 个活跃粉丝", len(analytics.ActiveFans))
	logrus.Debug("页面文本:", text[:min(len(text), 500)])

	return analytics
}

// ParseContentAnalytics 解析内容分析数据
func (p *HTMLParser) ParseContentAnalytics() *ContentAnalytics {
	analytics := &ContentAnalytics{Notes: []NoteMetrics{}}

	// 查找表格行或笔记卡片
	p.doc.Find("tr, [class*='note-row'], [class*='data-row']").Each(func(i int, s *goquery.Selection) {
		rowText := s.Text()

		// 跳过表头
		if strings.Contains(rowText, "标题") || strings.Contains(rowText, "发布时间") {
			return
		}

		// 提取笔记数据
		note := NoteMetrics{}

		// 提取标题（通常包含emoji）
		if strings.Contains(rowText, "💥") {
			// 找到标题部分
			titleStart := strings.Index(rowText, "💥")
			if titleStart >= 0 {
				titleEnd := strings.Index(rowText[titleStart:], "发布于")
				if titleEnd > 0 {
					note.Title = strings.TrimSpace(rowText[titleStart : titleStart+titleEnd])
				}
			}
		}

		// 提取发布时间
		if strings.Contains(rowText, "发布于") {
			// 匹配时间格式: 2026-01-21 00:32
			// 简化处理：查找"发布于"后的时间字符串
		}

		// 提取数字指标
		// 这里需要更精确的解析逻辑

		if note.Title != "" {
			analytics.Notes = append(analytics.Notes, note)
		}
	})

	return analytics
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func parseInt(s string) int {
	var num int
	for _, r := range s {
		if r >= '0' && r <= '9' {
			num = num*10 + int(r-'0')
		}
	}
	return num
}
