# 小红书 MCP 工具模块文档

## 项目概述

xiaohongshu-mcp 是一个基于 Model Context Protocol (MCP) 的小红书自动化服务，提供了 23 个工具模块，涵盖了登录、内容发布、数据浏览、互动功能、内容管理和数据分析等全方位的功能。

## 工具分类

### 1. 基础功能（4 个）

| 工具名称 | 功能描述 | 只读 | 危险 |
|---------|---------|------|------|
| `check_login_status` | 检查小红书登录状态，返回用户名 | ✅ | ❌ |
| `get_login_qrcode` | 获取登录二维码（Base64 图片 + 超时时间） | ✅ | ❌ |
| `sync_cookies` | 上传 cookies JSON 并写入服务端文件 | ❌ | ❌ |
| `delete_cookies` | 删除 cookies 文件，重置登录状态 | ❌ | ✅ |

#### 1.1 check_login_status

**功能描述**: 检查当前小红书账号的登录状态

**参数**: 无

**返回示例**:
```json
{
  "is_logged_in": true,
  "username": "用户名"
}
```

**使用场景**:
- 在执行需要登录的操作前验证状态
- 定期检查账号是否仍保持登录
- 诊断登录相关的问题

---

#### 1.2 get_login_qrcode

**功能描述**: 获取登录二维码图片和超时时间

**参数**: 无

**返回内容**:
- 文本提示：截止时间
- 图片：Base64 编码的二维码图片（PNG 格式）
- 已登录状态：如果已登录则不返回图片

**使用场景**:
- 首次登录账号
- Cookies 过期后重新登录
- 更换账号

**注意事项**:
- 二维码有效期为 4 分钟
- 需要使用小红书 App 扫码
- 扫码成功后 cookies 会自动保存

---

#### 1.3 sync_cookies

**功能描述**: 上传 cookies JSON 并写入服务端文件，适用于本地有头扫码后同步 cookies 到服务端。

**参数**:
- `cookies_base64` (string): Base64 编码的 cookies JSON（推荐）
- `cookies_json` (string): cookies JSON 字符串（备用）

**返回内容**:
- cookies 写入路径与字节大小

**使用场景**:
- 本地有头扫码登录后同步 cookies
- 无头服务器快速恢复登录状态

**注意事项**:
- 仅记录 cookies 字节长度，不打印内容
- 上传前请确认 cookies 文件来自可信环境

---

#### 1.4 delete_cookies

**功能描述**: 删除本地 cookies 文件，重置登录状态

**参数**: 无

**返回内容**:
- 操作结果提示
- 删除的 cookies 文件路径

**使用场景**:
- 账号被封禁或异常需要重新登录
- 切换账号
- 清理本地数据

**注意事项**:
- ⚠️ **危险操作**：删除后需要重新登录
- Cookies 保存位置：`~/.xiaohongshu/cookies.json`

---

### 2. 内容发布（2 个）

| 工具名称 | 功能描述 | 只读 | 危险 |
|---------|---------|------|------|
| `publish_content` | 发布小红书图文内容（支持 URL 或本地图片） | ❌ | ✅ |
| `publish_with_video` | 发布小红书视频内容（仅支持本地视频） | ❌ | ✅ |

#### 2.1 publish_content

**功能描述**: 发布图文内容到小红书

**参数**:
```json
{
  "title": "内容标题（最多20个中文字或英文单词）",
  "content": "正文内容（不包含#开头的标签）",
  "images": ["图片路径列表"],
  "tags": ["话题标签列表"],
  "schedule_at": "定时发布时间（可选，ISO8601格式）"
}
```

**参数说明**:
- `title`: 必填，标题限制 40 个单位长度（中文/日文/韩文占 2 单位，英文/数字占 1 单位）
- `content`: 必填，正文内容
- `images`: 必填，至少 1 张图片
  - 支持 HTTP/HTTPS 链接（会自动下载）
  - 支持本地绝对路径（推荐）
- `tags`: 可选，话题标签列表
- `schedule_at`: 可选，定时发布时间
  - 格式：ISO8601，如 `2024-01-20T10:30:00+08:00`
  - 时间范围：1 小时至 14 天内

**返回示例**:
```json
{
  "title": "标题",
  "content": "正文内容",
  "images": 3,
  "status": "发布完成",
  "post_id": "笔记ID"
}
```

