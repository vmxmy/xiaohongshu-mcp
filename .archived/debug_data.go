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

	fmt.Println("=== 检查个人主页数据结构 ===\n")

	// 导航到个人主页
	page.MustNavigate("https://www.xiaohongshu.com/user/profile/me")
	page.MustWaitDOMStable()
	time.Sleep(3 * time.Second)

	// 检查 __INITIAL_STATE__ 结构
	result := page.MustEval(`() => {
		if (window.__INITIAL_STATE__) {
			return JSON.stringify(window.__INITIAL_STATE__, null, 2);
		}
		return "No __INITIAL_STATE__ found";
	}`).String()

	fmt.Println("__INITIAL_STATE__ 结构:")
	fmt.Println(result[:min(2000, len(result))])
	fmt.Println("\n...")

	// 检查用户数据
	userData := page.MustEval(`() => {
		if (window.__INITIAL_STATE__ && window.__INITIAL_STATE__.user) {
			return JSON.stringify(window.__INITIAL_STATE__.user, null, 2);
		}
		return "No user data";
	}`).String()

	fmt.Println("\n\n=== User Data ===")
	fmt.Println(userData[:min(1000, len(userData))])

	// 检查页面上的统计数字
	stats := page.MustEval(`() => {
		const stats = {};

		// 查找所有可能包含统计数字的元素
		document.querySelectorAll('[class*="count"], [class*="num"], [class*="stat"]').forEach(el => {
			const text = el.textContent.trim();
			if (text && text.length < 20) {
				stats[el.className] = text;
			}
		});

		return JSON.stringify(stats, null, 2);
	}`).String()

	fmt.Println("\n\n=== 页面统计元素 ===")
	fmt.Println(stats)

	fmt.Println("\n\n浏览器将保持打开30秒，请手动检查页面...")
	time.Sleep(30 * time.Second)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
