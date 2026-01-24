#!/bin/bash
# Git 历史清理脚本 - 彻底移除大文件

set -e

echo "======================================"
echo "  Git 仓库历史清理"
echo "======================================"
echo ""

echo "⚠️  警告: 此操作会重写 Git 历史！"
echo "⚠️  所有协作者需要重新克隆仓库！"
echo ""
read -p "确认继续? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo "已取消"
    exit 0
fi

# 检查是否安装 git-filter-repo
if ! command -v git-filter-repo &> /dev/null; then
    echo ""
    echo "📦 安装 git-filter-repo..."
    pip3 install git-filter-repo || {
        echo "❌ 安装失败，请手动安装:"
        echo "  pip3 install git-filter-repo"
        exit 1
    }
fi

echo ""
echo "🗑️  清理大型 GIF 文件历史..."

# 备份
BACKUP_DIR="../xiaohongshu-mcp-backup-$(date +%Y%m%d%H%M%S)"
echo "💾 创建备份: $BACKUP_DIR"
cp -r . "$BACKUP_DIR"

# 清理历史中的大文件
git filter-repo --path assets/inspect_mcp_publish.gif --invert-paths --force
git filter-repo --path assets/claude_push.gif --invert-paths --force
git filter-repo --path assets/check_login.gif --invert-paths --force

echo ""
echo "✅ 清理完成!"
echo ""
echo "📊 清理前后对比:"
echo "之前: ~154MB"
du -sh .git | awk '{print "现在: " $1}'

echo ""
echo "🚀 推送到远程:"
echo "  git push origin main --force"
echo ""
echo "⚠️  通知所有协作者执行:"
echo "  cd /path/to/xiaohongshu-mcp"
echo "  git fetch origin"
echo "  git reset --hard origin/main"
echo ""
