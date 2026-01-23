package main

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/xpzouying/headless_browser"
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
)

type qrElement interface {
	Screenshot(format proto.PageCaptureScreenshotFormat, quality int) ([]byte, error)
}

type qrPage interface {
	Navigate(ctx context.Context, url string) error
	WaitLoad(ctx context.Context) error
	Has(ctx context.Context, selector string) (bool, error)
	HasR(ctx context.Context, selector, jsRegex string) (bool, error)
	Element(ctx context.Context, selector string) (qrElement, error)
	ElementR(ctx context.Context, selector, jsRegex string) (qrElement, error)
	Close() error
}

type rodPageAdapter struct {
	page *rod.Page
}

func (r *rodPageAdapter) Navigate(ctx context.Context, url string) error {
	return r.page.Context(ctx).Navigate(url)
}

func (r *rodPageAdapter) WaitLoad(ctx context.Context) error {
	return r.page.Context(ctx).WaitLoad()
}

func (r *rodPageAdapter) Has(ctx context.Context, selector string) (bool, error) {
	ok, _, err := r.page.Context(ctx).Has(selector)
	return ok, err
}

func (r *rodPageAdapter) HasR(ctx context.Context, selector, jsRegex string) (bool, error) {
	ok, _, err := r.page.Context(ctx).HasR(selector, jsRegex)
	return ok, err
}

func (r *rodPageAdapter) Element(ctx context.Context, selector string) (qrElement, error) {
	return r.page.Context(ctx).Element(selector)
}

func (r *rodPageAdapter) ElementR(ctx context.Context, selector, jsRegex string) (qrElement, error) {
	return r.page.Context(ctx).ElementR(selector, jsRegex)
}

func (r *rodPageAdapter) Close() error {
	return r.page.Close()
}

type rodLoginSession struct {
	browser     *headless_browser.Browser
	page        qrPage
	saveCookies func() error
	sleep       func(time.Duration)
}

func newRodLoginSession() (loginSession, error) {
	b := newBrowser()
	p := b.NewPage()
	return &rodLoginSession{
		browser:     b,
		page:        &rodPageAdapter{page: p},
		saveCookies: func() error { return saveCookies(p) },
		sleep:       time.Sleep,
	}, nil
}

func (s *rodLoginSession) Open(ctx context.Context) error {
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

func (s *rodLoginSession) LoggedIn(ctx context.Context) (bool, error) {
	if s.page == nil {
		return false, errors.New("login page not initialized")
	}
	ok, err := s.page.Has(ctx, loginStatusSelector)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func (s *rodLoginSession) QRCode(ctx context.Context) (loginQRCode, error) {
	if s.page == nil {
		return loginQRCode{}, errors.New("login page not initialized")
	}
	stage := "login"
	if s.hasSecurityHint(ctx) {
		stage = "security"
	}

	el, err := s.findQRCodeElement(ctx)
	if err != nil {
		return loginQRCode{}, err
	}

	img, err := el.Screenshot(proto.PageCaptureScreenshotFormatPng, 100)
	if err != nil {
		return loginQRCode{}, err
	}
	return loginQRCode{
		Image: base64.StdEncoding.EncodeToString(img),
		Stage: stage,
	}, nil
}

func (s *rodLoginSession) SaveCookies() error {
	if s.saveCookies == nil {
		return nil
	}
	return s.saveCookies()
}

func (s *rodLoginSession) Close() error {
	if s.page != nil {
		_ = s.page.Close()
	}
	if s.browser != nil {
		s.browser.Close()
	}
	return nil
}

func (s *rodLoginSession) hasSecurityHint(ctx context.Context) bool {
	ok, err := s.page.HasR(ctx, "body", securityHintRegexp)
	if err != nil {
		return false
	}
	return ok
}

func (s *rodLoginSession) findQRCodeElement(ctx context.Context) (qrElement, error) {
	for _, selector := range qrSelectors {
		el, err := s.page.Element(ctx, selector)
		if err == nil && el != nil {
			return el, nil
		}
	}

	el, err := s.page.ElementR(ctx, "div", qrFallbackRegex)
	if err == nil && el != nil {
		return el, nil
	}

	return nil, errors.New("login qrcode element not found")
}
