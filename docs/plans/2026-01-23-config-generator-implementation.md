# config.yaml 生成器 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 实现本地 CLI 生成器，自动登录并抓取页面结构，生成并覆盖 `config.yaml`（覆盖前备份、校验失败回滚）。

**Architecture:** 在 `internal/configgen` 中实现 Probe/Infer/Validate/Persist 四阶段流水线；CLI `cmd/config-gen` 负责参数解析与执行；使用 Playwright 完成页面抓取与验证。

**Tech Stack:** Go 1.24、Playwright-Go、YAML（gopkg.in/yaml.v3）。

---

### Task 1: 生成器参数与默认值

**Files:**
- Create: `internal/configgen/options.go`
- Create: `internal/configgen/options_test.go`

**Step 1: Write the failing test**

```go
package configgen

import "testing"

func TestOptions_Defaults(t *testing.T) {
	opt := DefaultOptions()
	if opt.OutputPath == "" || !opt.Backup {
		t.Fatalf("expected default output path and backup=true")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/configgen -run TestOptions_Defaults -v`  
Expected: FAIL with "undefined: DefaultOptions"

**Step 3: Write minimal implementation**

```go
package configgen

type Options struct {
	OutputPath string
	Backup     bool
	Headless   bool
	DryRun     bool
	VerifyOnly bool
}

func DefaultOptions() Options {
	return Options{
		OutputPath: "config.yaml",
		Backup:     true,
		Headless:   true,
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/configgen -run TestOptions_Defaults -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/configgen/options.go internal/configgen/options_test.go
git commit -m "feat: add config generator options"
```

---

### Task 2: 备份与写入（Persist）

**Files:**
- Create: `internal/configgen/persist.go`
- Create: `internal/configgen/persist_test.go`

**Step 1: Write the failing test**

```go
package configgen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPersist_BackupAndWrite(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(target, []byte("old: 1\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	p := Persister{Now: func() string { return "20260123-120000" }}
	if err := p.BackupAndWrite(target, []byte("new: 2\n")); err != nil {
		t.Fatalf("persist: %v", err)
	}
	if _, err := os.Stat(target + ".bak.20260123-120000"); err != nil {
		t.Fatalf("backup missing: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/configgen -run TestPersist_BackupAndWrite -v`  
Expected: FAIL with "undefined: Persister"

**Step 3: Write minimal implementation**

```go
package configgen

import (
	"os"
)

type Persister struct {
	Now func() string
}

func (p Persister) BackupAndWrite(path string, data []byte) error {
	if p.Now == nil {
		p.Now = func() string { return "unknown" }
	}
	if _, err := os.Stat(path); err == nil {
		backup := path + ".bak." + p.Now()
		old, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(backup, old, 0644); err != nil {
			return err
		}
	}
	return os.WriteFile(path, data, 0644)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/configgen -run TestPersist_BackupAndWrite -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/configgen/persist.go internal/configgen/persist_test.go
git commit -m "feat: add config generator persister"
```

---

### Task 3: 快照结构与选择器推断（Infer）

**Files:**
- Create: `internal/configgen/snapshot.go`
- Create: `internal/configgen/infer.go`
- Create: `internal/configgen/infer_test.go`

**Step 1: Write the failing test**

```go
package configgen

import "testing"

func TestInfer_UsesAriaLabel(t *testing.T) {
	snap := Snapshot{
		Nodes: []Node{
			{Tag: "button", Text: "发布", Attrs: map[string]string{"aria-label": "发布"}},
		},
	}
	inf := NewInfer()
	out := inf.Infer(snap)
	if out["submit"] == "" {
		t.Fatalf("expected submit selector")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/configgen -run TestInfer_UsesAriaLabel -v`  
Expected: FAIL with "undefined: Snapshot"

**Step 3: Write minimal implementation**

