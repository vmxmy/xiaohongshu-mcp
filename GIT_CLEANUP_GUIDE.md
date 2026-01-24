# Git 仓库瘦身方案

## 问题分析

当前仓库大小：**154MB**

主要占用空间的文件：
- `assets/inspect_mcp_publish.gif` - 29MB
- `assets/claude_push.gif` - 20MB
- `assets/check_login.gif` - 11MB
- 其他图片资源 - 约 10MB

总计约 **70MB** 的媒体文件

## 建议方案

### 方案 1: 使用 Git LFS（推荐）

Git LFS (Large File Storage) 专门用于管理大文件。

**优点**:
- 不需要删除历史记录
- 大文件不会被克隆到本地
- 适合持续添加媒体文件

**实施步骤**:

```bash
# 1. 安装 Git LFS
brew install git-lfs  # macOS
# sudo apt install git-lfs  # Ubuntu

# 2. 初始化 Git LFS
git lfs install

# 3. 追踪大文件类型
git lfs track "*.gif"
git lfs track "*.mp4"
git lfs track "assets/*.png"
git lfs track "examples/**/*.png"

# 4. 提交 .gitattributes
git add .gitattributes
git commit -m "chore: enable Git LFS for media files"

# 5. 迁移已有文件到 LFS
git lfs migrate import --include="*.gif,*.mp4" --everything

# 6. 推送
git push origin main --force
```

### 方案 2: 压缩 GIF 文件（简单快速）

将大型 GIF 转换为优化的 MP4 或 WebP。

**优点**:
- 不改变 Git 结构
- 文件大小减少 70-90%

**实施步骤**:

```bash
# 安装 ffmpeg
brew install ffmpeg  # macOS

# 转换 GIF 为 MP4（压缩比 ~90%）
ffmpeg -i assets/inspect_mcp_publish.gif -vf "fps=10,scale=1280:-1" -c:v libx264 -pix_fmt yuv420p assets/inspect_mcp_publish_opt.mp4

# 或转换为优化的 GIF（压缩比 ~50%）
ffmpeg -i assets/inspect_mcp_publish.gif -vf "fps=10,scale=1280:-1" assets/inspect_mcp_publish_opt.gif
```

### 方案 3: 移动到外部存储（最彻底）

将媒体文件移动到外部CDN或图床。

**优点**:
- 仓库最小化
- 加载速度可能更快

**实施步骤**:

```bash
# 1. 上传文件到图床（如 GitHub Releases、图床服务）
# 2. 在 README 中使用外部链接
# 3. 从仓库删除原文件
git rm assets/*.gif
git commit -m "chore: move media files to external storage"
```

### 方案 4: 清理 Git 历史（激进，不推荐）

⚠️ **警告**: 会重写历史，影响所有协作者

```bash
# 使用 git-filter-repo 清理大文件
pip install git-filter-repo

# 删除大文件历史
git filter-repo --path assets/inspect_mcp_publish.gif --invert-paths
git filter-repo --path assets/claude_push.gif --invert-paths
git filter-repo --path assets/check_login.gif --invert-paths

# 强制推送
git push origin main --force
```

## 推荐实施

### 立即执行（减少 50%+）

1. **压缩现有 GIF**:
   ```bash
   cd assets
   # 转换为优化的 MP4
   ffmpeg -i inspect_mcp_publish.gif -vf "fps=10,scale=1280:-1" -c:v libx264 inspect_mcp_publish.mp4
   ffmpeg -i claude_push.gif -vf "fps=10,scale=1024:-1" -c:v libx264 claude_push.mp4
   ffmpeg -i check_login.gif -vf "fps=10,scale=1024:-1" -c:v libx264 check_login.mp4

   # 删除原 GIF
   git rm inspect_mcp_publish.gif claude_push.gif check_login.gif

   # 提交
   git add *.mp4
   git commit -m "chore: optimize media files, convert GIF to MP4"
   git push
   ```

2. **更新 .gitignore**:
   ```bash
   # 添加到 .gitignore
   echo "# Large media files" >> .gitignore
   echo "*.gif" >> .gitignore
   echo "assets/*.gif" >> .gitignore
   ```

### 长期策略（减少 90%+）

启用 Git LFS，未来所有媒体文件自动管理。

## 预期效果

| 方案 | 仓库大小 | 克隆时间 | 复杂度 |
|------|---------|---------|--------|
| 当前 | 154MB | ~60s | - |
| 压缩GIF | ~80MB | ~30s | 低 |
| Git LFS | ~10MB | ~5s | 中 |
| 外部存储 | ~5MB | ~2s | 高 |

## 注意事项

1. **备份**: 任何操作前先备份
2. **通知**: 如果重写历史，通知所有协作者
3. **README**: 更新文档中的媒体文件链接
4. **CI/CD**: 如果使用 LFS，确保 CI 系统支持

## 快速决策

- **只有你一个人维护**: 方案2（压缩）或方案3（外部存储）
- **团队协作，未来还要加媒体**: 方案1（Git LFS）
- **只是文档项目**: 方案3（外部存储）
- **追求极致**: 方案4（清理历史）+ 方案1（Git LFS）
