#!/bin/bash

# 智能选择器诊断工具
# 用于检查选择器配置是否与实际页面匹配

echo "🔍 智能选择器诊断工具"
echo ""
echo "此工具会："
echo "1. 打开一个笔记详情页"
echo "2. 检查所有配置的选择器是否能找到元素"
echo "3. 输出诊断报告"
echo ""
echo "请确保已经登录小红书"
echo ""
read -p "按回车键继续..."

# 创建临时Go程序
cat > /tmp/selector_diagnostic.go << 'EOF'
package main

import (
	"fmt"
	"time"
	"github.com/xpzouying/xiaohongshu-mcp/browser"
	"gopkg.in/yaml.v3"
	"os"
)

type SelectorConfig struct {
	Elements map[string]*ElementSelectors `yaml:",inline"`
}

type ElementSelectors struct {
	Name     string         `yaml:"name"`
	Primary  []SelectorItem `yaml:"primary"`
	Fallback []SelectorItem `yaml:"fallback"`
}

type SelectorItem struct {
	Selector    string  `yaml:"selector"`
	Description string  `yaml:"description"`
}

func main() {
	// 读取配置
	data, err := os.ReadFile("configs/selectors.yaml")
	if err != nil {
		fmt.Printf("❌ 无法读取配置文件: %v\n", err)
		return
	}

	var config SelectorConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		fmt.Printf("❌ 无法解析配置文件: %v\n", err)
		return
	}

	// 打开浏览器
	b := browser.NewBrowser(false)
	defer b.Close()

	page := b.NewPage()

	// 导航到一个笔记详情页（使用示例URL）
	fmt.Println("📄 打开笔记详情页...")
	page.MustNavigate("https://www.xiaohongshu.com/explore/696872e7000000001a032e94?xsec_token=AB2AHlINujEqPed4q1rioho7X4rTy_TvlNRbc4EwiuO0k=&xsec_source=pc_feed")
	page.MustWaitDOMStable()
	time.Sleep(3 * time.Second)

	fmt.Println("\n🔍 开始检查选择器...\n")

	// 检查每个元素的选择器
	for name, element := range config.Elements {
		fmt.Printf("=== %s (%s) ===\n", name, element.Name)

		// 检查主选择器
		fmt.Println("主选择器:")
		for i, item := range element.Primary {
			found := checkSelector(page, item.Selector)
			status := "❌"
			if found {
				status = "✅"
			}
			fmt.Printf("  %d. %s %s - %s\n", i+1, status, item.Selector, item.Description)
		}

		// 检查备用选择器
		if len(element.Fallback) > 0 {
			fmt.Println("备用选择器:")
			for i, item := range element.Fallback {
				found := checkSelector(page, item.Selector)
				status := "❌"
				if found {
					status = "✅"
				}
				fmt.Printf("  %d. %s %s - %s\n", i+1, status, item.Selector, item.Description)
			}
		}

		fmt.Println()
	}

	fmt.Println("\n✅ 诊断完成！浏览器将保持打开10秒...")
	time.Sleep(10 * time.Second)
}

func checkSelector(page interface{}, selector string) bool {
	// 简化版本，实际应该使用 rod.Page
	return true // 占位符
}
EOF

echo "⚠️  此功能需要完整实现，当前仅为示例"
echo ""
echo "💡 建议："
echo "1. 智能选择器失败不影响功能（有fallback机制）"
echo "2. 如果功能正常工作，可以忽略此警告"
echo "3. 如果功能异常，需要更新 configs/selectors.yaml"
echo ""
echo "📝 手动检查方法："
echo "1. 打开小红书笔记详情页"
echo "2. 按F12打开开发者工具"
echo "3. 在Console中运行："
echo "   document.querySelector('#content-textarea')"
echo "4. 如果返回null，说明选择器需要更新"
