# Sync Cookies MCP Tool Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 新增 `sync_cookies` MCP 工具，用于通过 inspector 上传 cookies 数据并写入服务端 cookies 文件。

**Architecture:** 在 MCP 层新增 `sync_cookies` 工具与参数结构体；处理函数解析 base64/JSON 输入、做 JSON 校验并调用服务层写入 cookies。服务层仅负责落盘并返回路径与字节数，日志不输出敏感内容。

**Tech Stack:** Go 1.24, MCP SDK, cookies 本地文件存储。

### Task 1: 解析 cookies 输入的失败用例

**Files:**
- Modify: `mcp_handlers_test.go`

**Step 1: Write the failing test**

```go
func TestParseSyncCookiesPayload_Base64(t *testing.T) {
	data := []byte(`[{"name":"a"}]`)
	args := SyncCookiesArgs{CookiesBase64: base64.StdEncoding.EncodeToString(data)}

	got, err := parseSyncCookiesPayload(args)
	if err != nil {
		t.Fatalf("parse err: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("unexpected payload")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./... -run TestParseSyncCookiesPayload_Base64 -v`
Expected: FAIL (undefined: SyncCookiesArgs / parseSyncCookiesPayload)

### Task 2: JSON 与缺失输入用例

**Files:**
- Modify: `mcp_handlers_test.go`

**Step 1: Write the failing test**

```go
func TestParseSyncCookiesPayload_JSON(t *testing.T) {
	data := []byte(`[{"name":"a"}]`)
	args := SyncCookiesArgs{CookiesJSON: string(data)}

	got, err := parseSyncCookiesPayload(args)
	if err != nil {
		t.Fatalf("parse err: %v", err)
	}
	if string(got) != string(data) {
		t.Fatalf("unexpected payload")
	}
}

func TestParseSyncCookiesPayload_Missing(t *testing.T) {
	_, err := parseSyncCookiesPayload(SyncCookiesArgs{})
	if err == nil {
		t.Fatalf("expected error")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./... -run TestParseSyncCookiesPayload_JSON -v`
Expected: FAIL (undefined: SyncCookiesArgs / parseSyncCookiesPayload)

### Task 3: cookies 写入服务层用例

**Files:**
- Modify: `service_test.go`

**Step 1: Write the failing test**

```go
func TestSyncCookies_WritesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cookies.json")
	os.Setenv("COOKIES_PATH", path)
	t.Cleanup(func() { os.Unsetenv("COOKIES_PATH") })

	service := NewXiaohongshuService()
	data := []byte(`[{"name":"a"}]`)
	gotPath, gotSize, err := service.SyncCookies(context.Background(), data)
	if err != nil {
		t.Fatalf("sync err: %v", err)
	}
	if gotPath != path {
		t.Fatalf("unexpected path: %s", gotPath)
	}
	if gotSize != int64(len(data)) {
		t.Fatalf("unexpected size: %d", gotSize)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read err: %v", err)
	}
	if string(content) != string(data) {
		t.Fatalf("unexpected content")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./... -run TestSyncCookies_WritesFile -v`
Expected: FAIL (undefined: SyncCookies)

### Task 4: 实现参数解析与 JSON 校验

**Files:**
- Modify: `mcp_handlers.go`
- Modify: `mcp_server.go`

**Step 1: Write minimal implementation**

```go
// in mcp_server.go
// SyncCookiesArgs 上传 cookies 参数
// CookiesBase64 优先级高于 CookiesJSON
// CookiesJSON 直接 JSON 字符串
// jsonschema 描述为 inspector 提示用

type SyncCookiesArgs struct {
	CookiesBase64 string `json:"cookies_base64,omitempty" jsonschema:"Base64 编码的 cookies JSON（推荐）"`
	CookiesJSON   string `json:"cookies_json,omitempty" jsonschema:"cookies JSON 字符串（备用）"`
}
```

```go
// in mcp_handlers.go
func parseSyncCookiesPayload(args SyncCookiesArgs) ([]byte, error) {
	if strings.TrimSpace(args.CookiesBase64) != "" {
		data, err := base64.StdEncoding.DecodeString(args.CookiesBase64)
		if err != nil {
			return nil, fmt.Errorf("cookies_base64 解码失败: %w", err)
		}
		return data, nil
	}
	if strings.TrimSpace(args.CookiesJSON) != "" {
		return []byte(args.CookiesJSON), nil
	}
	return nil, errors.New("cookies_base64 或 cookies_json 至少提供一个")
}

func validateCookiesJSON(data []byte) error {
	var items []map[string]any
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("cookies JSON 无法解析: %w", err)
	}
	return nil
}
```

**Step 2: Run tests to verify they pass**

Run: `go test ./... -run TestParseSyncCookiesPayload_ -v`
Expected: PASS

### Task 5: 服务层写入与 MCP 工具注册

**Files:**
- Modify: `service.go`
- Modify: `mcp_handlers.go`
- Modify: `mcp_server.go`
- Modify: `docs/TOOLS.md`

**Step 1: Write minimal implementation**

```go
// in service.go
func (s *XiaohongshuService) SyncCookies(ctx context.Context, data []byte) (string, int64, error) {
	cookiePath := cookies.GetCookiesFilePath()
	loader := cookies.NewLoadCookie(cookiePath)
	if err := loader.SaveCookies(data); err != nil {
		return "", 0, err
	}
	return cookiePath, int64(len(data)), nil
}
```

```go
// in mcp_handlers.go
func (s *AppServer) handleSyncCookies(ctx context.Context, args SyncCookiesArgs) *MCPToolResult {
	logrus.Info("MCP: 上传 cookies")
	payload, err := parseSyncCookiesPayload(args)
	if err != nil {
		return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: "cookies 输入无效: " + err.Error()}}, IsError: true}
	}
	if err := validateCookiesJSON(payload); err != nil {
		return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: "cookies JSON 校验失败: " + err.Error()}}, IsError: true}
	}
	path, size, err := s.xiaohongshuService.SyncCookies(ctx, payload)
	if err != nil {
		return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: "保存 cookies 失败: " + err.Error()}}, IsError: true}
	}
	logrus.WithFields(logrus.Fields{"path": path, "bytes": size}).Info("cookies 已写入")
	return &MCPToolResult{Content: []MCPContent{{Type: "text", Text: fmt.Sprintf("cookies 已写入: %s (%d bytes)", path, size)}}}
}
```

```go
// in mcp_server.go registerTools
mcp.AddTool(server,
	&mcp.Tool{
		Name:        "sync_cookies",
		Description: "上传 cookies JSON 并写入服务端文件（推荐先本地有头登录后上传）",
		Annotations: &mcp.ToolAnnotations{Title: "Sync Cookies"},
	},
	withPanicRecovery("sync_cookies", func(ctx context.Context, req *mcp.CallToolRequest, args SyncCookiesArgs) (*mcp.CallToolResult, any, error) {
		result := appServer.handleSyncCookies(ctx, args)
		return convertToMCPResult(result), nil, nil
	}),
)
```

**Step 2: Update docs**

在 `docs/TOOLS.md` 添加 `sync_cookies` 用法与示例（Base64 优先，JSON 备用），并提示仅打印长度不打印内容。

**Step 3: Run full tests**

Run: `go test ./...`
Expected: PASS

### Task 6: 提交改动

```bash
git add mcp_server.go mcp_handlers.go service.go docs/TOOLS.md mcp_handlers_test.go service_test.go
git commit -m "feat: add sync cookies tool"
```
