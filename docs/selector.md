# 交互式选择器

## 概述

`SelectorModel` 是核心交互组件，实现为 Bubbletea Model，提供目录列表浏览、模糊搜索、创建新目录等功能。使用 Bubbles `list` 组件管理列表显示（光标、滚动、分页），搜索通过独立的 `textinput` 组件实现。

## 初始化

```go
type Config struct {
    SearchTerm    string   // 初始搜索词（多个 arg 用连字符合并）
    BasePath      string   // tries 根目录绝对路径
    ShipPath      string   // ship 目标目录绝对路径
    InitialInput  string   // --and-type 注入的搜索文本
    TestRenderOnce bool    // --and-exit 模式
    TestKeys      []string // 注入按键序列
    TestConfirm   string   // 注入确认输入
    ColorsEnabled bool     // 是否启用颜色
    Theme         string   // "dark" 或 "light"
}

func New(cfg Config) SelectorModel
```

初始化时自动创建 `BasePath` 目录（若不存在）。

### list 组件配置

```go
delegate := EntryDelegate{markedForDeletion: map[string]bool{}}
l := list.New([]list.Item{}, delegate, 0, 0)
l.SetShowTitle(false)
l.SetShowStatusBar(false)
l.SetShowHelp(false)
l.SetShowPagination(false)
l.SetShowFilter(false)
l.SetFilteringEnabled(false)
l.DisableQuitKeybindings()
```

> 禁用 list 内置的所有 UI 元素（标题、状态栏、帮助、分页、过滤输入），仅使用其光标追踪和滚动能力。搜索输入由独立的 `textinput` 管理。

## 共享常量

```go
// 匹配 name-YYYY-MM-DD 格式的日期后缀
var dateSuffixRe = regexp.MustCompile(`-\d{4}-\d{2}-\d{2}$`)
```

在多处使用：
- `loadAllTries`：判断是否有日期后缀以增加基础分
- ship 对话框：`dateSuffixRe.ReplaceAllString(name, "")` 去掉日期后缀得到项目名
- 条目渲染：拆分名称和日期后缀部分做差异化显示

## 目录条目类型

```go
// 从文件系统加载的原始条目
type Entry struct {
    Basename  string    // 目录名（os.DirEntry.Name()）
    Path      string    // 绝对路径
    Mtime     time.Time // 修改时间（用于评分和时间显示）
    BaseScore float64   // 预计算的基础分（时间权重 + 日期后缀加成）
}

// 匹配后的条目（包含评分和高亮信息）
// 实现 list.Item 接口以供 Bubbles list 组件使用
type MatchedEntry struct {
    Entry              Entry
    Score              float64
    HighlightPositions []int
}

// list.Item 接口实现
// FilterValue 返回空字符串，因为我们不使用 list 内置过滤
func (m MatchedEntry) FilterValue() string { return "" }
func (m MatchedEntry) Title() string       { return m.Entry.Basename }
func (m MatchedEntry) Description() string { return "" }

// 删除操作的条目信息
type DeleteItem struct {
    Path     string // 安全检查后的绝对路径（EvalSymlinks 解析后）
    Basename string // 目录名（用于错误消息和状态提示）
}
```

## 目录加载

### loadAllTries

单次遍历 `basePath` 目录，构建条目列表（懒加载 + 缓存）：

```go
func (m *SelectorModel) loadAllTries() []Entry {
    // 使用缓存：m.allTries 非空时直接返回
    entries := os.ReadDir(m.basePath)
    for _, entry := range entries {
        if strings.HasPrefix(entry.Name(), ".") { continue } // 跳过隐藏文件
        if !entry.IsDir() { continue } // 只处理目录，跳过文件和 symlink

        info, err := entry.Info()
        // 处理 ENOENT / EACCES

        mtime := info.ModTime()
        hoursSinceMod := time.Since(mtime).Hours()
        baseScore := 3.0 / math.Sqrt(hoursSinceMod + 1)
        if dateSuffixRe.MatchString(entry.Name()) {
            baseScore += 2.0
        }
    }
}
```

### refreshList

当搜索词变化时，重新计算匹配结果并更新 `list` 组件的 items：

