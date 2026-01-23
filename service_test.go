package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	apppublish "github.com/xpzouying/xiaohongshu-mcp/internal/app/publish"
	"github.com/xpzouying/xiaohongshu-mcp/internal/app/testkit"
	domainpublish "github.com/xpzouying/xiaohongshu-mcp/internal/domain/publish"
)

func TestPublishContent_UsesUsecase(t *testing.T) {
	gw := &testkit.FakePublishGateway{}
	uc := apppublish.Usecase{Gateway: gw, Limits: domainpublish.Limits{MaxTags: 10, MinImages: 1, MaxImages: 9}}
	service := NewXiaohongshuServiceWithUsecase(&uc)
	req := &PublishRequest{
		Title:   "t",
		Content: "c",
		Images:  []string{"/tmp/placeholder.jpg"},
	}
	if _, err := service.PublishContent(context.Background(), req); err != nil {
		t.Fatalf("publish err: %v", err)
	}
	if gw.ImageCalls != 1 {
		t.Fatalf("expected gateway call")
	}
}

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
