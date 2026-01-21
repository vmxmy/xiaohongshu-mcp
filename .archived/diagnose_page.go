package main

import (
	"fmt"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/xpzouying/xiaohongshu-mcp/cookies"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run diagnose_page.go <url>")
		fmt.Println("示例: go run diagnose_page.go https://www.xiaohongshu.com/user/profile/me/followers")
		os.Exit(1)
	}

	url := os.Args[1]

	// 启动浏览器
	l := launcher.New().Headless(false)
	defer l.Cleanup()
	controlURL := l.MustLaunch()

	browser := rod.New().ControlURL(controlURL).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage()

	// 加载cookies
	cookiesData, err := cookies.LoadCookies()
	if err != nil {
		fmt.Printf("加载cookies失败: %v\n", err)
		os.Exit(1)
	}

	if err := cookies.SetCookies(page, cookiesData); err != nil {
		fmt.Printf("设置cookies失败: %v\n", err)
		os.Exit(1)
	}

	// 导航到页面
	fmt.Printf("导航到: %s\n", url)
	page.MustNavigate(url)
	page.MustWaitLoad()
	time.Sleep(5 * time.Second)

	// 获取页面HTML
	html := page.MustHTML()
	fmt.Printf("\n页面HTML长度: %d\n", len(html))

	// 保存HTML到文件
	filename := "page_dump.html"
	if err := os.WriteFile(filename, []byte(html), 0644); err != nil {
		fmt.Printf("保存HTML失败: %v\n", err)
	} else {
		fmt.Printf("HTML已保存到: %s\n", filename)
	}

	// 获取页面文本
	text := page.MustEval(`() => document.body.innerText`).String()
	fmt.Printf("\n页面文本内容（前500字符）:\n%s\n", text[:min(len(text), 500)])

	// 尝试查找常见元素
	fmt.Println("\n=== 诊断信息 ===")

	// 检查是否有用户卡片
	userCards := page.MustElements("div, li, article")
	fmt.Printf("找到 %d 个 div/li/article 元素\n", len(userCards))

	// 检查class属性
	fmt.Println("\n前10个元素的class属性:")
	for i, elem := range userCards {
		if i >= 10 {
			break
		}
		if class, err := elem.Attribute("class"); err == nil && class != nil {
			fmt.Printf("%d: class=\"%s\"\n", i+1, *class)
		}
	}

	fmt.Println("\n按任意键退出...")
	fmt.Scanln()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
