package publish

import (
	"context"
	"testing"

	"github.com/xpzouying/xiaohongshu-mcp/internal/domain/publish"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser"
)

type fakePage struct {
	Calls []string
}

func (p *fakePage) Goto(url string) error {
	p.Calls = append(p.Calls, "goto")
	return nil
}

func (p *fakePage) Click(selector string) error {
	p.Calls = append(p.Calls, "click:"+selector)
	return nil
}

func (p *fakePage) Fill(selector, value string) error {
	p.Calls = append(p.Calls, "fill:"+selector)
	return nil
}

func (p *fakePage) SetFiles(selector string, files []string) error {
	p.Calls = append(p.Calls, "files:"+selector)
	return nil
}

func (p *fakePage) Text(selector string) (string, error) {
	return "", nil
}

func (p *fakePage) WaitVisible(selector string) error {
	return nil
}

func (p *fakePage) Close() error {
	return nil
}

type fakeEngine struct{ page *fakePage }

func (e *fakeEngine) Start() error {
	return nil
}

func (e *fakeEngine) NewPage() (browser.Page, error) {
	return e.page, nil
}

func (e *fakeEngine) Close() error {
	return nil
}

func TestGateway_PublishImage_UsesSelectors(t *testing.T) {
	engine := &fakeEngine{page: &fakePage{}}
	cfg := Config{
		PublishImageURL: "https://example.com",
		PublishVideoURL: "https://example.com",
		Selectors: map[string]string{
			"upload_input": "input[type=file]",
			"title_input":  "input[name=title]",
			"content":      "textarea[name=content]",
			"submit":       "button[type=submit]",
		},
	}
	gw, err := NewGateway(cfg, engine)
	if err != nil {
		t.Fatalf("new gateway err: %v", err)
	}
	err = gw.PublishImage(context.Background(), publish.ImageContent{
		Title:      "t",
		Content:    "c",
		ImagePaths: []string{"1.jpg"},
	})
	if err != nil {
		t.Fatalf("publish err: %v", err)
	}
	if len(engine.page.Calls) == 0 {
		t.Fatalf("expected page calls")
	}
}
