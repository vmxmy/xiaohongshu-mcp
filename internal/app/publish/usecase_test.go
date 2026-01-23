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
