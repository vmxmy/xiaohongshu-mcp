# MCP 登录二维码多阶段支持 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 让 `get_login_qrcode` 在无 GUI 环境下支持“登录二维码 + 安全认证二维码”的连续扫码，并通过 Inspector 显示当前二维码。

**Architecture:** 新增登录会话管理器，保持单一浏览器会话并按阶段返回二维码截图；登录成功后保存 cookies 并关闭会话。MCP 工具根据阶段输出不同提示文本。

**Tech Stack:** Go, Rod, Gin, MCP SDK

---

### Task 1: 登录会话管理器与单测

**Files:**
- Create: `login_session.go`
- Create: `login_session_test.go`

**Step 1: Write the failing test**

```go
func TestLoginManager_ReturnsQRCodeAndKeepsSession(t *testing.T) {
	clock := time.Date(2026, 1, 23, 10, 0, 0, 0, time.UTC)
	s := &fakeLoginSession{qr: loginQRCode{Image: "img", Stage: "login"}}
	m := NewLoginManager(func() (loginSession, error) { return s, nil }, 4*time.Minute)
	m.now = func() time.Time { return clock }

	got, err := m.GetQRCode(context.Background())
	if err != nil {
		t.Fatalf("GetQRCode err: %v", err)
	}
	if got.Image != "img" || got.Stage != "login" || got.IsLoggedIn {
		t.Fatalf("unexpected qr result: %+v", got)
	}
	if s.openCalls != 1 || s.qrCalls != 1 {
		t.Fatalf("expected session used once")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./... -run TestLoginManager_ReturnsQRCodeAndKeepsSession -v`  
Expected: FAIL with "undefined: NewLoginManager"

**Step 3: Write minimal implementation**

```go
type loginSession interface {
	Open(ctx context.Context) error
	LoggedIn(ctx context.Context) (bool, error)
	QRCode(ctx context.Context) (loginQRCode, error)
	SaveCookies() error
	Close() error
}

type loginQRCode struct {
	Image string
	Stage string
}

type loginQRResult struct {
	LoginQRCodeResponse
	Stage string
}

type LoginManager struct {
	mu         sync.Mutex
	session    loginSession
	newSession func() (loginSession, error)
	ttl        time.Duration
	now        func() time.Time
	openedAt   time.Time
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./... -run TestLoginManager_ReturnsQRCodeAndKeepsSession -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add login_session.go login_session_test.go
git commit -m "feat: add login session manager for qrcode flow"
```

---

### Task 2: 登录完成与过期逻辑单测 + 实现

**Files:**
- Modify: `login_session.go`
- Modify: `login_session_test.go`

**Step 1: Write the failing test**

```go
func TestLoginManager_SavesCookiesAndClosesOnLogin(t *testing.T) {
	s := &fakeLoginSession{loggedIn: true}
	m := NewLoginManager(func() (loginSession, error) { return s, nil }, 4*time.Minute)
	_, err := m.GetQRCode(context.Background())
	if err != nil {
		t.Fatalf("GetQRCode err: %v", err)
	}
	if !s.saved || !s.closed {
		t.Fatalf("expected cookies saved and session closed")
	}
}

func TestLoginManager_ExpiresSession(t *testing.T) {
	clock := time.Date(2026, 1, 23, 10, 0, 0, 0, time.UTC)
	s1 := &fakeLoginSession{qr: loginQRCode{Image: "a", Stage: "login"}}
	s2 := &fakeLoginSession{qr: loginQRCode{Image: "b", Stage: "login"}}
	calls := 0
	m := NewLoginManager(func() (loginSession, error) {
		calls++
		if calls == 1 {
			return s1, nil
		}
		return s2, nil
	}, 4*time.Minute)
	m.now = func() time.Time { return clock }
	_, _ = m.GetQRCode(context.Background())
	m.now = func() time.Time { return clock.Add(5 * time.Minute) }
	got, err := m.GetQRCode(context.Background())
	if err != nil {
		t.Fatalf("GetQRCode err: %v", err)
	}
	if got.Image != "b" || calls != 2 {
		t.Fatalf("expected new session after ttl")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./... -run TestLoginManager_SavesCookiesAndClosesOnLogin -v`  
Expected: FAIL with "expected cookies saved"

**Step 3: Write minimal implementation**