**使用场景**:
- 自动化内容发布
- 批量发布图文笔记
- 定时发布内容

**最佳实践**:
- 推荐使用本地图片路径，更稳定
- 标题简洁有力，符合小红书风格
- 添加合适的标签增加曝光

---

#### 2.2 publish_with_video

**功能描述**: 发布视频内容到小红书

**参数**:
```json
{
  "title": "内容标题（最多20个中文字或英文单词）",
  "content": "正文内容（不包含#开头的标签）",
  "video": "本地视频绝对路径",
  "tags": ["话题标签列表"],
  "schedule_at": "定时发布时间（可选，ISO8601格式）"
}
```

**参数说明**:
- `title`: 必填，同图文发布
- `content`: 必填，视频描述
- `video`: 必填，仅支持本地视频文件绝对路径
  - ⚠️ 不支持 HTTP 链接
  - 建议视频大小不超过 1GB
- `tags`: 可选，话题标签
- `schedule_at`: 可选，定时发布时间（同图文）

**返回示例**:
```json
{
  "title": "标题",
  "content": "正文内容",
  "video": "/path/to/video.mp4",
  "status": "发布完成",
  "post_id": "笔记ID"
}
```

**使用场景**:
- 发布短视频内容
- Vlog 分享
- 教程类视频发布

**注意事项**:
- 视频处理时间较长，请耐心等待
- 仅支持本地视频文件
- 建议提前测试视频格式兼容性

---

### 3. 内容浏览（4 个）

| 工具名称 | 功能描述 | 只读 | 危险 |
|---------|---------|------|------|
| `list_feeds` | 获取首页推荐 Feeds 列表 | ✅ | ❌ |
| `search_feeds` | 搜索小红书内容（支持筛选） | ✅ | ❌ |
| `get_feed_detail` | 获取笔记详情（含评论） | ✅ | ❌ |
| `user_profile` | 获取用户主页信息 | ✅ | ❌ |

#### 3.1 list_feeds

**功能描述**: 获取小红书首页推荐内容列表

**参数**: 无

**返回内容**:
```json
{
  "feeds": [
    {
      "id": "笔记ID",
      "xsecToken": "访问令牌",
      "modelType": "笔记类型",
      "noteCard": {
        "type": "normal",
        "displayTitle": "标题",
        "user": {
          "userId": "用户ID",
          "nickname": "昵称",
          "avatar": "头像URL"
        },
        "interactInfo": {
          "liked": false,
          "likedCount": "点赞数",
          "collected": false,
          "collectedCount": "收藏数",
          "commentCount": "评论数",
          "sharedCount": "分享数"
        },
        "cover": {
          "url": "封面URL",
          "width": 720,
          "height": 1080
        }
      }
    }
  ],
  "count": 10
}
```

**使用场景**:
- 浏览推荐内容
- 获取热门话题
- 发现优质创作者

---

#### 3.2 search_feeds

**功能描述**: 根据关键词搜索小红书内容

**参数**:
```json
{
  "keyword": "搜索关键词",
  "filters": {
    "sort_by": "综合|最新|最多点赞|最多评论|最多收藏",
    "note_type": "不限|视频|图文",
    "publish_time": "不限|一天内|一周内|半年内",
    "search_scope": "不限|已看过|未看过|已关注",
    "location": "不限|同城|附近"
  }
}
```

**参数说明**:
- `keyword`: 必填，搜索关键词
- `filters`: 可选，筛选条件
  - `sort_by`: 排序依据，默认"综合"
  - `note_type`: 笔记类型，默认"不限"
  - `publish_time`: 发布时间，默认"不限"
  - `search_scope`: 搜索范围，默认"不限"
  - `location`: 位置距离，默认"不限"

**返回内容**: 同 `list_feeds`

**使用场景**:
- 搜索特定话题内容
- 查找竞品笔记
- 研究热门内容趋势

---

#### 3.3 get_feed_detail

**功能描述**: 获取笔记完整详情，包括评论列表

**参数**:
```json
{
  "feed_id": "笔记ID（从Feed列表获取）",
  "xsec_token": "访问令牌（从Feed列表获取）",
  "load_all_comments": false,
  "limit": 20,
  "click_more_replies": false,
  "reply_limit": 10,
  "scroll_speed": "normal"
}
```

**参数说明**:
- `feed_id`: 必填，笔记 ID
- `xsec_token`: 必填，访问令牌
- `load_all_comments`: 可选，是否加载全部评论
  - `false`: 仅返回前 10 条一级评论（默认）
  - `true`: 滚动加载更多评论
