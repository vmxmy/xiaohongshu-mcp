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

func (u Usecase) SaveImageDraft(ctx context.Context, content publish.ImageContent) error {
	if err := publish.ValidateImageContent(content, u.Limits); err != nil {
		return err
	}
	return u.Gateway.SaveImageDraft(ctx, content)
}

func (u Usecase) SaveVideoDraft(ctx context.Context, content publish.VideoContent) error {
	return u.Gateway.SaveVideoDraft(ctx, content)
}
