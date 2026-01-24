# Clean Architecture + Playwright + 自动选择器自愈 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 完成全仓库 Clean Architecture 分层，替换 Rod 为 Playwright（Docker headless），并加入选择器自愈全自动落盘。

**Architecture:** 采用 `internal/domain`/`internal/app`/`internal/infra`/`internal/interfaces` 分层，应用层通过端口依赖基础设施，选择器自愈独立为 `infra/selector`，接口层统一接入 CLI/MCP。

**Tech Stack:** Go 1.24、Playwright-Go、Gin、MCP SDK、YAML 配置、Docker（headless）。

---

### Task 1: 新建 domain 发布模型与校验

**Files:**
- Create: `internal/domain/publish/content.go`
- Create: `internal/domain/publish/validation_test.go`

**Step 1: Write the failing test**

```go
package publish

import "testing"

func TestValidateImageContent_Limits(t *testing.T) {
	limits := Limits{MaxTags: 2, MinImages: 1, MaxImages: 3}
	err := ValidateImageContent(ImageContent{
		Title:      "t",
		Content:    "c",
		Tags:       []string{"a", "b", "c"},
		ImagePaths: []string{"1.jpg"},
	}, limits)
	if err == nil {
		t.Fatalf("expected tag limit error")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/domain/publish -v`  
Expected: FAIL with "undefined: ValidateImageContent"

**Step 3: Write minimal implementation**

```go
package publish

import (
	"fmt"
	"time"
)

type ImageContent struct {
	Title        string
	Content      string
	Tags         []string
	ImagePaths   []string
	ScheduleTime *time.Time
}

type VideoContent struct {
	Title        string
	Content      string
	Tags         []string
	VideoPath    string
	ScheduleTime *time.Time
}

type Limits struct {
	MaxTags   int
	MinImages int
	MaxImages int
}

func ValidateImageContent(c ImageContent, limits Limits) error {
	if len(c.ImagePaths) < limits.MinImages {
		return fmt.Errorf("图片数量不足: %d", len(c.ImagePaths))
	}
	if len(c.ImagePaths) > limits.MaxImages {
		return fmt.Errorf("图片数量过多: %d", len(c.ImagePaths))
	}
	if len(c.Tags) > limits.MaxTags {
		return fmt.Errorf("标签数量过多: %d", len(c.Tags))
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/domain/publish -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/domain/publish/content.go internal/domain/publish/validation_test.go
git commit -m "feat: add publish domain models"
```

---

### Task 2: 定义应用层端口与测试替身

**Files:**
- Create: `internal/app/ports/ports.go`
- Create: `internal/app/testkit/fakes.go`
- Create: `internal/app/testkit/fakes_test.go`

**Step 1: Write the failing test**

```go
package testkit

import (
	"testing"
	"github.com/xpzouying/xiaohongshu-mcp/internal/app/ports"
)

func TestFakesImplementPorts(t *testing.T) {
	var _ ports.PublishGateway = (*FakePublishGateway)(nil)
	var _ ports.SelectorStore = (*FakeSelectorStore)(nil)
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/app/testkit -v`  
Expected: FAIL with "undefined: ports.PublishGateway"

**Step 3: Write minimal implementation**

```go
package ports

import (
	"context"
	"github.com/xpzouying/xiaohongshu-mcp/internal/domain/publish"
)

type PublishGateway interface {
	PublishImage(ctx context.Context, content publish.ImageContent) error
	PublishVideo(ctx context.Context, content publish.VideoContent) error
}

type SelectorStore interface {
	Load() (map[string]string, error)
	Save(selectors map[string]string) error
	Snapshot() (string, error)
	Rollback(snapshot string) error
}
```

```go
package testkit

import (
	"context"
	"github.com/xpzouying/xiaohongshu-mcp/internal/domain/publish"
)

type FakePublishGateway struct {
	ImageCalls int
	VideoCalls int
	LastImage  publish.ImageContent
	LastVideo  publish.VideoContent
	Err        error
}

func (f *FakePublishGateway) PublishImage(ctx context.Context, content publish.ImageContent) error {
	f.ImageCalls++
	f.LastImage = content
	return f.Err
}

func (f *FakePublishGateway) PublishVideo(ctx context.Context, content publish.VideoContent) error {
	f.VideoCalls++
	f.LastVideo = content
	return f.Err
}

type FakeSelectorStore struct {
	Selectors map[string]string
}

func (f *FakeSelectorStore) Load() (map[string]string, error) {
	return f.Selectors, nil
}

func (f *FakeSelectorStore) Save(selectors map[string]string) error {
	f.Selectors = selectors
	return nil
}

func (f *FakeSelectorStore) Snapshot() (string, error) {
	return "snapshot", nil
}

func (f *FakeSelectorStore) Rollback(snapshot string) error {
	return nil
}
```

