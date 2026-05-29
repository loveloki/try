# 命令行解析与分派

## 概述

手工解析 `os.Args`，不使用 CLI 框架（如 cobra、urfave/cli）。保持轻量和对参数顺序的完全控制。

## 解析顺序

```
1. parseGlobalFlags：
   a. 提取颜色标志：--no-colors / --no-expand-tokens + NO_COLOR 环境变量
   b. 提取 --path VALUE
   c. 提取 --theme VALUE 和 --locale VALUE
   d. 提取测试参数：--and-exit、--and-type、--and-keys、--and-confirm
   e. LoadConfig 加载配置文件
   f. ResolvePaths / ResolveTheme / ResolveLocale 合并优先级
2. 检查 --help / -h → 输出帮助并退出
3. 检查 --version / -v → 输出版本并退出
4. shift 出第一个非选项参数作为 command
5. 根据 command 分派到对应处理函数
```

## 全局选项

| 选项 | 行为 |
|------|------|
| `--help` / `-h` | 输出帮助到 stderr 并退出 |
| `--version` / `-v` | 输出版本到 stderr 并退出 |
| `--no-colors` | 禁用 ANSI 颜色（Lipgloss 不渲染样式） |
| `--no-expand-tokens` | 与 `--no-colors` 等效（别名，保持向后兼容） |
| `--path PATH` | 覆盖 tries 根目录（支持 `--path=VALUE` 和 `--path VALUE` 两种形式） |
| `--theme dark\|light` | 配色主题（覆盖配置文件和环境变量） |
| `--locale en\|zh` | 界面语言（覆盖配置文件和环境变量） |

环境变量：
- `NO_COLOR`（非空时）等效于 `--no-colors`，遵循 [no-color.org](https://no-color.org/) 标准
- `TRY_THEME`：设置配色主题（`dark` / `light`），优先级低于 `--theme`

## --path 提取逻辑

从 args 切片中查找最后一个 `--path` 或 `--path=VALUE` 参数：

- 使用反向搜索（最后一个生效）
- 支持 `--path VALUE`（两个参数）和 `--path=VALUE`（单参数）两种形式
- 提取后从 args 中删除，不影响后续解析

## tries 路径解析优先级

```
1. --path 参数（显式指定）
2. TRY_PATH 环境变量
3. ~/.config/try/config.json 中的 path
4. 默认值 ~/src/tries
```

最终结果展开为绝对路径。详见 `config.md`。

## 命令分派

| command | 处理方式 | 说明 |
|---------|---------|------|
| 无参数 | `runSelector` | 无命令时直接进入交互式选择器 |
| `install` | `shell.Install` | 写入 Shell 配置文件 |
| `clone` | `cmdClone` | Git 仓库克隆 |
| `worktree` | `cmdWorktree` | 创建 worktree |
| `exec` | `cmdExec` 二级分派 | 包装函数内部调用入口 |
| 其他 | `runSelector` | 默认视为查询词（多个 arg 用连字符合并），进入选择器 |

### 查询词合并

当用户输入多个非选项参数时（如 `try foo bar`），将它们用连字符合并为单个搜索词（`"foo-bar"`）。这与目录命名规则一致（空格转连字符），使得 `try foo bar` 等效于搜索 `foo-bar`。

## exec 子命令

`exec` 是 Shell 包装函数的内部调用入口，支持二级分派：

```
try exec clone <url> [name]     → cmdClone
try exec worktree <dir> [name]  → cmdWorktree
try exec cd [query]             → runSelector
try exec [query]                → cmdExecDefault
```

### cmdExecDefault

`cmdExecDefault` 按顺序处理三种情况：

1. 参数看起来像 Git URL（`IsGitURI`）→ doClone（clone 工作流）
2. 参数以 `.` 开头 → handleDot（worktree / mkdir 快捷方式）
3. 其他 → runSelector（启动交互式选择器）

### "." 命令处理

```
try . name      → worktree（Git 仓库内）或 mkdir
try ./path name → 指定仓库路径
try .           → 报错退出（防误触，必须提供名称）
```

### Git URL 识别

满足任一条件即识别为 Git URL，自动进入 clone 工作流：

- 以 `https://` 或 `http://` 或 `git@` 开头
- 包含 `github.com`
- 包含 `gitlab.com`
- 以 `.git` 结尾

## 测试专用参数

这些参数不在 `--help` 中展示，仅供自动化测试使用。提取顺序在全局选项之后、命令分派之前。

| 参数 | 作用 |
|------|------|
| `--and-exit` | 渲染一次后立即退出（不进入交互循环），退出码始终为 1 |
| `--and-keys SPEC` | 注入按键序列，序列耗尽后自动发送 ESC |
| `--and-type TEXT` | 注入初始搜索文本（设置 textInput 初始值，优先于位置参数 search_term） |
| `--and-confirm TEXT` | 注入删除确认文本（跳过对话框中的逐字符输入） |

### --and-keys 解析

`parseTestKeys` 支持两种模式：

**Token 模式**（含逗号，或整个字符串仅由字母、数字、连字符和 `=` 组成）：

```
UP,DOWN,ENTER           → 方向键 + 回车
CTRL-D,TYPE=hello,ENTER → Ctrl-D + 逐字输入 "hello" + 回车
```

支持的 token：`UP`、`DOWN`、`LEFT`、`RIGHT`、`ENTER`、`ESC`、`BACKSPACE`、`CTRL-A` 到 `CTRL-W`、`TYPE=...`

`TYPE=text` 在解析阶段展开为逐字符 token，每个字符经 `keyToMsg` 转为 `tea.KeyPressMsg`。

**Raw 模式**（其他情况）：

逐字符解析，自动识别 `\e[X` 转义序列（`\e[A`=UP、`\e[B`=DOWN、`\e[C`=RIGHT、`\e[D`=LEFT）、`\r`/`\n`=ENTER、控制字符=CTRL-X。

## 启动选择器

CLI 解析完成后，构造 `selector.Config` 并启动 Bubbletea Program。关键字段映射：

| Config 字段 | 来源 |
|-------------|------|
| `SearchTerm` | 位置参数（连字符合并） |
| `BasePath` | `opts.triesPath`（已解析优先级） |
| `ShipPath` | `opts.shipPath` |
| `InitialInput` | `--and-type`（优先于 SearchTerm） |
| `TestRenderOnce` | `--and-exit` |
| `TestKeys` | `--and-keys`（已通过 `parseTestKeys` 解析） |
| `TestConfirm` | `--and-confirm` |
| `ColorsEnabled` | `--no-colors` / `NO_COLOR` 取反 |
| `Theme` | 已解析优先级 |
| `Messages` | `i18n.ForLocale(opts.locale)` |

创建 model 后通过 `SetDialogFactory` 注入对话框工厂，以 `tea.WithOutput(os.Stderr)` 启动 Program。

## 退出码

| 码 | 含义 |
|----|------|
| 0 | 成功选择/执行 |
| 1 | 错误或用户取消 |
| 2 | 无命令（显示帮助后退出） |
