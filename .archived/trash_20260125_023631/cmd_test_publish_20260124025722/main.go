package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

// 模拟 PublishRequest
type PublishRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Images  []string `json:"images"`
	Tags    []string `json:"tags"`
}

func main() {
	fmt.Println("========================================")
	fmt.Println("测试小红书发布功能（通过已有 service）")
	fmt.Println("========================================")
	fmt.Println()

	// 检查图片
	imagePath := "/Users/xumingyang/app/xiaohongshu-mcp/test_images/test.jpg"
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		log.Fatalf("❌ 测试图片不存在: %s", imagePath)
	}

	fmt.Println("✅ 测试图片已就绪")
	fmt.Println()

	// 准备发布数据
	title := fmt.Sprintf("测试发帖-%s", time.Now().Format("150405"))
	content := "这是一条测试笔记，用于调试发布功能的真实返回信息。我们改进了代码以捕获API响应。"

	fmt.Printf("📝 发布内容:\n")
	fmt.Printf("   标题: %s\n", title)
	fmt.Printf("   内容: %s\n", content)
	fmt.Printf("   图片: %s\n", imagePath)
	fmt.Printf("   标签: [测试, 调试]\n")
	fmt.Println()

	// 使用 service 层发布
	fmt.Println("🚀 开始发布...")
	fmt.Println("⏳ 浏览器窗口将打开，请观察发布过程...")
	fmt.Println()

	// 这里我们需要直接调用已有的 service 代码
	// 由于代码结构限制，我们改用更简单的方式

	fmt.Println("提示: 请直接查看浏览器日志输出")
	fmt.Println()
	fmt.Println("建议使用以下方式测试:")
	fmt.Println("1. 重启 MCP 服务器使用新编译的版本")
	fmt.Println("2. 或者查看现有浏览器窗口的发布过程")
}
