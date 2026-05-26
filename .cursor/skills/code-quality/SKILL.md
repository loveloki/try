---
name: code-quality
description: >-
  编码质量守则。适用于所有编码场景——新功能实现、bug 修复、重构、
  测试编写等。确保每次代码变更都通过质量检查、测试完备、代码简洁可维护。
---

# 编码质量守则

本规范适用于项目中所有编码活动。每次实现完成后必须执行质量检查流程。

## 质量检查流程（每次实现完必须执行）

每完成一个功能模块或修改后，按以下顺序逐项检查：

```
1. go build ./...          # 编译通过，零错误
2. go vet ./...            # 官方静态检查通过
3. staticcheck ./...       # 第三方静态检查通过
4. go test ./... -count=1  # 全量测试通过
5. ReadLints               # 无新增 linter 错误
```

**不允许跳过任何一步。** 如果某步失败，立即修复后重新检查。

## 文件体积控制

| 指标 | 上限 | 超限处理 |
|------|------|----------|
| 单文件行数 | 300 行 | 按职责拆分为多个文件 |
| 单测试文件行数 | 500 行 | 按测试主题拆分 |

拆分原则：
- 按职责边界拆分，而非机械地按行数切割
- 每个文件有一个清晰的主题（如 `model.go` 管状态，`view.go` 管渲染）
- 避免为拆分而拆分——如果逻辑紧密耦合，200+ 行的文件也可以接受

## 函数大小控制

| 指标 | 上限 | 超限处理 |
|------|------|----------|
| 单函数行数 | 40 行 | 提取子函数 |
| 函数参数个数 | 5 个 | 使用结构体参数 |
| 函数嵌套层级 | 3 层 | 提前 return 或提取函数 |

长函数的拆分策略：
- 识别独立的逻辑步骤，每步提取为函数
- 提取后的函数命名应自解释，减少注释需求
- switch/case 中每个 case 超过 5 行时考虑提取

## 纯函数优先

**优先编写纯函数**——给定相同输入总是返回相同输出，无副作用。

### 为什么

- 纯函数天然易于测试：输入 → 输出，不需要 mock
- 纯函数可以并行执行、缓存结果
- 纯函数容易理解，调用者无需担心隐藏的状态变更

### 如何实践

1. **分离计算与副作用**：先用纯函数计算结果，再由调用者执行 IO

```go
// ✅ 好：纯函数计算结果
func buildDestPath(shipPath, basename string) string {
    projectName := dateSuffixRe.ReplaceAllString(basename, "")
    return filepath.Join(shipPath, projectName)
}

// 调用者负责 IO
dest := buildDestPath(shipPath, entry.Basename)
os.Rename(src, dest)
```

```go
// ❌ 差：计算和 IO 混在一起
func shipEntry(entry Entry, shipPath string) error {
    projectName := dateSuffixRe.ReplaceAllString(entry.Basename, "")
    dest := filepath.Join(shipPath, projectName)
    return os.Rename(entry.Path, dest)
}
```

2. **依赖注入替代全局状态**：通过参数传入依赖，而非读取全局变量或环境变量

3. **IO 边界推到最外层**：CLI 层/main 负责读文件、读环境变量，内部模块接收已解析的值

### 项目中的参考

- `parseConfigData(data []byte) Config` — 纯函数，不读文件
- `LoadConfig()` — IO 层，读文件后调用纯函数
- `fuzzy.Match(entries, query, maxResults)` — 纯函数，不接触文件系统
- `ForLocale(locale string) *Messages` — 纯函数

## 测试要求

### 覆盖率目标

| 包类型 | 最低覆盖率 | 说明 |
|--------|-----------|------|
| 核心算法/逻辑 | 90% | fuzzy、config、git、i18n 等 |
| TUI 交互层 | 50% | selector、dialog（受终端环境限制） |
| IO/集成层 | 30% | cli、shell、script |

### 测试风格

- **表驱动测试**：使用 `[]struct{ name; input; want }` + `t.Run`
- **check 辅助函数**：封装断言逻辑，API 变更时只改一处
- **测试功能而非代码**：测试用户可观察的行为，而非内部实现细节
- **真实路径优先**：尽量走真实代码路径，仅在必要时使用 mock
- **环境隔离**：使用 `t.Setenv` 设置环境变量，使用 `t.TempDir` 创建临时目录

### 新代码必须有测试

- 新增的公开函数必须有对应测试
- 新增的配置项必须有解析测试和优先级测试
- bug 修复必须附带复现该 bug 的测试用例

## 注释规范

- 关键注释使用中文
- 不写显而易见的注释（如 `// 返回结果`）
- 注释解释"为什么"而非"做了什么"
- 公开函数必须有 GoDoc 注释

## 文档维护

代码变更后必须同步更新相关文档。文档是代码的一部分，不是事后补充。

### README.md 同步检查

每次涉及以下变更时，检查 README 是否需要更新：

| 变更类型 | README 需更新的部分 |
|----------|-------------------|
| 新增命令/子命令 | "使用" 部分的命令示例 |
| 新增快捷键 | "快捷键" 表格 |
| 新增配置项 | "配置" 部分的 JSON 示例和配置表格 |
| 新增包/模块 | "项目结构" 目录树 |
| 新增依赖 | "技术栈" 列表 |
| 新增功能特性 | "功能" 列表 |
| 开发工具链变更 | "开发" 部分的命令 |

### docs/ 设计文档同步

- 新增功能前先更新设计文档（参见 `new-feature` skill）
- 修改现有行为时同步更新对应文档
- 配置项变更时更新 `docs/config.md` 中的优先级链和示例
- 确保文档中的代码示例与实际代码一致

### 检查方法

实现完成后，快速浏览以下文件确认一致性：
1. `README.md` — 功能列表、命令示例、配置表格、项目结构
2. `docs/config.md` — 配置项列表和优先级
3. 涉及变更的其他 `docs/*.md` 文件

## 提交规范

- 每完成一个独立功能都要提交 commit
- commit message 格式：`feat/fix/refactor/test/chore/docs: 简要描述`
- 依赖变更（go.mod/go.sum）、测试、文档一并纳入对应 commit
- 提交前工作区无未提交变更