```go
package testkit

import (
	"testing"
	"github.com/xpzouying/xiaohongshu-mcp/internal/app/ports"
)

func TestFakesImplementPorts(t *testing.T) {
	var _ ports.PublishGateway = (*FakePublishGateway)(nil)
	var _ ports.SelectorStore = (*FakeSelectorStore)(nil)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app/testkit -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/app/ports/ports.go internal/app/testkit/fakes.go internal/app/testkit/fakes_test.go
git commit -m "feat: add app ports and fakes"
```

---

### Task 3: 配置存储（含限制配置）

**Files:**
- Create: `internal/infra/config/store.go`
- Create: `internal/infra/config/store_test.go`
- Create: `internal/infra/config/testdata/config.yaml`

**Step 1: Write the failing test**

```go
package config

import "testing"

func TestLoadConfig_File(t *testing.T) {
	cfg, err := LoadFromFile("testdata/config.yaml")
	if err != nil || cfg.URLs.Creator.PublishImage == "" || cfg.Limits.MaxTags == 0 {
		t.Fatalf("expected publish_image url and limits")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/infra/config -v`  
Expected: FAIL with "undefined: LoadFromFile"

**Step 3: Write minimal implementation**

```go
package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	URLs struct {
		Creator struct {
			PublishImage string `yaml:"publish_image"`
			PublishVideo string `yaml:"publish_video"`
		} `yaml:"creator"`
	} `yaml:"urls"`
	Limits struct {
		MaxTags   int `yaml:"max_tags"`
		MinImages int `yaml:"min_images"`
		MaxImages int `yaml:"max_images"`
	} `yaml:"limits"`
}

func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
```

```yaml
urls:
  creator:
    publish_image: "https://example.com/publish?target=image"
    publish_video: "https://example.com/publish?target=video"
limits:
  max_tags: 10
  min_images: 1
  max_images: 9
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/infra/config -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/infra/config
git commit -m "feat: add config store"
```

---

### Task 4: 选择器存储（快照/回滚）

**Files:**
- Create: `internal/infra/selector/store.go`
- Create: `internal/infra/selector/store_test.go`

**Step 1: Write the failing test**

```go
package selector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStore_SnapshotAndRollback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "selectors.yaml")
	if err := os.WriteFile(path, []byte("a: b\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	store := Store{Path: path}
	snap, err := store.Snapshot()
	if err != nil || snap == "" {
		t.Fatalf("snapshot err: %v", err)
	}
	if err := store.Save(map[string]string{"a": "c"}); err != nil {
		t.Fatalf("save err: %v", err)
	}
	if err := store.Rollback(snap); err != nil {
		t.Fatalf("rollback err: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/infra/selector -v`  
Expected: FAIL with "undefined: Store"

**Step 3: Write minimal implementation**

```go
package selector

import (
	"os"
	"path/filepath"
	"time"
	"gopkg.in/yaml.v3"
)

type Store struct {
	Path string
}

func (s Store) Load() (map[string]string, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string)
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (s Store) Save(selectors map[string]string) error {
	data, err := yaml.Marshal(selectors)
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, data, 0644)
}

func (s Store) Snapshot() (string, error) {
	ts := time.Now().Format("20060102-150405")
	dst := s.Path + "." + ts + ".bak"
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(dst, data, 0644); err != nil {
		return "", err
	}
	return filepath.Base(dst), nil
}

func (s Store) Rollback(snapshot string) error {
	src := filepath.Join(filepath.Dir(s.Path), snapshot)
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, data, 0644)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/infra/selector -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/infra/selector
git commit -m "feat: add selector store"
```

---

### Task 5: 引入 Playwright 依赖与引擎封装

**Files:**
- Modify: `go.mod`
- Create: `internal/infra/browser/engine.go`
- Create: `internal/infra/browser/playwright/engine.go`
- Create: `internal/infra/browser/playwright/engine_test.go`

**Step 1: Write the failing test**

```go
package playwright

import "testing"

func TestEngineConfig_Defaults(t *testing.T) {
	cfg := DefaultConfig()
	if !cfg.Headless {
		t.Fatalf("expected headless true")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/infra/browser/playwright -v`  
Expected: FAIL with "undefined: DefaultConfig"

**Step 3: Install dependency**

Run: `go get github.com/playwright-community/playwright-go@latest`

