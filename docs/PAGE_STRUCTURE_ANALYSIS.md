# 小红书页面结构分析文档

> 本文档用于记录小红书各页面的DOM结构和数据提取方法，为工具开发提供参考。

## 1. 登录页面

### 页面URL
- `https://www.xiaohongshu.com/explore` (会弹出登录弹窗)
- `https://www.xiaohongshu.com/login` (直接登录页)

### 二维码获取

#### 登录二维码
- **选择器**: `img.qrcode-img`
- **尺寸**: 128x128
- **数据格式**: Base64 PNG (`data:image/png;base64,...`)
- **提取代码**:
```javascript
() => {
  const img = document.querySelector('img.qrcode-img');
  if (img) {
    return {
      src: img.src,  // 完整的base64数据
      width: img.width,
      height: img.height
    };
  }
  return null;
}
```

#### 安全验证二维码
- **场景**: 首次登录新设备时需要扫码验证
- **选择器**: 第二个 `img.qrcode-img` 或按尺寸区分 (175x175)
- **尺寸**: 175x175
- **数据格式**: Base64 PNG
- **提取代码**:
```javascript
() => {
  const imgs = document.querySelectorAll('img.qrcode-img');
  // 安全验证二维码通常是第二个，且尺寸为175x175
  const securityQR = Array.from(imgs).find(img => img.width === 175);
  if (securityQR) {
    return {
      src: securityQR.src,
      width: securityQR.width,
      height: securityQR.height
    };
  }
  return null;
}
```

#### 创作者中心登录二维码
- **页面**: `https://creator.xiaohongshu.com/login`
- **触发方式**: 点击右上角的二维码图标切换到扫码登录
- **尺寸**: 160x160
- **数据格式**: Base64 PNG
- **提取代码**:
```javascript
() => {
  const imgs = document.querySelectorAll('img');
  // 创作者中心二维码尺寸为160x160
  const creatorQR = Array.from(imgs).find(img => img.width === 160 && img.height === 160);
  if (creatorQR && creatorQR.src.startsWith('data:image')) {
    return {
      src: creatorQR.src,
      width: creatorQR.width,
      height: creatorQR.height
    };
  }
  return null;
}
```

#### 创作者中心安全验证二维码
- **场景**: 登录创作者中心时需要安全验证
- **触发条件**: 登录新设备或长时间未登录
- **选择器**: 弹窗中尺寸为175x175的img
- **尺寸**: 175x175
- **数据格式**: Base64 PNG (约7000+字符)
- **提取代码**:
```javascript
() => {
  const imgs = document.querySelectorAll('img');
  // 查找尺寸为175x175的base64图片
  for (const img of imgs) {
    if (img.width === 175 && img.height === 175 && img.src.startsWith('data:image')) {
      return {
        src: img.src,
        width: img.width,
        height: img.height
      };
    }
  }
  return null;
}
```

### 二维码汇总

| 类型 | 位置 | 尺寸 | 选择器 |
|------|------|------|--------|
| 登录二维码 | 小红书主站 | 128x128 | `img.qrcode-img` |
| 安全验证(主站) | 小红书主站弹窗 | 175x175 | 第二个 `img.qrcode-img` |
| 登录二维码 | 创作者中心 | 160x160 | 尺寸过滤 |
| 安全验证(创作者) | 创作者中心弹窗 | 175x175 | 尺寸过滤 |

### 二维码刷新
- 二维码有效期: 1分钟
- 过期后显示: "已过期，点击二维码区域刷新"
- 刷新方法: 点击二维码区域

---

## 2. 创作者中心

### 页面URL
- 主页: `https://creator.xiaohongshu.com/new/home?source=official`
- 粉丝数据: `https://creator.xiaohongshu.com/statistics/fans-data?source=official`
- 内容分析: `https://creator.xiaohongshu.com/statistics/data-analysis?source=official`
- 账号概览: 侧边栏 "账号概览" 链接

### 侧边栏导航结构
```
- 首页
- 笔记管理
- 数据看板
  - 账号概览
  - 内容分析
  - 粉丝数据
- 活动中心
- 笔记灵感
- 创作学院
- 创作百科
```

