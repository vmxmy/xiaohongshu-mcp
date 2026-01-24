package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// 测试使用URL图片发布
func main() {
	ctx := context.Background()

	// 初始化发布用例
	publishUsecase := initPublishUsecase(false) // 非无头模式，方便观察
	if publishUsecase == nil {
		log.Fatal("初始化发布用例失败")
	}

	// 初始化服务
	service := NewXiaohongshuServiceWithUsecase(publishUsecase)

	// 准备发布请求 - 使用URL图片
	req := &PublishRequest{
		Title:   "URL图片测试",
		Content: "这是一个使用远程图片链接发布的测试。\n图片来自网络URL，会自动下载。",
		Images: []string{
			"https://pub-c918abf638c7475fa46f2a62c795106f.r2.dev/images/20260123-144120-313.png",
		},
		Tags: []string{"测试", "URL图片"},
	}

	fmt.Printf("开始发布测试...\n")
	fmt.Printf("标题: %s\n", req.Title)
	fmt.Printf("内容: %s\n", req.Content)
	fmt.Printf("图片URL: %v\n", req.Images)
	fmt.Printf("标签: %v\n\n", req.Tags)

	// 执行发布
	result, err := service.PublishContent(ctx, req)
	if err != nil {
		log.Fatalf("❌ 发布失败: %v", err)
	}

	// 输出结果
	resultJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("\n✅ 发布成功！\n")
	fmt.Printf("结果:\n%s\n", string(resultJSON))
}
