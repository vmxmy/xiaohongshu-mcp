#!/bin/bash
# 仅输出 cookies 的 base64 编码，方便复制粘贴

if [ ! -f cookies.json ]; then
    echo "错误: cookies.json 文件不存在" >&2
    echo "请先运行: ./bin/xiaohongshu-login" >&2
    exit 1
fi

echo "=== Cookies Base64 ==="
base64 -i cookies.json | tr -d '\n'
echo ""
echo ""
echo "💡 在远程服务器使用方法："
echo 'export COOKIES_BASE64="<上面的base64字符串>"'
echo 'curl -X POST http://localhost:18060/mcp -H "Content-Type: application/json" -d "{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"sync_cookies\",\"arguments\":{\"cookies_base64\":\"$COOKIES_BASE64\"}}}"'