### 基础用户信息区域
**DOM结构**:
```
用户头像 -> 用户名 -> 账号状态图标
├── {关注数}\n关注数
├── {粉丝数}\n粉丝数
└── {获赞与收藏}\n获赞与收藏
小红书账号: {账号ID}
{简介}
```

**提取方法 (JavaScript)**:
```javascript
() => {
  const text = document.body.innerText;
  const stats = {};

  // 匹配基础数据格式: 数字\n标签
  const followMatch = text.match(/(\d+)\n关注数/);
  const followerMatch = text.match(/(\d+)\n粉丝数/);
  const likedMatch = text.match(/(\d+)\n获赞与收藏/);
  const accountMatch = text.match(/小红书账号:\s*(\d+)/);

  if (followMatch) stats.follow_count = parseInt(followMatch[1]);
  if (followerMatch) stats.follower_count = parseInt(followerMatch[1]);
  if (likedMatch) stats.liked_count = parseInt(likedMatch[1]);
  if (accountMatch) stats.account_id = accountMatch[1];

  return stats;
}
```

### 笔记数据总览 (近7日/近30日)
**DOM结构**:
```
笔记数据总览 | 直播数据总览
统计周期 MM-DD 至 MM-DD
[近7日] [近30日]

曝光数        观看数        封面点击率      视频完播率
{数值}{单位}   {数值}{单位}   {数值}%        {数值}%
环比-         环比-         环比-          环比-

点赞数        评论数        收藏数         分享数
{数值}        {数值}        {数值}         {数值}
环比-         环比-         环比-          环比-

净涨粉        新增关注       取消关注       主页访客
{数值}        {数值}        {数值}         {数值}
环比-         环比-         环比-          环比-
```

**数据格式**:
- 大数值带单位: `10.5万`, `1.1万`
- 百分比: `10.8%`, `0%`
- 普通数值: `375`, `21`

**提取方法 (JavaScript)**:
```javascript
() => {
  const text = document.body.innerText;
  const parseNumber = (text) => {
    if (!text) return 0;
    text = text.replace(/,/g, '');
    if (text.includes('万')) {
      return Math.round(parseFloat(text.replace('万', '')) * 10000);
    } else if (text.includes('亿')) {
      return Math.round(parseFloat(text.replace('亿', '')) * 100000000);
    }
    return parseInt(text) || 0;
  };

  const stats = {};

  // 曝光数、观看数等
  const exposureMatch = text.match(/曝光数\n([\d.]+[万亿]?)/);
  const viewMatch = text.match(/观看数\n([\d.]+[万亿]?)/);
  const clickRateMatch = text.match(/封面点击率\n([\d.]+)%/);
  const completeRateMatch = text.match(/视频完播率\n([\d.]+)%/);
  const likeMatch = text.match(/点赞数\n(\d+)/);
  const commentMatch = text.match(/评论数\n(\d+)/);
  const collectMatch = text.match(/收藏数\n(\d+)/);
  const shareMatch = text.match(/分享数\n(\d+)/);
  const netGrowthMatch = text.match(/净涨粉\n(\d+)/);
  const newFollowerMatch = text.match(/新增关注\n(\d+)/);
  const unfollowMatch = text.match(/取消关注\n(\d+)/);
  const visitorMatch = text.match(/主页访客\n(\d+)/);

  if (exposureMatch) stats.exposure_count = parseNumber(exposureMatch[1]);
  if (viewMatch) stats.view_count = parseNumber(viewMatch[1]);
  if (clickRateMatch) stats.cover_click_rate = parseFloat(clickRateMatch[1]);
  if (completeRateMatch) stats.video_complete_rate = parseFloat(completeRateMatch[1]);
  if (likeMatch) stats.like_count_7d = parseInt(likeMatch[1]);
  if (commentMatch) stats.comment_count_7d = parseInt(commentMatch[1]);
  if (collectMatch) stats.collect_count_7d = parseInt(collectMatch[1]);
  if (shareMatch) stats.share_count_7d = parseInt(shareMatch[1]);
  if (netGrowthMatch) stats.net_follower_growth = parseInt(netGrowthMatch[1]);
  if (newFollowerMatch) stats.new_follower_count = parseInt(newFollowerMatch[1]);
  if (unfollowMatch) stats.unfollow_count = parseInt(unfollowMatch[1]);
  if (visitorMatch) stats.profile_visitor_count = parseInt(visitorMatch[1]);

  return stats;
}
```

