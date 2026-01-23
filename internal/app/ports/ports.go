package ports

import (
	"context"

	"github.com/xpzouying/xiaohongshu-mcp/internal/domain/publish"
)

type PublishGateway interface {
	PublishImage(ctx context.Context, content publish.ImageContent) error
	PublishVideo(ctx context.Context, content publish.VideoContent) error
}

type SelectorStore interface {
	Load() (map[string]string, error)
	Save(selectors map[string]string) error
	Snapshot() (string, error)
	Rollback(snapshot string) error
}
