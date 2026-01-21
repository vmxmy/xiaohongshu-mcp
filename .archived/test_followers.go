package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/xpzouying/xiaohongshu-mcp/cookies"
)

func main() {
	fmt.Println("=== 粉丝列表调试工具 ===\n")

	// 启动浏览器
	l := launcher.New().Headless(false)
	defer l.Cleanup()
	controlURL := l.MustLaunch()

	browser := rod.New().ControlURL(controlURL).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage()

	// 加载cookies
	fmt.Println("1. 加载cookies...")
	cookiePath := cookies.GetCookiesFilePath()
	cookieLoader := cookies.NewLoadCookie(cookiePath)

	cookiesData, err := cookieLoader.LoadCookies()
	if err != nil {
		fmt.Printf("❌ 加载cookies失败: %v\n", err)
		os.Exit(1)
	}

	// 解析cookies并设置到页面
	var cookieList []*proto.NetworkCookie
	if err := json.Unmarshal(cookiesData, &cookieList); err != nil {
		fmt.Printf("❌ 解析cookies失败: %v\n", err)
		os.Exit(1)
	}

	if err := page.SetCookies(cookieList); err != nil {
		fmt.Printf("❌ 设置cookies失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Cookies加载成功\n")

	// 导航到粉丝页面
	url := "https://www.xiaohongshu.com/user/profile/me/followers"
	fmt.Printf("2. 导航到: %s\n", url)
	page.MustNavigate(url)
	page.MustWaitLoad()
	time.Sleep(5 * time.Second)

	// 获取页面HTML
	html := page.MustHTML()
	fmt.Printf("✅ 页面HTML长度: %d\n\n", len(html))

	// 保存HTML
	filename := "test_followers_page.html"
	if err := os.WriteFile(filename, []byte(html), 0644); err == nil {
		fmt.Printf("✅ HTML已保存到: %s\n\n", filename)
	}

	// 获取页面文本
	text := page.MustEval(`() => document.body.innerText`).String()
	fmt.Printf("3. 页面文本内容（前500字符）:\n%s\n\n", text[:min(len(text), 500)])

	// 使用goquery解析
	fmt.Println("4. 使用goquery解析DOM结构...")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		fmt.Printf("❌ 解析HTML失败: %v\n", err)
		os.Exit(1)
	}

	// 尝试多种选择器
	selectors := []string{
		".user-card",
		"[class*='user-item']",
		"[class*='UserCard']",
		"[class*='follower']",
		"li[class*='user']",
		"div[class*='user']",
		"article",
		"li",
	}

	fmt.Println("\n测试各种选择器:")
	for _, selector := range selectors {
		count := doc.Find(selector).Length()
		fmt.Printf("  %s: %d 个元素\n", selector, count)
	}

	// 输出前20个div的class属性
	fmt.Println("\n5. 前20个div元素的class属性:")
	doc.Find("div").Each(func(i int, s *goquery.Selection) {
		if i >= 20 {
			return
		}
		if class, exists := s.Attr("class"); exists && class != "" {
			fmt.Printf("  %d: %s\n", i+1, class)
		}
	})

	// 输出前20个li的class属性
	fmt.Println("\n6. 前20个li元素的class属性:")
	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		if i >= 20 {
			return
		}
		if class, exists := s.Attr("class"); exists && class != "" {
			fmt.Printf("  %d: %s\n", i+1, class)
		}
	})

	fmt.Println("\n按任意键退出...")
	fmt.Scanln()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