```go
func (m *SelectorModel) refreshList() tea.Cmd {
    query := m.textInput.Value()
    if query == m.lastQuery && m.cachedResults != nil {
        return nil
    }

    m.loadAllTries()
    maxResults := max(m.height - 6, 3)

    // selector.Entry → fuzzy.Entry 转换（详见 fuzzy-matching.md）
    fuzzyEntries := m.toFuzzyEntries(m.allTries)
    results := fuzzy.Match(fuzzyEntries, query, maxResults)

    // fuzzy.MatchResult → MatchedEntry 转换
    matched := make([]MatchedEntry, len(results))
    for i, r := range results {
        matched[i] = MatchedEntry{
            Entry:              r.Entry.Data.(Entry),
            Score:              r.Score,
            HighlightPositions: r.Positions,
        }
    }
    m.cachedResults = matched
    m.lastQuery = query

    // MatchedEntry 实现 list.Item 接口，直接设置到 list
    items := make([]list.Item, len(matched))
    for i, me := range matched {
        items[i] = me
    }
    return m.list.SetItems(items)
}
```

> `MatchedEntry` 已实现 `list.Item` 接口（见上文类型定义）。`FilterValue()` 返回空字符串因为过滤逻辑在外部管理。

## Bubbletea 生命周期

### Init

```go
func (m SelectorModel) Init() tea.Cmd {
    cmds := []tea.Cmd{
        m.textInput.Focus(),                                   // 聚焦搜索框
        m.refreshList(),                                       // 初始加载目录列表
        func() tea.Msg { return tea.RequestWindowSize() },     // 获取终端尺寸
    }

    // --and-exit 模式：渲染一帧后退出
    if m.testRenderOnce {
        cmds = append(cmds, func() tea.Msg { return renderOnceMsg{} })
    }

    // 测试模式按键注入：为每个待注入按键排队一个 testKeyMsg
    if len(m.testKeys) > 0 {
        for range m.testKeys {
            cmds = append(cmds, func() tea.Msg { return testKeyMsg{} })
        }
    }

    return tea.Batch(cmds...)
}
```

### Update

```go
// testKeyMsg 用于测试模式按键注入
type testKeyMsg struct{}

func (m SelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case renderOnceMsg:
        // --and-exit 模式：View 已被调用至少一次，立即退出
        return m, tea.Quit

    case testKeyMsg:
        // 测试模式：从注入队列取出下一个按键并处理
        if len(m.testKeys) > 0 {
            key := m.testKeys[0]
            m.testKeys = m.testKeys[1:]
            // 递归调用 Update 让注入的按键走正常处理流程
            return m.Update(keyToMsg(key))
        }
        // 队列耗尽 → 自动发送 ESC 退出
        return m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        // TRY_WIDTH/TRY_HEIGHT 环境变量覆盖（测试用）
        if w := envInt("TRY_WIDTH"); w > 0 { m.width = w }
        if h := envInt("TRY_HEIGHT"); h > 0 { m.height = h }
        m.list.SetSize(m.width, m.height - headerLines - footerLines)
        m.cachedResults = nil // 使缓存失效（高度变化影响 maxResults）

    case tea.KeyPressMsg:
        if m.activeDialog != nil {
            return m.updateDialog(msg)
        }
        return m.handleKey(msg)
    }

    // 非按键消息（如 Blink）转发给子组件
    var cmds []tea.Cmd
    var cmd tea.Cmd
    m.textInput, cmd = m.textInput.Update(msg)
    cmds = append(cmds, cmd)
    m.list, cmd = m.list.Update(msg)
    cmds = append(cmds, cmd)
    return m, tea.Batch(cmds...)
}
```

测试按键通过自定义 `testKeyMsg` 消息驱动，在 `Init()` 中排队发送（见上文 Init 代码），注入按键走完整的 Update 流程（包括 dialog 路由），与真实用户输入行为一致。

### View

主循环由 Bubbletea 驱动，每次状态更新后自动调用 View()。

## 按键处理

按键分为三层：自定义拦截 → textinput 组件处理 → list 组件处理（仅当 textinput 未消费时）。

### handleKey 流程

