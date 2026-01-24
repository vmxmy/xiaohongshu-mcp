package playwright

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
)

const (
	defaultWaitDOMStableInterval = 200 * time.Millisecond
	minStableDuration            = 1 * time.Second
)

type page struct {
	p        playwright.Page
	ctx      playwright.BrowserContext
	boundCtx context.Context
	timeout  time.Duration
}

func newPage(p playwright.Page, ctx playwright.BrowserContext) *page {
	return &page{p: p, ctx: ctx}
}

func (p *page) clone() *page {
	cp := *p
	return &cp
}

func (p *page) Close() error {
	if p.ctx != nil {
		return p.ctx.Close()
	}
	return p.p.Close()
}

func (p *page) Goto(url string) error {
	_, err := p.p.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	return err
}

func (p *page) Reload() error {
	_, err := p.p.Reload(playwright.PageReloadOptions{
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	return err
}

func (p *page) WaitLoad() error {
	return p.p.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateLoad,
	})
}

func (p *page) WaitIdle() error {
	return p.p.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
}

func (p *page) WaitDOMStable(maxWait time.Duration, stabilityThreshold float64) error {
	if maxWait <= 0 {
		maxWait = 5 * time.Second
	}
	if stabilityThreshold <= 0 {
		stabilityThreshold = 0.05
	}
	ticker := time.NewTicker(defaultWaitDOMStableInterval)
	defer ticker.Stop()

	deadline := time.Now().Add(maxWait)
	var prev float64 = -1
	var stableFor time.Duration

	for time.Now().Before(deadline) {
		select {
		case <-ticker.C:
			if err := p.checkContext(); err != nil {
				// Context 被销毁通常是因为页面导航，这是正常情况
				if isContextDestroyedError(err) {
					return nil
				}
				return err
			}
			val, err := p.Eval(`() => document.body ? document.body.innerHTML.length : 0`)
			if err != nil {
				// Context 被销毁通常是因为页面导航，这是正常情况
				if isContextDestroyedError(err) {
					return nil
				}
				return err
			}
			current := toFloat(val)
			if prev < 0 {
				prev = current
				continue
			}
			diff := math.Abs(current - prev)
			base := math.Max(prev, 1)
			if diff/base <= stabilityThreshold {
				stableFor += defaultWaitDOMStableInterval
				if stableFor >= minStableDuration {
					return nil
				}
			} else {
				stableFor = 0
			}
			prev = current
		default:
			if err := p.checkContext(); err != nil {
				// Context 被销毁通常是因为页面导航，这是正常情况
				if isContextDestroyedError(err) {
					return nil
				}
				return err
			}
			time.Sleep(defaultWaitDOMStableInterval / 2)
		}
	}
	return errors.New("dom not stable before timeout")
}

// isContextDestroyedError 检查是否是 execution context destroyed 错误
func isContextDestroyedError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return regexp.MustCompile(`(?i)execution context.*destroyed|context.*destroyed`).MatchString(errMsg)
}

func (p *page) WaitVisible(selector string) error {
	return p.p.Locator(selector).First().WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
}

func (p *page) WaitHidden(selector string) error {
	return p.p.Locator(selector).First().WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateHidden,
	})
}

func (p *page) WaitForSelector(selector string, timeout time.Duration) error {
	return p.p.Locator(selector).First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: timeoutFloat(p.effectiveTimeout(timeout)),
	})
}

func (p *page) WaitForFunction(expression string, timeout time.Duration) error {
	handle, err := p.p.WaitForFunction(expression, nil, playwright.PageWaitForFunctionOptions{
		Timeout: timeoutFloat(p.effectiveTimeout(timeout)),
	})
	if handle != nil {
		defer handle.Dispose()
	}
	return err
}

func (p *page) Click(selector string) error {
	return p.p.Locator(selector).First().Click(playwright.LocatorClickOptions{
		Timeout: timeoutFloat(p.effectiveTimeout(0)),
	})
}

func (p *page) ClickForce(selector string) error {
	return p.p.Locator(selector).First().Click(playwright.LocatorClickOptions{
		Timeout: timeoutFloat(p.effectiveTimeout(0)),
		Force:   playwright.Bool(true),
	})
}

func (p *page) DoubleClick(selector string) error {
	return p.p.Locator(selector).First().Dblclick(playwright.LocatorDblclickOptions{
		Timeout: timeoutFloat(p.effectiveTimeout(0)),
	})
}

func (p *page) Fill(selector, value string) error {
	return p.p.Locator(selector).First().Fill(value)
}

func (p *page) Type(selector, value string) error {
	return p.p.Locator(selector).First().Type(value)
}

