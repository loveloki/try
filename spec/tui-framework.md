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

`--no-colors` / `NO_COLOR` 环境变量禁用所有样式。

```go
func newStyles(colorsEnabled bool, theme string) *styles
```

`colorsEnabled` 为 false 时所有样式设为空（无颜色、无加粗）。颜色降采样交由 bubbletea v2 内置渲染器处理，`newStyles` 不做额外 colorprofile 降采样以避免双重转换导致背景色丢失。

### 主题系统

支持 `dark` 和 `light` 两套配色，使用 GitHub 风格 256-color ANSI 码。通过 `themePalette` 结构定义色值，`newStyles` 根据 theme 参数选择色板。

```go
type themePalette struct {
    header     string // 标题/品牌色
    highlight  string // 选中箭头/强调前景色
    muted      string // 次要文本
    match      string // 模糊搜索命中高亮
    accent     string // 创建新目录等操作提示
    danger     string // 删除标记前景色
}
```

#### Dark 色板（GitHub Dark 风格）

| 名称 | ANSI 256 码 | 近似色 | 用途 |
|------|-------------|--------|------|
| header | 75 | #5fafff 浅蓝 | 标题文字 |
| highlight | 75 | #5fafff 浅蓝 | 选中箭头 |
| muted | 245 | #8a8a8a 灰 | 次要文本 |
| match | 215 | #ffaf5f 浅橙 | 搜索匹配高亮 |
| accent | 114 | #87d787 浅绿 | 操作提示 |
| danger | 196 | #ff0000 鲜红 | 删除标记前景色 |

#### Light 色板（GitHub Light 风格）

| 名称 | ANSI 256 码 | 近似色 | 用途 |
|------|-------------|--------|------|
| header | 26 | #005fd7 深蓝 | 标题文字 |
| highlight | 26 | #005fd7 深蓝 | 选中箭头 |
| muted | 242 | #6c6c6c 中灰 | 次要文本 |
| match | 130 | #af5f00 棕橙 | 搜索匹配高亮 |
| accent | 28 | #008700 深绿 | 操作提示 |
| danger | 160 | #d70000 深红 | 删除标记前景色 |

### 主题解析优先级

```
1. --theme 命令行参数（最高优先）
2. TRY_THEME 环境变量
3. ~/.config/try/config.json 中的 theme
4. auto（通过 COLORFGBG 环境变量推断，无法推断时默认 dark）
```

auto 检测逻辑：解析 `COLORFGBG` 环境变量（格式 `fg;bg`），背景色值 0-6 判定为浅色终端返回 "light"，其他返回 "dark"。

### 选中行与删除标记渲染

行样式的渲染通过扁平化的样式继承与组合来实现，避免样式嵌套导致的 ANSI 转义码解析错乱。

- **选中行**：应用 `selected` 样式进行加粗渲染。箭头、条目名称和元数据分别继承并组合 `selected` 的加粗样式。
- **删除标记行**：应用 `danger` 样式（带前景色与删除线）。为了确保稳定性，处于删除标记状态的行直接将其名称扁平化渲染为纯文本形式，不进行模糊匹配分段高亮。
- **删除模式底栏与删除确认弹窗**：底栏 `DELETE MODE`、取消提示、弹窗标题/列表/边框及 YES 选项均使用 `danger` 红色（`DeleteDialogStyles` / `NewDeleteDialogStyles`），与删除标记行一致。
- **样式继承**：在匹配高亮渲染时，高亮部分样式通过 `Inherit` 继承行基础样式（如加粗），确保渲染层级扁平且样式正确复合。
- **空白填充**：左右部分的对齐填充使用普通的空格填充，保持简单干净。

### styles 结构

```go
type styles struct {
    header    lipgloss.Style
    highlight lipgloss.Style
    muted     lipgloss.Style
    match     lipgloss.Style
    selected  lipgloss.Style // 仅包含加粗属性，供组件继承
    danger    lipgloss.Style // 包含前景色及删除线属性
    accent    lipgloss.Style
}

func (s *styles) render(style lipgloss.Style, text string) string
```

`render` 方法用 `style.Render(text)` 渲染文本，颜色降采样由 bubbletea 渲染器统一处理。

## 布局结构

```
Screen（Bubbletea View 输出）
  ├── Header → 标题、分隔线、搜索栏
  ├── Body   → 列表条目（可滚动）
  └── Footer → 状态栏、快捷键提示
```

Header 和 Footer 固定行数，Body 填充剩余空间。整帧由 Bubbletea 的渲染器在单次 flush 中输出（无闪烁）。

### 规格文档图示约定

- **允许**：模块树、数据流、优先级链、代码签名等 ASCII（不含模拟终端像素）。
- **禁止**：手写 TUI 界面示意图（固定列宽 `────`、`╭│╰` 弹窗框、emoji 与英文混排的 mock 屏幕）。此类图在 Markdown 与终端中宽度不一致，易误导实现。
- **界面行为**：用有序列表、表格或字段说明描述；实现以 `internal/selector/view.go`、`internal/dialog/` 为准。

### View 函数结构

`View()` 按 Header → Body（`list.View()`） → Footer 顺序拼接，通过 `tea.NewView()` 返回并声明式启用 alt screen。完整实现见 `selector.md`。

### 行布局

每行左侧为条目名称（含图标），右侧为元数据（相对时间与评分）。使用 `lipgloss.PlaceHorizontal` 拼接左右内容，配合 `lipgloss.Width()` 和 `lipgloss.Truncate()` 控制宽度。右侧内容在左侧过长时隐藏，保证左侧名称优先可见。

`maxContent = width - 1` 避免终端最后一列自动换行。

## 文本测量（Lipgloss 内置）

- `lipgloss.Width(text)`：ANSI-aware 宽度计算，内部使用 go-runewidth
- `lipgloss.Truncate(text, maxWidth)`：ANSI-aware 截断
- `lipgloss.TruncateLeft(text, maxWidth)`：从左侧截断（保留尾部）

## 终端尺寸（Bubbletea 内置）

Bubbletea 自动检测终端大小变化并发送 `tea.WindowSizeMsg`，只需在 Update 中存储 `width`/`height`。

唯一需要手动处理的：检查 `TRY_WIDTH` / `TRY_HEIGHT` 环境变量覆盖（测试用），在收到 WindowSizeMsg 时优先使用环境变量值。

## Bubbles 组件使用策略

### list 组件

使用 `charm.land/bubbles/v2/list` 管理光标追踪和滚动，禁用所有内置 UI，不使用内置过滤（搜索由独立 `textinput` 管理，排序混合时间权重需在外部控制）。自定义 `ItemDelegate` 实现条目渲染。初始化代码和使用策略见 `selector.md`。

### textinput 组件

搜索框和对话框各自拥有独立的 `textinput` 实例。搜索框渲染在 Header 区域，通过 `textinput.Validate` 实现各自的字符过滤规则。

## 填充线

分隔线使用 `strings.Repeat("─", width)` 填充，应用 muted 样式。

## Emoji 处理（Lipgloss 内置）

`lipgloss.Width()` 内部使用 `go-runewidth`，已正确处理 emoji 宽度（📁 = 2 列）。不需要单独引入 `go-runewidth` 或手动计算。