```go
func (m *LoginManager) GetQRCode(ctx context.Context) (loginQRResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.now == nil {
		m.now = time.Now
	}

	if m.session == nil || m.expiredLocked() {
		_ = m.closeLocked()
		s, err := m.newSession()
		if err != nil {
			return loginQRResult{}, err
		}
		m.session = s
		m.openedAt = m.now()
	}

	if err := m.session.Open(ctx); err != nil {
		_ = m.closeLocked()
		return loginQRResult{}, err
	}

	loggedIn, err := m.session.LoggedIn(ctx)
	if err != nil {
		_ = m.closeLocked()
		return loginQRResult{}, err
	}
	if loggedIn {
		_ = m.session.SaveCookies()
		_ = m.closeLocked()
		return loginQRResult{LoginQRCodeResponse: LoginQRCodeResponse{IsLoggedIn: true}}, nil
	}

	qr, err := m.session.QRCode(ctx)
	if err != nil {
		_ = m.closeLocked()
		return loginQRResult{}, err
	}

	remain := m.ttl - m.now().Sub(m.openedAt)
	if remain < 0 {
		remain = 0
	}
	return loginQRResult{
		LoginQRCodeResponse: LoginQRCodeResponse{
			Timeout:    remain.String(),
			IsLoggedIn: false,
			Img:        qr.Image,
		},
		Stage: qr.Stage,
	}, nil
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./... -run TestLoginManager_ -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add login_session.go login_session_test.go
git commit -m "feat: handle login completion and ttl in login manager"
```

---

### Task 3: Rod 登录会话实现（二维码截图）

**Files:**
- Create: `login_session_rod.go`
- Modify: `xiaohongshu/login.go`

**Step 1: Write the failing test**

这部分依赖浏览器，不写单测，改为保守实现 + 手动验证。

**Step 2: Implement minimal rod session**

```go
type rodLoginSession struct {
	browser *headless_browser.Browser
	page    *rod.Page
	opened  bool
}

func newRodLoginSession() (loginSession, error) {
	b := newBrowser()
	p := b.NewPage()
	return &rodLoginSession{browser: b, page: p}, nil
}
```

**Step 3: Implement QR detection**

```go
func (s *rodLoginSession) QRCode(ctx context.Context) (loginQRCode, error) {
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
		Image: "data:image/png;base64," + base64.StdEncoding.EncodeToString(img),
		Stage: stage,
	}, nil
}
```

**Step 4: Manual check**

Run: `go run .`  
Expected: MCP `get_login_qrcode` 返回图片；扫码后再次调用可出现安全认证二维码。

**Step 5: Commit**

```bash
git add login_session_rod.go xiaohongshu/login.go
git commit -m "feat: add rod login session qrcode capture"
```

---

### Task 4: 接入服务与 MCP 文案

**Files:**
- Modify: `service.go`
- Modify: `types.go`
- Modify: `mcp_handlers.go`
- Modify: `README.md`

**Step 1: Write the failing test**

```go
func TestLoginQrcodeHandler_TextForSecurityStage(t *testing.T) {
	// 伪造 loginManager 返回 Stage=security，断言提示文本包含“安全认证”
}
```

**Step 2: Implement wiring**

```go
type LoginQrcodeResponse struct {
	Timeout    string `json:"timeout"`
	IsLoggedIn bool   `json:"is_logged_in"`
	Img        string `json:"img,omitempty"`
	Stage      string `json:"stage,omitempty"`
}

func NewXiaohongshuServiceWithUsecase(publishUsecase *apppublish.Usecase) *XiaohongshuService {
	return &XiaohongshuService{
		publishUsecase: publishUsecase,
		loginManager:   NewLoginManager(newRodLoginSession, 4*time.Minute),
	}
}
```

**Step 3: Run tests**

Run: `go test ./... -run TestLoginQrcodeHandler_TextForSecurityStage -v`  
Expected: PASS

**Step 4: Commit**

```bash
git add service.go types.go mcp_handlers.go README.md
git commit -m "feat: expose security qrcode stage in mcp login"
```

---

### Task 5: 轻量文档补充（无 GUI + Inspector）

**Files:**
- Modify: `README.md`

**Step 1: Edit doc**

```markdown
- `get_login_qrcode` 可能返回“登录二维码 / 安全认证二维码”，扫码后请再次调用直到提示已登录。
- 无 GUI 环境可用 `npx @modelcontextprotocol/inspector` 查看二维码图片。
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add headless login qrcode guidance"
```
