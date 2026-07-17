# 对话框系统

## 概述

选择器中有三种对话框：删除确认、重命名、ship（发布为正式项目）。每种对话框实现为独立的 Bubbletea 子模型，通过 SelectorModel 的 `activeDialog` 字段切换。

所有对话框的用户可见文本（标题、提示、错误消息、底栏按键说明）均通过 `i18n.Get()` 全局语言包获取，支持中英文切换（`en` / `zh`）。图标（🗑️、✏️、🚀）与分隔线由实现按终端宽度渲染；spec 不以 ASCII 示意图描述界面布局（见 `tui-framework.md` 规格文档图示约定）。

## 接口与工厂

对话框接口在两个包中分别定义，通过工厂模式避免循环依赖。

```go
// selector/dialogs.go
type DialogInstance interface {
    tea.Model
    Result() *SelectionResult
    Done() bool
    ViewContent() string
    OverlaysMainUI() bool  // 三种对话框均为 true
}

type DialogFactory interface {
    NewDeleteDialog(items []DeleteItem, basePath, testConfirm string, width int, colorsEnabled bool) DialogInstance
    NewRenameDialog(entry *MatchedEntry, basePath string, width int, colorsEnabled bool) DialogInstance
    NewShipDialog(entry *MatchedEntry, basePath string, shipPaths []string, width int, colorsEnabled bool) DialogInstance
}
```

CLI 层实现工厂并通过 `SetDialogFactory` 注入。

### 对话框路由

`activeDialog != nil` 时，`tea.KeyPressMsg` 转发给对话框。对话框完成后检查 `Result()`：非 nil 则设置 `selected` 并退出，nil 则清除对话框回到主界面。

三种对话框均通过 `OverlaysMainUI() == true` 叠加在主选择器界面之上，背景保持渲染标题、搜索框、列表和底栏。

### 进入对话框

进入重命名和 ship 对话框前，先清除删除模式状态（删除对话框本身在删除模式下触发，需保留标记）。

### 通用按键处理

| 按键 | 动作 |
|------|------|
| Enter | 确认操作 |
| ESC / Ctrl-C | 取消，返回选择器 |

textinput 内置按键（Backspace、Ctrl-A/E/B/F/K/W 等 Emacs 快捷键）自动处理。

### 通用样式

所有对话框均为居中圆角弹窗（`internal/dialog/modal.go` 的 `renderModalBox`）：
- 外框宽度 40–64 列，随终端宽度推导。
- 边框使用 `line` 颜色；删除对话框因其危险语义，边框使用 `danger` 色。
- 标题含图标（🗑️/✏️/🚀）与文案，使用 `header` 样式。
- 分隔线使用 `line` 颜色。
- 底栏为 key-badge 形式的快捷键提示（如 `Enter`、`Esc`、`Tab`、`←/→`）。
- 错误行使用 `danger` 样式，前缀 `⚠`。

合成由 `SelectorModel.View()` 调用 `overlayModal`（`internal/selector/overlay.go`），使用 Lipgloss `Compositor` + `Canvas` 将弹窗按 `(x, y)` 叠到背景上。

### 渲染约束（反例）

下列做法会导致边框断裂、竖线错位或背景被空白盖住，**禁止**：

| 禁止 | 说明 |
|------|------|
| 在 spec 或代码里用手写 `╭│╰`、`strings.Repeat("─", N)` 拼弹窗外框 | 宽度与 emoji/ANSI 不一致时必然错位；spec 中的 ASCII 示意图也属于此类，**不得**当作实现参考 |
| 逐行把弹窗字符串插入背景行 | 破坏 Lipgloss 边框与样式序列 |
| 用 `Place` 生成整屏空白前景再叠层 | 空白格覆盖主界面标题与列表 |

实现只通过 Lipgloss `Style.Border` 与 `Compositor` 合成，尺寸来自终端 `width`/`height` 与 `lipgloss.Width` 测量。

