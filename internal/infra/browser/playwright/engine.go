package playwright

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
)

type Config struct {
	Headless          bool
	ActionTimeout     time.Duration
	NavigationTimeout time.Duration
	CookiePath        string
}

func DefaultConfig() Config {
	return Config{Headless: true}
}

type Engine struct {
	cfg     Config
	pw      *playwright.Playwright
	browser playwright.Browser
}

func New(cfg Config) *Engine {
	return &Engine{cfg: cfg}
}

func (e *Engine) Start() error {
	pw, err := playwright.Run()
	if err != nil {
		return wrapPlaywrightError(err)
	}
	b, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(e.cfg.Headless),
	})
	if err != nil {
		_ = pw.Stop()
		return err
	}
	e.pw = pw
	e.browser = b
	return nil
}

func (e *Engine) NewPage() (browser.Page, error) {
	if e.browser == nil {
		return nil, errors.New("browser not started")
	}

	// 创建上下文选项，设置视口大小（确保无头模式下有足够大的视口）
	contextOptions := playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
		// 设置 User-Agent，避免被检测为自动化
		UserAgent: playwright.String("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	}

	if e.cfg.CookiePath != "" {
		ctx, err := e.browser.NewContext(contextOptions)
		if err != nil {
			return nil, err
		}
		cookies, err := loadCookies(e.cfg.CookiePath)
		if err != nil {
			_ = ctx.Close()
			return nil, err
		}
		if len(cookies) > 0 {
			if err := ctx.AddCookies(cookies); err != nil {
				_ = ctx.Close()
				return nil, err
			}
		}
		p, err := ctx.NewPage()
		if err != nil {
			_ = ctx.Close()
			return nil, err
		}
		applyTimeouts(p, e.cfg)
		return &page{p: p, ctx: ctx}, nil
	}

	ctx, err := e.browser.NewContext(contextOptions)
	if err != nil {
		return nil, err
	}
	p, err := ctx.NewPage()
	if err != nil {
		_ = ctx.Close()
		return nil, err
	}
	applyTimeouts(p, e.cfg)
	return &page{p: p, ctx: ctx}, nil
}

func (e *Engine) Close() error {
	if e.browser != nil {
		_ = e.browser.Close()
	}
	if e.pw != nil {
		return e.pw.Stop()
	}
	return nil
}

type page struct {
	p   playwright.Page
	ctx playwright.BrowserContext
}

func (p *page) Goto(url string) error {
	_, err := p.p.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	})
	return err
}

func (p *page) Click(selector string) error {
	// 使用 first() 处理多个匹配的情况
	return p.p.Locator(selector).First().Click()
}

func (p *page) Fill(selector, value string) error {
	return p.p.Locator(selector).Fill(value)
}

func (p *page) SetFiles(selector string, files []string) error {
	return p.p.Locator(selector).SetInputFiles(files)
}

func (p *page) Text(selector string) (string, error) {
	text, err := p.p.Locator(selector).TextContent()
	if err != nil {
		return "", err
	}
	return text, nil
}

func (p *page) WaitVisible(selector string) error {
	return p.p.Locator(selector).WaitFor()
}

func (p *page) URL() string {
	return p.p.URL()
}

func (p *page) IsVisible(selector string) (bool, error) {
	isVisible, err := p.p.Locator(selector).IsVisible()
	if err != nil {
		return false, err
	}
	return isVisible, nil
}

func (p *page) ScrollIntoView(selector string) error {
	// 使用 Playwright 的 Locator API，避免 querySelector 的字符串转义问题
	err := p.p.Locator(selector).ScrollIntoViewIfNeeded()
	return err
}

func (p *page) ClickForce(selector string) error {
	// 使用 Playwright 的强制点击选项，更可靠
	err := p.p.Locator(selector).Click(playwright.LocatorClickOptions{
		Force:   playwright.Bool(true),   // 强制点击，即使元素被遮挡
		Timeout: playwright.Float(10000), // 10秒超时
	})
	return err
}

func (p *page) Close() error {
	if p.ctx != nil {
		return p.ctx.Close()
	}
	return p.p.Close()
}

func applyTimeouts(p playwright.Page, cfg Config) {
	if cfg.ActionTimeout > 0 {
		p.SetDefaultTimeout(float64(cfg.ActionTimeout.Milliseconds()))
	}
	if cfg.NavigationTimeout > 0 {
		p.SetDefaultNavigationTimeout(float64(cfg.NavigationTimeout.Milliseconds()))
	}
}

func wrapPlaywrightError(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "install the driver") {
		return fmt.Errorf("playwright driver not installed; run: go run github.com/playwright-community/playwright-go/cmd/playwright install: %w", err)
	}
	return err
}