- `limit`: 可选，限制加载的一级评论数量（仅在 `load_all_comments=true` 时生效），默认 20
- `click_more_replies`: 可选，是否展开二级回复（仅在 `load_all_comments=true` 时生效），默认 `false`
- `reply_limit`: 可选，跳过回复数过多的评论（仅在 `click_more_replies=true` 时生效），默认 10
- `scroll_speed`: 可选，滚动速度（仅在 `load_all_comments=true` 时生效）
  - `slow`: 慢速
  - `normal`: 正常（默认）
  - `fast`: 快速

**返回内容**:
```json
{
  "note": {
    "noteId": "笔记ID",
    "xsecToken": "访问令牌",
    "title": "标题",
    "desc": "正文描述",
    "type": "normal",
    "time": 1640000000,
    "ipLocation": "IP位置",
    "user": {
      "userId": "用户ID",
      "nickname": "昵称"
    },
    "interactInfo": {
      "liked": false,
      "likedCount": "点赞数",
      "collectedCount": "收藏数",
      "commentCount": "评论数"
    },
    "imageList": [
      {
        "width": 720,
        "height": 1080,
        "urlDefault": "图片URL"
      }
    ]
  },
  "comments": {
    "list": [
      {
        "id": "评论ID",
        "noteId": "笔记ID",
        "content": "评论内容",
        "likeCount": "点赞数",
        "createTime": 1640000000,
        "ipLocation": "IP位置",
        "liked": false,
        "userInfo": {
          "userId": "用户ID",
          "nickname": "昵称"
        },
        "subCommentCount": "子评论数",
        "subComments": []
      }
    ],
    "cursor": "游标",
    "hasMore": true
  }
}
```

**使用场景**:
- 获取笔记完整内容
- 研究热门评论
- 分析用户互动

**注意事项**:
- 默认只返回前 10 条评论
- 加载全部评论会较慢，建议只在需要时使用
- 需要登录才能使用此功能

---

#### 3.4 user_profile

**功能描述**: 获取指定用户的主页信息

**参数**:
```json
{
  "user_id": "小红书用户ID（从Feed列表获取）",
  "xsec_token": "访问令牌（从Feed列表获取）"
}
```

**返回内容**:
```json
{
  "userBasicInfo": {
    "gender": 1,
    "ipLocation": "IP位置",
    "desc": "用户简介",
    "imageb": "背景图",
    "nickname": "昵称",
    "images": "头像",
    "redId": "用户ID"
  },
  "interactions": [
    {
      "type": "follows",
      "name": "关注",
      "count": "关注数"
    },
    {
      "type": "fans",
      "name": "粉丝",
      "count": "粉丝数"
    },
    {
      "type": "interaction",
      "name": "获赞与收藏",
      "count": "获赞量"
    }
  ],
  "feeds": [
    {
      "id": "笔记ID",
      "noteCard": {
        "displayTitle": "标题",
        "interactInfo": {
          "likedCount": "点赞数",
          "commentCount": "评论数"
        }
      }
    }
  ]
}
```

**使用场景**:
- 了解创作者信息
- 研究竞品账号
- 查找合作对象

---

### 4. 互动功能（6 个）

| 工具名称 | 功能描述 | 只读 | 危险 |
|---------|---------|------|------|
| `post_comment_to_feed` | 发表评论到笔记 | ❌ | ✅ |
| `reply_comment_in_feed` | 回复笔记下的指定评论 | ❌ | ✅ |
| `like_feed` | 点赞/取消点赞笔记 | ❌ | ✅ |
| `favorite_feed` | 收藏/取消收藏笔记 | ❌ | ✅ |
| `follow_user` | 关注/取关用户 | ❌ | ✅ |
| `like_comment` | 点赞/取消点赞评论 | ❌ | ✅ |

#### 4.1 post_comment_to_feed

**功能描述**: 发表评论到小红书笔记

**参数**:
```json
{
  "feed_id": "笔记ID（从Feed列表获取）",
  "xsec_token": "访问令牌（从Feed列表获取）",
  "content": "评论内容"
}
```

**返回示例**:
```json
{
  "feed_id": "笔记ID",
  "success": true,
  "message": "评论发表成功"
}
```

**使用场景**:
- 互动交流
- 建立社区关系
- 营销推广

