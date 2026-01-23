package playwright

import (
	"errors"

	"github.com/playwright-community/playwright-go"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
)

type Config struct {
	Headless bool
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
		return err
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
	p, err := e.browser.NewPage()
	if err != nil {
		return nil, err
	}
	return &page{p: p}, nil
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
	p playwright.Page
}

func (p *page) Goto(url string) error {
	_, err := p.p.Goto(url)
	return err
}

func (p *page) Click(selector string) error {
	return p.p.Locator(selector).Click()
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

func (p *page) Close() error {
	return p.p.Close()
}
