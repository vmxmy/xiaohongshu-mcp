package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xpzouying/xiaohongshu-mcp/cookies"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser/playwright"
)

func main() {
	fmt.Println("🧪 测试 Playwright 导航修复...")
	fmt.Println()

	cfg := playwright.DefaultConfig()
	cfg.Headless = false
	cfg.CookiePath = cookies.GetCookiesFilePath()
	cfg.ActionTimeout = 30 * time.Second
	cfg.NavigationTimeout = 60 * time.Second

	engine := playwright.New(cfg)
	fmt.Println("✅ 创建引擎成功")

	if err := engine.Start(); err != nil {
		fmt.Printf("❌ 启动浏览器失败: %v\n", err)
		return
	}
	defer engine.Close()
	fmt.Println("✅ 启动浏览器成功")

	page, err := engine.NewPage()
	if err != nil {
		fmt.Printf("❌ 创建页面失败: %v\n", err)
		return
	}
	defer page.Close()
	fmt.Println("✅ 创建页面成功")

	ctx := context.Background()
	pp := page.WithContext(ctx)

	// 测试导航
	fmt.Println()
	fmt.Println("📍 开始导航到小红书...")
	start := time.Now()
	err = pp.Goto("https://www.xiaohongshu.com/explore")
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 导航失败 (耗时 %v): %v\n", elapsed, err)
		return
	}
	fmt.Printf("✅ 导航成功！耗时: %v\n", elapsed)

	// 等待页面加载
	fmt.Println("⏳ 等待页面加载...")
	if err := pp.WaitLoad(); err != nil {
		fmt.Printf("❌ 等待加载失败: %v\n", err)
		return
	}
	fmt.Println("✅ 页面加载完成")

	time.Sleep(1 * time.Second)

	// 检查登录元素
	fmt.Println()
	fmt.Println("🔍 检查登录元素...")
	selector := `.main-container .user .link-wrapper .channel`
	exists, err := pp.Has(selector)
	if err != nil {
		fmt.Printf("❌ 检查元素失败: %v\n", err)
		return
	}

	if exists {
		fmt.Println("✅ 找到登录元素 - 已登录")
	} else {
		fmt.Println("ℹ️  未找到登录元素 - 未登录")
	}

	fmt.Println()
	fmt.Println("🎉 所有测试通过！")
	fmt.Println("按 Ctrl+C 退出...")
	time.Sleep(5 * time.Second)
}
