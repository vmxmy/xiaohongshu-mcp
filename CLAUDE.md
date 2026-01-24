<!-- OPENSPEC:START -->
# OpenSpec Instructions

These instructions are for AI assistants working in this project.

Always open `@/openspec/AGENTS.md` when the request:
- Mentions planning or proposals (words like proposal, spec, change, plan)
- Introduces new capabilities, breaking changes, architecture shifts, or big performance/security work
- Sounds ambiguous and you need the authoritative spec before coding

Use `@/openspec/AGENTS.md` to learn:
- How to create and apply change proposals
- Spec format and conventions
- Project structure and guidelines

Keep this managed block so 'openspec update' can refresh the instructions.

<!-- OPENSPEC:END -->

- 要求每次修改完后,需要帮我格式化 Go 源码文件.
- 测试过程中产生的脚本和build中间文件,如果没有必要,则删除.
- 所有的feature变更,都需要使用分支进行开发.
- 在我未同意之前, 你不能推送到远程.
- 我需要: 1.本地 review; 2.远程 PR review.
- 不要过度设计, 保持代码的简洁和易读.
- 使用中文注释，一定要简洁明了.专业名词可以用英文.