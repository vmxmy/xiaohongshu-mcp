# Notion 工作空间访问指南

## 当前状态
你的 Notion 工作空间需要 API 访问权限才能通过 MCP 工具进行自动化访问。

## 方案一：设置 Notion API 集成（推荐）

### 步骤1：创建 Integration
1. 访问 https://www.notion.so/my-integrations
2. 点击 "New integration"
3. 填写信息：
   - Name: `Claude Assistant`
   - Associated workspace: 选择你的工作空间
   - Type: Internal integration
4. 点击 "Submit"
5. 复制生成的 **Internal Integration Token**

### 步骤2：共享页面给 Integration
1. 在 Notion 中选择要访问的页面
2. 点击页面右上角的 "Share"
3. 找到你创建的 integration
4. 点击 "Invite"

### 步骤3：提供 Token
```bash
# 环境变量方式
export NOTION_API_TOKEN="你的token在这里"

# 或直接提供给我（我会安全处理）
```

## 方案二：手动导出现有内容

### 快速导出方法
1. 在 Notion 中点击左上角的菜单
2. 选择 "Export"
3. 选择格式：
   - **Markdown**: 适合文本内容
   - **CSV**: 适合数据库内容
4. 包含子页面选择 "Include subpages"
5. 点击 "Export"

### 批量导出
如果内容很多，可以分批导出：
- 按页面类型导出（Database、Page、Collection）
- 按主题导出（项目、任务、笔记等）

## 方案三：使用现有工具查看

### 本地文件查看
如果你有已导出的 Notion 文件，我可以：
1. 读取和解析导出的文件
2. 分析页面结构和内容
3. 创建内容概览和分析

### 命令行查看
```bash
# 查看已导出的 Notion 文件
find ~/Documents -name "Notion*" -type f 2>/dev/null
ls ~/Downloads/*Notion* 2>/dev/null
```

## 下一步建议

1. **立即可行**: 提供导出的 Notion 文件，我可以立即分析
2. **长期解决方案**: 设置 Notion API 集成，获得自动化访问权限
3. **混合方案**: 导出 + API 设置两套方案同时进行

## 我能帮你做什么

- 📄 分析导出的 Notion 文件结构
- 📊 创建内容概览和分类
- 🔍 搜索特定主题内容
- 📋 提取和整理任务/项目信息
- 📈 分析数据趋势和模式
- 📝 生成内容摘要和行动计划

请告诉我你想通过哪种方式继续？