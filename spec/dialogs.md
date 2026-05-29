# 对话框系统

## 概述

选择器中有三种对话框：删除确认、重命名、ship（发布为正式项目）。每种对话框实现为独立的 Bubbletea 子模型，通过 SelectorModel 的 `activeDialog` 字段切换。

## 共享设计

### 对话框接口

对话框接口在两个包中分别定义，通过工厂模式避免循环依赖。

`selector/dialogs.go` 中定义消费方接口：

```go
// DialogInstance 对话框实例接口（导出供 CLI 层的工厂实现使用）
type DialogInstance interface {
    tea.Model
    Result() *SelectionResult
    Done() bool
    ViewContent() string
}
```

`dialog/dialog.go` 中定义生产方接口（语义相同，引用 selector 的类型）：

```go
type Dialog interface {
    tea.Model
    Result() *selector.SelectionResult
    Done() bool
    ViewContent() string
}
```

### DialogFactory（避免循环依赖）

selector 包不直接 import dialog 包，而是通过 `DialogFactory` 接口由 CLI 层注入：

```go
// selector/dialogs.go
type DialogFactory interface {
    NewDeleteDialog(items []DeleteItem, basePath, testConfirm string, width int, msgs *i18n.Messages) DialogInstance
    NewRenameDialog(entry *MatchedEntry, basePath string, width int, msgs *i18n.Messages) DialogInstance
    NewShipDialog(entry *MatchedEntry, basePath, shipPath string, width int, msgs *i18n.Messages) DialogInstance
}
```

CLI 层在 `cli.go` 中实现工厂并注入：

```go
model := selector.New(cfg)
model.SetDialogFactory(&dialogFactoryImpl{})
```

### 对话框路由

SelectorModel 在 `activeDialog != nil` 时，将所有 `tea.KeyPressMsg` 转发给对话框子模型：

```go
func (m SelectorModel) updateDialog(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
    dlg, cmd := m.activeDialog.Update(msg)
    m.activeDialog = dlg.(dialog)
    if m.activeDialog.Done() {
        if result := m.activeDialog.Result(); result != nil {
            m.selected = result
            return m, tea.Quit
        }
        m.activeDialog = nil
    }
    return m, cmd
}
```

### 进入对话框

进入任何对话框前，先清除删除模式状态：

```go
m.deleteMode = false
m.markedForDeletion = map[string]bool{}
```

### 按键处理

**自定义按键**（需手动拦截）：

| 按键 | 动作 |
|------|------|
| Enter | 确认操作 |
| ESC / Ctrl-C | 取消，返回选择器 |

**textinput 内置按键**（Backspace、Ctrl-A/E/B/F/K/W 等 Emacs 快捷键由 Bubbles `textinput` 自动处理，完整列表见 `selector.md`）。

### 对话框渲染

对话框通过 `ViewContent()` 返回渲染内容字符串（替换选择器的列表区域），居中显示在终端中。SelectorModel 的 `View()` 调用 `activeDialog.ViewContent()` 获取内容后包装为 `tea.View`：

```go
func (m SelectorModel) View() tea.View {
    // 对话框和主界面的渲染在 SelectorModel.View() 中统一处理
    // activeDialog != nil 时渲染对话框内容，否则渲染主界面
    // 最终通过 tea.NewView(content) 返回 tea.View
}
```

## 删除对话框

### 触发方式

1. 用 Ctrl-D 标记一个或多个条目
2. 按 Enter 确认进入删除对话框

### 标记机制

Ctrl-D 切换当前条目的标记状态：

```go
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
```

标记后列表底部显示红色 DELETE MODE 提示栏：
`" DELETE MODE  N marked  |  Ctrl-D: Toggle  Enter: Confirm  Esc: Cancel"`

标记的条目行显示 `dangerBg` 背景和 🗑️ 图标。

### 对话框界面

```
         🗑️  Delete 2 directories?
─────────────────────────────────────
🗑️ redis-experiment-2025-08-14
🗑️ test-project-2025-08-17


        Type YES to confirm: _

─────────────────────────────────────
        Enter: Confirm  Esc: Cancel
```

### 确认逻辑

```go
func (d *DeleteDialog) confirm() *SelectionResult {
    if d.confirmInput.Value() != "YES" {
        return nil // 非 YES → 取消
    }

    // 安全检查：每个路径必须在 basePath 内
    baseReal, _ := filepath.EvalSymlinks(d.basePath)
    var validated []DeleteItem
    for _, item := range d.markedItems {
        targetReal, _ := filepath.EvalSymlinks(item.Path)
        if !strings.HasPrefix(targetReal, baseReal + "/") {
            // 安全检查失败，拒绝删除
            return nil
        }
        validated = append(validated, DeleteItem{Path: targetReal, Basename: item.Basename})
    }

    return &SelectionResult{
        Type:     SelectDelete,
        Paths:    validated,
        BasePath: baseReal,
    }
}
```

安全措施：
- 必须精确输入 `YES`（大写）
- 使用 `filepath.EvalSymlinks` 解析符号链接后验证路径前缀
- 路径必须在 `basePath + "/"` 下（防目录穿越）

### 测试模式

