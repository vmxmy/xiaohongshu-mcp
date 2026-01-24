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

	fmt.Printf("开始完整发布流程测试...\n\n")

	// 启动 Playwright
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("启动 Playwright 失败: %v", err)
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false), // 不使用无头模式，方便观察
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
				browserCtx.AddCookies(cookieList)
			}
		}
	}

	page, err := browserCtx.NewPage()
	if err != nil {
		log.Fatalf("创建页面失败: %v", err)
	}
	defer page.Close()

	page.SetDefaultTimeout(60000)

	// 1. 导航
	fmt.Printf("【步骤1】导航到发布页面\n")
	if _, err := page.Goto(cfg.URLs.Creator.PublishImage, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
	}); err != nil {
		log.Fatalf("导航失败: %v", err)
	}
	time.Sleep(2 * time.Second)
	fmt.Printf("✅ 页面加载完成\n\n")

	// 2. 上传图片
	fmt.Printf("【步骤2】上传图片\n")
	uploadLoc := page.Locator(selectors["upload_input"])
	if err := uploadLoc.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(30000),
	}); err != nil {
		log.Fatalf("等待上传输入框失败: %v", err)
	}
	if err := uploadLoc.First().SetInputFiles("test_images/test.jpg"); err != nil {
		log.Fatalf("上传图片失败: %v", err)
	}
	fmt.Printf("✅ 图片已上传\n")
	time.Sleep(5 * time.Second) // 等待图片处理
	fmt.Println()

	// 3. 填写标题
	fmt.Printf("【步骤3】填写标题\n")
	titleLoc := page.Locator(selectors["title_input"])
	if err := titleLoc.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(30000),
	}); err != nil {
		log.Fatalf("等待标题输入框失败: %v", err)
	}
	if err := titleLoc.Fill("MCP测试发布"); err != nil {
		log.Fatalf("填写标题失败: %v", err)
	}
	fmt.Printf("✅ 已填写标题\n\n")

	// 4. 填写内容
	fmt.Printf("【步骤4】填写内容\n")
	contentLoc := page.Locator(selectors["content"])
	if err := contentLoc.WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(30000),
	}); err != nil {
		log.Fatalf("等待内容编辑器失败: %v", err)
	}
	if err := contentLoc.Fill("这是通过MCP工具发布的测试内容。\n发布时间: " + time.Now().Format("2006-01-02 15:04:05")); err != nil {
		log.Fatalf("填写内容失败: %v", err)
	}
	fmt.Printf("✅ 已填写内容\n\n")

	// 5. 点击发布按钮
	fmt.Printf("【步骤5】点击发布按钮\n")
	submitLoc := page.Locator(selectors["submit"])
	if err := submitLoc.First().Click(); err != nil {
		log.Fatalf("点击发布按钮失败: %v", err)
	}
	fmt.Printf("✅ 已点击发布按钮\n\n")

	// 6. 观察发布后的情况
	fmt.Printf("【步骤6】等待发布结果...\n")
	time.Sleep(3 * time.Second)

	// 检查是否有成功消息
	successSelectors := []string{
		".el-message--success",
		".d-message--success",
		"div:has-text('发布成功')",
		"div:has-text('发送成功')",
	}
	fmt.Printf("检查成功消息:\n")
	for _, sel := range successSelectors {
		count, _ := page.Locator(sel).Count()
		if count > 0 {
			fmt.Printf("  ✅ 找到成功消息: %s\n", sel)
			text, _ := page.Locator(sel).First().TextContent()
			fmt.Printf("     内容: %s\n", text)
		}
	}

	// 检查是否有错误消息
	errorSelectors := []string{
		".el-message--error",
		".d-message--error",
		"div:has-text('失败')",
		"div:has-text('错误')",
	}
	fmt.Printf("\n检查错误消息:\n")
	for _, sel := range errorSelectors {
		count, _ := page.Locator(sel).Count()
		if count > 0 {
			fmt.Printf("  ❌ 找到错误消息: %s\n", sel)
			text, _ := page.Locator(sel).First().TextContent()
			fmt.Printf("     内容: %s\n", text)
		}
	}

	// 检查是否有确认弹窗
	fmt.Printf("\n检查弹窗:\n")
	dialogSelectors := []string{
		".el-dialog",
		".d-modal",
		"div[role=\"dialog\"]",
	}
	for _, sel := range dialogSelectors {
		count, _ := page.Locator(sel).Count()
		if count > 0 {
			visible, _ := page.Locator(sel).First().IsVisible()
			if visible {
				fmt.Printf("  ⚠️  发现可见弹窗: %s\n", sel)
			}
		}
	}

	// 检查URL变化
	currentURL := page.URL()
	fmt.Printf("\n当前URL: %s\n", currentURL)

	// 保持浏览器打开60秒供观察
	fmt.Printf("\n浏览器将保持打开60秒，请检查页面状态...\n")
	fmt.Printf("请手动检查:\n")
	fmt.Printf("  1. 是否显示了发布成功的提示？\n")
	fmt.Printf("  2. 页面是否跳转到了其他地方？\n")
	fmt.Printf("  3. 是否有任何错误提示？\n")
	fmt.Printf("  4. 进入创作者中心查看笔记列表，是否有新发布的笔记？\n")
	time.Sleep(60 * time.Second)
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
