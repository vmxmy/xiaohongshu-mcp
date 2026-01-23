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
	if err := g.engine.Start(); err != nil {
		return err
	}
	defer g.engine.Close()

	page, err := g.engine.NewPage()
	if err != nil {
		return err
	}
	defer page.Close()

	if err := page.Goto(g.cfg.PublishImageURL); err != nil {
		return err
	}
	if err := page.SetFiles(g.cfg.Selectors["upload_input"], content.ImagePaths); err != nil {
		return err
	}
	if err := page.Fill(g.cfg.Selectors["title_input"], content.Title); err != nil {
		return err
	}
	if err := page.Fill(g.cfg.Selectors["content"], content.Content); err != nil {
		return err
	}
	return page.Click(g.cfg.Selectors["submit"])
}

func (g *Gateway) PublishVideo(ctx context.Context, content publish.VideoContent) error {
	return ErrNotReady
}
