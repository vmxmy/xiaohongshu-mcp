package main

import (
	"context"
	"encoding/json"
	"flag"
	"time"

	playwrightgo "github.com/playwright-community/playwright-go"
	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/configs"
	"github.com/xpzouying/xiaohongshu-mcp/cookies"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser/playwright"
	"github.com/xpzouying/xiaohongshu-mcp/xiaohongshu"
)

func main() {
	var (
		binPath string // 浏览器二进制文件路径
	)
	flag.StringVar(&binPath, "bin", "", "浏览器二进制文件路径")
	flag.Parse()

	// 登录的时候，需要界面，所以不能无头模式
	engine := newBrowserEngine()
	if err := engine.Start(); err != nil {
		logrus.Fatalf("failed to start browser: %v", err)
	}
	defer engine.Close()

	page, err := engine.NewPage()
	if err != nil {
		logrus.Fatalf("failed to create page: %v", err)
	}
	defer page.Close()

	action := xiaohongshu.NewLogin(page)

	status, err := action.CheckLoginStatus(context.Background())
	if err != nil {
		logrus.Fatalf("failed to check login status: %v", err)
	}

	logrus.Infof("当前登录状态: %v", status)

	if status {
		return
	}

	// 开始登录流程
	logrus.Info("开始登录流程...")
	if err = action.Login(context.Background()); err != nil {
		logrus.Fatalf("登录失败: %v", err)
	} else {
		if err := saveCookies(page); err != nil {
			logrus.Fatalf("failed to save cookies: %v", err)
		}
	}

	// 再次检查登录状态确认成功
	status, err = action.CheckLoginStatus(context.Background())
	if err != nil {
		logrus.Fatalf("failed to check login status after login: %v", err)
	}

	if status {
		logrus.Info("登录成功！")
	} else {
		logrus.Error("登录流程完成但仍未登录")
	}

}

// newBrowserEngine 创建 Playwright 浏览器引擎
func newBrowserEngine() browser.Engine {
	cfg := playwright.DefaultConfig()
	cfg.Headless = configs.IsHeadless()
	cfg.CookiePath = cookies.GetCookiesFilePath()
	cfg.ActionTimeout = 30 * time.Second
	cfg.NavigationTimeout = 60 * time.Second

	return playwright.New(cfg)
}

func saveCookies(page browser.Page) error {
	// 将 browser.Page 转换为 Playwright 的具体实现，获取 context
	type contextGetter interface {
		GetContext() playwrightgo.BrowserContext
	}

	pg, ok := page.(contextGetter)
	if !ok {
		logrus.Warn("无法获取 Playwright context，跳过保存 cookies")
		return nil
	}

	ctx := pg.GetContext()
	if ctx == nil {
		return nil
	}

	cks, err := ctx.Cookies()
	if err != nil {
		return err
	}

	data, err := json.Marshal(cks)
	if err != nil {
		return err
	}

	cookieLoader := cookies.NewLoadCookie(cookies.GetCookiesFilePath())
	return cookieLoader.SaveCookies(data)
}