```go
func (m SelectorModel) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
    // 1. 自定义按键拦截（不转发给 textinput/list）
    switch {
    case key.Matches(msg, m.keys.Enter):
        return m.handleEnter()
    case key.Matches(msg, m.keys.CtrlP):
        m.list.CursorUp()
        return m, nil
    case key.Matches(msg, m.keys.CtrlN):
        m.list.CursorDown()
        return m, nil
    case key.Matches(msg, m.keys.CtrlD):
        return m.toggleDelete()
    case key.Matches(msg, m.keys.CtrlT):
        return m.handleCreateNew()
    case key.Matches(msg, m.keys.CtrlR):
        return m.enterRenameDialog()
    case key.Matches(msg, m.keys.CtrlG):
        return m.enterShipDialog()
    case key.Matches(msg, m.keys.Quit):
        return m.handleQuit()
    }

    // 2. 转发给 textinput（搜索输入）
    prevValue := m.textInput.Value()
    var cmd tea.Cmd
    m.textInput, cmd = m.textInput.Update(msg)

    // 3. 如果搜索词变化，刷新列表并重置光标
    if m.textInput.Value() != prevValue {
        m.list.Select(0)
        refreshCmd := m.refreshList()
        return m, tea.Batch(cmd, refreshCmd)
    }

    // 4. 未被 textinput 消费的导航键转发给 list
    var listCmd tea.Cmd
    m.list, listCmd = m.list.Update(msg)
    return m, tea.Batch(cmd, listCmd)
}
```

### handleEnter

```go
func (m SelectorModel) handleEnter() (tea.Model, tea.Cmd) {
    // 优先级 1：删除模式 → 打开删除确认对话框
    if m.deleteMode && len(m.markedForDeletion) > 0 {
        return m.openDeleteDialog()
    }
    // 优先级 2：选中条目 → cd
    if entry := m.selectedEntry(); entry != nil {
        m.selected = &SelectionResult{Type: SelectCD, Path: entry.Entry.Path}
        return m, tea.Quit
    }
    // 优先级 3：有输入但无匹配 → 创建新目录
    if m.textInput.Value() != "" {
        return m.handleCreateNew()
    }
    return m, nil
}
```

### handleQuit

```go
func (m SelectorModel) handleQuit() (tea.Model, tea.Cmd) {
    // 删除模式下 ESC/Ctrl-C → 退出删除模式，不退出程序
    if m.deleteMode {
        m.deleteMode = false
        m.markedForDeletion = map[string]bool{}
        m.deleteStatus = "Delete cancelled"
        return m, nil
    }
    // 正常退出（selected 为 nil → 不生成脚本 → exit 1）
    return m, tea.Quit
}
```

### handleCreateNew

```go
func (m SelectorModel) handleCreateNew() (tea.Model, tea.Cmd) {
    input := strings.TrimSpace(m.textInput.Value())
    if input == "" { return m, nil }

    name := strings.ReplaceAll(input, " ", "-")
    dateSuffix := time.Now().Format("2006-01-02")
    // 处理同名目录冲突（同一天创建多次时自动追加 -2、-3...）
    name = git.ResolveUniqueName(m.basePath, name, dateSuffix)
    dirName := name + "-" + dateSuffix
    m.selected = &SelectionResult{
        Type: SelectMkdir,
        Path: filepath.Join(m.basePath, dirName),
    }
    return m, tea.Quit
}
```

### toggleDelete

```go
func (m SelectorModel) toggleDelete() (tea.Model, tea.Cmd) {
    entry := m.selectedEntry()
    if entry == nil { return m, nil }
    path := entry.Entry.Path
    if m.markedForDeletion[path] {
        delete(m.markedForDeletion, path)
    } else {
        m.markedForDeletion[path] = true
        m.deleteMode = true
    }
    if len(m.markedForDeletion) == 0 {
        m.deleteMode = false
    }
    return m, nil
}
```

### enterRenameDialog / enterShipDialog

```go
func (m SelectorModel) enterRenameDialog() (tea.Model, tea.Cmd) {
    entry := m.selectedEntry()
    if entry == nil { return m, nil }
    m.deleteMode = false
    m.markedForDeletion = map[string]bool{}
    dlg := NewRenameDialog(entry, m.basePath)
    m.activeDialog = dlg
    return m, dlg.Init()
}

func (m SelectorModel) enterShipDialog() (tea.Model, tea.Cmd) {
    entry := m.selectedEntry()
    if entry == nil { return m, nil }
    m.deleteMode = false
    m.markedForDeletion = map[string]bool{}
    dlg := NewShipDialog(entry, m.basePath, m.shipPath)
    m.activeDialog = dlg
    return m, dlg.Init()
}
```

### Enter 键行为优先级

```
1. deleteMode == true && len(markedForDeletion) > 0 → 进入删除确认对话框
2. list 有选中项（SelectedItem() != nil）           → 选择该项（cd）
3. textInput.Value() 非空 && list 为空               → 创建新目录
4. 其他情况                                          → 无操作
```

### 自定义按键