## 删除对话框

### 触发方式

1. 用 Space / Delete / Ctrl-D 标记一个或多个条目
2. 按 Enter 确认进入删除对话框

### 对话框界面

删除确认以**居中弹窗**叠放在主选择器界面之上。

弹窗内容包括：
- 标题：`🗑️  Delete {N} directories?`，`header` 样式。
- 待删目录列表：每项前缀 `✕`，`danger` 样式（删除线）。
- NO / YES 选择器：**默认高亮 NO**，使用 `header` 样式（反色）。
- YES 仅在选中时使用 `danger` 反色样式。
- 底栏 key badges：`←/→`、`Enter`、`Esc`。

### 确认逻辑

- 提供 **NO / YES** 两项选择器，**默认选中 NO**；用户须用 `←`/`→`（或 `Tab` 切换）选中 **YES** 后按 Enter 才会执行删除
- 选项文案在所有 locale 下均为英文 `NO` / `YES`（有意为之的安全设计）
- 默认 NO 时按 Enter 等同取消（`Result()` 为 nil，返回主界面）
- 使用 `filepath.EvalSymlinks` 解析符号链接后验证路径前缀
- 路径必须在 `basePath + "/"` 下（防目录穿越）
- 安全检查失败时拒绝删除（返回 nil）

### 测试模式

`testConfirm` 为 `"YES"` 时：Dialog 在 `Init()` 中将选择设为 YES 并自动投递 Enter。

## 重命名对话框

### 触发方式

Ctrl-R，对当前选中条目生效。

### 对话框界面

居中弹窗（`OverlaysMainUI() == true`）。内容包括：
- 标题：`✏️  Rename directory`，`header` 样式。
- 当前目录名：`muted` 样式。
- `New name:` 输入框（`textinput`，初始值为当前 `Basename`，光标在末尾），标签使用 `accent` 样式。
- 可选错误行：`⚠ {error}`，`danger` 样式。
- 底栏 key badges：`Enter`、`Esc`。

实现见 `internal/dialog/rename.go`。

### 允许字符

`[a-zA-Z0-9\-_.\s/]`。空格在最终确认时转为 `-`。通过 `textinput.Validate` 回调实现字符过滤。

### 确认逻辑

- 名称为空 → 错误消息
- 名称含 `/` → 错误消息
- 名称未变化 → 静默退出
- 目标已存在 → 错误消息
- 错误消息显示在输入框下方

## Ship 对话框

### 触发方式

Ctrl-G，将临时实验发布为正式项目。

### 目标目录

配置文件中的 `ships` 数组定义多个目标目录（默认 `~/src/ship` 和 `~/src/bug`）。对话框中通过 Tab / Shift-Tab 在目标目录之间切换。

自动推导项目名：去掉日期后缀，拼接当前选中的 ship 目录。

例如 `~/src/tries/redis-experiment-2025-08-14` → `~/src/ship/redis-experiment`。

### 对话框界面

居中弹窗（`OverlaysMainUI() == true`）。内容包括：
- 标题：`🚀  Ship try to project`，`header` 样式。
- 源目录名：`muted` 样式。
- 目标目录选项列表：`●` 标记当前选中项，`○` 标记未选中项。
- `Move to:` 可编辑目标路径（`textinput`，初始值为去掉日期后缀后的推导路径），标签使用 `accent` 样式。
- 可选错误行：`⚠ {error}`，`danger` 样式。
- 底栏 key badges：`Tab`、`Enter`、`Esc`。

实现见 `internal/dialog/ship.go`。

### 允许字符

`[a-zA-Z0-9\-_.\s/~]`。比重命名多允许 `~`（home 目录路径）。通过 `textinput.Validate` 实现。

### 确认逻辑

- 路径展开 `~` 后验证
- 目标为空 → 错误消息
- 目标已存在 → 错误消息
- 父目录不存在 → 错误消息
