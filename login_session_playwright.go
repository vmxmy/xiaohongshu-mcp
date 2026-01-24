package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/sirupsen/logrus"
	"github.com/xpzouying/xiaohongshu-mcp/cookies"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
)

const (
	xhsLoginURL         = "https://www.xiaohongshu.com/explore"
	loginStatusSelector = ".main-container .user .link-wrapper .channel"
)

var qrSelectors = []string{
	".login-container .qrcode-img",
	".login-container canvas",
	".login-container img",
	".qrcode-img",
	"canvas",
}

const (
	qrFallbackRegex    = "二维码|扫码"
	securityHintRegexp = "安全认证|安全验证|风险验证|二次验证|安全校验|扫码验证|验证身份|保护账号安全|二维码.*失效"
	scanSuccessRegexp  = "扫码成功|手机上确认|重新扫码"
)

// qrElement 二维码元素接口
type qrElement interface {
	Screenshot() ([]byte, error)
}

// qrFrame 二维码框架接口
type qrFrame interface {
	HasRegex(ctx context.Context, selector, jsRegex string) (bool, error)
	Element(ctx context.Context, selector string) (qrElement, error)
	ElementByRegex(ctx context.Context, selector, jsRegex string) (qrElement, error)
	Frames(ctx context.Context) ([]qrFrame, error)
}

// qrPage 二维码页面接口
type qrPage interface {
	qrFrame
	Navigate(ctx context.Context, url string) error
	WaitLoad(ctx context.Context) error
	Has(ctx context.Context, selector string) (bool, error)
	Close() error
}

// browserPageAdapter 将 browser.Page 适配为 qrPage 接口
type browserPageAdapter struct {
	page browser.Page
	ctx  playwright.BrowserContext
}

func (b *browserPageAdapter) Navigate(ctx context.Context, url string) error {
	return b.page.Goto(url)
}

func (b *browserPageAdapter) WaitLoad(ctx context.Context) error {
	return b.page.WaitLoad()
}

func (b *browserPageAdapter) Has(ctx context.Context, selector string) (bool, error) {
	return b.page.Has(selector)
}

func (b *browserPageAdapter) HasRegex(ctx context.Context, selector, jsRegex string) (bool, error) {
	return b.page.HasRegex(selector, jsRegex)
}

func (b *browserPageAdapter) Element(ctx context.Context, selector string) (qrElement, error) {
	el, err := b.page.Element(selector)
	if err != nil {
		return nil, err
	}
	return &browserElementAdapter{element: el}, nil
}

func (b *browserPageAdapter) ElementByRegex(ctx context.Context, selector, jsRegex string) (qrElement, error) {
	el, err := b.page.ElementByRegex(selector, jsRegex)
	if err != nil {
		return nil, err
	}
	return &browserElementAdapter{element: el}, nil
}

func (b *browserPageAdapter) Frames(ctx context.Context) ([]qrFrame, error) {
	return framesForPage(ctx, b.page)
}

func (b *browserPageAdapter) Close() error {
	return b.page.Close()
}

// browserElementAdapter 将 browser.Element 适配为 qrElement 接口
type browserElementAdapter struct {
	element browser.Element
}

func (b *browserElementAdapter) Screenshot() ([]byte, error) {
	// Playwright 元素截图通过 Eval 实现
	bbox, err := b.element.BoundingBox()
	if err != nil {
		return nil, err
	}
	if bbox == nil {
		return nil, errors.New("element has no bounding box")
	}

	// 通过 Eval 执行截图
	result, err := b.element.Eval(`(el) => {
		const canvas = document.createElement('canvas');
		const ctx = canvas.getContext('2d');
		const rect = el.getBoundingClientRect();
		canvas.width = rect.width;
		canvas.height = rect.height;

		// 如果是 canvas 元素，直接获取数据
		if (el.tagName === 'CANVAS') {
			return el.toDataURL('image/png');
		}

		// 如果是 img 元素，绘制到 canvas
		if (el.tagName === 'IMG') {
			ctx.drawImage(el, 0, 0);
			return canvas.toDataURL('image/png');
		}

		// 其他元素使用 html2canvas 或返回错误
		return null;
	}`)
	if err != nil {
		return nil, err
	}

	// 如果返回的是 data URL，解析它
	if dataURL, ok := result.(string); ok && len(dataURL) > 0 {
		// data:image/png;base64,iVBORw0KG...
		if len(dataURL) > 22 && dataURL[:22] == "data:image/png;base64," {
			return base64.StdEncoding.DecodeString(dataURL[22:])
		}
	}

	return nil, errors.New("failed to capture element screenshot")
}

// browserFrameAdapter 将 browser.Page (frame) 适配为 qrFrame 接口
type browserFrameAdapter struct {
	frame browser.Page
}

