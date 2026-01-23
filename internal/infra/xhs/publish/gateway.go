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
