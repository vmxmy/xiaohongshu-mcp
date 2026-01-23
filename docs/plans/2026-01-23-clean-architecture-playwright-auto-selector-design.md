# Clean Architecture + Playwright + 自动选择器自愈 设计方案

## 背景与目标
- 目标：将全仓库重构为 Clean Architecture 分层，引入 Playwright（无 GUI 环境优先），并实现 Agent 辅助的选择器自动更新与全自动落盘。
- 重点诉求：长期可维护性、页面更新的自愈能力、Docker 无 GUI 环境稳定运行。

## 范围与非目标
- 范围：`xiaohongshu/` 全量迁移，浏览器自动化替换为 Playwright，配置与选择器全配置化，新增选择器自愈流水线。
- 非目标：不做多平台适配（仅小红书），不做多租户治理，不引入复杂编排系统。

## 分层架构
- `internal/domain`：纯业务模型与规则（发布内容、标签、素材、定时策略、限制）。
- `internal/app`：用例编排（PublishImage/PublishVideo/Search/Feeds），只依赖 domain 与接口。
- `internal/infra`：Playwright 浏览器、配置加载、选择器存储、日志、网络拦截、重试策略。
- `internal/interfaces`：CLI/MCP/HTTP 适配层，负责入参解析与错误输出。
- `internal/core`：跨域能力（clock、retry、selector matcher、metrics、config cache）。

## 关键端口与实现
- `BrowserEngine`：启动/关闭、创建上下文/页面、等待策略、网络拦截、上传/下载。
- `SelectorStore`：读写选择器配置、版本快照、热加载。
- `ConfigStore`：读取配置与默认值、校验。
- `ProbeRunner`：页面探测、DOM/AX 树快照、截图采集。
- `SelectorLearner`：Agent 生成候选选择器与变更说明。
- `SelectorValidator`：最小关键路径验证（发布/搜索/feeds 冒烟）。

## 数据流
1. Interface 解析请求 -> 构造 `AppContext`（Config/Logger/Timeout/Retry）。
2. App 调用 Domain 校验与限额逻辑。
3. App 通过端口调用 Infra（Playwright、选择器、网络拦截）。
4. 结果向上返回，由 Interface 统一格式化输出。

## 选择器自动更新（Agent + 全自动落盘）
- 触发方式：运行失败或定时任务触发 `ProbeRunner`。
- 探测内容：DOM 快照、可访问性树、关键节点文本/属性指纹、页面截图。
- 生成策略：Agent 产出候选选择器 + 风险等级 + 变更理由。
- 验证策略：使用最小关键路径用例，所有用例通过才进入落盘。
- 落盘策略：写入新配置前生成版本快照（diff + hash + 时间戳），失败自动回滚。

## 配置与热加载
- 配置源：`config.yaml`（可扩展为多环境配置）。
- 热加载：双缓冲策略，加载失败自动回退上一个版本。
- 配置校验：启动时完整校验，运行期按需校验关键字段。

## Docker 无 GUI 支持
- 镜像内预装 Playwright 浏览器与依赖（字体、渲染库）。
- 启动时做依赖自检并输出缺失提示。
- 默认 headless，保留可切换开关以便本地调试。

## 错误处理与回滚
- 选择器失败：触发自愈流程；自愈失败进入降级模式（停写配置，保留旧版本）。
- 网络拦截异常：退回 UI 检测策略。
- 发布失败：保留原始失败上下文（页面快照/请求摘要）供诊断。

## 迁移策略
- 阶段 1：抽离端口接口 + 引入 Playwright 适配层。
- 阶段 2：发布流程迁移到 app/domain，选择器/超时从配置获取。
- 阶段 3：搜索/feeds 迁移并完成自动自愈闭环。

## 风险与缓解
- Playwright 依赖体积大：通过 Docker 镜像缓存与分层减少影响。
- 选择器误更新：全自动落盘前必须通过完整冒烟验证。
- 过度抽象：只抽当前用例所需接口，避免 YAGNI。

## 原则评估（KISS / YAGNI / SOLID / DRY）
- KISS：分层带来复杂度，使用最小接口与最小用例缓解。
- YAGNI：仅对当前发布/搜索/feeds 提供端口与自动自愈能力。
- SOLID：依赖倒置与接口隔离为核心，infra 替换不影响 domain。
- DRY：选择器/超时/日志/重试统一下沉至 core 与 infra。

## 验证与里程碑
- 冒烟：Docker headless 启动 + 发布图文 + 搜索成功。
- 回归：选择器自愈流程至少完成一次自动落盘。
- 可观测：失败时保留完整快照与变更记录。
