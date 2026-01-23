package wiring

import (
	apppublish "github.com/xpzouying/xiaohongshu-mcp/internal/app/publish"
	domainpublish "github.com/xpzouying/xiaohongshu-mcp/internal/domain/publish"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/config"
	xhspublish "github.com/xpzouying/xiaohongshu-mcp/internal/infra/xhs/publish"
)

func BuildPublishUsecase(cfg *config.Config, selectors map[string]string, engine browser.Engine) (*apppublish.Usecase, error) {
	gw, err := xhspublish.NewGateway(xhspublish.Config{
		PublishImageURL: cfg.URLs.Creator.PublishImage,
		PublishVideoURL: cfg.URLs.Creator.PublishVideo,
		Selectors:       selectors,
	}, engine)
	if err != nil {
		return nil, err
	}
	return &apppublish.Usecase{
		Gateway: gw,
		Limits: domainpublish.Limits{
			MaxTags:   cfg.Limits.MaxTags,
			MinImages: cfg.Limits.MinImages,
			MaxImages: cfg.Limits.MaxImages,
		},
	}, nil
}
