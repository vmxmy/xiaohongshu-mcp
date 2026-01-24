package wiring

import (
	"testing"

	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/config"
)

type fakeEngine struct{}

func (fakeEngine) Start() error                   { return nil }
func (fakeEngine) NewPage() (browser.Page, error) { return nil, nil }
func (fakeEngine) Close() error                   { return nil }

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
