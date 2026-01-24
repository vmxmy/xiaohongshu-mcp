package browser

import (
	"context"
	"time"
)

// Page 定义浏览器页面需要暴露的核心能力，供业务无差别使用。
type Page interface {
	// --- 导航 & 生命周期 ---
	Goto(url string) error
	Reload() error
	WaitLoad() error
	WaitIdle() error
	WaitDOMStable(maxWait time.Duration, stabilityThreshold float64) error
	URL() string
	Close() error

	// --- 选择器等待 ---
	WaitVisible(selector string) error
	WaitHidden(selector string) error
	WaitForSelector(selector string, timeout time.Duration) error
	WaitForFunction(expression string, timeout time.Duration) error

	// --- 基础交互 ---
	Click(selector string) error
	ClickForce(selector string) error
	DoubleClick(selector string) error
	Fill(selector, value string) error
	Type(selector, value string) error
	SetFiles(selector string, files []string) error
	Hover(selector string) error
	Focus(selector string) error
	Press(key string) error
	ScrollIntoView(selector string) error
	ScrollBy(deltaX, deltaY float64) error
	IsVisible(selector string) (bool, error)
	Text(selector string) (string, error)
	HTML(selector string) (string, error)
	Attribute(selector, name string) (string, error)

	// --- Eval / 截图 ---
	Eval(expression string, args ...interface{}) (interface{}, error)
	EvalOnSelector(selector, expression string, args ...interface{}) (interface{}, error)
	Screenshot(path string) error
	ScreenshotFullPage(path string) error

	// --- 元素查询 ---
	Element(selector string) (Element, error)
	ElementByRegex(selector, jsRegex string) (Element, error)
	Elements(selector string) ([]Element, error)
	Has(selector string) (bool, error)
	HasRegex(selector, jsRegex string) (bool, error)

	// --- 输入设备 ---
	Mouse() Mouse
	Keyboard() Keyboard

	// --- 上下文控制 ---
	WithContext(ctx context.Context) Page
	WithTimeout(timeout time.Duration) Page
}

// Element 抽象单个 DOM 元素，封装常用操作。
type Element interface {
	Click() error
	ClickForce() error
	DoubleClick() error
	Hover() error
	Focus() error
	Fill(value string) error
	Type(value string) error
	Press(key string) error
	Input(value string) error
	SetFiles(files []string) error
	ScrollIntoView() error
	WaitVisible() error
	WaitHidden() error
	WaitStable(duration time.Duration) error
	IsVisible() (bool, error)
	Text() (string, error)
	HTML() (string, error)
	Attribute(name string) (string, error)
	Value() (string, error)
	Eval(expression string, args ...interface{}) (interface{}, error)
	Element(selector string) (Element, error)
	Elements(selector string) ([]Element, error)
	Remove() error
	BoundingBox() (*BoundingBox, error)
	Frame() (Page, error)
}

// BoundingBox 描述元素的位置信息，方便鼠标定位等场景。
type BoundingBox struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// Mouse 提供鼠标级别的交互封装。
type Mouse interface {
	MoveTo(x, y float64) error
	Click(button MouseButton, opts ...MouseClickOption) error
	Down(button MouseButton) error
	Up(button MouseButton) error
	Wheel(deltaX, deltaY float64) error
}

// MouseButton 表示鼠标按键。
type MouseButton string

const (
	MouseButtonLeft   MouseButton = "left"
	MouseButtonRight  MouseButton = "right"
	MouseButtonMiddle MouseButton = "middle"
)

// MouseClickOption 自定义点击参数。
type MouseClickOption struct {
	ClickCount int
	Delay      time.Duration
	Force      bool
}

// Keyboard 封装键盘输入能力。
type Keyboard interface {
	Type(text string) error
	InsertText(text string) error
	Press(key string) error
	Down(key string) error
	Up(key string) error
}

// Engine 负责启动/关闭浏览器实例并创建 Page。
type Engine interface {
	Start() error
	NewPage() (Page, error)
	Close() error
}