**Step 4: Write minimal implementation**

```go
package browser

type Page interface {
	Goto(url string) error
	Click(selector string) error
	Fill(selector, value string) error
	SetFiles(selector string, files []string) error
	Text(selector string) (string, error)
	WaitVisible(selector string) error
	Close() error
}

type Engine interface {
	Start() error
	NewPage() (Page, error)
	Close() error
}
```

```go
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
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(e.cfg.Headless),
	})
	if err != nil {
		_ = pw.Stop()
		return err
	}
	e.pw = pw
	e.browser = browser
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
	return p.p.Locator(selector).SetInputFiles(playwright.InputFiles{Files: files})
}

func (p *page) Text(selector string) (string, error) {
	text, err := p.p.Locator(selector).TextContent()
	if err != nil {
		return "", err
	}
	if text == nil {
		return "", nil
	}
	return *text, nil
}

func (p *page) WaitVisible(selector string) error {
	_, err := p.p.Locator(selector).WaitFor()
	return err
}

func (p *page) Close() error {
	return p.p.Close()
}
```

**Step 5: Run test to verify it passes**

Run: `go test ./internal/infra/browser/playwright -v`  
Expected: PASS

**Step 6: Commit**

```bash
git add go.mod go.sum internal/infra/browser internal/infra/browser/playwright
git commit -m "feat: add playwright engine"
```

---

### Task 6: 应用层发布用例（调用端口）

**Files:**
- Create: `internal/app/publish/usecase.go`
- Create: `internal/app/publish/usecase_test.go`

**Step 1: Write the failing test**

```go
package publish

import (
	"context"
	"testing"
	"github.com/xpzouying/xiaohongshu-mcp/internal/app/testkit"
	"github.com/xpzouying/xiaohongshu-mcp/internal/domain/publish"
)

func TestPublishImage_UsesGateway(t *testing.T) {
	gw := &testkit.FakePublishGateway{}
	uc := Usecase{Gateway: gw, Limits: publish.Limits{MaxTags: 10, MinImages: 1, MaxImages: 9}}
	err := uc.PublishImage(context.Background(), publish.ImageContent{ImagePaths: []string{"1.jpg"}})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if gw.ImageCalls != 1 {
		t.Fatalf("expected gateway call")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/app/publish -v`  
Expected: FAIL with "undefined: Usecase"

**Step 3: Write minimal implementation**

```go
package publish

import (
	"context"
	"github.com/xpzouying/xiaohongshu-mcp/internal/app/ports"
	"github.com/xpzouying/xiaohongshu-mcp/internal/domain/publish"
)

type Usecase struct {
	Gateway ports.PublishGateway
	Limits  publish.Limits
}

func (u Usecase) PublishImage(ctx context.Context, content publish.ImageContent) error {
	if err := publish.ValidateImageContent(content, u.Limits); err != nil {
		return err
	}
	return u.Gateway.PublishImage(ctx, content)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app/publish -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/app/publish/usecase.go internal/app/publish/usecase_test.go
git commit -m "feat: add publish usecase"
```

---

### Task 7: 发布网关配置校验与骨架

**Files:**
- Create: `internal/infra/xhs/publish/gateway.go`
- Create: `internal/infra/xhs/publish/gateway_test.go`

**Step 1: Write the failing test**

```go
package publish

import "testing"

func TestNewGateway_ValidatesConfig(t *testing.T) {
	_, err := NewGateway(Config{})
	if err == nil {
		t.Fatalf("expected config error")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/infra/xhs/publish -v`  
Expected: FAIL with "undefined: NewGateway"

**Step 3: Write minimal implementation**

```go
package publish

import (
	"context"
	"errors"
	"github.com/xpzouying/xiaohongshu-mcp/internal/domain/publish"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
)

var ErrNotReady = errors.New("publish not implemented")

type Config struct {
	PublishImageURL string
	PublishVideoURL string
	Selectors       map[string]string
}

type Gateway struct {
	cfg    Config
	engine browser.Engine
}

func NewGateway(cfg Config, engine browser.Engine) (*Gateway, error) {
	if cfg.PublishImageURL == "" || cfg.PublishVideoURL == "" {
		return nil, errors.New("publish url missing")
	}
	if engine == nil {
		return nil, errors.New("engine missing")
	}
	return &Gateway{cfg: cfg, engine: engine}, nil
}

func (g *Gateway) PublishImage(ctx context.Context, content publish.ImageContent) error {
	return ErrNotReady
}

func (g *Gateway) PublishVideo(ctx context.Context, content publish.VideoContent) error {
	return ErrNotReady
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/infra/xhs/publish -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/infra/xhs/publish
git commit -m "feat: add publish gateway skeleton"
```

