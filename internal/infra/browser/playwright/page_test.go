package playwright

import (
	"context"
	"testing"
	"time"

	"github.com/xpzouying/xiaohongshu-mcp/cookies"
)

// TestGotoWithLoad 测试 Goto 方法使用 Load 等待策略
func TestGotoWithLoad(t *testing.T) {
	t.Skip("需要真实浏览器环境，仅在本地调试时运行")

	cfg := DefaultConfig()
	cfg.Headless = true
	cfg.CookiePath = cookies.GetCookiesFilePath()
	cfg.ActionTimeout = 30 * time.Second
	cfg.NavigationTimeout = 60 * time.Second

	engine := New(cfg)
	if err := engine.Start(); err != nil {
		t.Fatalf("启动浏览器失败: %v", err)
	}
	defer engine.Close()

	page, err := engine.NewPage()
	if err != nil {
		t.Fatalf("创建页面失败: %v", err)
	}
	defer page.Close()

	// 测试访问小红书
	start := time.Now()
	err = page.Goto("https://www.xiaohongshu.com/explore")
	elapsed := time.Since(start)

	t.Logf("导航耗时: %v", elapsed)

	if err != nil {
		t.Errorf("导航失败: %v", err)
	}

	// 应该在合理时间内完成（不超过20秒）
	if elapsed > 20*time.Second {
		t.Errorf("导航耗时过长: %v", elapsed)
	}

	// 等待页面加载
	if err := page.WaitLoad(); err != nil {
		t.Errorf("等待页面加载失败: %v", err)
	}

	// 验证页面 URL
	url := page.URL()
	t.Logf("当前页面: %s", url)

	if url == "" {
		t.Error("页面 URL 为空")
	}
}

// TestGotoTimeout 测试导航超时处理
func TestGotoTimeout(t *testing.T) {
	t.Skip("需要真实浏览器环境，仅在本地调试时运行")

	cfg := DefaultConfig()
	cfg.Headless = true
	cfg.NavigationTimeout = 5 * time.Second // 设置很短的超时

	engine := New(cfg)
	if err := engine.Start(); err != nil {
		t.Fatalf("启动浏览器失败: %v", err)
	}
	defer engine.Close()

	page, err := engine.NewPage()
	if err != nil {
		t.Fatalf("创建页面失败: %v", err)
	}
	defer page.Close()

	// 测试超时情况
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	pp := page.WithContext(ctx)
	err = pp.Goto("https://www.xiaohongshu.com/explore")

	// 应该返回超时错误
	if err == nil {
		t.Log("注意: 导航意外成功，可能网络很快")
	} else {
		t.Logf("预期的超时错误: %v", err)
	}
}

// TestCheckLoginStatusFlow 测试完整的登录检查流程
func TestCheckLoginStatusFlow(t *testing.T) {
	t.Skip("需要真实浏览器环境，仅在本地调试时运行")

	cfg := DefaultConfig()
	cfg.Headless = false // 有头模式便于观察
	cfg.CookiePath = cookies.GetCookiesFilePath()
	cfg.ActionTimeout = 30 * time.Second
	cfg.NavigationTimeout = 60 * time.Second

	engine := New(cfg)
	if err := engine.Start(); err != nil {
		t.Fatalf("启动浏览器失败: %v", err)
	}
	defer engine.Close()

	page, err := engine.NewPage()
	if err != nil {
		t.Fatalf("创建页面失败: %v", err)
	}
	defer page.Close()

	ctx := context.Background()
	pp := page.WithContext(ctx)

	// 1. 导航到小红书
	t.Log("步骤 1: 导航到小红书...")
	start := time.Now()
	if err := pp.Goto("https://www.xiaohongshu.com/explore"); err != nil {
		t.Fatalf("导航失败: %v", err)
	}
	t.Logf("导航成功，耗时: %v", time.Since(start))

	// 2. 等待页面加载
	t.Log("步骤 2: 等待页面加载...")
	if err := pp.WaitLoad(); err != nil {
		t.Fatalf("等待加载失败: %v", err)
	}

	// 3. 等待 1 秒
	time.Sleep(1 * time.Second)

	// 4. 检查登录元素
	t.Log("步骤 3: 检查登录元素...")
	selector := `.main-container .user .link-wrapper .channel`
	exists, err := pp.Has(selector)
	if err != nil {
		t.Errorf("检查元素失败: %v", err)
	}

	t.Logf("登录元素存在: %v", exists)

	// 5. 截图保存
	// screenshot, err := page.Screenshot()
	// if err == nil {
	// 	os.WriteFile("/tmp/xhs_test.png", screenshot, 0644)
	// 	t.Log("截图已保存到 /tmp/xhs_test.png")
	// }
}