### 最新笔记区域
**DOM结构**:
```
最新笔记 [查看详情]
[封面图] {笔记标题}
观看 | 点赞 | 收藏 | 评论
```

### 数据解析注意事项
1. **换行符分隔**: 创作者中心使用`\n`分隔标签和数值
2. **单位处理**: 万=10000, 亿=100000000
3. **环比数据**: 目前显示为`环比-`，可能需要展开查看
4. **时间周期**: 可切换近7日/近30日

---

## 3. 用户主页

### 页面URL
- 自己主页: `https://www.xiaohongshu.com/user/profile/me` (会重定向到实际user_id)
- 他人主页: `https://www.xiaohongshu.com/user/profile/{user_id}`

### 页面结构

```
┌─────────────────────────────────────────┐
│  用户头像  用户名                        │
│           小红书号: 414757977            │
│           IP属地: 广东                   │
│           个人简介...                    │
├─────────────────────────────────────────┤
│  [35 关注]  [96 粉丝]  [796 获赞与收藏]  │
├─────────────────────────────────────────┤
│  [笔记]  [收藏]  [点赞]                  │
├─────────────────────────────────────────┤
│  ┌─────┐  ┌─────┐  ┌─────┐  ┌─────┐    │
│  │封面  │  │封面  │  │封面  │  │封面  │    │
│  │图片  │  │图片  │  │图片  │  │图片  │    │
│  ├─────┤  ├─────┤  ├─────┤  ├─────┤    │
│  │标题  │  │标题  │  │标题  │  │标题  │    │
│  │作者 ♥│  │作者 ♥│  │作者 ♥│  │作者 ♥│    │
│  └─────┘  └─────┘  └─────┘  └─────┘    │
│           (瀑布流笔记列表)               │
└─────────────────────────────────────────┘
```

### 用户信息提取

**DOM结构**:
```html
<!-- 用户统计数据区域 -->
<div>
  <div>
    <div>35</div>
    <div>关注</div>
  </div>
  <div>
    <div>96</div>
    <div>粉丝</div>
  </div>
  <div>
    <div>796</div>
    <div>获赞与收藏</div>
  </div>
</div>
```

**提取代码**:
```javascript
() => {
  const text = document.body.innerText;
  const stats = {};

  // 匹配格式: 数字\n关注/粉丝/获赞与收藏
  const followMatch = text.match(/(\d+)\n关注/);
  const followerMatch = text.match(/(\d+)\n粉丝/);
  const likedMatch = text.match(/(\d+)\n获赞与收藏/);
  const accountMatch = text.match(/小红书号:\s*(\d+)/);
  const ipMatch = text.match(/IP属地:\s*(\S+)/);

  if (followMatch) stats.follow_count = parseInt(followMatch[1]);
  if (followerMatch) stats.follower_count = parseInt(followerMatch[1]);
  if (likedMatch) stats.liked_count = parseInt(likedMatch[1]);
  if (accountMatch) stats.account_id = accountMatch[1];
  if (ipMatch) stats.ip_location = ipMatch[1];

  return stats;
}
```

### 笔记列表结构

**笔记卡片URL格式**:
```
/user/profile/{user_id}/{note_id}?xsec_token={token}&xsec_source=pc_user
```

**DOM结构**:
```html
<a href="/user/profile/{user_id}/{note_id}?xsec_token=...">
  <img src="封面图URL" />
  <div>笔记标题</div>
  <div>
    <a href="/user/profile/{author_id}">作者名</a>
    <span>♥ 点赞数</span>
  </div>
</a>
```

**提取代码**:
```javascript
() => {
  const notes = [];
  // 查找所有笔记链接
  document.querySelectorAll('a[href*="/user/profile/"]').forEach(link => {
    const href = link.getAttribute('href');
    // 排除用户主页链接，只要笔记链接
    if (href && href.includes('xsec_token')) {
      const match = href.match(/\/user\/profile\/(\w+)\/(\w+)\?/);
      if (match) {
        const [_, userId, noteId] = match;
        const title = link.querySelector('div')?.textContent || '';
        const img = link.querySelector('img')?.src || '';
        notes.push({
          note_id: noteId,
          user_id: userId,
          title: title.trim(),
          cover: img,
          url: href
        });
      }
    }
  });
  return notes;
}
```