```go
package configgen

type Node struct {
	Tag   string
	Text  string
	Attrs map[string]string
}

type Snapshot struct {
	Nodes []Node
}
```

```go
package configgen

type Infer struct{}

func NewInfer() Infer { return Infer{} }

func (i Infer) Infer(s Snapshot) map[string]string {
	out := make(map[string]string)
	for _, n := range s.Nodes {
		if n.Attrs["aria-label"] == "发布" {
			out["submit"] = "[aria-label='发布']"
		}
	}
	return out
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/configgen -run TestInfer_UsesAriaLabel -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/configgen/snapshot.go internal/configgen/infer.go internal/configgen/infer_test.go
git commit -m "feat: add selector inference skeleton"
```

---

### Task 4: 选择器校验器（Validate）

**Files:**
- Create: `internal/configgen/validator.go`
- Create: `internal/configgen/validator_test.go`

**Step 1: Write the failing test**

```go
package configgen

import "testing"

type fakePage struct{ calls int }
func (p *fakePage) WaitVisible(selector string) error { p.calls++; return nil }

func TestValidator_ChecksRequiredSelectors(t *testing.T) {
	v := Validator{Required: []string{"submit"}}
	if err := v.Validate(map[string]string{"submit": "button"}, &fakePage{}); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/configgen -run TestValidator_ChecksRequiredSelectors -v`  
Expected: FAIL with "undefined: Validator"

**Step 3: Write minimal implementation**

```go
package configgen

import "fmt"

type PageWaiter interface {
	WaitVisible(selector string) error
}

type Validator struct {
	Required []string
}

func (v Validator) Validate(selectors map[string]string, page PageWaiter) error {
	for _, key := range v.Required {
		selector := selectors[key]
		if selector == "" {
			return fmt.Errorf("missing selector: %s", key)
		}
		if err := page.WaitVisible(selector); err != nil {
			return err
		}
	}
	return nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/configgen -run TestValidator_ChecksRequiredSelectors -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/configgen/validator.go internal/configgen/validator_test.go
git commit -m "feat: add selector validator"
```

---

### Task 5: 自动登录与 cookies 保存（Login）

**Files:**
- Create: `internal/configgen/login.go`
- Create: `internal/configgen/login_test.go`

**Step 1: Write the failing test**

```go
package configgen

import "testing"

type fakeCookies struct{ saved bool }
func (f *fakeCookies) Save(path string) error { f.saved = true; return nil }

func TestLoginFlow_SavesCookies(t *testing.T) {
	f := &fakeCookies{}
	lf := LoginFlow{CookieSaver: f}
	if err := lf.SaveCookies("/tmp/cookies.json"); err != nil {
		t.Fatalf("save err: %v", err)
	}
	if !f.saved {
		t.Fatalf("expected cookies saved")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/configgen -run TestLoginFlow_SavesCookies -v`  
Expected: FAIL with "undefined: LoginFlow"

**Step 3: Write minimal implementation**

```go
package configgen

type CookieSaver interface {
	Save(path string) error
}

type LoginFlow struct {
	CookieSaver CookieSaver
}

func (l LoginFlow) SaveCookies(path string) error {
	if l.CookieSaver == nil {
		return nil
	}
	return l.CookieSaver.Save(path)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/configgen -run TestLoginFlow_SavesCookies -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/configgen/login.go internal/configgen/login_test.go
git commit -m "feat: add login flow skeleton"
```

---

### Task 6: 生成器流水线（Runner）

**Files:**
- Create: `internal/configgen/runner.go`
- Create: `internal/configgen/runner_test.go`

**Step 1: Write the failing test**