---

### Task 8: 实现 Playwright 发布流程（图文/视频）

**Files:**
- Modify: `internal/infra/xhs/publish/gateway.go`
- Create: `internal/infra/xhs/publish/gateway_playwright_test.go`

**Step 1: Write the failing test**

```go
package publish

import (
	"context"
	"testing"
	"github.com/xpzouying/xiaohongshu-mcp/internal/domain/publish"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
)

type fakePage struct {
	Calls []string
}

func (p *fakePage) Goto(url string) error               { p.Calls = append(p.Calls, "goto"); return nil }
func (p *fakePage) Click(selector string) error         { p.Calls = append(p.Calls, "click:"+selector); return nil }
func (p *fakePage) Fill(selector, value string) error   { p.Calls = append(p.Calls, "fill:"+selector); return nil }
func (p *fakePage) SetFiles(selector string, files []string) error {
	p.Calls = append(p.Calls, "files:"+selector)
	return nil
}
func (p *fakePage) Text(selector string) (string, error) { return "", nil }
func (p *fakePage) WaitVisible(selector string) error    { return nil }
func (p *fakePage) Close() error                         { return nil }

type fakeEngine struct{ page *fakePage }
func (e *fakeEngine) Start() error                 { return nil }
func (e *fakeEngine) NewPage() (browser.Page, error) { return e.page, nil }
func (e *fakeEngine) Close() error                 { return nil }

func TestGateway_PublishImage_UsesSelectors(t *testing.T) {
	engine := &fakeEngine{page: &fakePage{}}
	cfg := Config{
		PublishImageURL: "https://example.com",
		PublishVideoURL: "https://example.com",
		Selectors: map[string]string{
			"upload_input": "input[type=file]",
			"title_input":  "input[name=title]",
			"content":      "textarea[name=content]",
			"submit":       "button[type=submit]",
		},
	}
	gw, err := NewGateway(cfg, engine)
	if err != nil {
		t.Fatalf("new gateway err: %v", err)
	}
	err = gw.PublishImage(context.Background(), publish.ImageContent{
		Title:      "t",
		Content:    "c",
		ImagePaths: []string{"1.jpg"},
	})
	if err != nil {
		t.Fatalf("publish err: %v", err)
	}
	if len(engine.page.Calls) == 0 {
		t.Fatalf("expected page calls")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/infra/xhs/publish -run TestGateway_PublishImage_UsesSelectors -v`  
Expected: FAIL with "ErrNotReady"

**Step 3: Write minimal implementation**

```go
func (g *Gateway) PublishImage(ctx context.Context, content publish.ImageContent) error {
	if err := g.engine.Start(); err != nil {
		return err
	}
	defer g.engine.Close()

	page, err := g.engine.NewPage()
	if err != nil {
		return err
	}
	defer page.Close()

	if err := page.Goto(g.cfg.PublishImageURL); err != nil {
		return err
	}
	if err := page.SetFiles(g.cfg.Selectors["upload_input"], content.ImagePaths); err != nil {
		return err
	}
	if err := page.Fill(g.cfg.Selectors["title_input"], content.Title); err != nil {
		return err
	}
	if err := page.Fill(g.cfg.Selectors["content"], content.Content); err != nil {
		return err
	}
	return page.Click(g.cfg.Selectors["submit"])
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/infra/xhs/publish -run TestGateway_PublishImage_UsesSelectors -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/infra/xhs/publish/gateway.go internal/infra/xhs/publish/gateway_playwright_test.go
git commit -m "feat: implement publish image via playwright engine"
```

---

### Task 9: 选择器自愈流水线（全自动落盘）

**Files:**
- Create: `internal/infra/selector/pipeline.go`
- Create: `internal/infra/selector/pipeline_test.go`

**Step 1: Write the failing test**

```go
package selector

import (
	"context"
	"testing"
)

type fakeStore struct {
	Selectors map[string]string
	Saved     bool
}

func (f *fakeStore) Load() (map[string]string, error) { return f.Selectors, nil }
func (f *fakeStore) Save(s map[string]string) error   { f.Saved = true; return nil }
func (f *fakeStore) Snapshot() (string, error)        { return "snap", nil }
func (f *fakeStore) Rollback(string) error            { return nil }

type fakeLearner struct{}
func (fakeLearner) Learn(ctx context.Context, current map[string]string) (map[string]string, error) {
	return map[string]string{"a": "b"}, nil
}

type fakeValidator struct{}
func (fakeValidator) Validate(ctx context.Context, selectors map[string]string) error { return nil }

func TestPipeline_Run_SavesSelectors(t *testing.T) {
	p := Pipeline{
		Store:    &fakeStore{Selectors: map[string]string{"a": "a"}},
		Learner:  fakeLearner{},
		Validate: fakeValidator{},
	}
	if err := p.Run(context.Background()); err != nil {
		t.Fatalf("run err: %v", err)
	}
	if !p.Store.(*fakeStore).Saved {
		t.Fatalf("expected Save to be called")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/infra/selector -run TestPipeline_Run_SavesSelectors -v`  