---

#### 4.2 reply_comment_in_feed

**功能描述**: 回复笔记下的指定评论

**参数**:
```json
{
  "feed_id": "笔记ID（从Feed列表获取）",
  "xsec_token": "访问令牌（从Feed列表获取）",
  "comment_id": "目标评论ID（从评论列表获取）",
  "user_id": "目标评论用户ID（从评论列表获取）",
  "content": "回复内容"
}
```

**注意事项**:
- `comment_id` 和 `user_id` 至少提供一个
- 通常两者都从评论列表中获取

**返回示例**:
```json
{
  "feed_id": "笔记ID",
  "target_comment_id": "评论ID",
  "target_user_id": "用户ID",
  "success": true,
  "message": "评论回复成功"
}
```

---

#### 4.3 like_feed

**功能描述**: 为指定笔记点赞或取消点赞

**参数**:
```json
{
  "feed_id": "笔记ID（从Feed列表获取）",
  "xsec_token": "访问令牌（从Feed列表获取）",
  "unlike": false
}
```

**参数说明**:
- `unlike`: 可选，是否取消点赞
  - `false` 或未设置：点赞
  - `true`：取消点赞

**行为说明**:
- 如已点赞则跳过点赞操作
- 如未点赞则跳过取消点赞操作

**返回示例**:
```json
{
  "feed_id": "笔记ID",
  "success": true,
  "message": "点赞成功或已点赞"
}
```

---

#### 4.4 favorite_feed

**功能描述**: 收藏指定笔记或取消收藏

**参数**:
```json
{
  "feed_id": "笔记ID（从Feed列表获取）",
  "xsec_token": "访问令牌（从Feed列表获取）",
  "unfavorite": false
}
```

**参数说明**:
- `unfavorite`: 可选，是否取消收藏
  - `false` 或未设置：收藏
  - `true`：取消收藏

**行为说明**:
- 如已收藏则跳过收藏操作
- 如未收藏则跳过取消收藏操作

**返回示例**:
```json
{
  "feed_id": "笔记ID",
  "success": true,
  "message": "收藏成功或已收藏"
}
```

---

#### 4.5 follow_user

**功能描述**: 关注指定用户或取关

**参数**:
```json
{
  "user_id": "小红书用户ID（从Feed列表或用户主页获取）",
  "xsec_token": "访问令牌（从Feed列表获取）",
  "unfollow": false
}
```

**参数说明**:
- `unfollow`: 可选，是否取关
  - `false` 或未设置：关注
  - `true`：取关

**行为说明**:
- 如已关注则跳过关注操作
- 如未关注则跳过取关操作

**返回示例**:
```json
{
  "feed_id": "用户ID",
  "success": true,
  "message": "关注成功"
}
```

---

#### 4.6 like_comment

**功能描述**: 点赞指定评论或取消点赞

**参数**:
```json
{
  "feed_id": "笔记ID（从Feed列表获取）",
  "xsec_token": "访问令牌（从Feed列表获取）",
  "comment_id": "目标评论ID（从评论列表获取）",
  "user_id": "目标评论用户ID（从评论列表获取）",
  "unlike": false
}
```

**参数说明**:
- `comment_id` 和 `user_id` 至少提供一个
- `unlike`: 可选，是否取消点赞

**行为说明**:
- 如已点赞则跳过点赞操作
- 如未点赞则跳过取消点赞操作

**返回示例**:
```json
{
  "feed_id": "笔记ID",
  "success": true,
  "message": "评论点赞成功"
}
```

---

### 5. 内容管理（3 个）

| 工具名称 | 功能描述 | 只读 | 危险 |
|---------|---------|------|------|
| `share_feed` | 分享笔记获取链接 | ❌ | ❌ |
| `delete_feed` | 删除自己的笔记 | ❌ | ✅ |
| `delete_comment` | 删除自己的评论 | ❌ | ✅ |

#### 5.1 share_feed

**功能描述**: 分享指定笔记，获取分享链接

**参数**:
```json
{
  "feed_id": "笔记ID（从Feed列表获取）",
  "xsec_token": "访问令牌（从Feed列表获取）"
}
```

**返回示例**:
```json
{
  "success": true,
  "message": "分享成功 - 分享链接: https://www.xiaohongshu.com/explore/xxx"
}
```

**使用场景**:
- 获取笔记分享链接
- 分享到其他平台
- 统计分享数据

