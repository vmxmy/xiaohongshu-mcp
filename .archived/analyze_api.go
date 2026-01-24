package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

func main() {
	fmt.Println("=== 启动浏览器分析API ===")

	// 启动有头浏览器
	u := launcher.New().
		Headless(false).
		Devtools(true).
		MustLaunch()

	browser := rod.New().
		ControlURL(u).
		MustConnect()
	defer browser.MustClose()

	page := browser.MustPage()

	// 先加载cookies
	fmt.Println("=== 加载cookies ===")
	cookies, err := LoadCookiesFromFile("cookies.json")
	if err != nil {
		fmt.Printf("警告: 无法加载cookies: %v\n", err)
	} else {
		fmt.Printf("加载了 %d 个cookies\n", len(cookies))
		page.MustSetCookies(cookies...)
	}

	// 启用网络监听
	router := page.HijackRequests()
	defer router.MustStop()

	router.MustAdd("*", func(ctx *rod.Hijack) {
		// 打印请求
		fmt.Printf("\n[REQUEST] %s %s\n", ctx.Request.Method(), ctx.Request.URL().String())

		// 打印请求头中的关键信息
		headers := ctx.Request.Headers()
		if xsCommon, ok := headers["X-S-Common"]; ok {
			fmt.Printf("  X-S-Common: %v\n", xsCommon)
		}
		if xs, ok := headers["X-S"]; ok {
			fmt.Printf("  X-S: %v\n", xs)
		}
		if xt, ok := headers["X-T"]; ok {
			fmt.Printf("  X-T: %v\n", xt)
		}

		ctx.MustLoadResponse()

		// 如果是API请求，打印响应
		url := ctx.Request.URL().String()
		if contains(url, "/api/") || contains(url, "/web_api/") {
			fmt.Printf("[RESPONSE] Status: %d\n", ctx.Response.Payload().ResponseCode)

			body := ctx.Response.Body()
			if len(body) > 0 && len(body) < 10000 {
				var prettyJSON interface{}
				if err := json.Unmarshal([]byte(body), &prettyJSON); err == nil {
					formatted, _ := json.MarshalIndent(prettyJSON, "  ", "  ")
					fmt.Printf("  Body: %s\n", string(formatted))
				}
			}
		}
	})
	go router.Run()

	// 导航到数据分析页面
	fmt.Println("\n=== 正在打开数据分析页面 ===")
	page.MustNavigate("https://creator.xiaohongshu.com/statistics/data-analysis?source=official")
	page.MustWaitLoad()

	fmt.Println("\n=== 页面已加载，等待20秒观察网络请求 ===")
	fmt.Println("请在浏览器中滚动页面，触发数据加载...")

	time.Sleep(20 * time.Second)

	fmt.Println("\n=== 分析完成，按Ctrl+C退出 ===")
	select {} // 永久阻塞，直到用户手动退出
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(findInString(s, substr) >= 0)))
}

func findInString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// LoadCookiesFromFile 从JSON文件加载cookies
func LoadCookiesFromFile(path string) ([]*proto.NetworkCookieParam, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var rawCookies []struct {
		Name     string  `json:"name"`
		Value    string  `json:"value"`
		Domain   string  `json:"domain"`
		Path     string  `json:"path"`
		Expires  float64 `json:"expires"`
		HTTPOnly bool    `json:"httpOnly"`
		Secure   bool    `json:"secure"`
		SameSite string  `json:"sameSite"`
	}

	if err := json.Unmarshal(data, &rawCookies); err != nil {
		return nil, err
	}

	cookies := make([]*proto.NetworkCookieParam, 0, len(rawCookies))
	for _, c := range rawCookies {
		sameSite := proto.NetworkCookieSameSiteNone
		switch c.SameSite {
		case "Lax":
			sameSite = proto.NetworkCookieSameSiteLax
		case "Strict":
			sameSite = proto.NetworkCookieSameSiteStrict
		}

		cookies = append(cookies, &proto.NetworkCookieParam{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain,
			Path:     c.Path,
			HTTPOnly: c.HTTPOnly,
			Secure:   c.Secure,
			SameSite: sameSite,
		})
	}

	return cookies, nil
}
