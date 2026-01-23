package publish

import (
	"context"

	"github.com/xpzouying/xiaohongshu-mcp/internal/app/ports"
	"github.com/xpzouying/xiaohongshu-mcp/internal/domain/publish"
)

type Usecase struct {
	Gateway ports.PublishGateway
	Limits  publish.Limits
}

func (u Usecase) PublishImage(ctx context.Context, content publish.ImageContent) error {
	if err := publish.ValidateImageContent(content, u.Limits); err != nil {
		return err
	}
	return u.Gateway.PublishImage(ctx, content)
}
