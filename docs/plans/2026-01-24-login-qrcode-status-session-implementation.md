# Login QR Status & Session Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为 `get_login_qrcode` 输出补充 `status` 与 `session_id`，并在 MCP 文本中明确展示登录阶段与状态，便于诊断二次安全认证流程。

**Architecture:** 在登录会话管理器内生成并持有稳定 `session_id`，每次调用返回 `status` 与 `stage`；服务层补齐 JSON 字段，MCP 层负责格式化输出。所有状态通过常量统一，避免散落。

**Tech Stack:** Go, Rod, MCP SDK, Gin

---

### Task 1: 登录会话结果字段测试（status/session_id）

**Files:**
- Modify: `login_session_test.go`

**Step 1: Write the failing test**

```go
func TestLoginManager_ReturnsStatusAndSessionID(t *testing.T) {
	clock := time.Date(2026, 1, 24, 10, 0, 0, 0, time.UTC)
	s := &fakeLoginSession{qr: loginQRCode{Image: "img", Stage: "login"}}
	m := NewLoginManager(func() (loginSession, error) { return s, nil }, 4*time.Minute)
	m.now = func() time.Time { return clock }
	m.newSessionID = func() string { return "sess-1" }

	got, err := m.GetQRCode(context.Background())
	if err != nil {
		t.Fatalf("GetQRCode err: %v", err)
	}
	if got.Status != loginStatusLoginRequired || got.SessionID != "sess-1" {
		t.Fatalf("unexpected status/session: %+v", got)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./... -run TestLoginManager_ReturnsStatusAndSessionID -v`  
Expected: FAIL with "undefined: loginStatusLoginRequired" or missing fields

**Step 3: Write minimal implementation**

Implementation in Task 2.

**Step 4: Run test to verify it passes**

Run: `go test ./... -run TestLoginManager_ReturnsStatusAndSessionID -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add login_session_test.go
git commit -m "test: cover login qrcode status and session id"
```

---

### Task 2: 登录会话管理器增加 status/session_id

**Files:**
- Modify: `login_session.go`
- Modify: `login_session_test.go`

**Step 1: Write the failing test**

```go
func TestLoginManager_SessionIDChangesOnExpire(t *testing.T) {
	clock := time.Date(2026, 1, 24, 10, 0, 0, 0, time.UTC)
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
	ids := []string{"sess-1", "sess-2"}
	m.newSessionID = func() string {
		id := ids[0]
		ids = ids[1:]
		return id
	}
	m.now = func() time.Time { return clock }

	_, _ = m.GetQRCode(context.Background())
	m.now = func() time.Time { return clock.Add(5 * time.Minute) }
	got, err := m.GetQRCode(context.Background())
	if err != nil {
		t.Fatalf("GetQRCode err: %v", err)
	}
	if got.SessionID != "sess-2" {
		t.Fatalf("expected new session id")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./... -run TestLoginManager_SessionIDChangesOnExpire -v`  
Expected: FAIL with missing `SessionID`

**Step 3: Write minimal implementation**

```go
const (
	loginStatusLoggedIn       = "logged_in"
	loginStatusLoginRequired  = "login_required"
	loginStatusSecurityNeeded = "security_required"
)

type LoginManager struct {
	// ...
	sessionID   string
	newSessionID func() string
}
```

- 新会话创建时生成 `sessionID`
- 关闭/过期时清空
- `GetQRCode` 返回 `Status` 与 `SessionID`

**Step 4: Run test to verify it passes**

Run: `go test ./... -run TestLoginManager_ -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add login_session.go login_session_test.go
git commit -m "feat: add status and session id to login qrcode flow"
```

---

### Task 3: 服务层与 MCP 文案展示 status/session_id

**Files:**
- Modify: `service.go`
- Modify: `mcp_handlers.go`
- Modify: `mcp_handlers_test.go`

**Step 1: Write the failing test**

```go
func TestLoginQrcodeHandler_TextIncludesStatusAndSession(t *testing.T) {
	service := &XiaohongshuService{
		loginManager: fakeLoginProvider{
			result: loginQRResult{
				LoginQrcodeResponse: LoginQrcodeResponse{
					Timeout:    "4m0s",
					IsLoggedIn: false,
					Img:        "img",
					Status:     loginStatusSecurityNeeded,
					SessionID:  "sess-1",
				},
				Stage: "security",
			},
		},
	}
	app := &AppServer{xiaohongshuService: service}

	result := app.handleGetLoginQrcode(context.Background())
	if result == nil || len(result.Content) == 0 {
		t.Fatalf("expected content")
	}
	if !strings.Contains(result.Content[0].Text, "状态:") {
		t.Fatalf("expected status text")
	}
	if !strings.Contains(result.Content[0].Text, "sess-1") {
		t.Fatalf("expected session id")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./... -run TestLoginQrcodeHandler_TextIncludesStatusAndSession -v`  
Expected: FAIL with missing fields

**Step 3: Write minimal implementation**

- `LoginQrcodeResponse` 增加 `Status`、`SessionID`
- `GetLoginQrcode` 组装 status/session_id
- `mcp_handlers.go` 文本追加 `状态` 与 `会话`

**Step 4: Run test to verify it passes**

Run: `go test ./... -run TestLoginQrcodeHandler_TextIncludesStatusAndSession -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add service.go mcp_handlers.go mcp_handlers_test.go
git commit -m "feat: show login status and session id in qrcode tool"
```

---

### Task 4: 文档补充

**Files:**
- Modify: `README.md`

**Step 1: Edit doc**

```markdown
- `get_login_qrcode` 返回 `stage/status/session_id`，可用于判断是否进入安全认证阶段。
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: mention login qrcode status and session id"
```
