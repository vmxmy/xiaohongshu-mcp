package testkit

import (
	"context"

	"github.com/xpzouying/xiaohongshu-mcp/internal/domain/publish"
)

type FakePublishGateway struct {
	ImageCalls int
	VideoCalls int
	LastImage  publish.ImageContent
	LastVideo  publish.VideoContent
	Err        error
}

func (f *FakePublishGateway) PublishImage(ctx context.Context, content publish.ImageContent) error {
	f.ImageCalls++
	f.LastImage = content
	return f.Err
}

func (f *FakePublishGateway) PublishVideo(ctx context.Context, content publish.VideoContent) error {
	f.VideoCalls++
	f.LastVideo = content
	return f.Err
}

type FakeSelectorStore struct {
	Selectors map[string]string
}

func (f *FakeSelectorStore) Load() (map[string]string, error) {
	return f.Selectors, nil
}

func (f *FakeSelectorStore) Save(selectors map[string]string) error {
	f.Selectors = selectors
	return nil
}

func (f *FakeSelectorStore) Snapshot() (string, error) {
	return "snapshot", nil
}

func (f *FakeSelectorStore) Rollback(snapshot string) error {
	return nil
}
