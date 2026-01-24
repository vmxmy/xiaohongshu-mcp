package config

import "testing"

func TestLoadConfig_File(t *testing.T) {
	cfg, err := LoadFromFile("testdata/config.yaml")
	if err != nil || cfg.URLs.Creator.PublishImage == "" || cfg.Limits.MaxTags == 0 {
		t.Fatalf("expected publish_image url and limits")
	}
	if cfg.Timeouts.Navigate == 0 || cfg.Timeouts.ElementWait == 0 {
		t.Fatalf("expected timeouts")
	}
}
