package rod

import (
	"errors"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/xpzouying/headless_browser"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
)

type Config struct {
	Headless          bool
	ActionTimeout     time.Duration
	NavigationTimeout time.Duration
	CookiePath        string
}

func DefaultConfig() Config {
	return Config{
		Headless:          true,
		ActionTimeout:     30 * time.Second,
		NavigationTimeout: 60 * time.Second,
	}
}

type Engine struct {
	cfg     Config
	manager *headless_browser.Browser // headless_browser 管理器
}

func New(cfg Config) *Engine {
	return &Engine{cfg: cfg}
}

func (e *Engine) Start() error {
	// 使用 headless_browser 包来管理浏览器实例
	e.manager = headless_browser.New(headless_browser.WithHeadless(e.cfg.Headless))
	return nil
}

func (e *Engine) NewPage() (browser.Page, error) {
	if e.manager == nil {
		return nil, errors.New("browser not started")
	}

	rodPage := e.manager.NewPage()

	// 设置超时
	if e.cfg.ActionTimeout > 0 {
		rodPage = rodPage.Timeout(e.cfg.ActionTimeout)
	}

	// 如果有 cookie 路径，加载 cookies
	if e.cfg.CookiePath != "" {
		cookies, err := loadCookies(e.cfg.CookiePath)
		if err == nil && len(cookies) > 0 {
			// 转换并设置 cookies
			if err := setCookies(rodPage, cookies); err != nil {
				_ = rodPage.Close()
				return nil, err
			}
		}
	}

	return &page{
		p:                 rodPage,
		actionTimeout:     e.cfg.ActionTimeout,
		navigationTimeout: e.cfg.NavigationTimeout,
	}, nil
}

func (e *Engine) Close() error {
	if e.manager != nil {
		e.manager.Close()
	}
	return nil
}

type page struct {
	p                 *rod.Page
	actionTimeout     time.Duration
	navigationTimeout time.Duration
}

func (p *page) Goto(url string) error {
	timeout := p.navigationTimeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	page := p.p.Timeout(timeout)
	if err := page.Navigate(url); err != nil {
		return err
	}

	// 等待页面加载
	return page.WaitLoad()
}

func (p *page) Click(selector string) error {
	page := p.withActionTimeout()

	// 等待元素可见
	elem, err := page.Element(selector)
	if err != nil {
		return err
	}

	// 滚动到视图中
	if err := elem.ScrollIntoView(); err != nil {
		return err
	}

	// 等待元素可交互
	if err := elem.WaitVisible(); err != nil {
		return err
	}

	return elem.Click(proto.InputMouseButtonLeft, 1)
}

func (p *page) Fill(selector, value string) error {
	page := p.withActionTimeout()

	elem, err := page.Element(selector)
	if err != nil {
		return err
	}

	// 先清空
	if err := elem.SelectAllText(); err != nil {
		return err
	}

	// 输入新值
	return elem.Input(value)
}

func (p *page) SetFiles(selector string, files []string) error {
	page := p.withActionTimeout()

	elem, err := page.Element(selector)
	if err != nil {
		return err
	}

	return elem.SetFiles(files)
}

func (p *page) Text(selector string) (string, error) {
	page := p.withActionTimeout()

	elem, err := page.Element(selector)
	if err != nil {
		return "", err
	}

	return elem.Text()
}

func (p *page) WaitVisible(selector string) error {
	page := p.withActionTimeout()

	elem, err := page.Element(selector)
	if err != nil {
		return err
	}

	return elem.WaitVisible()
}

func (p *page) URL() string {
	info := p.p.MustInfo()
	return info.URL
}

func (p *page) IsVisible(selector string) (bool, error) {
	page := p.withActionTimeout()

	has, elem, err := page.Has(selector)
	if err != nil || !has {
		return false, err
	}

	visible, err := elem.Visible()
	if err != nil {
		return false, err
	}

	return visible, nil
}

func (p *page) ScrollIntoView(selector string) error {
	page := p.withActionTimeout()

	elem, err := page.Element(selector)
	if err != nil {
		return err
	}

	return elem.ScrollIntoView()
}

func (p *page) ClickForce(selector string) error {
	page := p.withActionTimeout()

	elem, err := page.Element(selector)
	if err != nil {
		return err
	}

	// Rod 的 MustClick 会强制点击
	return elem.Click(proto.InputMouseButtonLeft, 1)
}

func (p *page) Close() error {
	return p.p.Close()
}

func (p *page) withActionTimeout() *rod.Page {
	if p.actionTimeout > 0 {
		return p.p.Timeout(p.actionTimeout)
	}
	return p.p
}
