package main

import (
	"fmt"
	"time"

	"github.com/xpzouying/xiaohongshu-mcp/browser"
)

func main() {
	b := browser.NewBrowser(false)
	defer b.Close()

	page := b.NewPage()

	fmt.Println("=== 小红书页面功能分析 ===\n")

	// 1. 分析首页
	fmt.Println("1. 分析首页功能...")
	page.MustNavigate("https://www.xiaohongshu.com/explore")
	page.MustWaitDOMStable()
	time.Sleep(3 * time.Second)

	homeFeatures := page.MustEval(`() => {
		const features = {
			navigation: [],
			interactions: [],
			filters: [],
			userActions: []
		};

		// 导航元素
		document.querySelectorAll('nav a, [class*="nav"] a').forEach(el => {
			features.navigation.push({
				text: el.textContent.trim(),
				href: el.href,
				class: el.className
			});
		});

		// 互动按钮
		document.querySelectorAll('button, [role="button"]').forEach(el => {
			const text = el.textContent.trim();
			if (text && text.length < 20) {
				features.interactions.push({
					text: text,
					class: el.className,
					id: el.id
				});
			}
		});

		// 筛选器
		document.querySelectorAll('[class*="filter"], [class*="tab"]').forEach(el => {
			features.filters.push({
				text: el.textContent.trim().substring(0, 30),
				class: el.className
			});
		});

		return JSON.stringify(features, null, 2);
	}`).String()

	fmt.Println("首页功能:")
	fmt.Println(homeFeatures)

	// 2. 分析帖子详情页
	fmt.Println("\n2. 分析帖子详情页功能...")
	page.MustNavigate("https://www.xiaohongshu.com/explore/696872e7000000001a032e94?xsec_token=AB2AHlINujEqPed4q1rioho7X4rTy_TvlNRbc4EwiuO0k=&xsec_source=pc_feed")
	page.MustWaitDOMStable()
	time.Sleep(3 * time.Second)

	detailFeatures := page.MustEval(`() => {
		const features = {
			interactions: [],
			userInfo: {},
			contentInfo: {},
			comments: {
				actions: [],
				features: []
			},
			sharing: []
		};

		// 互动按钮
		document.querySelectorAll('[class*="interact"], [class*="action"]').forEach(el => {
			const text = el.textContent.trim();
			if (text && text.length < 30) {
				features.interactions.push({
					text: text,
					class: el.className,
					tag: el.tagName
				});
			}
		});

		// 用户信息区域
		const userArea = document.querySelector('[class*="user"], [class*="author"]');
		if (userArea) {
			features.userInfo = {
				hasFollowButton: !!userArea.querySelector('[class*="follow"]'),
				hasAvatar: !!userArea.querySelector('img'),
				hasUsername: !!userArea.querySelector('[class*="name"]')
			};
		}

		// 评论区功能
		document.querySelectorAll('[class*="comment"]').forEach(el => {
			const buttons = el.querySelectorAll('button');
			buttons.forEach(btn => {
				const text = btn.textContent.trim();
				if (text && !features.comments.actions.includes(text)) {
					features.comments.actions.push(text);
				}
			});
		});

		// 分享功能
		document.querySelectorAll('[class*="share"]').forEach(el => {
			features.sharing.push({
				text: el.textContent.trim().substring(0, 20),
				class: el.className
			});
		});

		return JSON.stringify(features, null, 2);
	}`).String()

	fmt.Println("详情页功能:")
	fmt.Println(detailFeatures)

	// 3. 分析用户主页
	fmt.Println("\n3. 分析用户主页功能...")
	page.MustNavigate("https://www.xiaohongshu.com/user/profile/5e5e19e7000000000100373e")
	page.MustWaitDOMStable()
	time.Sleep(3 * time.Second)

	profileFeatures := page.MustEval(`() => {
		const features = {
			tabs: [],
			actions: [],
			stats: [],
			content: []
		};

		// 标签页
		document.querySelectorAll('[role="tab"], [class*="tab"]').forEach(el => {
			features.tabs.push(el.textContent.trim());
		});

		// 操作按钮
		document.querySelectorAll('button').forEach(el => {
			const text = el.textContent.trim();
			if (text && text.length < 20 && text.length > 0) {
				features.actions.push(text);
			}
		});

		// 统计信息
		document.querySelectorAll('[class*="stat"], [class*="count"]').forEach(el => {
			const text = el.textContent.trim();
			if (text && text.length < 30) {
				features.stats.push(text);
			}
		});

		return JSON.stringify(features, null, 2);
	}`).String()

	fmt.Println("用户主页功能:")
	fmt.Println(profileFeatures)

	// 4. 分析搜索页
	fmt.Println("\n4. 分析搜索功能...")
	page.MustNavigate("https://www.xiaohongshu.com/search_result?keyword=咖啡")
	page.MustWaitDOMStable()
	time.Sleep(3 * time.Second)

	searchFeatures := page.MustEval(`() => {
		const features = {
			filters: [],
			sortOptions: [],
			searchBar: null
		};

		// 筛选选项
		document.querySelectorAll('[class*="filter"], [class*="option"]').forEach(el => {
			const text = el.textContent.trim();
			if (text && text.length < 30) {
				features.filters.push(text);
			}
		});

		// 排序选项
		document.querySelectorAll('[class*="sort"]').forEach(el => {
			features.sortOptions.push(el.textContent.trim());
		});

		// 搜索框
		const searchInput = document.querySelector('input[type="search"], input[placeholder*="搜索"]');
		if (searchInput) {
			features.searchBar = {
				placeholder: searchInput.placeholder,
				value: searchInput.value
			};
		}

		return JSON.stringify(features, null, 2);
	}`).String()

	fmt.Println("搜索页功能:")
	fmt.Println(searchFeatures)

	// 5. 检查可用的 API 和数据
	fmt.Println("\n5. 检查页面数据结构...")
	dataStructure := page.MustEval(`() => {
		const data = {
			hasInitialState: !!window.__INITIAL_STATE__,
			initialStateKeys: window.__INITIAL_STATE__ ? Object.keys(window.__INITIAL_STATE__) : [],
			hasUserData: !!window.__INITIAL_STATE__?.user,
			hasNoteData: !!window.__INITIAL_STATE__?.note,
			localStorage: Object.keys(localStorage),
			cookies: document.cookie.split(';').length
		};

		return JSON.stringify(data, null, 2);
	}`).String()

	fmt.Println("页面数据:")
	fmt.Println(dataStructure)

	fmt.Println("\n\n分析完成！浏览器将保持打开30秒...")
	time.Sleep(30 * time.Second)
}
