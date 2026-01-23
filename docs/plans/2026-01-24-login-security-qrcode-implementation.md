# 登录安全认证二维码识别实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在安全认证阶段优先返回 iframe 内二维码，并准确标记 stage=security。

**Architecture:** 在二维码获取时先判定安全阶段，再按阶段决定搜索优先级；安全提示与二维码元素均递归扫描 iframe 树。

**Tech Stack:** Go 1.24, Rod, existing login session abstraction.

### Task 1: 为 iframe 安全阶段优先二维码添加失败用例

**Files:**
- Modify: `login_session_rod_test.go`

**Step 1: Write the failing test**

```go
func TestRodLoginSession_QRCode_SecurityPrefersFrameQRCode(t *testing.T) {
	frame := &fakeQRFrame{
		text: "扫码验证",
		elements: map[string]*fakeQRElement{
			".login-container .qrcode-img": {image: []byte("frame")},
		},
	}
	page := &fakeQRPage{
		text: "",
		elements: map[string]*fakeQRElement{
			".login-container .qrcode-img": {image: []byte("page")},
		},
		frames: []qrFrame{frame},
	}
	session := rodLoginSession{page: page, sleep: func(time.Duration) {}}

	got, err := session.QRCode(context.Background())
	if err != nil {
		t.Fatalf("QRCode err: %v", err)
	}
	if got.Stage != "security" {
		t.Fatalf("expected security stage")
	}
	if got.Image != base64.StdEncoding.EncodeToString([]byte("frame")) {
		t.Fatalf("expected frame qrcode")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./... -run TestRodLoginSession_QRCode_SecurityPrefersFrameQRCode -v`
Expected: FAIL (当前会返回主文档二维码或仍为 login stage)

### Task 2: 扩展测试桩以支持 iframe 结构

**Files:**
- Modify: `login_session_rod_test.go`

**Step 1: Write the failing test**

```go
type fakeQRFrame struct {
	text     string
	elements map[string]*fakeQRElement
	frames   []qrFrame
}

func (f *fakeQRFrame) HasR(_ context.Context, _ string, jsRegex string) (bool, error) {
	if f.text == "" {
		return false, nil
	}
	re, err := regexp.Compile(jsRegex)
	if err != nil {
		return false, err
	}
	return re.MatchString(f.text), nil
}

func (f *fakeQRFrame) Element(_ context.Context, selector string) (qrElement, error) {
	if f.elements == nil {
		return nil, errors.New("not found")
	}
	el, ok := f.elements[selector]
	if !ok {
		return nil, errors.New("not found")
	}
	return el, nil
}

func (f *fakeQRFrame) ElementR(_ context.Context, _ string, _ string) (qrElement, error) {
	return nil, errors.New("not found")
}

func (f *fakeQRFrame) Frames(_ context.Context) ([]qrFrame, error) {
	return f.frames, nil
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./... -run TestRodLoginSession_QRCode_SecurityPrefersFrameQRCode -v`
Expected: FAIL（生产代码尚未支持 iframe）

### Task 3: 生产代码加入 iframe 扫描与安全阶段优先策略

**Files:**
- Modify: `login_session_rod.go`

**Step 1: Write minimal implementation**

```go
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
```

```go
func (s *rodLoginSession) hasSecurityHint(ctx context.Context) bool {
	ok, err := s.page.HasR(ctx, "body", securityHintRegexp)
	if err == nil && ok {
		return true
	}
	return s.frameHasSecurityHint(ctx, s.page)
}

func (s *rodLoginSession) findQRCodeElement(ctx context.Context, preferFrames bool) (qrElement, error) {
	if preferFrames {
		if el, ok := s.findQRCodeElementInFrame(ctx, s.page); ok {
			return el, nil
		}
	}
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
	if !preferFrames {
		if el, ok := s.findQRCodeElementInFrame(ctx, s.page); ok {
			return el, nil
		}
	}
	return nil, errors.New("login qrcode element not found")
}
```

**Step 2: Run test to verify it passes**

Run: `go test ./... -run TestRodLoginSession_QRCode_SecurityPrefersFrameQRCode -v`
Expected: PASS

### Task 4: 回归测试

**Step 1: Run full tests**

Run: `go test ./...`
Expected: PASS

### Task 5: 提交改动

```bash
git add login_session_rod.go login_session_rod_test.go
git commit -m "fix: detect security qrcode inside iframes"
```
