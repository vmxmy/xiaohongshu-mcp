package main

import (
	"context"
	"fmt"
	"time"

	"github.com/xpzouying/xiaohongshu-mcp/cookies"
	"github.com/xpzouying/xiaohongshu-mcp/internal/infra/browser/playwright"
)

func main() {
	fmt.Println("🔍 分析小红书数据分析页面结构...")
	fmt.Println()

	cfg := playwright.DefaultConfig()
	cfg.Headless = false // 有头模式，便于观察
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

	// 导航到数据分析页面
	url := "https://creator.xiaohongshu.com/statistics/data-analysis?source=official"
	fmt.Printf("📍 导航到: %s\n", url)

	if err := pp.Goto(url); err != nil {
		fmt.Printf("❌ 导航失败: %v\n", err)
		return
	}

	fmt.Println("⏳ 等待页面加载...")
	time.Sleep(5 * time.Second)

	// 截图
	screenshotPath := "/tmp/xhs_data_analysis.png"
	if err := page.Screenshot(screenshotPath); err != nil {
		fmt.Printf("⚠️  截图失败: %v\n", err)
	} else {
		fmt.Printf("📸 截图已保存: %s\n", screenshotPath)
	}

	// 获取页面HTML结构
	fmt.Println("\n🔍 分析页面结构...")

	// 1. 查找所有可能包含笔记数据的容器
	result, err := pp.Eval(`() => {
		const containers = [];

		// 查找所有可能的容器类名
		const possibleSelectors = [
			'[class*="note"]',
			'[class*="card"]',
			'[class*="item"]',
			'[class*="content"]',
			'[class*="data"]',
			'[class*="row"]',
			'[class*="list"]',
			'table tr',
			'tbody tr'
		];

		const found = {};
		possibleSelectors.forEach(selector => {
			const elements = document.querySelectorAll(selector);
			if (elements.length > 0) {
				found[selector] = {
					count: elements.length,
					sampleClasses: elements[0] ? elements[0].className : '',
					sampleText: elements[0] ? elements[0].textContent.substring(0, 100) : ''
				};
			}
		});

		return JSON.stringify(found, null, 2);
	}`)

	if err != nil {
		fmt.Printf("❌ 分析失败: %v\n", err)
	} else {
		fmt.Println("找到的元素：")
		fmt.Println(result)
	}

	// 2. 查找表格结构
	fmt.Println("\n📊 查找表格结构...")
	tableResult, err := pp.Eval(`() => {
		const tables = document.querySelectorAll('table');
		const tableInfo = [];

		tables.forEach((table, index) => {
			const headers = [];
			const headerCells = table.querySelectorAll('thead th, thead td');
			headerCells.forEach(cell => headers.push(cell.textContent.trim()));

			const firstRow = table.querySelector('tbody tr');
			const firstRowData = [];
			if (firstRow) {
				const cells = firstRow.querySelectorAll('td');
				cells.forEach(cell => firstRowData.push(cell.textContent.trim().substring(0, 50)));
			}

			tableInfo.push({
				index: index,
				headers: headers,
				firstRowSample: firstRowData,
				rowCount: table.querySelectorAll('tbody tr').length
			});
		});

		return JSON.stringify(tableInfo, null, 2);
	}`)

	if err != nil {
		fmt.Printf("❌ 查找表格失败: %v\n", err)
	} else {
		fmt.Println(tableResult)
	}

	// 3. 获取页面完整的 class 名称列表
	fmt.Println("\n📋 获取所有唯一的 class 名称（前50个）...")
	classResult, err := pp.Eval(`() => {
		const classes = new Set();
		document.querySelectorAll('[class]').forEach(el => {
			el.className.split(' ').forEach(c => {
				if (c.trim()) classes.add(c.trim());
			});
		});
		return JSON.stringify(Array.from(classes).slice(0, 50));
	}`)

	if err != nil {
		fmt.Printf("❌ 获取 class 失败: %v\n", err)
	} else {
		fmt.Println(classResult)
	}

	// 4. 查找包含数字的元素（可能是数据）
	fmt.Println("\n🔢 查找包含数字的元素...")
	numberResult, err := pp.Eval(`() => {
		const elements = [];
		const allElements = document.querySelectorAll('*');

		allElements.forEach(el => {
			const text = el.textContent.trim();
			// 查找包含数字的文本
			if (/\d{2,}/.test(text) && el.children.length === 0) {
				const parent = el.parentElement;
				elements.push({
					text: text.substring(0, 100),
					tag: el.tagName,
					classes: el.className,
					parentTag: parent ? parent.tagName : '',
					parentClasses: parent ? parent.className : ''
				});
			}
		});

		return JSON.stringify(elements.slice(0, 20), null, 2);
	}`)

	if err != nil {
		fmt.Printf("❌ 查找数字元素失败: %v\n", err)
	} else {
		fmt.Println(numberResult)
	}

	fmt.Println("\n✅ 分析完成！")
	fmt.Println("📸 请查看截图和上面的输出来确定正确的选择器")
	fmt.Println("\n按 Ctrl+C 退出，浏览器窗口将保持打开30秒...")
	time.Sleep(30 * time.Second)
}
