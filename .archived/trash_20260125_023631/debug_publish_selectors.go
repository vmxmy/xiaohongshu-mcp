package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/xpzouying/xiaohongshu-mcp/cookies"
	infraconfig "github.com/xpzouying/xiaohongshu-mcp/internal/infra/config"
	"gopkg.in/yaml.v3"
)

// 诊断脚本：测试发布页面的选择器是否可用
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

	fmt.Printf("发布页面URL: %s\n", cfg.URLs.Creator.PublishImage)
	fmt.Printf("选择器配置:\n")
	for k, v := range selectors {
		fmt.Printf("  %s: %s\n", k, v)
	}

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
	ctxOptions := playwright.BrowserNewContextOptions{}
	browserCtx, err := browser.NewContext(ctxOptions)
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
	fmt.Printf("\n正在导航到发布页面...\n")
	if _, err := page.Goto(cfg.URLs.Creator.PublishImage); err != nil {
		log.Fatalf("导航失败: %v", err)
	}

	// 等待页面加载
	time.Sleep(3 * time.Second)

	// 测试上传选择器
	fmt.Printf("\n步骤1: 测试上传选择器\n")
	testSelector(ctx, page, "upload_input", selectors["upload_input"])

	// 尝试上传图片（如果test_images目录存在）
	testImagePath := "test_images/test.jpg"
	if _, err := os.Stat(testImagePath); err == nil {
		fmt.Printf("\n步骤2: 上传测试图片 %s\n", testImagePath)
		loc := page.Locator(selectors["upload_input"])
		if err := loc.SetInputFiles(testImagePath); err != nil {
			fmt.Printf("❌ 上传失败: %v\n", err)
		} else {
			fmt.Printf("✅ 图片已上传\n")
			// 等待图片处理
			fmt.Printf("等待10秒让页面加载...\n")
			time.Sleep(10 * time.Second)
		}
	} else {
		fmt.Printf("\n步骤2: 跳过 - 未找到测试图片 %s\n", testImagePath)
	}

	// 测试其他选择器
	fmt.Printf("\n步骤3: 获取页面结构信息\n")

	// 获取所有可能的输入元素
	inputElements := []string{
		"input[type=\"text\"]",
		"input",
		"textarea",
		"div[contenteditable]",
		"div.input",
		"div[role=\"textbox\"]",
	}
	fmt.Printf("查找可能的输入元素:\n")
	for _, sel := range inputElements {
		count, _ := page.Locator(sel).Count()
		if count > 0 {
			fmt.Printf("  找到 %d 个 %s\n", count, sel)
		}
	}

	// 获取所有按钮
	buttonElements := []string{
		"button",
		"div[role=\"button\"]",
		"div.button",
		"div.submit",
	}
	fmt.Printf("\n查找可能的按钮元素:\n")
	for _, sel := range buttonElements {
		count, _ := page.Locator(sel).Count()
		if count > 0 {
			fmt.Printf("  找到 %d 个 %s\n", count, sel)
		}
	}

	// 尝试获取页面主要容器的类名
	fmt.Printf("\n查找主要容器:\n")
	containers := []string{
		"div[class*=\"publish\"]",
		"div[class*=\"editor\"]",
		"div[class*=\"content\"]",
		"div[class*=\"upload\"]",
	}
	for _, sel := range containers {
		count, _ := page.Locator(sel).Count()
		if count > 0 {
			fmt.Printf("  找到 %d 个 %s\n", count, sel)
		}
	}

	fmt.Printf("\n步骤4: 测试原始选择器可用性\n")
	testSelector(ctx, page, "upload_input", selectors["upload_input"])
	testSelector(ctx, page, "title_input", selectors["title_input"])
	testSelector(ctx, page, "content", selectors["content"])
	testSelector(ctx, page, "submit", selectors["submit"])

	// 尝试备用选择器
	fmt.Printf("\n步骤5: 测试备用内容选择器\n")
	alternatives := []string{
		"div.ql-editor",
		"[role=\"textbox\"]",
		"div[contenteditable=\"true\"]",
		".ql-editor",
		"#editor",
		"textarea",
		"div.content-input",
	}
	for _, sel := range alternatives {
		testSelector(ctx, page, "备用", sel)
	}

	// 保持浏览器打开以便手动检查
	fmt.Printf("\n浏览器将保持打开30秒，请手动检查页面元素...\n")
	time.Sleep(30 * time.Second)
}

func testSelector(ctx context.Context, page playwright.Page, name, selector string) {
	fmt.Printf("  测试 %s (%s): ", name, selector)
	loc := page.Locator(selector)
	count, err := loc.Count()
	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		return
	}
	if count == 0 {
		fmt.Printf("❌ 未找到\n")
		return
	}
	visible, err := loc.First().IsVisible()
	if err != nil {
		fmt.Printf("⚠️  找到 %d 个，但无法检查可见性: %v\n", count, err)
		return
	}
	if visible {
		fmt.Printf("✅ 找到 %d 个，可见\n", count)
	} else {
		fmt.Printf("⚠️  找到 %d 个，但不可见\n", count)
	}
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
