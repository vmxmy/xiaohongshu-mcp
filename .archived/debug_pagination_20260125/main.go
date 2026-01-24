package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xpzouying/xiaohongshu-mcp/cookies"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser/playwright"
)

func main() {
	fmt.Println("🔍 分析小红书数据分析页面的翻页结构...")
	fmt.Println()

	cfg := playwright.DefaultConfig()
	cfg.Headless = false
	cfg.CookiePath = cookies.GetCookiesFilePath()
	cfg.ActionTimeout = 30 * time.Second
	cfg.NavigationTimeout = 60 * time.Second

	engine := playwright.New(cfg)
	if err := engine.Start(); err != nil {
		fmt.Printf("❌ 启动浏览器失败: %v\n", err)
		return
	}
	defer engine.Close()

	page, err := engine.NewPage()
	if err != nil {
		fmt.Printf("❌ 创建页面失败: %v\n", err)
		return
	}
	defer page.Close()

	ctx := context.Background()
	pp := page.WithContext(ctx)

	url := "https://creator.xiaohongshu.com/statistics/data-analysis?source=official"
	fmt.Printf("📍 导航到: %s\n", url)

	if err := pp.Goto(url); err != nil {
		fmt.Printf("❌ 导航失败: %v\n", err)
		return
	}

	fmt.Println("⏳ 等待页面加载...")
	time.Sleep(5 * time.Second)

	// 分析翻页相关元素
	fmt.Println("\n🔢 查找翻页相关元素...")
	result, err := pp.Eval(`() => {
		const pagination = {};

		// 查找所有可能的翻页相关元素
		const possibleSelectors = [
			'.el-pagination',
			'.pagination',
			'[class*="pagination"]',
			'[class*="pager"]',
			'.el-pager',
			'button[class*="next"]',
			'button[class*="prev"]',
			'li.number',
			'.el-pagination__total',
			'.el-pagination__sizes',
			'.el-pagination__jump'
		];

		possibleSelectors.forEach(selector => {
			const elements = document.querySelectorAll(selector);
			if (elements.length > 0) {
				pagination[selector] = {
					count: elements.length,
					sampleHTML: elements[0] ? elements[0].outerHTML.substring(0, 200) : '',
					sampleText: elements[0] ? elements[0].textContent.trim() : ''
				};
			}
		});

		// 查找包含"下一页"文本的按钮
		const allButtons = document.querySelectorAll('button');
		const nextButtons = [];
		allButtons.forEach(btn => {
			const text = btn.textContent.trim();
			if (text.includes('下一页') || text.includes('next') || btn.className.includes('next')) {
				nextButtons.push({
					text: text,
					className: btn.className,
					disabled: btn.disabled,
					outerHTML: btn.outerHTML.substring(0, 150)
				});
			}
		});
		pagination['nextButtons'] = nextButtons;

		// 查找页码数字
		const pageNumbers = [];
		document.querySelectorAll('li, button, a').forEach(el => {
			const text = el.textContent.trim();
			if (/^\d+$/.test(text) && text.length <= 3) {
				pageNumbers.push({
					text: text,
					tag: el.tagName,
					className: el.className,
					outerHTML: el.outerHTML.substring(0, 150)
				});
			}
		});
		pagination['pageNumbers'] = pageNumbers.slice(0, 10);

		// 查找总条数信息
		const totalInfo = [];
		document.querySelectorAll('*').forEach(el => {
			const text = el.textContent.trim();
			if ((text.includes('共') && text.includes('条')) || text.includes('total')) {
				if (el.children.length === 0) {
					totalInfo.push({
						text: text,
						tag: el.tagName,
						className: el.className
					});
				}
			}
		});
		pagination['totalInfo'] = totalInfo.slice(0, 5);

		return JSON.stringify(pagination, null, 2);
	}`)

	if err != nil {
		fmt.Printf("❌ 分析失败: %v\n", err)
	} else {
		fmt.Println(result)
	}

	// 测试点击下一页
	fmt.Println("\n🖱️  尝试查找并点击下一页按钮...")
	hasNext, err := pp.Has("button.btn-next:not([disabled])")
	if err != nil {
		fmt.Printf("⚠️  查找下一页按钮失败: %v\n", err)
	} else if hasNext {
		fmt.Println("✅ 找到可点击的下一页按钮")

		// 获取当前页码
		currentPage, _ := pp.Eval(`() => {
			const current = document.querySelector('.number.active, li.active');
			return current ? current.textContent.trim() : '1';
		}`)
		fmt.Printf("📄 当前页码: %v\n", currentPage)

		// 点击下一页
		fmt.Println("⏭️  点击下一页...")
		if err := pp.Click("button.btn-next"); err != nil {
			fmt.Printf("⚠️  点击失败: %v\n", err)
		} else {
			time.Sleep(2 * time.Second)
			newPage, _ := pp.Eval(`() => {
				const current = document.querySelector('.number.active, li.active');
				return current ? current.textContent.trim() : '?';
			}`)
			fmt.Printf("📄 新页码: %v\n", newPage)
		}
	} else {
		fmt.Println("ℹ️  未找到可点击的下一页按钮（可能已经是最后一页）")
	}

	fmt.Println("\n✅ 分析完成！浏览器窗口将保持打开30秒...")
	time.Sleep(30 * time.Second)
}
