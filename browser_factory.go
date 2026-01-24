package main

import (
	"time"

	"github.com/xpzouying/xiaohongshu-mcp/configs"
	"github.com/xpzouying/xiaohongshu-mcp/cookies"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser/playwright"
)

// newBrowserEngine 创建 Playwright 浏览器引擎
func newBrowserEngine() browser.Engine {
	cfg := playwright.DefaultConfig()
	cfg.Headless = configs.IsHeadless()
	cfg.CookiePath = cookies.GetCookiesFilePath()
	cfg.ActionTimeout = 30 * time.Second
	cfg.NavigationTimeout = 60 * time.Second

	return playwright.New(cfg)
}

// withBrowserPage 执行需要浏览器页面的操作的通用函数
func withBrowserPage(fn func(browser.Page) error) error {
	engine := newBrowserEngine()
	if err := engine.Start(); err != nil {
		return err
	}
	defer engine.Close()

	page, err := engine.NewPage()
	if err != nil {
		return err
	}
	defer page.Close()

	return fn(page)
}