测试模式下有两种确认输入来源：
1. `testKeys` — 从注入按键序列中逐字符读取直到 Enter（走正常的 Update 路径，逐字符送入 textinput）
2. `testConfirm` — 直接使用预设确认文本（跳过逐字符输入）

#### testConfirm 消费路径

`testConfirm` 由 `SelectorModel` 持有，在打开删除对话框时传入。对话框 `Init()` 检测到非空值后，直接将确认文本设置到 textinput 并立即提交：

```go
// SelectorModel 通过工厂创建删除对话框，收集标记条目列表
func (m SelectorModel) openDeleteDialog() (tea.Model, tea.Cmd) {
    var items []DeleteItem
    for _, entry := range m.cachedResults {
        if m.markedForDeletion[entry.Entry.Path] {
            items = append(items, DeleteItem{Path: entry.Entry.Path, Basename: entry.Entry.Basename})
        }
    }
    if m.dialogFactory != nil {
        dlg := m.dialogFactory.NewDeleteDialog(items, m.basePath, m.testConfirm, m.width, m.messages)
        m.activeDialog = dlg
        return m, dlg.Init()
    }
    return m, nil
}

// DeleteDialog 初始化：testConfirm 非空时跳过手动输入
func (d *DeleteDialog) Init() tea.Cmd {
    cmds := []tea.Cmd{d.confirmInput.Focus()}
    if d.testConfirm != "" {
        d.confirmInput.SetValue(d.testConfirm)
        cmds = append(cmds, func() tea.Msg {
            return tea.KeyPressMsg{Code: tea.KeyEnter}
        })
    }
    return tea.Batch(cmds...)
}
```

这样 `testConfirm` 不需要经过 SelectorModel 的 Update 拦截，而是由 Dialog 自身在初始化时消费，保持了职责内聚。

## 重命名对话框

### 触发方式

Ctrl-R，对当前选中条目生效。

### 对话框界面

```
          ✏️  Rename directory
─────────────────────────────────────
📁 redis-experiment-2025-08-14


        New name: redis-experiment-2025-08-14█

─────────────────────────────────────
        Enter: Confirm  Esc: Cancel
```

初始值为当前目录名，光标在末尾。

### 允许字符

```
[a-zA-Z0-9\-_\.\s\/]
```

比选择器搜索框多允许 `/` 和空格。空格在最终确认时转为 `-`。

> 可通过 `textinput` 的 `Validate` 回调函数实现字符过滤，不需要手动拦截每个按键。

### 确认逻辑

```go
func (d *RenameDialog) confirmRename() (*selector.SelectionResult, string) {
    newName := strings.TrimSpace(d.input.Value())
    newName = whitespaceRe.ReplaceAllString(newName, "-")
    oldName := d.entry.Entry.Basename

    if newName == "" { return nil, d.msgs.RenameEmpty }
    if strings.Contains(newName, "/") { return nil, d.msgs.RenameSlash }
    if newName == oldName { return nil, "" }
    if selector.DirExists(filepath.Join(d.basePath, newName)) {
        return nil, d.msgs.RenameExists + newName
    }

    return &selector.SelectionResult{
        Type:     selector.SelectRename,
        Old:      oldName,
        New:      newName,
        BasePath: d.basePath,
    }, ""
}
```

错误消息显示在输入框下方。

## Ship 对话框

### 触发方式

Ctrl-G，将临时实验发布为正式项目。

### 目标目录

ship 目标目录（`shipPath`）由 CLI 层按优先级解析后传入 SelectorModel，优先级详见 `config.md`。

```go
projectName := dateSuffix.ReplaceAllString(currentName, "")  // 去掉日期后缀
shipBuffer := filepath.Join(shipPath, projectName)            // 默认目标路径
```

例如 `~/src/tries/redis-experiment-2025-08-14`，默认 ship 到 `~/src/ship/redis-experiment`。

### 对话框界面

```
         🚀  Ship try to project
─────────────────────────────────────
📁 redis-experiment-2025-08-14

   Destination: /home/user/src/ship
   Move to: /home/user/src/ship/redis-experiment█

   The directory will be moved to the destination

─────────────────────────────────────
        Enter: Confirm  Esc: Cancel
```

用户可编辑完整目标路径。

### 允许字符

```
[a-zA-Z0-9\-_\.\s\/~]
```

比重命名多允许 `~`（用于输入 home 目录路径）。同样通过 `textinput.Validate` 实现。

### 确认逻辑

```go
func (d *ShipDialog) confirmShip() (*selector.SelectionResult, string) {
    dest := config.ExpandPath(strings.TrimSpace(d.input.Value()))

    if dest == "" { return nil, d.msgs.ShipEmptyErr }
    if selector.FileExists(dest) { return nil, d.msgs.ShipExistsErr + dest }
    parent := filepath.Dir(dest)
    if !selector.DirExists(parent) { return nil, d.msgs.ShipNoParentErr + parent }

    return &selector.SelectionResult{
        Type:     selector.SelectShip,
        Source:   d.entry.Entry.Path,
        Dest:     dest,
        Basename: d.entry.Entry.Basename,
        BasePath: d.basePath,
    }, ""
}
```

### Ship 后的行为

Go 直接执行：
1. 移动目录（Git worktree 用 `git worktree move`，普通目录用 `os.Rename`）
2. 输出提示消息到 stderr
3. 输出 cd 脚本到 stdout
