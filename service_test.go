package main

import (
	"context"
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