---

#### 5.2 delete_feed

**功能描述**: 删除自己的笔记

**参数**:
```json
{
  "feed_id": "笔记ID（从Feed列表获取）",
  "xsec_token": "访问令牌（从Feed列表获取）"
}
```

**返回示例**:
```json
{
  "feed_id": "笔记ID",
  "success": true,
  "message": "笔记删除成功"
}
```

**注意事项**:
- ⚠️ **危险操作**：删除后无法恢复
- 只能删除自己发布的笔记

---

#### 5.3 delete_comment

**功能描述**: 删除自己的评论

**参数**:
```json
{
  "feed_id": "笔记ID（从Feed列表获取）",
  "xsec_token": "访问令牌（从Feed列表获取）",
  "comment_id": "目标评论ID（从评论列表获取）",
  "user_id": "目标评论用户ID（从评论列表获取）"
}
```

**参数说明**:
- `comment_id` 和 `user_id` 至少提供一个

**返回示例**:
```json
{
  "feed_id": "笔记ID",
  "success": true,
  "message": "评论删除成功"
}
```

**注意事项**:
- ⚠️ **危险操作**：删除后无法恢复
- 只能删除自己的评论

---

### 6. 数据获取（2 个）

| 工具名称 | 功能描述 | 只读 | 危险 |
|---------|---------|------|------|
| `get_fan_analytics` | 获取粉丝分析数据 | ✅ | ❌ |
| `get_content_analytics` | 获取内容分析数据 | ✅ | ❌ |

> **注意**: 以下工具已暂时禁用，待修复后重新启用
> - `get_my_stats` - 获取个人统计数据（已禁用）
> - `get_my_feeds` - 获取自己发布的笔记列表（已禁用）

#### 6.1 get_fan_analytics

**功能描述**: 获取粉丝分析数据

**参数**:
```json
{
  "period": "7d"
}
```

**参数说明**:
- `period`: 可选，统计周期
  - `"7d"`: 近 7 天（默认）
  - `"30d"`: 近 30 天

**返回示例**:
```json
{
  "overview": {
    "totalFans": "总粉丝数",
    "newFans": "新增粉丝数",
    "lostFans": "流失粉丝数",
    "growthRate": "增长率"
  },
  "portrait": {
    "gender": {
      "male": "男性比例",
      "female": "女性比例",
      "unknown": "未知比例"
    },
    "interests": [
      {
        "category": "兴趣类别",
        "percentage": "占比"
      }
    ]
  },
  "activeFans": [
    {
      "userId": "用户ID",
      "nickname": "昵称",
      "avatar": "头像",
      "interactionCount": "互动次数"
    }
  ]
}
```

**使用场景**:
- 分析粉丝画像
- 了解粉丝增长趋势
- 发现活跃粉丝

---

#### 6.2 get_content_analytics

**功能描述**: 获取内容分析数据

**参数**:
```json
{
  "limit": 20
}
```

**参数说明**:
- `limit`: 可选，限制返回笔记数量
  - 默认：20
  - 最大：100

**返回示例**:
```json
{
  "notes": [
    {
      "noteId": "笔记ID",
      "title": "标题",
      "publishTime": "发布时间",
      "exposure": "曝光量",
      "views": "观看量",
      "likes": "点赞数",
      "comments": "评论数",
      "favorites": "收藏数",
      "newFans": "涨粉数"
    }
  ]
}
```

**使用场景**:
- 分析内容表现
- 找出爆款笔记
- 优化内容策略

---

## 使用建议

### 操作频率限制

为避免账号被封，建议遵循以下频率：

| 操作类型 | 频率建议 | 间隔时间 |
|---------|---------|---------|
| 点赞 | 每小时 ≤ 60 次 | ≥ 30 秒 |
| 评论 | 每小时 ≤ 20 次 | ≥ 2 分钟 |
| 关注 | 每小时 ≤ 30 次 | ≥ 1 分钟 |
| 发布 | 每天 ≤ 5 条 | ≥ 2 小时 |

**重要提示**:
- 🚫 **不要使用批量操作**
- ✅ 让 AI 控制调用频率和间隔时间
- ✅ 模拟真实用户行为
- ✅ 避免高频重复操作

### 最佳实践

1. **登录管理**
   - 定期检查登录状态
   - Cookies 过期后及时更新
   - 不要多端同时登录（会踢出）