---

## 4. 粉丝/关注列表

### ⚠️ 重要发现: 网页版不支持粉丝/关注列表

**测试结果**:
- URL `/user/profile/me/followers` → **404 错误**
- URL `/user/profile/{user_id}/followers` → **404 错误**
- URL `/user/profile/me/follows` → **404 错误**
- 点击粉丝数字不会弹出模态框

**结论**: 小红书网页版不提供粉丝/关注列表页面，这是**仅限移动端**的功能。

### 替代方案

使用**创作者中心**获取粉丝相关数据:
- 粉丝数据页面: `https://creator.xiaohongshu.com/statistics/fans-data?source=official`
- 提供: 粉丝总数、新增粉丝、流失粉丝、性别分布、兴趣分布、活跃粉丝列表

### 已废弃的选择器
以下选择器在网页版无法使用（页面不存在）:
```go
selectors := []string{
    ".user-card",
    "[class*='user-item']",
    "[class*='UserCard']",
    "[class*='follower']",
    "li[class*='user']",
}
```

---

## 5. 笔记详情页

### 页面URL
- 探索页入口: `https://www.xiaohongshu.com/explore/{note_id}`
- 用户主页入口: `https://www.xiaohongshu.com/user/profile/{user_id}/{note_id}?xsec_token={token}`
- 发现页入口: `https://www.xiaohongshu.com/discovery/item/{note_id}`

### 页面结构

```
┌─────────────────────────────────────────┐
│  ┌─────────────────┐  ┌───────────────┐ │
│  │                 │  │ 用户头像 用户名│ │
│  │    笔记图片      │  │ 发布时间      │ │
│  │    (轮播)       │  │               │ │
│  │                 │  │ 笔记正文内容  │ │
│  │                 │  │ #话题标签     │ │
│  │                 │  │               │ │
│  │                 │  │ ♥点赞 ⭐收藏   │ │
│  │                 │  │ 💬评论 ↗分享  │ │
│  └─────────────────┘  └───────────────┘ │
│                                         │
│  ─────────── 评论区 ───────────         │
│  [共 X 条评论]                          │
│  ┌─────────────────────────────────┐    │
│  │ 头像 昵称         时间          │    │
│  │ 评论内容                        │    │
│  │ ♥点赞 回复                      │    │
│  │   └─ 子评论...                  │    │
│  └─────────────────────────────────┘    │
└─────────────────────────────────────────┘
```

### 数据提取

**笔记内容提取**:
```javascript
() => {
  const data = {};
  const text = document.body.innerText;

  // 提取互动数据
  const likeMatch = text.match(/(\d+)\s*点赞/);
  const collectMatch = text.match(/(\d+)\s*收藏/);
  const commentMatch = text.match(/共\s*(\d+)\s*条评论/);

  if (likeMatch) data.likes = parseInt(likeMatch[1]);
  if (collectMatch) data.collects = parseInt(collectMatch[1]);
  if (commentMatch) data.comments = parseInt(commentMatch[1]);

  return data;
}
```

### 评论列表结构
> 评论采用懒加载，需要滚动页面触发加载更多

---

## 附录

### A. 通用工具函数

#### 等待页面加载
```go
page.MustWaitLoad()
page.MustWaitDOMStable()
time.Sleep(3 * time.Second) // 等待动态内容渲染
```

#### 滚动加载更多
```javascript
window.scrollBy(0, window.innerHeight);
```

### B. 常用库
- **go-rod**: 浏览器自动化
- **goquery**: HTML解析 (jQuery风格)
- **regexp**: 正则匹配

### C. 注意事项
1. 小红书页面大量使用动态渲染，需要等待足够时间
2. 二维码使用Base64编码，无头浏览器需要提取src属性
3. 数字可能带有"万"、"亿"单位，需要转换
4. Cookie有效期有限，需要定期刷新