func (p *page) SetFiles(selector string, files []string) error {
	if len(files) == 0 {
		return errors.New("files cannot be empty")
	}
	return p.p.Locator(selector).First().SetInputFiles(files)
}

func (p *page) Hover(selector string) error {
	return p.p.Locator(selector).First().Hover()
}

func (p *page) Focus(selector string) error {
	return p.p.Locator(selector).First().Focus()
}

func (p *page) Press(key string) error {
	return p.p.Keyboard().Press(key)
}

func (p *page) ScrollIntoView(selector string) error {
	return p.p.Locator(selector).First().ScrollIntoViewIfNeeded()
}

func (p *page) ScrollBy(deltaX, deltaY float64) error {
	_, err := p.Eval(`(delta) => window.scrollBy(delta.x, delta.y)`, map[string]float64{
		"x": deltaX,
		"y": deltaY,
	})
	return err
}

func (p *page) IsVisible(selector string) (bool, error) {
	return p.p.Locator(selector).First().IsVisible()
}

func (p *page) Text(selector string) (string, error) {
	return p.p.Locator(selector).First().InnerText()
}

func (p *page) HTML(selector string) (string, error) {
	return p.p.Locator(selector).First().InnerHTML()
}

func (p *page) Attribute(selector, name string) (string, error) {
	return p.p.Locator(selector).First().GetAttribute(name)
}

func (p *page) Eval(expression string, args ...interface{}) (interface{}, error) {
	switch len(args) {
	case 0:
		return p.p.Evaluate(expression)
	case 1:
		return p.p.Evaluate(expression, args[0])
	default:
		return nil, errors.New("Eval supports at most one argument")
	}
}

func (p *page) EvalOnSelector(selector, expression string, args ...interface{}) (interface{}, error) {
	var arg interface{}
	if len(args) > 0 {
		arg = args[0]
	}
	return p.p.EvalOnSelector(selector, expression, arg)
}

func (p *page) Screenshot(path string) error {
	if path == "" {
		return errors.New("screenshot path cannot be empty")
	}
	_, err := p.p.Screenshot(playwright.PageScreenshotOptions{
		Path: playwright.String(path),
	})
	return err
}

func (p *page) ScreenshotFullPage(path string) error {
	if path == "" {
		return errors.New("screenshot path cannot be empty")
	}
	_, err := p.p.Screenshot(playwright.PageScreenshotOptions{
		Path:     playwright.String(path),
		FullPage: playwright.Bool(true),
	})
	return err
}

func (p *page) Element(selector string) (browser.Element, error) {
	handle, err := p.p.Locator(selector).First().ElementHandle()
	if err != nil {
		return nil, err
	}
	if handle == nil {
		return nil, fmt.Errorf("element %s not found", selector)
	}
	return newElement(handle, p), nil
}

func (p *page) ElementByRegex(selector, jsRegex string) (browser.Element, error) {
	locators, err := p.p.Locator(selector).All()
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(jsRegex)
	if err != nil {
		return nil, err
	}
	for _, loc := range locators {
		text, err := loc.InnerText()
		if err != nil {
			continue
		}
		if re.MatchString(text) {
			handle, err := loc.ElementHandle()
			if err != nil {
				return nil, err
			}
			return newElement(handle, p), nil
		}
	}
	return nil, fmt.Errorf("element %s matching regex %s not found", selector, jsRegex)
}

func (p *page) Elements(selector string) ([]browser.Element, error) {
	handles, err := p.p.Locator(selector).ElementHandles()
	if err != nil {
		return nil, err
	}
	result := make([]browser.Element, 0, len(handles))
	for _, handle := range handles {
		result = append(result, newElement(handle, p))
	}
	return result, nil
}

func (p *page) Has(selector string) (bool, error) {
	count, err := p.p.Locator(selector).Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (p *page) HasRegex(selector, jsRegex string) (bool, error) {
	locators, err := p.p.Locator(selector).All()
	if err != nil {
		return false, err
	}
	re, err := regexp.Compile(jsRegex)
	if err != nil {
		return false, err
	}
	for _, loc := range locators {
		text, err := loc.InnerText()
		if err != nil {
			continue
		}
		if re.MatchString(text) {
			return true, nil
		}
	}
	return false, nil
}

func (p *page) Mouse() browser.Mouse {
	return &mouse{m: p.p.Mouse()}
}

func (p *page) Keyboard() browser.Keyboard {
	return &keyboard{k: p.p.Keyboard()}
}

func (p *page) URL() string {
	return p.p.URL()
}

func (p *page) WithContext(ctx context.Context) browser.Page {
	cp := p.clone()
	cp.boundCtx = ctx
	return cp
}

func (p *page) WithTimeout(timeout time.Duration) browser.Page {
	cp := p.clone()
	cp.timeout = timeout
	return cp
}

// GetContext 返回底层的 Playwright BrowserContext
func (p *page) GetContext() playwright.BrowserContext {
	return p.ctx
}

func (p *page) checkContext() error {
	if p.boundCtx == nil {
		return nil
	}
	select {
	case <-p.boundCtx.Done():
		return p.boundCtx.Err()
	default:
		return nil
	}
}

func (p *page) effectiveTimeout(explicit time.Duration) time.Duration {
	if explicit > 0 {
		return explicit
	}
	return p.timeout
}

func timeoutFloat(d time.Duration) *float64 {
	if d <= 0 {
		return nil
	}
	v := float64(d.Milliseconds())
	return &v
}

func toFloat(value interface{}) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	default:
		return 0
	}
}