2. **内容发布**
   - 推荐使用本地图片路径
   - 控制发布频率，避免被风控
   - 定时发布功能合理使用
   - 标题简洁有力（20 字以内）
   - 添加合适的标签

3. **互动操作**
   - 点赞、评论、关注分散进行
   - 避免短时间内大量操作
   - 互动要有质量，避免刷屏

4. **数据分析**
   - 定期获取数据进行分析
   - 关注粉丝增长趋势
   - 研究内容表现数据
   - 优化内容策略

### 常见问题

**Q: Cookies 保存在哪里？**

A: 默认保存在 `~/.xiaohongshu/cookies.json`

**Q: 如何重新登录？**

A:
```bash
# 删除 cookies
rm ~/.xiaohongshu/cookies.json

# 或使用 delete_cookies 工具
# 然后使用 get_login_qrcode 重新扫码登录
```

**Q: 为什么显示发布成功但实际没有？**

A: 排查步骤：
1. 使用非无头模式重新发布
2. 更换不同的内容重新发布
3. 检查账号是否被风控
4. 检查图片大小是否过大
5. 确认图片路径中没有中文字符

**Q: 显示用户名是 xiaghgngshu-mcp？**

A: 用户名是写死的，不影响功能使用

**Q: Cookies 多久过期？**

A: 通常几周到几个月不等，建议定期检查登录状态

**Q: 可以同时运行多个账号吗？**

A: 可以，但需要为每个账号运行独立的 MCP 实例，并确保 cookies 不冲突

---

## 技术架构

### 项目结构

```
xiaohongshu-mcp/
├── main.go                 # 服务入口
├── mcp_server.go          # MCP 工具注册
├── mcp_handlers.go        # MCP 工具处理函数
├── routes.go             # HTTP 路由配置
├── service.go            # 业务服务层
├── xiaohongshu/         # 核心业务逻辑
│   ├── login.go          # 登录相关
│   ├── publish.go        # 图文发布
│   ├── publish_video.go  # 视频发布
│   ├── feeds.go          # Feeds 列表
│   ├── search.go         # 搜索功能
│   ├── feed_detail.go    # Feed 详情
│   ├── user_profile.go   # 用户主页
│   ├── comment_feed.go   # 评论相关
│   ├── comment_like.go   # 评论点赞
│   ├── like_favorite.go   # 点赞收藏
│   ├── follow.go         # 关注功能
│   ├── share.go          # 分享功能
│   ├── delete.go         # 删除功能
│   ├── data.go           # 数据分析
│   └── types.go         # 数据结构定义
├── browser/             # 浏览器封装
├── cookies/             # Cookies 管理
├── configs/             # 配置文件
└── pkg/                 # 通用工具
```

### 技术栈

- **语言**: Go 1.24
- **Web 框架**: Gin
- **浏览器自动化**: Rod
- **MCP 协议**: 官方 Go SDK
- **日志**: Logrus

### MCP 协议支持

- 支持 HTTP Streamable MCP 协议
- 标准的 JSON-RPC 2.0 消息格式
- 完整的工具注册和调用机制
- Panic 恢复和错误处理

---

## 支持的客户端

### 官方推荐

- **Claude Code CLI**: 官方命令行工具
- **Cursor**: 代码编辑器集成
- **VSCode**: 代码编辑器集成
- **Cline**: AI 编程助手

### 其他支持 HTTP MCP 的客户端

任何支持 HTTP MCP 协议的客户端都可以连接到 `http://localhost:18060/mcp`

---

## 更新日志

### 最新更新（2026-01-21）

- **修复**: 重写数据提取工具，使用 JavaScript 页面解析替代不可靠的 DOM 选择器
- **优化**: 移除仅移动端可用的 `get_followers_list` 和 `get_following_list` 工具（网页版不支持）
- **新增**: `get_fan_analytics` 粉丝分析数据工具
- **新增**: `get_content_analytics` 内容分析数据工具
- **新增**: 页面结构分析文档 `docs/PAGE_STRUCTURE_ANALYSIS.md`

---

## 许可证

本项目仅供学习和研究使用。禁止任何违法行为。

---

## 联系方式

- GitHub Issues: https://github.com/xpzouying/xiaohongshu-mcp/issues
- 飞书群：[加入群聊]
- 微信群：[加入群聊]

---

**注意**: 使用本工具前请务必阅读完整文档和常见问题解答。
