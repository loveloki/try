# 交互式选择器

## 概述

`SelectorModel` 是核心交互组件，实现为 Bubbletea Model，提供目录列表浏览、模糊搜索、创建新目录、删除/重命名/ship 等功能。使用 Bubbles `list` 组件管理列表显示（光标、滚动、分页），搜索通过独立的 `textinput` 组件实现。

## 初始化

```go
type Config struct {
    SearchTerm     string
    BasePath       string
    ShipPaths      []string
    InitialInput   string   // --and-type 注入的搜索文本
    TestRenderOnce bool     // --and-exit 模式
    TestKeys       []string // 注入按键序列
    TestConfirm    string   // 注入确认输入
    ColorsEnabled  bool
}

func New(cfg Config) SelectorModel
```

初始化时自动创建 `BasePath` 和所有 `ShipPaths` 目录（若不存在）。`InitialInput` 非空时优先于 `SearchTerm` 设置 `textInput` 的初始值。

### list 组件配置

禁用 list 内置的所有 UI 元素（标题、状态栏、帮助、分页、过滤输入），仅使用其光标追踪和滚动能力。搜索输入由独立的 `textinput` 管理。

## 共享常量

```go
var DateSuffixRe = regexp.MustCompile(`-\d{4}-\d{2}-\d{2}$`)
```

在多处使用：
- `loadAllTries`：判断是否有日期后缀以增加基础分
- ship 对话框：去掉日期后缀得到项目名
- 条目渲染：拆分名称和日期后缀部分做差异化显示

## 目录条目类型

```go
type Entry struct {
    Basename  string
    Path      string
    Mtime     time.Time
    BaseScore float64   // 预计算的基础分（时间权重 + 日期后缀加成）
    Source    string    // "tries" 或 ship 目录的 basename
}

type MatchedEntry struct {
    Entry              Entry
    Score              float64
    HighlightPositions []int
}

// list.Item 接口实现（FilterValue 返回空字符串，不使用 list 内置过滤）
func (m MatchedEntry) FilterValue() string { return "" }
func (m MatchedEntry) Title() string       { return m.Entry.Basename }
func (m MatchedEntry) Description() string { return "" }

type DeleteItem struct {
    Path     string // 安全检查后的绝对路径
    Basename string
}
```

## 目录加载

### loadAllTries

单次遍历 `basePath` 和所有 `ShipPaths` 目录，构建条目列表。行为：
- 懒加载 + 缓存（`allTries` 非空时直接返回）
- 跳过隐藏文件（`.` 开头）和非目录
- 基础分公式：`3.0 / sqrt(hoursSinceMod + 1)`
- 日期后缀匹配时额外加 2.0
- 同时计算 `sourceCounts`，供 Header 来源标签栏显示数量

### refreshList

当搜索词变化时重新计算匹配结果。行为：
- 查询未变化且缓存存在时直接返回
- 通过 `fuzzy.Match` 执行匹配（详见 `fuzzy-matching.md`）
- `maxResults = max(bodyHeight, 3)`，动态限制结果数
- selector.Entry → fuzzy.Entry → fuzzy.MatchResult → MatchedEntry 的类型转换链

## Bubbletea 生命周期

### Init

Init 排队以下命令：聚焦搜索框、初始加载目录列表、请求终端尺寸。`--and-exit` 模式额外排队 `renderOnceMsg`，`--and-keys` 模式为每个按键排队 `testKeyMsg`。

### Update

消息路由：

| 消息类型 | 处理 |
|----------|------|
| `renderOnceMsg` | 立即退出（--and-exit 模式） |
| `testKeyMsg` | 从注入队列取出按键走正常 Update 流程，队列耗尽后自动发送 ESC |
| `tea.WindowSizeMsg` | 更新尺寸（TRY_WIDTH/TRY_HEIGHT 覆盖），重设 list 大小，使缓存失效 |
| `tea.KeyPressMsg` | 有活跃对话框时转发给对话框，否则走 `handleKey` |
| 其他 | 转发给 textinput 和 list 子组件 |

### View

`activeDialog != nil` 时渲染对话框内容（若 `OverlaysMainUI()` 为 true 则叠加在主界面之上），否则渲染 Header + Body + Footer。通过 `tea.NewView` 声明式启用 alt screen。

## 按键处理

按键分为三层：自定义拦截 → textinput 组件处理 → list 组件处理（仅当 textinput 未消费时）。搜索词变化时重置光标到首项并刷新列表。

### Enter 键行为优先级

```
1. deleteMode == true && len(markedForDeletion) > 0 → 进入删除确认对话框
2. list 有选中项                                     → 选择该项（cd）
3. textInput.Value() 非空                            → 创建新目录
4. 其他情况                                          → 无操作
```

### 自定义按键