| 按键 | 动作 |
|------|------|
| Enter (`\r`) | 按上述优先级处理 |
| Ctrl-P | 光标上移（list DefaultKeyMap 不含此键，需自定义拦截后调用 `list.CursorUp()`） |
| Ctrl-N | 光标下移（list DefaultKeyMap 不含此键，需自定义拦截后调用 `list.CursorDown()`） |
| Ctrl-D | 切换标记删除 |
| Ctrl-T | 立即创建新目录 |
| Ctrl-R | 进入重命名对话框 |
| Ctrl-G | 进入 ship 对话框 |
| Ctrl-C / ESC | 取消（删除模式下仅退出删除模式） |

### list 组件处理的按键

使用 `list.DefaultKeyMap()` 内置的按键绑定，不做额外定制。由于已通过 `SetShowFilter(false)`、`SetShowHelp(false)`、`DisableQuitKeybindings()` 等禁用了对应功能，内置的 `/`（过滤）、`?`（帮助）、`q`（退出）等键即使触发也是空操作。

| 按键 | 动作 |
|------|------|
| ↑ / k | 光标上移 |
| ↓ / j | 光标下移 |
| Home / g | 跳转首项 |
| End / G | 跳转末项 |

> **注意**：`j`/`k`/`g`/`G` 等可打印字符在我们的架构中会先被 `textinput` 消费（添加到搜索框），不会到达 list 组件。实际生效的导航键只有方向键和 Home/End。这是自然的行为优先级，无需手动处理。

### textinput 组件处理的按键

| 按键 | 动作 |
|------|------|
| Backspace | 删除光标前字符 |
| Ctrl-A | 光标移到行首 |
| Ctrl-E | 光标移到行尾 |
| Ctrl-B | 光标左移一字符 |
| Ctrl-F | 光标右移一字符 |
| Ctrl-K | 删除光标到行尾 |
| Ctrl-W | 删除前一个单词 |
| ← / → | 光标左右移动 |
| 可打印字符 | 插入到输入缓冲区 |

textinput 还内置了光标闪烁、Placeholder 显示、光标位置追踪，这些都不需要手动实现。

### 测试模式消息

测试模式使用两种自定义消息，已集成到上文的 Init 和 Update 中：

- `renderOnceMsg`：--and-exit 模式，渲染一帧后退出（exit 1，stderr 含 TUI 帧，stdout 为空）
- `testKeyMsg`：--and-keys 模式，从注入队列取出按键后走完整 Update 流程

> 详见 Init 中的排队逻辑和 Update 中的 case 分支。

#### keyToMsg 映射

`keyToMsg` 将 `parseTestKeys` 输出的字符串 token 转换为 `tea.KeyPressMsg`：

```go
func keyToMsg(token string) tea.KeyPressMsg {
    switch token {
    case "UP":
        return tea.KeyPressMsg{Code: tea.KeyUp}
    case "DOWN":
        return tea.KeyPressMsg{Code: tea.KeyDown}
    case "LEFT":
        return tea.KeyPressMsg{Code: tea.KeyLeft}
    case "RIGHT":
        return tea.KeyPressMsg{Code: tea.KeyRight}
    case "ENTER":
        return tea.KeyPressMsg{Code: tea.KeyEnter}
    case "ESC":
        return tea.KeyPressMsg{Code: tea.KeyEscape}
    case "BACKSPACE":
        return tea.KeyPressMsg{Code: tea.KeyBackspace}
    default:
        // CTRL-X → 使用 Mod 字段表示修饰键
        if strings.HasPrefix(token, "CTRL-") {
            ch := strings.ToLower(token[5:])
            return tea.KeyPressMsg{Code: rune(ch[0]), Mod: tea.ModCtrl}
        }
        // 单个可打印字符
        return tea.KeyPressMsg{Code: rune(token[0]), Text: token}
    }
}
```

`TYPE=...` 在 `parseTestKeys` 阶段已展开为逐字符 token，因此 `keyToMsg` 不需要处理 `TYPE=` 前缀。

## 渲染

### View 结构

```go
func (m SelectorModel) View() tea.View {
    var b strings.Builder

    if m.activeDialog != nil {
        b.WriteString(m.activeDialog.ViewContent())
    } else {
        b.WriteString(renderHeader(m))   // 标题 + 分隔线 + 搜索栏 + 分隔线
        b.WriteString(m.list.View())    // list 组件渲染（返回 string，非 tea.View）
        b.WriteString(renderFooter(m))   // 状态栏/快捷键
    }

    v := tea.NewView(b.String())
    v.AltScreen = true  // 声明式启用 alt screen
    return v
}
```

### Header

