# AGENTS.md

## 项目概述

try 是一个 Go 语言编写的临时实验目录管理 TUI 工具。使用 Bubbletea v2 + Lipgloss v2 + Bubbles v2 构建交互界面。

## 必须遵循的规则

在修改本项目代码前，**必须**阅读并遵循 `.cursor/rules/` 目录下的所有 `.mdc` 规则文件。

当前规则：

1. `.cursor/rules/code-quality.mdc` — 质量检查流程、体积控制、文档同步、i18n 要求
2. `.cursor/rules/go-coding-style.mdc` — 编码风格、错误处理、函数设计
3. `.cursor/rules/go-patterns.mdc` — 设计模式、依赖注入、并发
4. `.cursor/rules/go-testing.mdc` — 测试规范、表驱动测试
5. `.cursor/rules/go-security.mdc` — 安全规范、超时控制

## Skills

`.cursor/skills/` 目录包含可复用的工作流技能。执行特定类型任务时，先检查是否有匹配的 skill 文件，有则读取并按其指引操作。

## 设计文档

所有模块的详细设计见 `spec/` 目录。代码变更时必须检查对应 spec 是否需要同步更新。

关键 spec：
- `spec/architecture.md` — 项目结构与模块职责
- `spec/config.md` — 配置文件格式与解析逻辑
- `spec/selector.md` — 选择器模型、按键处理、渲染
- `spec/dialogs.md` — 对话框系统（删除/重命名/ship）
- `spec/i18n.md` — 国际化字段定义与覆盖范围

## 工作流程

每次代码变更后必须按顺序执行：

```bash
go build ./...
go vet ./...
staticcheck ./...
go test ./... -count=1
```

## 关键约束

- 所有用户可见文本通过 `i18n.Messages` 获取，禁止硬编码字符串
- 单文件 ≤ 300 行，单函数 ≤ 40 行，参数 ≤ 5 个
- 新增公开函数必须有测试
- 配置变更同步更新 `spec/config.md` 和 `README.md`
- 快捷键变更同步更新 `spec/selector.md` 和 `README.md`
