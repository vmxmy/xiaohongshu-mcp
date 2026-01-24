package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/xpzouying/xiaohongshu-mcp/internal/domain/publish"
	infraplaywright "github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser/playwright"
	infraconfig "github.com/xpzouying/xiaohongshu-mcp/internal/infra/config"
	publishgateway "github.com/xpzouying/xiaohongshu-mcp/internal/infra/xhs/publish"
	"gopkg.in/yaml.v3"
)

// 完整测试发布流程
func main() {
	ctx := context.Background()

	// 加载配置
	configPath := os.Getenv("XHS_CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := infraconfig.LoadFromFile(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	selectors, err := loadSelectors(configPath)
	if err != nil {
		log.Fatalf("加载选择器失败: %v", err)
	}

	fmt.Printf("配置加载成功\n")
	fmt.Printf("发布URL: %s\n", cfg.URLs.Creator.PublishImage)
	fmt.Printf("选择器: %+v\n", selectors)

	// 创建 engine
	engineCfg := infraplaywright.DefaultConfig()
	engineCfg.Headless = false
	engineCfg.ActionTimeout = 60 * time.Second
	engineCfg.NavigationTimeout = 120 * time.Second

	// 设置 cookie 路径
	cookiePath := os.Getenv("HOME") + "/.xhs-mcp/cookies.yaml"
	if _, err := os.Stat(cookiePath); err == nil {
		engineCfg.CookiePath = cookiePath
		fmt.Printf("Cookie路径: %s\n", cookiePath)
	}

	engine := infraplaywright.New(engineCfg)

	// 创建 gateway
	gatewayCfg := publishgateway.Config{
		PublishImageURL: cfg.URLs.Creator.PublishImage,
		PublishVideoURL: cfg.URLs.Creator.PublishVideo,
		Selectors:       selectors,
	}

	gateway, err := publishgateway.NewGateway(gatewayCfg, engine)
	if err != nil {
		log.Fatalf("创建 gateway 失败: %v", err)
	}

	// 测试发布
	content := publish.ImageContent{
		Title:      "发布测试",
		Content:    "这是一个测试内容，用于验证选择器是否正确。",
		ImagePaths: []string{"test_images/test.jpg"},
	}

	fmt.Printf("\n开始发布测试...\n")
	fmt.Printf("标题: %s\n", content.Title)
	fmt.Printf("内容: %s\n", content.Content)
	fmt.Printf("图片: %v\n", content.ImagePaths)

	if err := gateway.PublishImage(ctx, content); err != nil {
		log.Fatalf("发布失败: %v", err)
	}

	fmt.Printf("\n✅ 发布成功！\n")
}

type selectorConfig struct {
	Selectors struct {
		Publish struct {
			UploadInput     string `yaml:"upload_input"`
			TitleInput      string `yaml:"title_input"`
			ContentEditorQL string `yaml:"content_editor_ql"`
			SubmitButton    string `yaml:"submit_button"`
		} `yaml:"publish"`
	} `yaml:"selectors"`
}

func loadSelectors(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg selectorConfig
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
