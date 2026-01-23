package main

import (
	"os"
	"testing"

	infraconfig "github.com/xpzouying/xiaohongshu-mcp/internal/infra/config"
)

func TestBuildPublishUsecase_ValidConfig(t *testing.T) {
	cfg := &infraconfig.Config{}
	cfg.URLs.Creator.PublishImage = "https://example.com/publish?target=image"
	cfg.URLs.Creator.PublishVideo = "https://example.com/publish?target=video"
	cfg.Limits.MaxTags = 10
	cfg.Limits.MinImages = 1
	cfg.Limits.MaxImages = 9
	selectors := map[string]string{
		"upload_input": "input[type=file]",
		"title_input":  "input[name=title]",
		"content":      "div.editor",
		"submit":       "button[type=submit]",
	}

	uc, err := buildPublishUsecase(cfg, selectors, true)
	if err != nil || uc == nil {
		t.Fatalf("expected usecase, err=%v", err)
	}
}

func TestLoadPublishSelectors(t *testing.T) {
	path := t.TempDir() + "/config.yaml"
	data := []byte("selectors:\n  publish:\n    upload_input: \"input[type=file]\"\n    title_input: \"input[name=title]\"\n    content_editor_ql: \"div.editor\"\n    submit_button: \"button[type=submit]\"\n")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	selectors, err := loadPublishSelectors(path)
	if err != nil {
		t.Fatalf("load selectors err: %v", err)
	}
	if selectors["upload_input"] == "" || selectors["submit"] == "" {
		t.Fatalf("expected selectors to be loaded")
	}
}

func TestLoadPublishUsecase(t *testing.T) {
	path := t.TempDir() + "/config.yaml"
	data := []byte("urls:\n  creator:\n    publish_image: \"https://example.com/publish?target=image\"\n    publish_video: \"https://example.com/publish?target=video\"\nlimits:\n  max_tags: 10\n  min_images: 1\n  max_images: 9\nselectors:\n  publish:\n    upload_input: \"input[type=file]\"\n    title_input: \"input[name=title]\"\n    content_editor_ql: \"div.editor\"\n    submit_button: \"button[type=submit]\"\n")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Setenv("XHS_CONFIG_PATH", path)
	uc, err := loadPublishUsecase(true)
	if err != nil || uc == nil {
		t.Fatalf("expected usecase, err=%v", err)
	}
}

func TestInitPublishUsecase_ReturnsUsecase(t *testing.T) {
	path := t.TempDir() + "/config.yaml"
	data := []byte("urls:\n  creator:\n    publish_image: \"https://example.com/publish?target=image\"\n    publish_video: \"https://example.com/publish?target=video\"\nlimits:\n  max_tags: 10\n  min_images: 1\n  max_images: 9\nselectors:\n  publish:\n    upload_input: \"input[type=file]\"\n    title_input: \"input[name=title]\"\n    content_editor_ql: \"div.editor\"\n    submit_button: \"button[type=submit]\"\n")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	t.Setenv("XHS_CONFIG_PATH", path)
	uc := initPublishUsecase(true)
	if uc == nil {
		t.Fatalf("expected usecase")
	}
}
