package main

import (
	"context"
	"encoding/base64"
	"errors"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"
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

type qrFrame interface {
	HasR(ctx context.Context, selector, jsRegex string) (bool, error)
	Element(ctx context.Context, selector string) (qrElement, error)
	ElementR(ctx context.Context, selector, jsRegex string) (qrElement, error)
	Frames(ctx context.Context) ([]qrFrame, error)
}

type qrPage interface {
	qrFrame
	Navigate(ctx context.Context, url string) error
	WaitLoad(ctx context.Context) error
	Has(ctx context.Context, selector string) (bool, error)
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

func (r *rodPageAdapter) Frames(ctx context.Context) ([]qrFrame, error) {
	return framesForPage(ctx, r.page)
}

func (r *rodPageAdapter) Close() error {
	return r.page.Close()
}

type rodFrameAdapter struct {
	page *rod.Page
}

func (r *rodFrameAdapter) HasR(ctx context.Context, selector, jsRegex string) (bool, error) {
	ok, _, err := r.page.Context(ctx).HasR(selector, jsRegex)
	return ok, err
}

func (r *rodFrameAdapter) Element(ctx context.Context, selector string) (qrElement, error) {
	return r.page.Context(ctx).Element(selector)
}

func (r *rodFrameAdapter) ElementR(ctx context.Context, selector, jsRegex string) (qrElement, error) {
	return r.page.Context(ctx).ElementR(selector, jsRegex)
}

func (r *rodFrameAdapter) Frames(ctx context.Context) ([]qrFrame, error) {
	return framesForPage(ctx, r.page)
}

func framesForPage(ctx context.Context, page *rod.Page) ([]qrFrame, error) {
	frames := []qrFrame{}
	if page == nil {
		return frames, nil
	}
	iframes, err := page.Context(ctx).Elements("iframe")
	if err != nil {
		return frames, err
	}
	for _, el := range iframes {
		framePage, err := el.Context(ctx).Frame()
		if err != nil {
			continue
		}
		frames = append(frames, &rodFrameAdapter{page: framePage})
	}
	return frames, nil
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

	logrus.WithField("stage", stage).Info("login qrcode stage detect")

	el, err := s.findQRCodeElement(ctx, stage == "security")
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
	logrus.WithFields(logrus.Fields{
		"match": ok,
		"err":   err,
	}).Info("login qrcode security hint on page")
	if err == nil && ok {
		return true
	}
	return s.frameHasSecurityHint(ctx, s.page)
}

func (s *rodLoginSession) findQRCodeElement(ctx context.Context, preferFrames bool) (qrElement, error) {
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

	el, err := s.page.ElementR(ctx, "div", qrFallbackRegex)
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

func (s *rodLoginSession) frameHasSecurityHint(ctx context.Context, frame qrFrame) bool {
	frames, err := frame.Frames(ctx)
	logrus.WithFields(logrus.Fields{
		"count": len(frames),
		"err":   err,
	}).Info("login qrcode scan frames")
	if err != nil {
		return false
	}
	for _, child := range frames {
		ok, err := child.HasR(ctx, "body", securityHintRegexp)
		logrus.WithFields(logrus.Fields{
			"match": ok,
			"err":   err,
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

func (s *rodLoginSession) findQRCodeElementInFrame(ctx context.Context, frame qrFrame) (qrElement, bool) {
	for _, selector := range qrSelectors {
		el, err := frame.Element(ctx, selector)
		if err == nil && el != nil {
			return el, true
		}
	}

	el, err := frame.ElementR(ctx, "div", qrFallbackRegex)
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

func (s *rodLoginSession) findQRCodeElementInChildFrames(ctx context.Context, frame qrFrame) (qrElement, bool) {
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
