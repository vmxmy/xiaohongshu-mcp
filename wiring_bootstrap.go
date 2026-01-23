package main

import (
	"errors"
	"os"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	apppublish "github.com/xpzouying/xiaohongshu-mcp/internal/app/publish"
	infraconfig "github.com/xpzouying/xiaohongshu-mcp/internal/infra/config"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser/playwright"
	"github.com/xpzouying/xiaohongshu-mcp/internal/interfaces/wiring"
)

func buildPublishUsecase(cfg *infraconfig.Config, selectors map[string]string, headless bool) (*apppublish.Usecase, error) {
	if cfg == nil {
		return nil, errors.New("config missing")
	}
	engineCfg := playwright.DefaultConfig()
	engineCfg.Headless = headless
	engine := playwright.New(engineCfg)
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
	}, nil
}