Expected: FAIL with "undefined: Pipeline"

**Step 3: Write minimal implementation**

```go
package selector

import "context"

type Learner interface {
	Learn(ctx context.Context, current map[string]string) (map[string]string, error)
}

type Validator interface {
	Validate(ctx context.Context, selectors map[string]string) error
}

type Store interface {
	Load() (map[string]string, error)
	Save(selectors map[string]string) error
	Snapshot() (string, error)
	Rollback(snapshot string) error
}

type Pipeline struct {
	Store    Store
	Learner  Learner
	Validate Validator
}

func (p Pipeline) Run(ctx context.Context) error {
	current, err := p.Store.Load()
	if err != nil {
		return err
	}
	snap, err := p.Store.Snapshot()
	if err != nil {
		return err
	}
	next, err := p.Learner.Learn(ctx, current)
	if err != nil {
		return err
	}
	if err := p.Validate.Validate(ctx, next); err != nil {
		_ = p.Store.Rollback(snap)
		return err
	}
	return p.Store.Save(next)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/infra/selector -run TestPipeline_Run_SavesSelectors -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/infra/selector/pipeline.go internal/infra/selector/pipeline_test.go
git commit -m "feat: add selector healing pipeline"
```

---

### Task 10: 接口层接入新用例 + Docker headless

**Files:**
- Create: `internal/interfaces/wiring/wiring.go`
- Create: `internal/interfaces/wiring/wiring_test.go`
- Modify: `main.go`
- Modify: `mcp_server.go`
- Modify: `Dockerfile`

**Step 1: Write the failing test**

```go
package wiring

import (
	"testing"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/config"
)

type fakeEngine struct{}
func (fakeEngine) Start() error                 { return nil }
func (fakeEngine) NewPage() (browser.Page, error) { return nil, nil }
func (fakeEngine) Close() error                 { return nil }

func TestBuildPublishUsecase(t *testing.T) {
	cfg := &config.Config{}
	cfg.URLs.Creator.PublishImage = "https://example.com/publish?target=image"
	cfg.URLs.Creator.PublishVideo = "https://example.com/publish?target=video"
	cfg.Limits.MaxTags = 10
	cfg.Limits.MinImages = 1
	cfg.Limits.MaxImages = 9
	engine := fakeEngine{}
	uc, err := BuildPublishUsecase(cfg, map[string]string{}, engine)
	if err != nil || uc == nil {
		t.Fatalf("expected usecase, err=%v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/interfaces/wiring -v`  
Expected: FAIL with "undefined: BuildPublishUsecase"

**Step 3: Write minimal implementation**

```go
package wiring

import (
	"github.com/xpzouying/xiaohongshu-mcp/internal/app/publish"
	"github.com/xpzouying/xiaohongshu-mcp/internal/domain/publish"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/config"
	xhspublish "github.com/xpzouying/xiaohongshu-mcp/internal/infra/xhs/publish"
)

func BuildPublishUsecase(cfg *config.Config, selectors map[string]string, engine browser.Engine) (*publish.Usecase, error) {
	gw, err := xhspublish.NewGateway(xhspublish.Config{
		PublishImageURL: cfg.URLs.Creator.PublishImage,
		PublishVideoURL: cfg.URLs.Creator.PublishVideo,
		Selectors:       selectors,
	}, engine)
	if err != nil {
		return nil, err
	}
	return &publish.Usecase{
		Gateway: gw,
		Limits: publish.Limits{
			MaxTags:   cfg.Limits.MaxTags,
			MinImages: cfg.Limits.MinImages,
			MaxImages: cfg.Limits.MaxImages,
		},
	}, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/interfaces/wiring -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/interfaces/wiring main.go mcp_server.go Dockerfile
git commit -m "feat: wire usecases and docker headless"
```

---

## 注意事项
- 每一步都保持可编译/可测试，避免大规模一次性迁移。
- Playwright 实际交互逻辑建议按“发布图文 -> 发布视频 -> 搜索/feeds”分批替换。