```go
package configgen

import "testing"

type fakeProbe struct{ called bool }
func (f *fakeProbe) Probe() Snapshot { f.called = true; return Snapshot{} }

type fakeInfer struct{ called bool }
func (f *fakeInfer) Infer(s Snapshot) map[string]string { f.called = true; return map[string]string{} }

type fakePersist struct{ called bool }
func (f *fakePersist) Write(path string, data []byte) error { f.called = true; return nil }

func TestRunner_ExecutesPipeline(t *testing.T) {
	r := Runner{
		Probe:   &fakeProbe{},
		Infer:   &fakeInfer{},
		Persist: &fakePersist{},
	}
	if err := r.Run("config.yaml"); err != nil {
		t.Fatalf("run err: %v", err)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/configgen -run TestRunner_ExecutesPipeline -v`  
Expected: FAIL with "undefined: Runner"

**Step 3: Write minimal implementation**

```go
package configgen

type Prober interface {
	Probe() Snapshot
}

type Inferer interface {
	Infer(Snapshot) map[string]string
}

type Persist interface {
	Write(path string, data []byte) error
}

type Runner struct {
	Probe   Prober
	Infer   Inferer
	Persist Persist
}

func (r Runner) Run(path string) error {
	snap := r.Probe.Probe()
	selectors := r.Infer.Infer(snap)
	_ = selectors
	return r.Persist.Write(path, []byte("selectors: {}\n"))
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/configgen -run TestRunner_ExecutesPipeline -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/configgen/runner.go internal/configgen/runner_test.go
git commit -m "feat: add config generator runner"
```

---

### Task 7: CLI 入口（cmd/config-gen）

**Files:**
- Create: `cmd/config-gen/main.go`
- Create: `cmd/config-gen/main_test.go`

**Step 1: Write the failing test**

```go
package main

import "testing"

func TestParseFlags_Defaults(t *testing.T) {
	opt := parseFlags([]string{})
	if opt.OutputPath == \"\" || !opt.Backup {
		t.Fatalf(\"expected defaults\")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd/config-gen -run TestParseFlags_Defaults -v`  
Expected: FAIL with "undefined: parseFlags"

**Step 3: Write minimal implementation**

```go
package main

import (
	"flag"
	"os"

	"github.com/xpzouying/xiaohongshu-mcp/internal/configgen"
)

func parseFlags(args []string) configgen.Options {
	fs := flag.NewFlagSet("config-gen", flag.ContinueOnError)
	opt := configgen.DefaultOptions()
	fs.StringVar(&opt.OutputPath, "output", opt.OutputPath, "输出配置文件路径")
	fs.BoolVar(&opt.Backup, "backup", opt.Backup, "是否备份旧配置")
	fs.BoolVar(&opt.Headless, "headless", opt.Headless, "是否无头模式")
	fs.BoolVar(&opt.DryRun, "dry-run", opt.DryRun, "仅生成报告不写入")
	fs.BoolVar(&opt.VerifyOnly, "verify-only", opt.VerifyOnly, "仅校验不写入")
	_ = fs.Parse(args)
	return opt
}

func main() {
	_ = parseFlags(os.Args[1:])
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./cmd/config-gen -run TestParseFlags_Defaults -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/config-gen/main.go cmd/config-gen/main_test.go
git commit -m "feat: add config generator cli"
```

---

### Task 8: 文档与使用说明

**Files:**
- Modify: `README_CONFIG.md`

**Step 1: Write the failing test**

```text
无自动化测试，新增手动验证说明。
```

**Step 2: Run test to verify it fails**

Skip

**Step 3: Write minimal documentation**

```markdown
## 配置生成器
go run cmd/config-gen/main.go --headless=false --output=config.yaml
```

**Step 4: Manual verification**

Run: `go run cmd/config-gen/main.go --headless=false --output=config.yaml`  
Expected: 浏览器打开并提示扫码登录，最终写入并备份配置

**Step 5: Commit**

```bash
git add README_CONFIG.md
git commit -m "docs: add config generator usage"
```

---

## 注意事项
- 生成器运行需要本地 Playwright 浏览器依赖；Docker 模式需提前安装。
- 覆盖写入前务必备份；校验失败需回滚。
