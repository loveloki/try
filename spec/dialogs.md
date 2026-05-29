# 对话框系统

## 概述

选择器中有三种对话框：删除确认、重命名、ship（发布为正式项目）。每种对话框实现为独立的 Bubbletea 子模型，通过 SelectorModel 的 `activeDialog` 字段切换。

## 接口与工厂

对话框接口在两个包中分别定义，通过工厂模式避免循环依赖。

```go
// selector/dialogs.go
type DialogInstance interface {
    tea.Model
    Result() *SelectionResult
    Done() bool
    ViewContent() string
}

type DialogFactory interface {
    NewDeleteDialog(items []DeleteItem, basePath, testConfirm string, width int, msgs *i18n.Messages) DialogInstance
    NewRenameDialog(entry *MatchedEntry, basePath string, width int, msgs *i18n.Messages) DialogInstance
    NewShipDialog(entry *MatchedEntry, basePath, shipPath string, width int, msgs *i18n.Messages) DialogInstance
}
```

CLI 层实现工厂并通过 `SetDialogFactory` 注入。

### 对话框路由

`activeDialog != nil` 时，`tea.KeyPressMsg` 转发给对话框。对话框完成后检查 `Result()`：非 nil 则设置 `selected` 并退出，nil 则清除对话框回到主界面。

### 进入对话框

进入任何对话框前，先清除删除模式状态。

### 按键处理

| 按键 | 动作 |
|------|------|
| Enter | 确认操作 |
| ESC / Ctrl-C | 取消，返回选择器 |

textinput 内置按键（Backspace、Ctrl-A/E/B/F/K/W 等 Emacs 快捷键）自动处理。

## 删除对话框

### 触发方式

1. 用 Ctrl-D 标记一个或多个条目
2. 按 Enter 确认进入删除对话框

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

- 必须精确输入 `YES`（大写）
- 使用 `filepath.EvalSymlinks` 解析符号链接后验证路径前缀
- 路径必须在 `basePath + "/"` 下（防目录穿越）
- 安全检查失败时拒绝删除（返回 nil）

### 测试模式

`testConfirm` 由 Dialog 自身在 `Init()` 时消费：非空时预填充确认文本并自动提交 Enter。

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

自动推导：去掉日期后缀得到项目名，拼接 `shipPath`。

例如 `~/src/tries/redis-experiment-2025-08-14` → `~/src/ship/redis-experiment`。

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

`[a-zA-Z0-9\-_.\s/~]`。比重命名多允许 `~`（home 目录路径）。通过 `textinput.Validate` 实现。

### 确认逻辑

- 路径展开 `~` 后验证
- 目标为空 → 错误消息
- 目标已存在 → 错误消息
- 父目录不存在 → 错误消息
