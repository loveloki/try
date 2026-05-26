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
// 通过 colorprofile.Writer 控制颜色输出
// 禁用颜色时将 profile 设为 colorprofile.Ascii

func newStyles(colorsEnabled bool, theme string) *styles {
    var profile colorprofile.Profile
    if colorsEnabled {
        w := colorprofile.NewWriter(os.Stderr, os.Environ())
        profile = w.Profile
    } else {
        profile = colorprofile.Ascii
    }
    // 根据 theme 选择色板，构建样式...
}
```

### 主题系统

支持 `dark` 和 `light` 两套配色，使用 GitHub 风格 256-color ANSI 码。通过 `themePalette` 结构定义色值，`newStyles` 根据 theme 参数选择色板。

```go
// 色板定义
type themePalette struct {
    header     string // 标题/品牌色
    highlight  string // 选中箭头/强调前景色
    muted      string // 次要文本
    match      string // 模糊搜索命中高亮
    selectedBg string // 选中行背景
    dangerBg   string // 删除标记背景
    accent     string // 创建新目录等操作提示
}
```

#### Dark 色板（GitHub Dark 风格）

| 名称 | ANSI 256 码 | 近似色 | 用途 |
|------|-------------|--------|------|
| header | 75 | #5fafff 浅蓝 | 标题文字 |
| highlight | 75 | #5fafff 浅蓝 | 选中箭头 |
| muted | 245 | #8a8a8a 灰 | 次要文本 |
| match | 215 | #ffaf5f 浅橙 | 搜索匹配高亮 |
| selectedBg | 237 | #3a3a3a 深灰 | 选中行背景 |
| dangerBg | 52 | #5f0000 暗红 | 删除标记背景 |
| accent | 114 | #87d787 浅绿 | 操作提示 |

#### Light 色板（GitHub Light 风格）

| 名称 | ANSI 256 码 | 近似色 | 用途 |
|------|-------------|--------|------|
| header | 26 | #005fd7 深蓝 | 标题文字 |
| highlight | 26 | #005fd7 深蓝 | 选中箭头 |
| muted | 242 | #6c6c6c 中灰 | 次要文本 |
| match | 130 | #af5f00 棕橙 | 搜索匹配高亮 |
| selectedBg | 254 | #e4e4e4 浅灰 | 选中行背景 |
| dangerBg | 217 | #ffafaf 浅红 | 删除标记背景 |
| accent | 28 | #008700 深绿 | 操作提示 |

### 主题解析优先级

```
1. --theme 命令行参数（最高优先）
2. TRY_THEME 环境变量
3. ~/.config/try/config.json 中的 theme
4. auto（通过 COLORFGBG 环境变量推断，无法推断时默认 dark）
```

auto 检测逻辑：解析 `COLORFGBG` 环境变量（格式 `fg;bg`），背景色值 0-6 判定为浅色终端返回 "light"，其他返回 "dark"。

### 选中行背景色渲染

选中行或删除标记行的背景色通过将背景色融入每个组件的样式中实现（而非事后包裹整行）。原因：lipgloss 渲染产生的 ANSI 重置序列会清除外层背景色。

```go
// 为样式附加行背景色
withBg := func(s lipgloss.Style) lipgloss.Style {
    if !hasRowBg { return s }
    return s.Background(bgOnly.GetBackground())
}

// 每个组件分别应用：
arrow = styles.render(withBg(styles.highlight), "→ ")
icon  = styles.render(bgOnly, "📁 ")
name  = renderNameWithBg(entry, bgOnly, hasRowBg)
meta  = styles.render(withBg(styles.muted), timeStr)
```

非高亮的普通文本和 padding 空格也需要用 `bgOnly` 样式渲染，确保整行背景连续。

### styles 结构

```go
type styles struct {
    header     lipgloss.Style
    highlight  lipgloss.Style
    muted      lipgloss.Style
    match      lipgloss.Style
    selectedBg lipgloss.Style // 仅存储背景色，供 delegate 提取
    dangerBg   lipgloss.Style // 同上
    accent     lipgloss.Style
    profile    colorprofile.Profile
}

// 通过 colorprofile 降采样后渲染
func (s *styles) render(style lipgloss.Style, text string) string
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