// element implements browser.Element。
type element struct {
	handle playwright.ElementHandle
	parent *page
}

func newElement(handle playwright.ElementHandle, parent *page) browser.Element {
	return &element{handle: handle, parent: parent}
}

func (e *element) Click() error {
	return e.handle.Click()
}

func (e *element) ClickForce() error {
	return e.handle.Click(playwright.ElementHandleClickOptions{
		Force: playwright.Bool(true),
	})
}

func (e *element) DoubleClick() error {
	return e.handle.Dblclick()
}

func (e *element) Hover() error {
	return e.handle.Hover()
}

func (e *element) Focus() error {
	return e.handle.Focus()
}

func (e *element) Fill(value string) error {
	return e.handle.Fill(value)
}

func (e *element) Type(value string) error {
	return e.handle.Type(value)
}

func (e *element) Press(key string) error {
	return e.handle.Press(key)
}

func (e *element) Input(value string) error {
	return e.handle.Type(value)
}

func (e *element) SetFiles(files []string) error {
	if len(files) == 0 {
		return errors.New("files cannot be empty")
	}
	return e.handle.SetInputFiles(files)
}

func (e *element) ScrollIntoView() error {
	return e.handle.ScrollIntoViewIfNeeded()
}

func (e *element) WaitVisible() error {
	return e.handle.WaitForElementState(elementStateValue(playwright.ElementStateVisible))
}

func (e *element) WaitHidden() error {
	return e.handle.WaitForElementState(elementStateValue(playwright.ElementStateHidden))
}

func (e *element) WaitStable(duration time.Duration) error {
	return e.handle.WaitForElementState(elementStateValue(playwright.ElementStateStable), playwright.ElementHandleWaitForElementStateOptions{
		Timeout: timeoutFloat(duration),
	})
}

func (e *element) IsVisible() (bool, error) {
	return e.handle.IsVisible()
}

func (e *element) Text() (string, error) {
	return e.handle.InnerText()
}

func (e *element) HTML() (string, error) {
	return e.handle.InnerHTML()
}

func (e *element) Attribute(name string) (string, error) {
	return e.handle.GetAttribute(name)
}

func (e *element) Value() (string, error) {
	val, err := e.handle.InputValue()
	if err == nil {
		return val, nil
	}
	return e.Attribute("value")
}

func (e *element) Eval(expression string, args ...interface{}) (interface{}, error) {
	switch len(args) {
	case 0:
		return e.handle.Evaluate(expression)
	case 1:
		return e.handle.Evaluate(expression, args[0])
	default:
		return nil, errors.New("Eval supports at most one argument")
	}
}

func (e *element) Element(selector string) (browser.Element, error) {
	child, err := e.handle.QuerySelector(selector)
	if err != nil {
		return nil, err
	}
	if child == nil {
		return nil, fmt.Errorf("selector %s not found", selector)
	}
	return newElement(child, e.parent), nil
}

func (e *element) Elements(selector string) ([]browser.Element, error) {
	children, err := e.handle.QuerySelectorAll(selector)
	if err != nil {
		return nil, err
	}
	result := make([]browser.Element, 0, len(children))
	for _, child := range children {
		result = append(result, newElement(child, e.parent))
	}
	return result, nil
}

func (e *element) Remove() error {
	_, err := e.handle.Evaluate(`(node) => node.remove()`)
	return err
}

func (e *element) BoundingBox() (*browser.BoundingBox, error) {
	rect, err := e.handle.BoundingBox()
	if err != nil {
		return nil, err
	}
	if rect == nil {
		return nil, errors.New("element has no bounding box")
	}
	return &browser.BoundingBox{
		X:      rect.X,
		Y:      rect.Y,
		Width:  rect.Width,
		Height: rect.Height,
	}, nil
}

func (e *element) Frame() (browser.Page, error) {
	frame, err := e.handle.ContentFrame()
	if err != nil {
		return nil, err
	}
	if frame == nil {
		return nil, errors.New("element has no content frame")
	}
	child := newPage(frame.Page(), e.parent.ctx)
	child.boundCtx = e.parent.boundCtx
	child.timeout = e.parent.timeout
	return child, nil
}

