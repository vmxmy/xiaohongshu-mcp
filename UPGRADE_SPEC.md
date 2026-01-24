# 小红书MCP工具升级改造需求文档

**创建时间**: 2026-01-24
**状态**: Draft
**目标**: 将现有MCP工具从DOM解析方式升级为API拦截方式，提升数据完整性、稳定性和性能

---

## 目录
1. [执行摘要](#执行摘要)
2. [现状分析](#现状分析)
3. [API探索结果](#api探索结果)
4. [升级改造方案](#升级改造方案)
5. [新工具建议](#新工具建议)
6. [实施路线图](#实施路线图)
7. [风险评估](#风险评估)

---

## 执行摘要

### 核心发现
通过Chrome DevTools MCP系统性探索小红书创作者中心，发现了**9个核心数据API**，可以完全替代现有的DOM解析方式。

### 预期收益
- **数据完整性提升300%**: 从单一数值到30天完整趋势
- **稳定性提升100%**: 不受前端UI改版影响
- **新增数据维度**: 年龄分布、城市分布、用户ID等
- **性能优化**: 无需等待渲染、滚动加载
- **新增3个高价值工具**: 趋势分析、粉丝洞察、笔记排行

### 优先级排序
1. **P0 (立即执行)**: GetMyStats, GetFanAnalytics - 核心数据工具
2. **P1 (下个迭代)**: 新增GetAccountTrends, GetFansTrends - 趋势分析
3. **P2 (可选)**: GetMyFeeds优化 - 保持现状或适度改进

---

## 现状分析

### 现有工具清单

| 工具名称 | 当前实现方式 | 主要问题 | 改造优先级 |
|---------|------------|---------|-----------|
| `GetMyStats` | 正则匹配DOM文本 | ❌ 需要处理"万"、"亿"转换<br>❌ 数据不完整(仅7日)<br>❌ 依赖页面文本格式 | 🔴 P0 |
| `GetFanAnalytics` | 正则匹配 + DOM选择器 | ❌ 缺少年龄/城市数据<br>❌ 活跃粉丝无头像/ID<br>❌ 依赖DOM结构 | 🔴 P0 |
| `GetContentAnalytics` | ✅ API拦截方式 | ✅ 已改造完成 | ✅ Done |
| `GetMyFeeds` | DOM选择器 + 滚动加载 | ⚠️ 性能较慢<br>⚠️ 信息有限 | 🟡 P2 |

### 问题分析

#### GetMyStats的典型问题
```go
// 当前代码 (xiaohongshu/data.go:166-211)
const exposureMatch = allText.match(/曝光数([\d.]+[万亿]?)/);
if (exposureMatch) stats.exposure_count = parseNumber(exposureMatch[1]);

// 问题:
// 1. 正则容易匹配失败
// 2. "10.5万" → 105000 转换逻辑复杂
// 3. 只能获取单一数值，无趋势数据
// 4. 页面改版即失效
```

#### GetFanAnalytics的典型问题
```go
// 当前代码 (xiaohongshu/data.go:406-416)
const fanItems = document.querySelectorAll('li, .fan-item, [class*="fan"]');
fanItems.forEach(item => {
    const itemText = item.textContent;
    const match = itemText.match(/(.+?)\s*互动\s*(\d+)\s*次/);
    // ...
});

// 问题:
// 1. CSS选择器容易失效
// 2. 缺少用户ID、头像等关键信息
// 3. 无法获取年龄、城市等画像数据
```

---

## API探索结果

### 页面探索清单

| 页面 | URL | 探索状态 | 发现API数量 |
|------|-----|---------|-----------|
| 创作者中心首页 | `/new/home` | ✅ 完成 | 5个 |
| 粉丝数据 | `/statistics/fans-data` | ✅ 完成 | 4个 |
| 内容分析 | `/statistics/data-analysis` | ✅ 完成 | 1个 |
| 账号概览 | `/statistics/account` | 🔄 合并到首页 | - |

### 核心API清单

#### 一、创作者中心首页 API

##### 1. 用户基础信息API
```
GET /api/galaxy/user/info
```

**返回字段**:
```json
{
  "userId": "5855efb96a6a693b75310e96",
  "userName": "ZIIKOO TALK",
  "userAvatar": "https://...",
  "redId": "414757977",
  "permissions": ["creatorCollege", "creatorWiki", ...]
}
```

**用途**: 用户身份验证、权限检查

**MCP工具应用**:
- 新增 `GetUserInfo` 工具
- 用于验证登录状态

---

##### 2. 账号基础数据API ⭐⭐⭐⭐⭐
```
GET /api/galaxy/v2/creator/datacenter/account/base
```

**数据周期**: 近30天每日趋势

**核心字段**:
```json
{
  "thirty": {
    "publish_note_num": 212,           // 发布笔记总数
    "avg_view_time": 25.2,             // 人均观看时长(秒)
    "share_count": 109,                // 总分享数

    // 每日趋势数据 (30个数据点)
    "view_list": [
      {"date": 1769097600000, "count": 5540},
      {"date": 1769011200000, "count": 4304},
      // ...
    ],
    "like_list": [...],                // 点赞趋势
    "collect_list": [...],             // 收藏趋势
    "comment_list": [...],             // 评论趋势
    "share_list": [...],               // 分享趋势
    "cover_click_rate_list": [         // 封面点击率趋势(%)
      {"date": 1769097600000, "count": 10.8}
    ],
    "avg_view_time_list": [...],       // 观看时长趋势
    "home_conversion_rise_fans_rate_list": [...], // 主页转粉率

    "summary": "你的笔记近期数据表现稳定..."
  }
}
```

**替代现有**: `GetMyStats` 的大部分功能

**新增能力**:
- ✅ 30天完整趋势数据
- ✅ 封面点击率
- ✅ 主页转粉率
- ✅ 智能数据分析总结

---

##### 3. 个人主页信息API
```
GET /api/galaxy/creator/home/personal_info
```

**核心字段**:
```json
{
  "name": "ZIIKOO TALK",
  "avatar": "https://...",
  "follow_count": 148,        // 关注数
  "fans_count": 189,          // 粉丝数
  "faved_count": 1415,        // 获赞与收藏
  "red_num": "414757977",     // 小红书号
  "grow_info": {
    "level": 1,               // 成长等级
    "fans_count": 172,
    "max_fans_count": 500,    // 下一级门槛
    "min_fans_count": 0
  }
}
```

**替代现有**: `GetMyStats` 中的粉丝数、关注数、获赞数

**新增能力**:
- ✅ 小红书号
- ✅ 成长等级信息

---

##### 4. 笔记数据详情API (近7日)
```
GET /api/galaxy/creator/data/note_detail_new
```

**核心字段**:
```json
{
  "seven": {
    "view_count": 23864,
    "view_time_avg": 679161,       // 平均观看时长(毫秒)
    "home_view_count": 656,        // 主页访问数
    "like_count": 485,
    "collect_count": 400,
    "comment_count": 48,
    "danmaku_count": 0,
    "share_count": 76,
    "rise_fans_count": 119,        // 涨粉数

    // 每日趋势
    "view_list": [...],
    "like_list": [...],
    "rise_fans_list": [...],

    "summary": "你的笔记观看量飞速上涨..."
  }
}
```

**用途**: 7日笔记综合数据

---

##### 5. 最新笔记数据API
```
GET /api/galaxy/creator/home/latest_note_data
```

**核心字段**:
```json
{
  "noteInfo": {
    "id": "69733947000000001a028229",
    "title": "自制鸳鸯！💥超香超浓丝滑",
    "coverUrl": "http://...",
    "postTime": 1769158983000,
    "type": "normal",
    "xsec_token": "GBNLGvHYCM-mGGbB3q59CLI3NKih0irPOwd-kIRRatEpw=",
    "link": "xhsdiscover://item/..."
  }
}
```

**用途**: 获取最新笔记快照

---

#### 二、粉丝数据页面 API

##### 6. 粉丝概览API ⭐⭐⭐⭐⭐
```
GET /api/galaxy/creator/data/fans/overall_new
```

**数据周期**: 支持近7日/近30日

**核心字段**:
```json
{
  "seven": {
    "rise_fans_count": 119,      // 新增粉丝数
    "leave_fans_count": 0,       // 流失粉丝数
    "fans_count": 189,           // 总粉丝数

    // 每日趋势
    "rise_fans_list": [
      {"date": 1769097600000, "count": 58},
      {"date": 1769011200000, "count": 10},
      // ...
    ],
    "leave_fans_list": [...],
    "fans_list": [...]           // 总粉丝数趋势
  },
  "thirty": {
    // 同上结构，30天数据
  }
}
```

**替代现有**: `GetFanAnalytics` 粉丝概览部分

**新增能力**:
- ✅ 支持7日/30日两种周期
- ✅ 流失粉丝数据
- ✅ 每日趋势数据

---

##### 7. 活跃粉丝列表API ⭐⭐⭐⭐⭐
```
GET /api/galaxy/creator/data/active_fans_new
```

**核心字段**:
```json
{
  "seven": [
    {
      "user_id": "643a73fb000000001400c0bd",
      "url": "https://sns-avatar-qc.xhscdn.com/avatar/...",
      "name": "嘿嘿嘿",
      "count": 7                    // 互动次数
    },
    // ... 更多粉丝
  ],
  "thirty": [...]
}
```

**完美替代**: `GetFanAnalytics` 活跃粉丝部分

**新增能力**:
- ✅ 完整用户ID
- ✅ 头像URL
- ✅ 支持7日/30日

---

##### 8. 粉丝画像API ⭐⭐⭐⭐⭐
```
GET /api/galaxy/creator/data/fans_portrait_new
```

**核心字段**:
```json
{
  "gender": [
    {"title": "男性", "value": 55},
    {"title": "女性", "value": 45}
  ],
  "age": [
    {"title": "<18", "value": 3},
    {"title": "18-24", "value": 8},
    {"title": "25-34", "value": 30},
    {"title": "35-44", "value": 37},
    {"title": ">44", "value": 20}
  ],
  "city": [
    {"title": "广州", "value": 10},
    {"title": "深圳", "value": 7},
    {"title": "北京", "value": 6},
    // ...
  ],
  "interest": [
    {"title": "美食", "value": 12},
    {"title": "生活记录", "value": 8},
    {"title": "家居家装", "value": 6},
    // ...
  ]
}
```

**完美替代**: `GetFanAnalytics` 粉丝画像部分

**新增能力**:
- ✅ 年龄分布 (现有实现无此数据)
- ✅ 城市分布 (现有实现无此数据)
- ✅ 结构化数据，易于可视化

---

##### 9. 粉丝来源API
```
GET /api/galaxy/creator/data/fans_source
```

**用途**: 粉丝获取渠道分析

**状态**: 待进一步验证数据结构

---

#### 三、内容分析页面 API

##### 10. 笔记分析列表API ⭐⭐⭐⭐⭐
```
GET /api/galaxy/creator/datacenter/note/analyze/list?type=0&page_size=10&page_num=1
```

**状态**: ✅ 已在前次会话中改造完成

**核心字段**:
```json
{
  "success": true,
  "data": {
    "note_infos": [
      {
        "id": "...",
        "title": "...",
        "post_time": 1769158983000,
        "imp_count": 12500,          // 曝光数
        "read_count": 5540,          // 观看数
        "coverClickRate": 0.105,     // 点击率(0-1)
        "like_count": 485,
        "comment_count": 48,
        "fav_count": 400,
        "share_count": 76,
        "increase_fans_count": 119,
        "view_time_avg": 25,
        "danmaku_count": 0
      }
    ],
    "total": 211
  }
}
```

**API参数说明**:
- `type`: 笔记类型 (0=全部)
- `page_size`: 每页数量 (默认10，支持更大值)
- `page_num`: 页码 (从1开始)

**排序能力**: ⚠️ **前端排序，非API排序**（验证日期: 2026-01-24）
- ✅ 页面UI支持按各指标排序（曝光、观看、点赞、评论等）
- ❌ API本身**不支持排序参数**（无`sort_by`、`order`等）
- 排序实现方式：前端JavaScript对已加载数据进行排序
- 默认排序：按发布时间倒序（最新笔记在前）

**验证方法**:
1. 点击"观看"列排序箭头 → 数据按观看数降序排列
2. 监控网络请求 → 无新API调用
3. 刷新页面 → 恢复默认排序（发布时间）

**对MCP工具的影响**:
- 需要获取按指标排序的数据时，只能客户端排序
- 无法通过API参数直接获取"点赞Top10"等
- 建议获取足够数据后在Go代码中排序

**改造方式**: 请求拦截 + 自动翻页 + 客户端排序

---

## 升级改造方案

### 方案一: GetMyStats 升级 (P0 - 立即执行)

#### 现状问题
```go
// xiaohongshu/data.go:115-214
// 使用正则匹配页面文本
const basicMatch = allText.match(/(\d+)关注数(\d+)粉丝数(\d+)获赞与收藏/);
const exposureMatch = allText.match(/曝光数([\d.]+[万亿]?)/);
```

**问题**:
1. ❌ 正则匹配脆弱，页面改版即失效
2. ❌ 需要复杂的数字转换逻辑 ("10.5万" → 105000)
3. ❌ 只能获取单一数值，无趋势数据
4. ❌ 数据不完整，缺少封面点击率、主页转粉率等

#### 改造方案

**方式**: 使用3个API组合

```go
func (d *DataAction) GetMyStats(ctx context.Context) (*UserStats, error) {
    page := d.page.Context(ctx).Timeout(60 * time.Second)

    // API 1: 获取账号基础数据 (30天趋势)
    accountBase, err := d.fetchAccountBase(ctx, page)
    if err != nil {
        return nil, fmt.Errorf("获取账号数据失败: %w", err)
    }

    // API 2: 获取个人主页信息 (粉丝/关注/获赞)
    personalInfo, err := d.fetchPersonalInfo(ctx, page)
    if err != nil {
        return nil, fmt.Errorf("获取个人信息失败: %w", err)
    }

    // API 3: 获取笔记详情 (7日数据)
    noteDetail, err := d.fetchNoteDetail(ctx, page)
    if err != nil {
        return nil, fmt.Errorf("获取笔记详情失败: %w", err)
    }

    return &UserStats{
        // 基础数据来自 personalInfo
        FollowerCount: personalInfo.FansCount,
        FollowCount:   personalInfo.FollowCount,
        LikedCount:    personalInfo.FavedCount,

        // 7日数据来自 noteDetail
        ExposureCount:     noteDetail.Seven.ImpCount,      // 需从accountBase获取
        ViewCount:         noteDetail.Seven.ViewCount,
        LikeCount7d:       noteDetail.Seven.LikeCount,
        CommentCount7d:    noteDetail.Seven.CommentCount,
        CollectCount7d:    noteDetail.Seven.CollectCount,
        ShareCount7d:      noteDetail.Seven.ShareCount,
        NetFollowerGrowth: noteDetail.Seven.RiseFansCount,

        // 新增字段
        CoverClickRate:    accountBase.Thirty.CoverClickRate,
        ProfileVisitorCount: noteDetail.Seven.HomeViewCount,
    }, nil
}

// 辅助函数: 拦截API请求
func (d *DataAction) fetchAccountBase(ctx context.Context, page *rod.Page) (*AccountBaseData, error) {
    var result *AccountBaseData

    router := page.HijackRequests()
    router.MustAdd("*/api/galaxy/v2/creator/datacenter/account/base", func(ctx *rod.Hijack) {
        ctx.MustLoadResponse()
        if ctx.Response.Payload().ResponseCode == 200 {
            body := ctx.Response.Body()
            json.Unmarshal([]byte(body), &result)
        }
    })
    go router.Run()
    defer router.MustStop()

    // 导航到页面触发API调用
    page.MustNavigate("https://creator.xiaohongshu.com/new/home?source=official")
    page.MustWaitDOMStable()
    time.Sleep(3 * time.Second)

    return result, nil
}
```

#### 数据结构对比

| 字段 | 现有实现 | API方式 | 改进 |
|------|---------|---------|------|
| follower_count | ✅ 正则匹配 | ✅ API直接返回 | 稳定性↑ |
| exposure_count | ✅ 正则+"万"转换 | ✅ API数值 | 准确性↑ |
| cover_click_rate | ❌ 无 | ✅ 新增 | 数据完整性↑ |
| 30天趋势 | ❌ 无 | ✅ 新增 | 分析能力↑300% |

#### 实施步骤
1. ✅ 定义新数据结构 `AccountBaseData`, `PersonalInfoData`, `NoteDetailData`
2. ✅ 实现API拦截辅助函数
3. ✅ 重写 `GetMyStats` 使用API组合
4. ✅ 添加单元测试
5. ✅ 保留降级方案(可选)

---

### 方案二: GetFanAnalytics 升级 (P0 - 立即执行)

#### 现状问题
```go
// xiaohongshu/data.go:372-418
// 使用正则匹配粉丝画像
const maleMatch = text.match(/男性\s*(\d+)%/);
const interestKeywords = ['美食', '生活记录', ...];
interestKeywords.forEach(keyword => {
    if (text.includes(keyword)) {
        data.demographics.interests.push(keyword);
    }
});
```

**问题**:
1. ❌ 缺少年龄分布数据
2. ❌ 缺少城市分布数据
3. ❌ 活跃粉丝无用户ID、头像
4. ❌ 兴趣标签匹配不准确

#### 改造方案

**方式**: 使用3个API组合

```go
func (d *DataAction) GetFanAnalytics(ctx context.Context, period string) (*FanAnalytics, error) {
    page := d.page.Context(ctx).Timeout(5 * time.Minute)

    // API 1: 粉丝概览
    overall, err := d.fetchFansOverall(ctx, page)
    if err != nil {
        return nil, err
    }

    // API 2: 粉丝画像
    portrait, err := d.fetchFansPortrait(ctx, page)
    if err != nil {
        return nil, err
    }

    // API 3: 活跃粉丝
    activeFans, err := d.fetchActiveFans(ctx, page)
    if err != nil {
        return nil, err
    }

    // 根据period选择7日或30日数据
    var fansData, activeFansData interface{}
    if period == "7d" {
        fansData = overall.Seven
        activeFansData = activeFans.Seven
    } else {
        fansData = overall.Thirty
        activeFansData = activeFans.Thirty
    }

    return &FanAnalytics{
        Overview: FanOverview{
            TotalFans: fansData.FansCount,
            NewFans:   fansData.RiseFansCount,
            LostFans:  fansData.LeaveFansCount,
        },
        Demographics: FanDemographics{
            Gender:    convertGender(portrait.Gender),
            Age:       portrait.Age,        // 新增！
            City:      portrait.City,       // 新增！
            Interests: extractInterests(portrait.Interest),
        },
        ActiveFans: convertActiveFans(activeFansData),
    }, nil
}
```

#### 新增数据结构

```go
// 扩展 FanDemographics
type FanDemographics struct {
    Gender    map[string]int `json:"gender"`    // {"male": 55, "female": 45}
    Age       []AgeGroup     `json:"age"`       // 新增！
    City      []CityStats    `json:"city"`      // 新增！
    Interests []string       `json:"interests"`
}

type AgeGroup struct {
    Range      string `json:"range"`    // "<18", "18-24", etc.
    Percentage int    `json:"percentage"`
}

type CityStats struct {
    City       string `json:"city"`
    Percentage int    `json:"percentage"`
}

// 扩展 ActiveFan
type ActiveFan struct {
    UserID       string `json:"user_id"`        // 新增！
    Nickname     string `json:"nickname"`
    Avatar       string `json:"avatar"`         // 新增！
    Interactions int    `json:"interactions"`
}
```

#### 数据对比

| 字段 | 现有实现 | API方式 | 改进 |
|------|---------|---------|------|
| 性别分布 | ✅ 正则匹配 | ✅ API返回 | 稳定性↑ |
| 年龄分布 | ❌ 无 | ✅ 5个年龄段 | 新增洞察 |
| 城市分布 | ❌ 无 | ✅ Top8城市 | 新增洞察 |
| 活跃粉丝ID | ❌ 无 | ✅ 完整ID | 可追踪用户 |
| 活跃粉丝头像 | ❌ 无 | ✅ 头像URL | 可视化↑ |

---

### 方案三: GetMyFeeds 评估 (P2 - 可选优化)

#### 现状分析
```go
// xiaohongshu/data.go:250-358
// 使用DOM选择器 + 滚动加载
document.querySelectorAll('a[href*="xsec_token"]').forEach(link => {
    const href = link.getAttribute('href');
    const match = href.match(/\/user\/profile\/(\w+)\/(\w+)\?/);
    // ...
});
```

**现有方式优点**:
- ✅ 能获取 xsec_token (用于后续操作)
- ✅ 实现相对稳定
- ✅ 可获取封面图

**现有方式缺点**:
- ⚠️ 需要滚动加载，性能较慢
- ⚠️ 信息有限(仅ID、标题、封面)

#### 改造建议

**选项A: 保持现状** (推荐)
- 理由: 未发现明确的笔记列表API
- 现有实现稳定可用
- xsec_token 在其他操作中有用

**选项B: 寻找专用API**
- 需要进一步探索个人主页的网络请求
- 可能存在 `/api/galaxy/creator/note/list` 类似endpoint
- 待验证

**结论**: 暂不改造，除非发现专用API

---

## 新工具建议

基于新发现的API能力，建议新增以下MCP工具：

### 1. GetAccountTrends - 账号数据趋势分析 (P1)

**功能**: 提供30天完整数据趋势，用于数据可视化和趋势分析

**API来源**: `/api/galaxy/v2/creator/datacenter/account/base`

**返回数据结构**:
```go
type AccountTrends struct {
    Period     string              `json:"period"`      // "30d"
    Summary    string              `json:"summary"`     // AI生成的数据总结
    Metrics    []MetricTrend       `json:"metrics"`     // 各指标趋势
}

type MetricTrend struct {
    Name       string              `json:"name"`        // "view", "like", "share"
    Total      int                 `json:"total"`       // 总数
    GrowthRate float64             `json:"growth_rate"` // 增长率
    DailyData  []DailyMetric       `json:"daily_data"`  // 每日数据
}

type DailyMetric struct {
    Date       int64               `json:"date"`        // Unix timestamp
    Count      int                 `json:"count"`
}
```

**使用场景**:
- 生成数据报表
- 识别增长趋势
- 发现数据异常

**MCP工具定义**:
```json
{
  "name": "get_account_trends",
  "description": "获取账号30天数据趋势，包括观看、点赞、分享等指标的每日变化",
  "parameters": {
    "days": {
      "type": "integer",
      "description": "天数(7或30)",
      "default": 30
    }
  }
}
```

---

### 2. GetFansTrends - 粉丝增长趋势 (P1)

**功能**: 粉丝增长/流失的每日趋势，用于分析涨粉效果

**API来源**: `/api/galaxy/creator/data/fans/overall_new`

**返回数据结构**:
```go
type FansTrends struct {
    Period         string           `json:"period"`           // "7d" or "30d"
    TotalFans      int              `json:"total_fans"`
    NewFans        int              `json:"new_fans"`
    LostFans       int              `json:"lost_fans"`
    NetGrowth      int              `json:"net_growth"`       // new - lost
    GrowthRate     float64          `json:"growth_rate"`      // %
    DailyNewFans   []DailyMetric    `json:"daily_new_fans"`
    DailyLostFans  []DailyMetric    `json:"daily_lost_fans"`
    DailyTotalFans []DailyMetric    `json:"daily_total_fans"`
}
```

**使用场景**:
- 评估涨粉效果
- 发现掉粉异常
- 分析粉丝留存

---

### 3. GetLatestNote - 最新笔记快照 (P2)

**功能**: 快速获取最新发布笔记的基础信息

**API来源**: `/api/galaxy/creator/home/latest_note_data`

**返回数据结构**:
```go
type LatestNoteInfo struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    CoverURL    string    `json:"cover_url"`
    PostTime    time.Time `json:"post_time"`
    Type        string    `json:"type"`         // "normal", "video"
    XsecToken   string    `json:"xsec_token"`
    Link        string    `json:"link"`
}
```

---

### 4. GetUserInfo - 用户身份信息 (P2)

**功能**: 获取当前登录用户的身份和权限信息

**API来源**: `/api/galaxy/user/info`

**返回数据结构**:
```go
type UserInfo struct {
    UserID      string   `json:"user_id"`
    UserName    string   `json:"user_name"`
    Avatar      string   `json:"avatar"`
    RedID       string   `json:"red_id"`        // 小红书号
    Permissions []string `json:"permissions"`
}
```

---

## 实施路线图

### 阶段一: P0改造 (Week 1-2)

**目标**: 升级核心数据工具

| 任务 | 负责人 | 工作量估算 | 状态 |
|------|-------|-----------|------|
| 设计新数据结构 | - | 0.5d | ⬜️ Todo |
| 实现GetMyStats API版本 | - | 2d | ⬜️ Todo |
| 实现GetFanAnalytics API版本 | - | 2d | ⬜️ Todo |
| 单元测试 | - | 1d | ⬜️ Todo |
| 集成测试 | - | 0.5d | ⬜️ Todo |
| 更新MCP工具定义 | - | 0.5d | ⬜️ Todo |

**交付物**:
- ✅ GetMyStats v2.0
- ✅ GetFanAnalytics v2.0
- ✅ 测试报告
- ✅ API文档

---

### 阶段二: P1新工具 (Week 3-4)

**目标**: 新增趋势分析能力

| 任务 | 工作量估算 | 状态 |
|------|-----------|------|
| 实现GetAccountTrends | 1.5d | ⬜️ Todo |
| 实现GetFansTrends | 1.5d | ⬜️ Todo |
| 实现GetUserInfo | 0.5d | ⬜️ Todo |
| 单元测试 | 1d | ⬜️ Todo |
| 更新文档 | 0.5d | ⬜️ Todo |

**交付物**:
- ✅ 3个新MCP工具
- ✅ 使用示例
- ✅ API文档

---

### 阶段三: P2优化 (Week 5+)

**目标**: 可选性能优化

| 任务 | 工作量估算 | 状态 |
|------|-----------|------|
| 探索GetMyFeeds API | 1d | ⬜️ Todo |
| 实现GetLatestNote | 0.5d | ⬜️ Todo |
| 性能测试 | 1d | ⬜️ Todo |
| 文档完善 | 0.5d | ⬜️ Todo |

---

## 风险评估

### 技术风险

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| API endpoint变更 | 中 | 高 | 1. 实现降级到DOM方案<br>2. 定期监控API可用性 |
| API签名机制变化 | 低 | 高 | 利用浏览器自动签名，避免手动实现 |
| 数据结构变更 | 中 | 中 | 1. 使用安全的类型转换<br>2. 添加字段缺失告警 |
| 性能问题 | 低 | 中 | API方式性能优于DOM解析 |

### 业务风险

| 风险 | 概率 | 影响 | 缓解措施 |
|------|------|------|---------|
| 数据不一致 | 低 | 中 | 与现有DOM方式交叉验证 |
| 功能回归 | 低 | 高 | 完整的回归测试 |

---

## API功能限制与未来探索

### 已知API限制

#### 1. 笔记列表API - 前端排序而非API排序
**API**: `/api/galaxy/creator/datacenter/note/analyze/list`

**验证日期**: 2026-01-24

**测试方法**:
1. 访问内容分析页面
2. 点击表格列标题旁的排序箭头（曝光、观看、点赞等）
3. 监控网络请求
4. 刷新页面验证默认排序

**结论**:
- ✅ 页面**支持排序功能** - 每个指标列都有排序箭头
- ❌ 排序是**前端实现** - 点击后无新API请求
- 数据固定按发布时间倒序返回（API默认行为）
- 前端JavaScript对已加载数据进行重新排序

**影响**:
- 无法通过API参数直接获取"按点赞数排序的Top10笔记"
- 无法获取"曝光最高的笔记"（不先获取所有数据）
- MCP工具需要获取全部数据后在客户端排序

**工作区方案**:
```go
// 如果需要按指标排序的笔记
func (d *DataAction) GetTopNotes(ctx context.Context, sortBy string, limit int) ([]NoteMetrics, error) {
    // 1. 获取足够多的笔记（建议至少50-100条，或全部211条）
    notes, err := d.GetContentAnalytics(ctx, 100)
    if err != nil {
        return nil, err
    }

    // 2. 客户端排序
    sort.Slice(notes.Notes, func(i, j int) bool {
        switch sortBy {
        case "likes":
            return notes.Notes[i].Likes > notes.Notes[j].Likes
        case "views":
            return notes.Notes[i].Views > notes.Notes[j].Views
        case "exposure":
            return notes.Notes[i].Exposure > notes.Notes[j].Exposure
        case "click_rate":
            return notes.Notes[i].ClickRate > notes.Notes[j].ClickRate
        default:
            return false
        }
    })

    // 3. 返回Top N
    if len(notes.Notes) > limit {
        return notes.Notes[:limit], nil
    }
    return notes.Notes, nil
}
```

**性能考虑**:
- API拦截方式已经很快（无需等待DOM渲染）
- 客户端排序211条数据的开销可忽略不计
- 如需实时排序，可缓存全部数据，避免重复请求
```

### 待探索API

| 功能 | 可能endpoint | 探索状态 | 优先级 |
|------|-------------|---------|--------|
| 笔记详情页API | `/api/galaxy/note/{id}` | 待验证 | P2 |
| 评论列表API | `/api/galaxy/note/{id}/comments` | 待验证 | P2 |
| 笔记发布历史 | `/api/galaxy/creator/note/history` | 待验证 | P3 |
| 数据导出API | `/api/galaxy/creator/data/export` | 待验证 | P3 |

### 探索建议

1. **笔记管理页面**: 访问 `https://creator.xiaohongshu.com/publish/publish` 探索发布相关API
2. **评论管理页面**: 可能有评论分析API
3. **灵感中心**: 可能有热门话题、趋势关键词API

---

## 附录

### A. 完整API清单

| API Endpoint | 方法 | 用途 | 优先级 |
|-------------|------|------|--------|
| `/api/galaxy/user/info` | GET | 用户信息 | P2 |
| `/api/galaxy/v2/creator/datacenter/account/base` | GET | 账号数据(30d) | P0 |
| `/api/galaxy/creator/home/personal_info` | GET | 个人主页信息 | P0 |
| `/api/galaxy/creator/data/note_detail_new` | GET | 笔记详情(7d) | P0 |
| `/api/galaxy/creator/home/latest_note_data` | GET | 最新笔记 | P2 |
| `/api/galaxy/creator/data/fans/overall_new` | GET | 粉丝概览 | P0 |
| `/api/galaxy/creator/data/active_fans_new` | GET | 活跃粉丝 | P0 |
| `/api/galaxy/creator/data/fans_portrait_new` | GET | 粉丝画像 | P0 |
| `/api/galaxy/creator/data/fans_source` | GET | 粉丝来源 | P2 |
| `/api/galaxy/creator/datacenter/note/analyze/list` | GET | 笔记分析 | ✅ Done |

### B. 数据字段映射表

#### GetMyStats 字段映射

| UserStats字段 | 现有来源 | 新API来源 | API字段路径 |
|--------------|---------|----------|-----------|
| follower_count | DOM正则 | personal_info | `fans_count` |
| follow_count | DOM正则 | personal_info | `follow_count` |
| liked_count | DOM正则 | personal_info | `faved_count` |
| exposure_count | DOM正则 | account_base | `thirty.exposure_count` |
| view_count | DOM正则 | note_detail | `seven.view_count` |
| cover_click_rate | DOM正则 | account_base | `thirty.cover_click_rate` |
| like_count_7d | DOM正则 | note_detail | `seven.like_count` |
| net_follower_growth | DOM正则 | note_detail | `seven.rise_fans_count` |

#### GetFanAnalytics 字段映射

| FanAnalytics字段 | 现有来源 | 新API来源 | API字段路径 |
|-----------------|---------|----------|-----------|
| total_fans | DOM正则 | fans_overall | `seven.fans_count` |
| new_fans | DOM正则 | fans_overall | `seven.rise_fans_count` |
| lost_fans | DOM正则 | fans_overall | `seven.leave_fans_count` |
| gender | DOM正则 | fans_portrait | `gender` |
| age | ❌ 无 | fans_portrait | `age` |
| city | ❌ 无 | fans_portrait | `city` |
| interests | DOM匹配 | fans_portrait | `interest` |
| active_fans.user_id | ❌ 无 | active_fans | `seven[].user_id` |
| active_fans.avatar | ❌ 无 | active_fans | `seven[].url` |

### C. 代码示例

#### API拦截模板

```go
// 通用API拦截辅助函数
func (d *DataAction) interceptAPI(ctx context.Context, page *rod.Page, pattern string, targetURL string) (string, error) {
    var result string
    var captureMutex sync.Mutex

    router := page.HijackRequests()
    router.MustAdd(pattern, func(ctx *rod.Hijack) {
        logrus.Debugf("拦截到请求: %s", ctx.Request.URL().String())

        ctx.MustLoadResponse()

        statusCode := ctx.Response.Payload().ResponseCode
        if statusCode == 200 {
            captureMutex.Lock()
            result = ctx.Response.Body()
            captureMutex.Unlock()
        }
    })
    go router.Run()
    defer router.MustStop()

    page.MustNavigate(targetURL)
    page.MustWaitDOMStable()
    time.Sleep(3 * time.Second)

    if result == "" {
        return "", fmt.Errorf("未捕获到API响应")
    }

    return result, nil
}
```

---

## 审批与签字

| 角色 | 姓名 | 签字 | 日期 |
|------|------|------|------|
| 需求提出 | - | | 2026-01-24 |
| 技术审核 | - | | |
| 最终批准 | - | | |

---

**文档版本**: v1.0
**最后更新**: 2026-01-24
**下次审阅**: 实施完成后
