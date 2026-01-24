package main

import (
	"errors"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/xpzouying/xiaohongshu-mcp/cookies"
	apppublish "github.com/xpzouying/xiaohongshu-mcp/internal/app/publish"
	browserrod "github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser/rod"
	infraconfig "github.com/xpzouying/xiaohongshu-mcp/internal/infra/config"
	"github.com/xpzouying/xiaohongshu-mcp/internal/interfaces/wiring"
)

func buildPublishUsecase(cfg *infraconfig.Config, selectors map[string]string, headless bool) (*apppublish.Usecase, error) {
	if cfg == nil {
		return nil, errors.New("config missing")
	}
	engineCfg := browserrod.DefaultConfig()
	engineCfg.Headless = headless
	engineCfg.CookiePath = cookies.GetCookiesFilePath()
	if cfg.Timeouts.Navigate > 0 {
		engineCfg.NavigationTimeout = time.Duration(cfg.Timeouts.Navigate) * time.Second
	}
	if cfg.Timeouts.ElementWait > 0 {
		engineCfg.ActionTimeout = time.Duration(cfg.Timeouts.ElementWait) * time.Second
	}
	if cfg.Timeouts.ImageUpload > 0 {
		uploadTimeout := time.Duration(cfg.Timeouts.ImageUpload) * time.Second
		if uploadTimeout > engineCfg.ActionTimeout {
			engineCfg.ActionTimeout = uploadTimeout
		}
	}
	engine := browserrod.New(engineCfg)
	return wiring.BuildPublishUsecase(cfg, selectors, engine)
}

func loadPublishUsecase(headless bool) (*apppublish.Usecase, error) {
	configPath := os.Getenv("XHS_CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	cfg, err := infraconfig.LoadFromFile(configPath)
	if err != nil {
		return nil, err
	}
	selectors, err := loadPublishSelectors(configPath)
	if err != nil {
		return nil, err
	}
	return buildPublishUsecase(cfg, selectors, headless)
}

func initPublishUsecase(headless bool) *apppublish.Usecase {
	usecase, err := loadPublishUsecase(headless)
	if err != nil {
		logrus.Warnf("初始化发布用例失败: %v", err)
		return nil
	}
	return usecase
}

type publishSelectorConfig struct {
	Selectors struct {
		Publish struct {
			UploadInput     string `yaml:"upload_input"`
			TitleInput      string `yaml:"title_input"`
			ContentEditorQL string `yaml:"content_editor_ql"`
			SubmitButton    string `yaml:"submit_button"`
			SaveDraftButton string `yaml:"save_draft_button"`
		} `yaml:"publish"`
	} `yaml:"selectors"`
}

func loadPublishSelectors(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg publishSelectorConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return map[string]string{
		"upload_input": cfg.Selectors.Publish.UploadInput,
		"title_input":  cfg.Selectors.Publish.TitleInput,
		"content":      cfg.Selectors.Publish.ContentEditorQL,
		"submit":       cfg.Selectors.Publish.SubmitButton,
		"save_draft":   cfg.Selectors.Publish.SaveDraftButton,
	}, nil
}