func (b *browserFrameAdapter) HasRegex(ctx context.Context, selector, jsRegex string) (bool, error) {
	return b.frame.HasRegex(selector, jsRegex)
}

func (b *browserFrameAdapter) Element(ctx context.Context, selector string) (qrElement, error) {
	el, err := b.frame.Element(selector)
	if err != nil {
		return nil, err
	}
	return &browserElementAdapter{element: el}, nil
}

func (b *browserFrameAdapter) ElementByRegex(ctx context.Context, selector, jsRegex string) (qrElement, error) {
	el, err := b.frame.ElementByRegex(selector, jsRegex)
	if err != nil {
		return nil, err
	}
	return &browserElementAdapter{element: el}, nil
}

func (b *browserFrameAdapter) Frames(ctx context.Context) ([]qrFrame, error) {
	return framesForPage(ctx, b.frame)
}

func framesForPage(ctx context.Context, page browser.Page) ([]qrFrame, error) {
	frames := []qrFrame{}
	if page == nil {
		return frames, nil
	}

	// 获取所有 iframe 元素
	iframes, err := page.Elements("iframe")
	if err != nil {
		return frames, err
	}

	// 对于每个 iframe，获取其 frame page
	for _, el := range iframes {
		framePage, err := el.Frame()
		if err != nil {
			continue
		}
		frames = append(frames, &browserFrameAdapter{frame: framePage})
	}
	return frames, nil
}

type playwrightLoginSession struct {
	engine      browser.Engine
	page        qrPage
	pwPage      *playwrightPageWrapper // 保存包含 context 的包装器
	saveCookies func() error
	sleep       func(time.Duration)
}

// playwrightPageWrapper 包装 browser.Page 并保存 playwright context
type playwrightPageWrapper struct {
	page browser.Page
	ctx  playwright.BrowserContext
}

func newPlaywrightLoginSession() (loginSession, error) {
	engine := newBrowserEngine()
	if err := engine.Start(); err != nil {
		return nil, err
	}

	page, err := engine.NewPage()
	if err != nil {
		_ = engine.Close()
		return nil, err
	}

	// 尝试从 page 中获取 playwright context
	var pwCtx playwright.BrowserContext
	type pwPageInterface interface {
		GetContext() playwright.BrowserContext
	}
	if pwPage, ok := page.(pwPageInterface); ok {
		pwCtx = pwPage.GetContext()
	}

	wrapper := &playwrightPageWrapper{
		page: page,
		ctx:  pwCtx,
	}

	return &playwrightLoginSession{
		engine:      engine,
		page:        &browserPageAdapter{page: page},
		pwPage:      wrapper,
		saveCookies: func() error { return saveCookiesFromWrapper(wrapper) },
		sleep:       time.Sleep,
	}, nil
}

func (s *playwrightLoginSession) Open(ctx context.Context) error {
	if s.page == nil {
		return errors.New("login page not initialized")
	}
	if err := s.page.Navigate(ctx, xhsLoginURL); err != nil {
		return err
	}
	if err := s.page.WaitLoad(ctx); err != nil {
		return err
	}
	if s.sleep != nil {
		s.sleep(2 * time.Second)
	}
	return nil
}