```
🏠 Try Directory Selection
────────────────────────────────────
Search: user-query█
────────────────────────────────────
```

### 滚动管理

由 `list` 组件自动处理。光标驱动滚动：光标超出可见范围时自动调整视口。`list.SetSize(width, bodyHeight)` 在窗口大小变化时同步更新。

### ItemDelegate（自定义条目渲染）

```go
type EntryDelegate struct {
    markedForDeletion map[string]bool
    width             int
}

func (d EntryDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
    entry := item.(MatchedEntry)
    isSelected := index == m.Index()

    // 1. 选中箭头（2字符）
    // 2. 图标（2字符）：🗑️ / 🔗 / 📁
    // 3. 名称（含日期后缀拆分 + 模糊高亮）
    // 4. 右对齐元数据（时间 + 评分）
}

func (d EntryDelegate) Height() int  { return 1 }
func (d EntryDelegate) Spacing() int { return 0 }
func (d EntryDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
```

### 条目行渲染

```
→ 📁 redis-server-2025-08-14                    2h ago, 3.2
  📁 another-project-2025-08-10                  5d ago, 1.5
  📁 quick-test                                 just now, 4.0
```

每行组成：选中箭头(2) + 图标(2) + 空格(1) + 名称 + 右对齐元数据（时间 + 评分）。

### 条目图标

根据条目状态选择不同图标：

| 状态 | 图标 | 说明 |
|------|------|------|
| 标记删除 | 🗑️ | `markedForDeletion` 中的条目 |
| 普通目录 | 📁 | 默认 |

### 条目名称格式化

日期后缀目录拆分为正常名称 + 暗淡日期：

```
"redis-server-2025-08-14" →
   highlight_name("redis-server") + muted/highlight("-") + muted("2025-08-14")
```

模糊匹配高亮位置应用：连续的匹配字符批量高亮（减少样式切换次数）。

### Footer 三种状态

| 条件 | Footer 内容 |
|------|-------------|
| `deleteStatus` 非空 | 粗体显示状态消息（如 "Deleted: xxx" 或 "Delete cancelled"），渲染后在下一次按键处理开头清除（`handleKey` 入口处 `m.deleteStatus = ""`） |
| `deleteMode` 为 true | 红色背景 DELETE MODE 栏：标记数量 + 操作提示 |
| 默认 | 居中暗淡快捷键提示 |

### "Create new" 行

输入非空时在列表下方显示。此行**不作为 list.Item 加入列表**，而是在 `renderFooter` 中、状态栏之前单独渲染：

```
📂 Create new: my-experiment-2025-08-17
```

与列表条目之间有一行空行分隔。当光标在 list 最后一项上继续按 ↓ 时，光标不移动（list 组件行为），用户需按 Ctrl-T 或 Enter 创建新目录。

### 创建新目录（handleCreateNew）

Ctrl-T 或在"Create new"行上按 Enter 触发：

- **输入框非空**：使用当前输入作为目录名（`{input}-YYYY-MM-DD`），通过 `git.ResolveUniqueName` 处理同名冲突后设置 selected = `{type: mkdir, path: fullPath}`
- **输入框为空**：不执行操作

### 空目录场景

当 tries 根目录下没有任何子目录时，列表区域为空。此时仅在输入非空时显示"Create new"行。用户可直接输入名称并按 Enter 或 Ctrl-T 创建新目录。

## 输出接口：选择结果

Bubbletea 退出后，通过 Model 获取 `selected` 结果。消费方根据 `Type` 调用对应的脚本生成函数。

```go
type SelectionResult struct {
    Type     SelectionType
    Path     string              // :cd, :mkdir
    Paths    []DeleteItem        // :delete
    Old, New string              // :rename
    Source, Dest string          // :ship
    Basename string              // :ship
    BasePath string              // :delete, :rename, :ship
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

## 获取当前选中项

通过 `list.SelectedItem()` 获取当前选中的条目：

```go
func (m *SelectorModel) selectedEntry() *MatchedEntry {
    item := m.list.SelectedItem()
    if item == nil { return nil }
    entry := item.(MatchedEntry)
    return &entry
}
```

所有需要操作当前项的功能（Enter 选择、Ctrl-D 标记、Ctrl-R 重命名、Ctrl-G ship）都通过此方法获取。

## 时间格式化

```
seconds < 60    → "just now"
minutes < 60    → "Xm ago"
hours < 24      → "Xh ago"
days < 7        → "Xd ago"
else            → "Xw ago"
```

显示在条目右侧元数据区域。
