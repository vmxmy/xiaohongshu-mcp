#!/bin/bash
cd /Users/xumingyang/app/xiaohongshu-mcp
echo "=== 启动API分析工具 ==="
echo "浏览器将保持打开，请："
echo "1. 在浏览器中滚动数据分析页面"
echo "2. 观察控制台输出的API请求"
echo "3. 按Ctrl+C停止"
echo ""
go run analyze_api.go
