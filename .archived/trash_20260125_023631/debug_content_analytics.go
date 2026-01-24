package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/sirupsen/logrus"
)

type NoteData struct {
	Title          string  `json:"title"`
	PublishTime    string  `json:"publish_time"`
	Exposure       int     `json:"exposure"`
	Views          int     `json:"views"`
	ClickRate      float64 `json:"click_rate"`
	Likes          int     `json:"likes"`
	Comments       int     `json:"comments"`
	Collects       int     `json:"collects"`
	Shares         int     `json:"shares"`
	FollowerGrowth int     `json:"follower_growth"`
	Status         string  `json:"status"`
}

type ContentAnalytics struct {
	Notes []NoteData `json:"notes"`
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	// 启动有头浏览器
	fmt.Println("========================================")
	fmt.Println("调试内容分析数据获取")
	fmt.Println("========================================")
	fmt.Println()

	l := launcher.New().
		Headless(false).
		Devtools(true)

	url := l.MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()
	defer browser.MustClose()

	// 加载 cookies
	page := browser.MustPage()
	cookiesData, err := loadCookies()
	if err != nil {
		logrus.Fatalf("加载 cookies 失败: %v", err)
	}

	// 设置 cookies
	for _, cookie := range cookiesData {
		page.MustSetCookies(cookie)
	}

	// 访问内容分析页面
	analyticsURL := "https://creator.xiaohongshu.com/statistics/data-analysis?source=official"
	fmt.Printf("访问页面: %s\n\n", analyticsURL)

	page.MustNavigate(analyticsURL)
	page.MustWaitStable()
	time.Sleep(5 * time.Second)

	fmt.Println("页面已加载，等待观察...")
	fmt.Println()

	// 先尝试提取 DOM 结构信息
	fmt.Println("=== 检查页面结构 ===")
	tableInfo := page.MustEval(`() => {
		const tables = document.querySelectorAll('table');
		const rows = document.querySelectorAll('tr');
		const noteRows = document.querySelectorAll('[class*="note"]');
		
		return {
			tables_count: tables.length,
			tr_count: rows.length,
			note_elements_count: noteRows.length,
			sample_row_html: rows.length > 0 ? rows[0].outerHTML.substring(0, 500) : 'no rows',
			sample_row_text: rows.length > 0 ? rows[0].textContent.substring(0, 200) : 'no text'
		};
	}`).String()

	fmt.Printf("页面结构信息:\n%s\n\n", tableInfo)

	// 提取所有行的文本内容
	fmt.Println("=== 提取所有行文本 ===")
	allRowsText := page.MustEval(`() => {
		const rows = document.querySelectorAll('tr');
		const texts = [];
		rows.forEach((row, index) => {
			if (index < 20) {  // 只显示前20行
				texts.push({
					index: index,
					text: row.textContent.trim().substring(0, 300)
				});
			}
		});
		return JSON.stringify(texts, null, 2);
	}`).String()

	fmt.Printf("前20行内容:\n%s\n\n", allRowsText)

	// 尝试原始的提取逻辑
	fmt.Println("=== 使用原始逻辑提取 (limit=100) ===")
	result := page.MustEval(`(limit) => {
		const notes = [];
		const rows = document.querySelectorAll('tr, .note-row, [class*="note"]');
		
		console.log('Total rows found:', rows.length);

		rows.forEach((row, index) => {
			const text = row.textContent;
			
			// 匹配笔记标题和发布时间
			const titleMatch = text.match(/💥(.+?)发布于(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2})/);
			if (!titleMatch) {
				if (index < 10) {
					console.log('Row', index, 'no match:', text.substring(0, 100));
				}
				return;
			}
			
			console.log('Row', index, 'matched:', titleMatch[0]);

			const title = '💥' + titleMatch[1].trim();
			const publishTime = titleMatch[2];

			// 提取数字数据
			const numbers = text.match(/\d+/g) || [];
			console.log('Row', index, 'numbers:', numbers);
			
			if (numbers.length < 8) {
				console.log('Row', index, 'insufficient numbers:', numbers.length);
				return;
			}

			notes.push({
				title: title,
				publish_time: publishTime,
				exposure: parseInt(numbers[2]) || 0,
				views: parseInt(numbers[3]) || 0,
				click_rate: parseFloat(numbers[4]) || 0,
				likes: parseInt(numbers[5]) || 0,
				comments: parseInt(numbers[6]) || 0,
				collects: parseInt(numbers[7]) || 0,
				shares: parseInt(numbers[8]) || 0,
				follower_growth: parseInt(numbers[9]) || 0,
				status: text.includes('违规') ? 'violation' : 'normal'
			});

			if (notes.length >= limit) return;
		});

		return JSON.stringify({notes: notes.slice(0, limit)}, null, 2);
	}`, 100).String()

	fmt.Printf("提取结果:\n%s\n\n", result)

	var analytics ContentAnalytics
	if err := json.Unmarshal([]byte(result), &analytics); err != nil {
		logrus.Errorf("解析结果失败: %v", err)
	} else {
		fmt.Printf("✅ 成功提取 %d 条笔记\n\n", len(analytics.Notes))
	}

	// 保持浏览器打开，方便手动检查
	fmt.Println("========================================")
	fmt.Println("浏览器保持打开，按 Ctrl+C 退出")
	fmt.Println("请手动检查页面，看看是否需要滚动加载")
	fmt.Println("========================================")

	select {}
}

func loadCookies() ([]*rod.CookiesFromJSON, error) {
	// 这里简化处理，实际应该从 cookies.json 加载
	return nil, nil
}
