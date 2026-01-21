# 小红书创作者中心可用数据汇总

## 数据来源页面

### 1. 首页 (https://creator.xiaohongshu.com/new/home)

**基础账号数据**：
- 关注数
- 粉丝数
- 获赞与收藏总数
- 小红书账号ID

**笔记数据总览（近7日/近30日）**：
- 曝光数
- 观看数
- 封面点击率 (%)
- 视频完播率 (%)
- 点赞数
- 评论数
- 收藏数
- 分享数
- 净涨粉
- 新增关注
- 取消关注
- 主页访客

### 2. 数据看板 - 账号概览 (https://creator.xiaohongshu.com/statistics/account/v2)

**账号诊断数据（近7日/近30日）**：
- 观看数（含同类创作者排名百分比）
- 涨粉数（含同类创作者排名百分比）
- 主页访客数（含同类创作者排名百分比）
- 发布数（含同类创作者排名百分比）
- 互动数（含同类创作者排名百分比）

**详细观看数据**：
- 曝光数
- 观看数
- 封面点击率
- 平均观看时长（秒）
- 观看总时长（小时）
- 视频完播率
- 观看来源分布
- 观看时段分布

### 3. 数据看板 - 粉丝数据 (https://creator.xiaohongshu.com/statistics/fans-data)

**粉丝概览（近7天/近30天）**：
- 总粉丝数
- 新增粉丝数
- 流失粉丝数

**粉丝数据趋势**：
- 总粉丝数趋势图
- 新增粉丝数趋势图
- 流失粉丝数趋势图

**粉丝画像**：
- 性别分布（男性/女性百分比）
- 年龄分布
- 兴趣分布（美食、生活记录、社科、娱乐、家居家装、影视、科技数码、职场等）
- 地域分布
- 粉丝来源

**活跃粉丝列表（近7天/近30天）**：
- 粉丝昵称
- 互动次数
- Top 10活跃粉丝

### 4. 数据看板 - 内容分析 (https://creator.xiaohongshu.com/statistics/data-analysis)

**笔记数据列表**：
- 笔记标题
- 发布时间
- 曝光数
- 观看数
- 封面点击率
- 点赞数
- 评论数
- 收藏数
- 涨粉数
- 分享数
- 人均观看时长
- 违规状态

**直播数据**：
- 直播场次
- 观看人数
- 互动人数
- 人均停留时长
- 支付金额
- 支付订单数
- 送礼人数
- 薯钻

### 5. 笔记管理 (https://creator.xiaohongshu.com/new/note-manager)

**笔记列表管理**：
- 所有已发布笔记
- 笔记状态（正常/违规/审核中）
- 笔记基本信息

## 推荐的MCP工具设计

### 工具1: get_creator_stats (获取创作者统计数据)
**参数**：
- `period`: 统计周期 (7d/30d)

**返回数据**：
```json
{
  "basic": {
    "follower_count": 96,
    "follow_count": 34,
    "liked_count": 785,
    "account_id": "414757977"
  },
  "overview_7d": {
    "exposure_count": 105000,
    "view_count": 11000,
    "cover_click_rate": 10.8,
    "video_complete_rate": 0,
    "like_count": 375,
    "comment_count": 21,
    "collect_count": 286,
    "share_count": 48,
    "net_follower_growth": 63,
    "new_follower_count": 63,
    "unfollow_count": 0,
    "profile_visitor_count": 328
  },
  "ranking": {
    "view_rank_percent": 99,
    "follower_growth_rank_percent": 99,
    "visitor_rank_percent": 99,
    "publish_rank_percent": 99,
    "interaction_rank_percent": 99
  }
}
```

### 工具2: get_fan_analytics (获取粉丝分析数据)
**参数**：
- `period`: 统计周期 (7d/30d)

**返回数据**：
```json
{
  "overview": {
    "total_fans": 96,
    "new_fans": 63,
    "lost_fans": 0
  },
  "demographics": {
    "gender": {"male": 59, "female": 41},
    "age_distribution": {...},
    "interests": ["美食", "生活记录", "社科", ...],
    "regions": [...]
  },
  "active_fans": [
    {"nickname": "偏方", "interactions": 32},
    ...
  ]
}
```

### 工具3: get_content_analytics (获取内容分析数据)
**参数**：
- `limit`: 返回笔记数量
- `start_date`: 开始日期
- `end_date`: 结束日期

**返回数据**：
```json
{
  "notes": [
    {
      "title": "💥三件五折！艾丽达新品来袭！",
      "publish_time": "2026-01-21 00:32",
      "exposure": 0,
      "views": 3,
      "click_rate": 0,
      "likes": 0,
      "comments": 0,
      "collects": 0,
      "shares": 0,
      "follower_growth": 0,
      "status": "normal"
    },
    ...
  ]
}
```

### 工具4: get_detailed_metrics (获取详细指标)
**参数**：
- `period`: 统计周期 (7d/30d)

**返回数据**：
```json
{
  "view_metrics": {
    "exposure_count": 105000,
    "view_count": 11000,
    "cover_click_rate": 10.8,
    "avg_view_duration": 28,
    "total_view_duration_hours": 105.5,
    "video_complete_rate": 0,
    "view_sources": {...},
    "view_time_distribution": {...}
  }
}
```

## 实现优先级

### 高优先级（立即实现）：
1. ✅ get_creator_stats - 基础统计数据（已在v1.0.3中实现）
2. get_fan_analytics - 粉丝分析数据
3. get_content_analytics - 内容分析数据

### 中优先级（后续实现）：
4. get_detailed_metrics - 详细观看指标
5. get_live_analytics - 直播数据分析

### 低优先级（可选）：
6. export_data - 数据导出功能
7. get_trending_topics - 热门话题推荐

## 技术实现要点

1. **统一入口**：所有数据都通过创作者中心获取
2. **登录状态**：需要保持创作者中心的登录状态
3. **数据提取**：使用JavaScript从页面DOM提取数据
4. **缓存策略**：考虑缓存数据，避免频繁请求
5. **错误处理**：处理页面加载失败、数据缺失等情况
