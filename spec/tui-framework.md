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

`NewStyles(colorsEnabled)` 根据 `colorsEnabled` 与 `config.DetectTheme()` 返回带完整样式或纯文本样式。十六进制色值由 Bubbletea/Lipgloss 渲染器根据终端 `colorprofile` 自动降采样，避免业务层做二次转换。

### 主题系统

支持 `dark` 和 `light` 两套配色，色值与 `internal/selector/styles.go` 及 GUI 主题 token 保持一致。

```go
type colorToken struct {
    dark  string
    light string
}
```

#### 语义色板

| 名称 | Dark | Light | 用途 |
|------|------|-------|------|
| background | `#0d1117` | `#ffffff` | 页面背景（少用） |
| foreground | `#e6edf3` | `#1f2328` | 主文字、选中行前景 |
| surface | `#161b22` | `#f6f8fa` | 面板、footer、输入框背景 |
| surfaceHover | `#1f242c` | `#f3f4f6` | hover 背景 |
| surfaceSelected | `#21262d` | `#eef0f2` | 选中行背景 |
| line | `#30363d` | `#d0d7de` | 分隔线、边框 |
| header | `#5fafff` | `#005fd7` | 标题、选中箭头、活动标签 |
| highlight | `#5fafff` | `#005fd7` | 强调前景色 |
| match | `#ffaf5f` | `#af5f00` | 模糊搜索命中高亮 |
| accent | `#87d787` | `#008700` | 创建/成功提示 |
| danger | `#ff3b30` | `#d70000` | 删除标记、危险按钮 |
| dangerSurface | `#3f1e1c` | `#ffe5e5` | 标记删除行背景 |
| onDanger | `#ffffff` | `#ffffff` | 危险背景上的文字 |
| muted | `#8b949e` | `#6e7781` | 次要文本 |
| disabled | `#6e7681` | `#aeb4ba` | 不可用元素 |

### 主题检测

通过 `COLORFGBG` 环境变量推断终端亮暗（背景色 `7`/`15` 判定为 light），无法推断时默认 dark。

auto 检测逻辑：解析 `COLORFGBG` 环境变量（格式 `fg;bg`），背景色为 `7` 或 `15` 时判定为浅色终端返回 `"light"`，其他返回 `"dark"`。

### Styles 结构

```go
type Styles struct {
    Background, Foreground, Surface, SurfaceHover, SurfaceSelected lipgloss.Style
    Line, Header, Highlight, Match, Accent, Danger, DangerSurface, Muted, Disabled lipgloss.Style
    SelectedArrow, MarkedIcon, FolderIcon lipgloss.Style
    ScoreBarFilled, ScoreBarEmpty lipgloss.Style
    SourcePill, KeyBadge, DeleteModeBadge lipgloss.Style
    MarkedName, MarkedMeta lipgloss.Style
}
```

样式渲染统一通过 `Styles.Render(style, text)`，禁止在业务代码中直接构造 `lipgloss.NewStyle()`。

## 布局结构

```
Screen（Bubbletea View 输出）
  ├── Header → 标题、分隔线、搜索栏、分隔线、来源标签栏
  ├── Body   → 列表条目（可滚动）或空状态面板
  └── Footer → 创建输入预览（可选）、分隔线、状态栏/快捷键提示
```

Header 与 Footer 行数动态计算（Footer 在创建预览出现时增加 1 行），Body 填充剩余空间。整帧由 Bubbletea 渲染器在单次 flush 中输出。

### 规格文档图示约定

- **允许**：模块树、数据流、优先级链、代码签名等 ASCII（不含模拟终端像素）。
- **禁止**：手写 TUI 界面示意图（固定列宽 `────`、`╭│╰` 弹窗框、emoji 与英文混排的 mock 屏幕）。此类图在 Markdown 与终端中宽度不一致，易误导实现。
- **界面行为**：用有序列表、表格或字段说明描述；实现以 `internal/selector/view.go`、`internal/selector/layout.go`、`internal/dialog/` 为准。

## 列表行

每行占用 **2 行**：1 行内容 + 1 行水平分隔线。

### 内容行结构

左侧：选中箭头 `›`（仅选中行）+ 图标 + 名称（含日期后缀拆分）+ 右对齐元数据。

| 状态 | 图标 | 背景 | 文字样式 |
|------|------|------|----------|
| 普通 | `📁` | 无 | `foreground` |
| 选中 | `📁` | `surfaceSelected` | `foreground` 加粗 |
| 标记删除 | `✕` | `dangerSurface` | 白色删除线 |

### 元数据

右侧显示：
1. 评分条：5 段 ASCII 块（`█`/`░`），主色填充。
2. 相对时间：just now / Xm ago / Xh ago / Xd ago / Xmo ago / Xy ago。
3. 来源 pill（仅非 `tries` 来源）：`[source]` 样式。

### 模糊匹配高亮

命中字符使用 `match` 颜色（橙色）加粗。标记删除行上的命中字符保留删除线，以维持删除状态可读性。

## 来源标签栏

Header 第 5 行渲染来源过滤标签：`all`、`tries` 以及每个 ship 目录的 basename。

- 活动标签：`surfaceSelected` 背景 + `header` 前景，右侧显示数量徽章。
- 非活动标签：`surfaceHover` 背景 + `muted` 前景，右侧显示数量徽章。

数量在 `loadAllTries` 时计算并缓存到 `sourceCounts`。

## Footer

### 左侧

- 默认模式：`{N} items`。
- 删除模式：`DELETE {N}` 危险徽章。

### 右侧

快捷键以 key-badge 形式渲染（小背景块），例如 `Ctrl-T`、`Ctrl-D`、`Tab`、`Esc`。

### 创建输入预览

搜索框非空时，在分隔线上方额外渲染一行：`Create new: {input}`（`accent` 样式）。

## 空状态

- 加载中（`allTries == nil`）：居中显示 `⏳ Loading directories...`。
- 无目录：居中显示 `📁 No directories yet.` + `Ctrl-T: create the directory`。
- 搜索无匹配：居中显示 `🔍 No matches for '{query}'.` + 创建提示。

空状态面板使用 `surface` 背景与 `muted` 前景，高度由 `bodyHeight` 推导。

## 文本测量

- `lipgloss.Width(text)`：ANSI-aware 宽度计算。
- `lipgloss.Place(width, height, ...)`：居中放置内容。

## 终端尺寸

Bubbletea 自动检测终端大小变化并发送 `tea.WindowSizeMsg`。收到时优先使用 `TRY_WIDTH` / `TRY_HEIGHT` 环境变量覆盖。

## Bubbles 组件使用策略

### list 组件

使用 `charm.land/bubbles/v2/list` 管理光标追踪和滚动，禁用所有内置 UI，不使用内置过滤（搜索由独立 `textinput` 管理）。自定义 `ItemDelegate` 实现条目渲染。

### textinput 组件

搜索框和对话框各自拥有独立的 `textinput` 实例。搜索框渲染在 Header 区域，通过 `textinput.Validate` 实现各自的字符过滤规则。

## Emoji 处理

`lipgloss.Width()` 内部使用 `go-runewidth`，已正确处理 emoji 宽度（📁 = 2 列）。
