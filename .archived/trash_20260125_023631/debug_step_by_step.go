package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/xpzouying/xiaohongshu-mcp/cookies"
	infraconfig "github.com/xpzouying/xiaohongshu-mcp/internal/infra/config"
	"gopkg.in/yaml.v3"
)

func main() {
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

	fmt.Printf("发布URL: %s\n", cfg.URLs.Creator.PublishImage)
	fmt.Printf("选择器: %+v\n\n", selectors)

	// 启动 Playwright
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("启动 Playwright 失败: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		log.Fatalf("启动浏览器失败: %v", err)
	}
	defer browser.Close()

	// 创建带 cookies 的上下文
	browserCtx, err := browser.NewContext()
	if err != nil {
		log.Fatalf("创建上下文失败: %v", err)
	}
	defer browserCtx.Close()

	// 加载 cookies
	cookiePath := cookies.GetCookiesFilePath()
	if cookiePath != "" {
		cookieData, err := os.ReadFile(cookiePath)
		if err == nil {
			var rawCookies []map[string]interface{}
			if err := yaml.Unmarshal(cookieData, &rawCookies); err == nil && len(rawCookies) > 0 {
				var cookieList []playwright.OptionalCookie
				for _, c := range rawCookies {
					cookie := playwright.OptionalCookie{}
					if name, ok := c["name"].(string); ok {
						cookie.Name = name
					}
					if value, ok := c["value"].(string); ok {
						cookie.Value = value
					}
					if domain, ok := c["domain"].(string); ok {
						cookie.Domain = &domain
					}
					if path, ok := c["path"].(string); ok {
						cookie.Path = &path
					}
					cookieList = append(cookieList, cookie)
				}
				if err := browserCtx.AddCookies(cookieList); err != nil {
					log.Printf("添加 cookies 失败: %v", err)
				} else {
					fmt.Printf("已加载 %d 个 cookies\n", len(cookieList))
				}
			}
		}
	}

	page, err := browserCtx.NewPage()
	if err != nil {
		log.Fatalf("创建页面失败: %v", err)
	}
	defer page.Close()

	// 设置超时
	page.SetDefaultTimeout(60000)

	// 导航到发布页面
	fmt.Printf("\n【步骤1】导航到发布页面...\n")
	if _, err := page.Goto(cfg.URLs.Creator.PublishImage, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		log.Fatalf("导航失败: %v", err)
	}
	fmt.Printf("✅ 页面加载完成\n")

	// 额外等待2秒确保 JS 执行完成
	time.Sleep(2 * time.Second)

	// 测试上传输入框
	fmt.Printf("\n【步骤2】检查上传输入框 %s\n", selectors["upload_input"])
	uploadLoc := page.Locator(selectors["upload_input"])
	count, err := uploadLoc.Count()
	if err != nil {
		log.Fatalf("检查上传输入框失败: %v", err)
	}
	fmt.Printf("找到 %d 个上传输入框\n", count)

	if count == 0 {
		fmt.Printf("❌ 未找到上传输入框，尝试备用选择器\n")
		alternatives := []string{
			"input[type=\"file\"]",
			".upload-input",
			"input.upload-input",
			"#upload-input",
			"input[accept*=\"image\"]",
		}
		for _, alt := range alternatives {
			altCount, _ := page.Locator(alt).Count()
			if altCount > 0 {
				fmt.Printf("  找到 %d 个 %s\n", altCount, alt)
			}
		}
		log.Fatalf("无法找到上传输入框")
	}

	visible, err := uploadLoc.First().IsVisible()
	if err != nil {
		log.Fatalf("检查可见性失败: %v", err)
	}
	fmt.Printf("可见性: %v\n", visible)

	// 上传图片
	testImage := "test_images/test.jpg"
	fmt.Printf("\n【步骤3】上传图片 %s\n", testImage)
	if err := uploadLoc.First().SetInputFiles(testImage); err != nil {
		log.Fatalf("上传图片失败: %v", err)
	}
	fmt.Printf("✅ 图片已上传\n")

	// 等待页面响应
	fmt.Printf("\n【步骤4】等待页面处理图片...\n")
	time.Sleep(5 * time.Second)

	// 测试标题输入框
	fmt.Printf("\n【步骤5】检查标题输入框 %s\n", selectors["title_input"])
	titleLoc := page.Locator(selectors["title_input"])
	if err := titleLoc.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(30000),
	}); err != nil {
		log.Fatalf("等待标题输入框失败: %v", err)
	}
	fmt.Printf("✅ 标题输入框已可见\n")

	if err := titleLoc.Fill("测试标题"); err != nil {
		log.Fatalf("填写标题失败: %v", err)
	}
	fmt.Printf("✅ 已填写标题\n")

	// 测试内容编辑器
	fmt.Printf("\n【步骤6】检查内容编辑器 %s\n", selectors["content"])
	contentLoc := page.Locator(selectors["content"])
	if err := contentLoc.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(30000),
	}); err != nil {
		log.Fatalf("等待内容编辑器失败: %v", err)
	}
	fmt.Printf("✅ 内容编辑器已可见\n")

	if err := contentLoc.Fill("这是测试内容"); err != nil {
		log.Fatalf("填写内容失败: %v", err)
	}
	fmt.Printf("✅ 已填写内容\n")

	// 测试提交按钮
	fmt.Printf("\n【步骤7】检查提交按钮 %s\n", selectors["submit"])
	submitLoc := page.Locator(selectors["submit"])
	submitCount, _ := submitLoc.Count()
	fmt.Printf("找到 %d 个提交按钮\n", submitCount)

	fmt.Printf("\n✅ 所有步骤完成！浏览器将保持打开30秒供检查...\n")
	time.Sleep(30 * time.Second)
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
