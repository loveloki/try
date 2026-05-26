# TUI 渲染

## 概述

TUI 基于 Bubbletea v2 + Lipgloss v2 构建。Bubbletea 提供 Elm Architecture 的运行时（Model/Update/View），Lipgloss 提供声明式样式，不需要手动管理 ANSI 转义码。

## 依赖

```
charm.land/bubbletea/v2    # TUI 框架
charm.land/bubbles/v2      # 预置组件（list、textinput 等）
charm.land/lipgloss/v2     # 样式
```

## 颜色系统

### 全局开关

`--no-colors` / `NO_COLOR` 环境变量禁用所有样式。实现方式：

```go
// 通过 Writer 的 ColorProfile 控制颜色输出
// Lipgloss 样式渲染需要一个 Writer（携带终端能力信息）
// 禁用颜色时将 Writer 的 profile 设为 lipgloss.NoColor

var writer *lipgloss.Writer

func initStyles(colorsEnabled bool) {
    if colorsEnabled {
        writer = lipgloss.DefaultWriter()  // 自动检测终端能力
    } else {
        writer = lipgloss.NewWriter(os.Stderr, lipgloss.WithColorProfile(lipgloss.NoColor))
    }
}

// 所有样式通过 writer 渲染，禁用颜色时自动剥离 ANSI 转义码
func render(style lipgloss.Style, text string) string {
    return writer.Render(style, text)
}
```

### 色彩主题

使用 Lipgloss 定义样式常量：

| 名称 | 用途 | 样式 |
|------|------|------|
| `headerStyle` | 标题文字 | Bold + 前景色 114（绿色） |
| `accentStyle` | 强调文字 | Bold + 前景色 214（橙色） |
| `highlightStyle` | 匹配高亮/选中箭头 | Bold + 黄色 |
| `mutedStyle` | 暗淡文字 | 前景色 245（灰色） |
| `matchStyle` | 匹配字符 | Bold + 前景色 226 |
| `selectedBgStyle` | 选中行背景 | 背景色 238（深灰） |
| `dangerBgStyle` | 危险操作背景 | 背景色 52（深红） |

```go
var (
    highlightStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33"))
    mutedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
    selectedBg     = lipgloss.NewStyle().Background(lipgloss.Color("238"))
    dangerBg       = lipgloss.NewStyle().Background(lipgloss.Color("52"))
)
```

## 布局结构

```
Screen（Bubbletea View 输出）
  ├── Header → 标题、分隔线、搜索栏
  ├── Body   → 列表条目（可滚动）
  └── Footer → 状态栏、快捷键提示
```

Header 和 Footer 固定行数，Body 填充剩余空间。整帧由 Bubbletea 的渲染器在单次 flush 中输出（无闪烁）。

### View 函数结构

`View()` 按 Header → Body（list.View()） → Footer 顺序拼接，通过 `tea.NewView()` 返回并声明式启用 alt screen。完整实现见 `selector.md`。

### 行布局

每行支持左对齐和右对齐内容：

```
[左侧内容]                              [右侧内容]
```

使用 `lipgloss.PlaceHorizontal(width, lipgloss.Left, left) + lipgloss.PlaceHorizontal(width, lipgloss.Right, right)` 拼接，配合 `lipgloss.Width()` 和 `lipgloss.Truncate()` 控制内容宽度。右侧内容在左侧过长时隐藏，保证左侧名称优先可见。

`maxContent = width - 1` 避免终端最后一列自动换行。

## 文本测量（Lipgloss 内置，无需手动实现）

### 可见宽度

```go
lipgloss.Width(text)  // ANSI-aware 宽度计算，内部使用 go-runewidth
```

### 文本截断

```go
lipgloss.Truncate(text, maxWidth)  // ANSI-aware 截断 + 省略号
```

从左侧截断（保留尾部）：`lipgloss.TruncateLeft(text, maxWidth)`。

## 终端尺寸（Bubbletea 内置，仅需存储值）

Bubbletea 自动检测终端大小变化并发送 `tea.WindowSizeMsg`，只需在 Update 中存储：

```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
```

唯一需要手动处理的：检查 `TRY_WIDTH` / `TRY_HEIGHT` 环境变量覆盖（测试用），在收到 WindowSizeMsg 时优先使用环境变量值。

## Bubbles 组件使用策略

### list 组件

使用 `charm.land/bubbles/v2/list` 管理光标追踪和滚动，禁用所有内置 UI，不使用内置过滤（搜索由独立 `textinput` 管理，排序混合时间权重需在外部控制）。自定义 `ItemDelegate` 实现条目渲染。初始化代码和使用策略见 `selector.md`。

### textinput 组件

搜索框和对话框各自拥有独立的 `textinput` 实例。搜索框渲染在 Header 区域，通过 `textinput.Validate` 实现各自的字符过滤规则。

## 填充线

分隔线使用字符重复填充：

```go
separator := mutedStyle.Render(strings.Repeat("─", m.width))
```

## Emoji 处理（Lipgloss 内置，无需手动实现）

`lipgloss.Width()` 内部使用 `go-runewidth`，已正确处理 emoji 宽度（📁 = 2 列）。不需要单独引入 `go-runewidth` 或手动计算。