// mouse implements browser.Mouse。
type mouse struct {
	m           playwright.Mouse
	lastX       float64
	lastY       float64
	hasPosition bool
}

func (m *mouse) MoveTo(x, y float64) error {
	if err := m.m.Move(x, y); err != nil {
		return err
	}
	m.lastX = x
	m.lastY = y
	m.hasPosition = true
	return nil
}

func (m *mouse) Click(button browser.MouseButton, opts ...browser.MouseClickOption) error {
	option := playwright.MouseClickOptions{
		Button: mouseButtonPtr(mapMouseButton(button)),
	}
	if len(opts) == 1 {
		if opts[0].ClickCount > 0 {
			option.ClickCount = playwright.Int(opts[0].ClickCount)
		}
		if opts[0].Delay > 0 {
			option.Delay = playwright.Float(float64(opts[0].Delay.Milliseconds()))
		}
	}
	x, y := 0.0, 0.0
	if m.hasPosition {
		x, y = m.lastX, m.lastY
	}
	return m.m.Click(x, y, option)
}

func (m *mouse) Down(button browser.MouseButton) error {
	return m.m.Down(playwright.MouseDownOptions{
		Button: mouseButtonPtr(mapMouseButton(button)),
	})
}

func (m *mouse) Up(button browser.MouseButton) error {
	return m.m.Up(playwright.MouseUpOptions{
		Button: mouseButtonPtr(mapMouseButton(button)),
	})
}

func (m *mouse) Wheel(deltaX, deltaY float64) error {
	return m.m.Wheel(deltaX, deltaY)
}

// keyboard implements browser.Keyboard。
type keyboard struct {
	k playwright.Keyboard
}

func (k *keyboard) Type(text string) error {
	return k.k.Type(text)
}

func (k *keyboard) InsertText(text string) error {
	return k.k.InsertText(text)
}

func (k *keyboard) Press(key string) error {
	return k.k.Press(key)
}

func (k *keyboard) Down(key string) error {
	return k.k.Down(key)
}

func (k *keyboard) Up(key string) error {
	return k.k.Up(key)
}

func mapMouseButton(button browser.MouseButton) playwright.MouseButton {
	switch button {
	case browser.MouseButtonRight:
		return derefMouseButton(playwright.MouseButtonRight, "right")
	case browser.MouseButtonMiddle:
		return derefMouseButton(playwright.MouseButtonMiddle, "middle")
	default:
		return derefMouseButton(playwright.MouseButtonLeft, "left")
	}
}

func mouseButtonPtr(btn playwright.MouseButton) *playwright.MouseButton {
	value := btn
	return &value
}

func derefMouseButton(ptr *playwright.MouseButton, fallback string) playwright.MouseButton {
	if ptr != nil {
		return *ptr
	}
	return playwright.MouseButton(fallback)
}

func elementStateValue(ptr *playwright.ElementState) playwright.ElementState {
	if ptr != nil {
		return *ptr
	}
	return ""
}

// Route 注册网络请求拦截器
func (p *page) Route(urlPattern string, handler browser.RouteHandler) error {
	return p.p.Route(urlPattern, func(route playwright.Route) {
		// 包装 Playwright Route 为 browser.Route 接口
		wrappedRoute := &playwrightRoute{route: route}
		handler(wrappedRoute)
	})
}

// UnrouteAll 移除所有路由拦截
func (p *page) UnrouteAll() error {
	return p.p.Unroute("**/*")
}

// playwrightRoute 实现 browser.Route 接口
type playwrightRoute struct {
	route playwright.Route
}

func (r *playwrightRoute) Request() browser.Request {
	return &playwrightRequest{req: r.route.Request()}
}

func (r *playwrightRoute) Continue() error {
	return r.route.Continue()
}

func (r *playwrightRoute) Abort() error {
	return r.route.Abort()
}

func (r *playwrightRoute) Fulfill(options browser.FulfillOptions) error {
	return r.route.Fulfill(playwright.RouteFulfillOptions{
		Status:  playwright.Int(options.Status),
		Headers: options.Headers,
		Body:    options.Body,
	})
}

// playwrightRequest 实现 browser.Request 接口
type playwrightRequest struct {
	req playwright.Request
}

func (r *playwrightRequest) URL() string {
	return r.req.URL()
}

func (r *playwrightRequest) Method() string {
	return r.req.Method()
}

func (r *playwrightRequest) Headers() map[string]string {
	return r.req.Headers()
}

func (r *playwrightRequest) PostData() string {
	data, _ := r.req.PostData()
	return data
}