| 按键 | 动作 |
|------|------|
| Enter | 按上述优先级处理 |
| ↑ / ↓ | 上下移动（list 默认绑定） |
| Ctrl-P | 光标上移，到顶部后循环到末项 |
| Ctrl-N | 光标下移，到末项后循环到首项 |
| Space / Delete | 切换当前条目的删除标记 |
| Ctrl-D | 切换删除模式；无标记时标记当前项，再次按下清空标记并退出模式 |
| Ctrl-A | 标记当前过滤结果中的所有条目 |
| Ctrl-T | 创建新目录 |
| Ctrl-R | 进入重命名对话框 |
| Ctrl-G | 进入 ship 对话框 |
| Tab | 正向切换来源过滤（all → tries → ship → bug → all） |
| Shift-Tab | 反向切换来源过滤 |
| / / Ctrl-F | 清空搜索框以重新输入（搜索框始终聚焦） |
| Ctrl-C / ESC | 关闭对话框 → 清空搜索（若非空） → 退出删除模式 → 退出程序 |

### handleCreateNew 行为

- 输入框非空：空格替换为连字符，通过 `git.ResolveUniqueName` 处理同名冲突，生成 `{name}-YYYY-MM-DD` 格式目录名
- 输入框为空：显示提示消息

### toggleDelete / toggleDeleteMark 行为

`toggleDeleteMark` 切换当前条目的删除标记。`toggleDelete` 在有标记时清空并退出删除模式，无标记时调用 `toggleDeleteMark`。

所有标记集合的修改都通过 `copyMarkedMap` 深拷贝实现不可变性，并同步更新 delegate 的标记状态。

### 进入对话框

进入重命名和 ship 对话框前先清除删除模式。通过 `DialogFactory` 创建对话框实例（避免循环依赖）。

### list 组件处理的按键

方向键由 selector 拦截处理：`↑` 到顶部后循环到末项，`↓` 到末项后循环到首项。Home/End 等其他列表按键仍使用 `list.DefaultKeyMap()` 内置绑定。

### textinput 组件处理的按键

内置 Emacs 快捷键（Ctrl-A/E/B/F/K/W）、Backspace、方向键、可打印字符插入。

### 测试模式

- `renderOnceMsg`：--and-exit 模式，渲染一帧后退出（exit 1）
- `testKeyMsg`：--and-keys 模式，注入按键走完整 Update 流程
- `KeyToMsg`：将字符串 token（UP/DOWN/ENTER/ESC/SPACE/DELETE/SHIFT-TAB/CTRL-X/单字符）转换为 `tea.KeyPressMsg`
- `TYPE=...` 在 `ParseTestKeys` 阶段已展开为逐字符 token

## 渲染

### Header

自上而下（`renderHeader`，固定 5 行）：

1. 标题行：`msgs().Title`，`header` 样式。
2. `line` 分隔线。
3. 搜索行：搜索图标 + `msgs().SearchPrefix` + `textinput.View()`；右侧显示 `/` 快捷键徽章。
4. 第二条 `line` 分隔线。
5. 来源标签栏（`renderSourceTabs`）：活动标签使用 `surfaceSelected`+`header`，非活动标签使用 `surfaceHover`+`muted`，均显示数量徽章。

实现见 `internal/selector/layout.go`。

### 条目行

由 `EntryDelegate.Render` 生成，每行 2 行高。

- 第 1 行：选中箭头 + 图标 + 名称 + 右对齐元数据（评分条、时间、来源 pill）。
- 第 2 行：`line` 分隔线。

选中行：背景 `surfaceSelected`，前景 `foreground`，加粗。
标记删除行：左侧图标 `✕`，背景 `dangerSurface`，前景 `onDanger`，删除线，保留模糊匹配高亮。
普通行：前景 `foreground`，无选中背景；仅选中/标记行套用状态背景。

### Footer 三种状态

| 条件 | Footer 内容 |
|------|-------------|
| `deleteStatus` 非空 | 状态消息（下一次按键处理开头清除） |
| `deleteMode` 为 true | `DELETE {N}` 危险徽章 + key badges |
| 默认 | `{N} items` + 快捷键 key badges |

### "Create new" 行

输入非空时在列表下方、分隔线上方单独渲染（不作为 list.Item），提示文案使用 `accent` 样式，输入内容使用 `highlight` 样式。

## 输出接口

```go
type SelectionResult struct {
    Type     SelectionType
    Path     string       // :cd, :mkdir
    Paths    []DeleteItem // :delete
    Old, New string       // :rename
    Source   string       // :ship
    Dest     string       // :ship
    Basename string       // :ship
    BasePath string       // :delete, :rename, :ship
}

type SelectionType int
const (
    SelectCD SelectionType = iota
    SelectMkdir
    SelectDelete
    SelectRename
    SelectShip
)
```

返回 nil 时不生成脚本，以 exit 1 退出。

## 时间格式化

```
seconds < 60  → "just now"
minutes < 60  → "Xm ago"
hours < 24    → "Xh ago"
days < 30     → "Xd ago"
months < 12   → "Xmo ago"
else          → "Xy ago"
```
