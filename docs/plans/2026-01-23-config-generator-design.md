# config.yaml 生成器（全量抓取 + 校验覆盖）设计方案

## 背景与目标
- 目标：在发布者本地通过 CLI + Playwright 自动登录并抓取页面结构，生成/覆盖 `config.yaml`，减少手工维护成本。
- 关键约束：全量覆盖所有工具页面；自动登录并保存 cookies；生成前备份，校验失败自动回滚。

## 范围与非目标
- 范围：新增 CLI 子命令 `cmd/config-gen`，完成“抓取→推断→校验→覆盖写入→报告”闭环。
- 非目标：不做多平台适配；不做多租户治理；不引入远程云端生成。

## 架构与流程
- 模块：`Probe`（采集）、`Infer`（候选生成）、`Validate`（关键路径校验）、`Persist`（备份与写入）。
- 流程：读取/备份配置 → 自动登录 → 全量页面导航 → DOM/AX/文本采集 → 选择器推断与评分 → 关键路径校验 → 覆盖写入 → 生成报告。

## 选择器生成策略
- 优先级：语义属性（aria-label、placeholder、role、可访问性树名） > 稳定结构路径 > class 前缀/局部匹配。
- 评分维度：稳定性（语义强度）、唯一性（定位唯一）、可操作性（可点击/可输入）、历史稳定性（与旧配置对比）。
- 生成结果统一落盘到 `config.yaml:selectors`。

## URL 与其他配置字段
- URLs：以已知页面清单为基准，失败页面记录并标记。
- 超时/间隔/限制：使用默认模板值，若校验阶段触发超时，会记录建议值。

## 校验与回滚
- 校验路径：发布图文/视频、搜索、feeds、互动（点赞/评论/关注等）。
- 校验失败：终止写入并回滚至备份；保留失败快照。
- 备份策略：`config.yaml.bak.<timestamp>`。

## 自动登录与 cookies
- 登录方式：二维码 + 人工确认。
- 结果：登录成功自动保存 cookies（默认 `~/.xiaohongshu/cookies.json`）。

## CLI 设计
- `go run cmd/config-gen/main.go --headless=false --output=config.yaml --backup=true`
- `--dry-run`：仅生成报告不写入。
- `--verify-only`：仅校验现有配置。

## 运行与依赖
- Playwright + Chromium（本地/容器均支持无 GUI）。
- 运行期依赖自动检测；缺失时提示并退出。

## 风险与缓解
- 风险：页面结构大改导致误生成。
- 缓解：关键路径校验 + 回滚 + 报告输出。

## 原则评估（KISS / YAGNI / SOLID / DRY）
- KISS：流程清晰但模块较多，限制在最小闭环。
- YAGNI：只生成已验证通过的必要字段。
- SOLID：职责拆分清晰，依赖注入便于测试。
- DRY：单一 `config.yaml` 作为配置事实源。
