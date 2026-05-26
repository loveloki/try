# 命令行解析与分派

## 概述

手工解析 `os.Args`，不使用 CLI 框架（如 cobra、urfave/cli）。保持轻量和对参数顺序的完全控制。

## 解析顺序

```
1. 提取颜色标志：--no-colors / --no-expand-tokens
2. 检查 NO_COLOR 环境变量
3. 检查 --help / -h
4. 检查 --version / -v
5. 提取 --path VALUE 和 --theme VALUE（可出现在任何位置）
6. 提取测试参数：--and-type、--and-exit、--and-keys、--and-confirm
7. 解析配置（LoadConfig）并合并路径和主题
8. shift 出第一个非选项参数作为 command
9. 根据 command 分派到对应处理函数
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
| `nil` | `runSelector` | 无命令时直接进入交互式选择器 |
| `clone` | `cmdClone` | Git 仓库克隆 |
| `install` | `cmdInstall` | 写入 Shell 配置文件 |
| `exec` | 二级分派 | 包装函数内部调用入口 |
| `worktree` | `cmdWorktree` | 创建 worktree |
| 其他 | `cmdCd` | 默认视为查询词，进入选择器 |

### 查询词合并

当用户输入多个非选项参数时（如 `try foo bar`），将它们用连字符合并为单个搜索词：

```go
searchTerm := strings.Join(remainingArgs, "-")  // "foo bar" → "foo-bar"
```

这与目录命名规则一致（空格转连字符），使得 `try foo bar` 等效于搜索 `foo-bar`。

## exec 子命令

`exec` 是 Shell 包装函数的内部调用入口，支持二级分派：

```
try exec clone <url> [name]     → cmdClone
try exec worktree <dir> [name]  → cmdWorktree
try exec cd [query]             → cmdCd
try exec [query]                → cmdCd（默认）
```

## cmdCd 内部逻辑

`cmdCd` 是最复杂的分派点，按顺序处理四种情况：

```
1. 参数以 "clone" 开头 → 转发到 cmdClone
2. 参数以 "." 开头 → 处理 worktree / mkdir 快捷方式
3. 参数看起来像 Git URL → clone 工作流
4. 其他 → 启动交互式选择器
```

### "." 命令处理

```
try . name      → worktree（Git 仓库内）或 mkdir
try ./path name → 指定仓库路径
try .           → 报错退出（防误触，必须提供名称）
```

### Git URL 识别

满足任一条件即识别为 Git URL，自动进入 clone 工作流：

```
- 以 https:// 或 http:// 或 git@ 开头
- 包含 github.com
- 包含 gitlab.com
- 以 .git 结尾
```

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

**Token 模式**（含逗号，或整个字符串仅由大写字母和连字符组成）：

```
UP,DOWN,ENTER           → 方向键 + 回车
CTRL-D,TYPE=hello,ENTER → Ctrl-D + 逐字输入 "hello" + 回车
```

支持的 token：`UP`、`DOWN`、`LEFT`、`RIGHT`、`ENTER`、`ESC`、`BACKSPACE`、`CTRL-A` 到 `CTRL-W`、`TYPE=...`

#### TYPE=... 展开

`TYPE=text` 在解析阶段被展开为逐字符 token，keyToMsg 不需要处理 TYPE= 前缀：

```go
// parseTestKeys 中处理 TYPE= token
if strings.HasPrefix(token, "TYPE=") {
    text := token[5:]
    for _, ch := range text {
        keys = append(keys, string(ch))  // 每个字符作为独立 token
    }
    continue
}
keys = append(keys, token)
```

例如 `TYPE=hello` → `["h", "e", "l", "l", "o"]`，每个字符经 keyToMsg 转为 `tea.KeyPressMsg{Code: 'h', Text: "h"}` 等。

**Raw 模式**（其他情况）：

逐字符解析，自动识别 `\e[X` 转义序列（箭头键等）：

```go
// Raw 模式解析
for i := 0; i < len(spec); {
    if spec[i] == '\x1b' && i+2 < len(spec) && spec[i+1] == '[' {
        // ANSI 转义序列：\e[A=UP, \e[B=DOWN, \e[C=RIGHT, \e[D=LEFT
        switch spec[i+2] {
        case 'A': keys = append(keys, "UP")
        case 'B': keys = append(keys, "DOWN")
        case 'C': keys = append(keys, "RIGHT")
        case 'D': keys = append(keys, "LEFT")
        }
        i += 3
    } else if spec[i] == '\x1b' {
        keys = append(keys, "ESC")
        i++
    } else if spec[i] == '\r' || spec[i] == '\n' {
        keys = append(keys, "ENTER")
        i++
    } else if spec[i] < 0x20 {
        // 控制字符 → CTRL-X
        keys = append(keys, "CTRL-"+string(rune(spec[i]+'A'-1)))
        i++
    } else {
        keys = append(keys, string(spec[i]))
        i++
    }
}
```

## 启动选择器

CLI 解析完成后，构造 SelectorModel 并启动 Bubbletea Program：

```go
model := selector.New(selector.Config{
    SearchTerm:    searchTerm,    // ARGV 中的查询词（多个 arg 用连字符合并）
    BasePath:      triesPath,     // tries 根目录绝对路径
    ShipPath:      shipPath,      // ship 目标目录绝对路径
    InitialInput:  andType,       // --and-type 值（覆盖 searchTerm）
    TestRenderOnce: andExit,      // --and-exit 模式
    TestKeys:      andKeys,       // 注入的按键序列
    TestConfirm:   andConfirm,    // --and-confirm 值
    ColorsEnabled: colorsEnabled, // 是否启用颜色
    Theme:         theme,         // "dark" 或 "light"
})

p := tea.NewProgram(model, tea.WithOutput(os.Stderr))
// Alt screen 在 View() 返回的 tea.View 中声明式设置
result, err := p.Run()
```

`InitialInput` 非空时优先于 `SearchTerm` 设置 `textInput` 的初始值。

## 退出码

| 码 | 含义 |
|----|------|
| 0 | 成功选择/执行 |
| 1 | 错误或用户取消 |
| 2 | 无命令（显示帮助后退出） |