func (s *playwrightLoginSession) LoggedIn(ctx context.Context) (bool, error) {
	if s.page == nil {
		return false, errors.New("login page not initialized")
	}
	ok, err := s.page.Has(ctx, loginStatusSelector)
	loginVisible, loginErr := s.page.Has(ctx, ".login-container")
	logrus.WithFields(logrus.Fields{
		"login_status_selector": loginStatusSelector,
		"login_status_match":    ok,
		"login_status_err":      err,
		"login_container_match": loginVisible,
		"login_container_err":   loginErr,
	}).Info("login status selector check")
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (s *playwrightLoginSession) QRCode(ctx context.Context) (loginQRCode, error) {
	if s.page == nil {
		return loginQRCode{}, errors.New("login page not initialized")
	}
	stage := "login"
	if s.hasSecurityHint(ctx) {
		stage = "security"
	}

	logrus.WithField("stage", stage).Info("login qrcode stage detect")

	el, err := s.findQRCodeElement(ctx, stage == "security")
	if err != nil {
		return loginQRCode{}, err
	}

	img, err := el.Screenshot()
	if err != nil {
		return loginQRCode{}, err
	}

	return loginQRCode{
		Image: base64.StdEncoding.EncodeToString(img),
		Stage: stage,
	}, nil
}

func (s *playwrightLoginSession) SaveCookies() error {
	if s.saveCookies == nil {
		return nil
	}
	return s.saveCookies()
}

func (s *playwrightLoginSession) Close() error {
	if s.page != nil {
		_ = s.page.Close()
	}
	if s.engine != nil {
		s.engine.Close()
	}
	return nil
}

func (s *playwrightLoginSession) hasSecurityHint(ctx context.Context) bool {
	ok, err := s.page.HasRegex(ctx, "body", securityHintRegexp)
	scanOK, scanErr := s.page.HasRegex(ctx, "body", scanSuccessRegexp)
	logrus.WithFields(logrus.Fields{
		"match":       ok,
		"err":         err,
		"scan_match":  scanOK,
		"scan_err":    scanErr,
		"scan_regex":  scanSuccessRegexp,
		"security_re": securityHintRegexp,
	}).Info("login qrcode security hint on page")
	if err == nil && ok {
		return true
	}
	return s.frameHasSecurityHint(ctx, s.page)
}

func (s *playwrightLoginSession) findQRCodeElement(ctx context.Context, preferFrames bool) (qrElement, error) {
	if preferFrames {
		if el, ok := s.findQRCodeElementInChildFrames(ctx, s.page); ok {
			logrus.WithField("source", "frame").Info("login qrcode element found")
			return el, nil
		}
	}

	for _, selector := range qrSelectors {
		el, err := s.page.Element(ctx, selector)
		if err == nil && el != nil {
			logrus.WithFields(logrus.Fields{
				"source":   "page",
				"selector": selector,
			}).Info("login qrcode element found")
			return el, nil
		}
	}

	el, err := s.page.ElementByRegex(ctx, "div", qrFallbackRegex)
	if err == nil && el != nil {
		logrus.WithField("source", "page_fallback").Info("login qrcode element found")
		return el, nil
	}

	if !preferFrames {
		if el, ok := s.findQRCodeElementInChildFrames(ctx, s.page); ok {
			logrus.WithField("source", "frame").Info("login qrcode element found")
			return el, nil
		}
	}

	logrus.WithFields(logrus.Fields{
		"prefer_frames": preferFrames,
	}).Info("login qrcode element not found")
	return nil, errors.New("login qrcode element not found")
}

func (s *playwrightLoginSession) frameHasSecurityHint(ctx context.Context, frame qrFrame) bool {
	frames, err := frame.Frames(ctx)
	logrus.WithFields(logrus.Fields{
		"count": len(frames),
		"err":   err,
	}).Info("login qrcode scan frames")
	if err != nil {
		return false
	}
	for _, child := range frames {
		ok, err := child.HasRegex(ctx, "body", securityHintRegexp)
		scanOK, scanErr := child.HasRegex(ctx, "body", scanSuccessRegexp)
		logrus.WithFields(logrus.Fields{
			"match":       ok,
			"err":         err,
			"scan_match":  scanOK,
			"scan_err":    scanErr,
			"scan_regex":  scanSuccessRegexp,
			"security_re": securityHintRegexp,
		}).Info("login qrcode security hint on frame")
		if err == nil && ok {
			return true
		}
		if s.frameHasSecurityHint(ctx, child) {
			return true
		}
	}
	return false
}

func (s *playwrightLoginSession) findQRCodeElementInFrame(ctx context.Context, frame qrFrame) (qrElement, bool) {
	for _, selector := range qrSelectors {
		el, err := frame.Element(ctx, selector)
		if err == nil && el != nil {
			return el, true
		}
	}

	el, err := frame.ElementByRegex(ctx, "div", qrFallbackRegex)
	if err == nil && el != nil {
		return el, true
	}

	frames, err := frame.Frames(ctx)
	if err != nil {
		return nil, false
	}
	for _, child := range frames {
		if el, ok := s.findQRCodeElementInFrame(ctx, child); ok {
			return el, true
		}
	}
	return nil, false
}

func (s *playwrightLoginSession) findQRCodeElementInChildFrames(ctx context.Context, frame qrFrame) (qrElement, bool) {
	frames, err := frame.Frames(ctx)
	if err != nil {
		return nil, false
	}
	for _, child := range frames {
		if el, ok := s.findQRCodeElementInFrame(ctx, child); ok {
			return el, true
		}
	}
	return nil, false
}

// saveCookiesFromWrapper 保存 Playwright 的 cookies
func saveCookiesFromWrapper(wrapper *playwrightPageWrapper) error {
	if wrapper == nil || wrapper.ctx == nil {
		return errors.New("playwright context is nil")
	}

	cks, err := wrapper.ctx.Cookies()
	if err != nil {
		return err
	}

	data, err := json.Marshal(cks)
	if err != nil {
		return err
	}

	cookiePath := cookies.GetCookiesFilePath()
	return os.WriteFile(cookiePath, data, 0644)
}
