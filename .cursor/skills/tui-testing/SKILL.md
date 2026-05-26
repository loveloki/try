---
name: tui-testing
description: >-
  审查和编写 Go TUI 应用测试，遵循 matklad 测试原则。用于发现测试盲区、
  检查值接收器/指针接收器陷阱、验证真实用户交互路径是否被覆盖。
  当用户提到 TUI 测试、Bubbletea 测试、检查测试覆盖、或 matklad 测试原则时触发。
---

# TUI 测试审查与编写

## 核心原则（matklad）

1. **测试功能而非代码**：测试用户可观测的行为（"输入文字后按 Ctrl-T 创建目录"），而非内部实现细节
2. **data-driven check 函数**：每个模块用 `check` 辅助函数封装被测 API，API 变更只改一处
3. **neural network test**：即使内部实现完全替换，只要行为不变测试仍应通过
4. **不 mock，用真实依赖**：上层测试直接调用真实的下层模块，用 temp dir 做文件操作

## 常见陷阱检查清单

### 1. 值接收器 vs 指针接收器

Bubbletea v2 的 `Init()` / `Update()` / `View()` 使用值接收器。调用子组件的指针方法时，
修改的是副本而非原始 model。

```go
// 错误：Focus() 是指针方法，修改的是 Init() 参数的副本
func (m Model) Init() tea.Cmd {
    return m.textInput.Focus() // m.textInput.focus = true 不会保留
}

// 正确：在 New() 中直接调用
func New() Model {
    ti := textinput.New()
    ti.Focus()  // 直接修改 ti，状态保留到 Model 中
    return Model{textInput: ti}
}
```

**审查规则**：在 `Init()` 和 `Update()` 中搜索所有对子组件指针方法的调用，
确认状态修改会被返回值携带回去。

### 2. 测试绕过了真实路径

测试基础设施（如 `driveModel`）可能绕过 `Init()`，导致初始化相关的 bug 不被发现。

```go
// 危险：InitialInput 通过 SetValue 直接设值，不经过 textInput.Update
sm := driveModel(t, Config{InitialInput: "text", TestKeys: []string{"CTRL-T"}})

// 安全：通过 TestKeys 模拟真实打字，经过 textInput.Update 路径
sm := driveModel(t, Config{TestKeys: []string{"h", "i", "CTRL-T"}})
```

**审查规则**：对每个测试问"这个测试走的是用户真实操作的代码路径吗？"

### 3. 预填值 vs 真实输入

`SetValue()` 不需要焦点，`Update(KeyPressMsg)` 需要焦点。
如果所有测试都用 `SetValue()` 预填，则焦点相关的 bug 永远不会被发现。

**审查规则**：确保至少有一个测试通过模拟按键输入文字。

## 审查流程

1. **列出所有 `driveModel` / 测试辅助函数**，检查它们跳过了哪些真实路径（Init、Focus、WindowSize 等）
2. **检查每个 `InitialInput` / `SetValue` 的使用**，确认是否有对应的真实输入测试
3. **搜索值接收器方法中的指针方法调用**：`func (m Model) Init/Update/View` 中调用 `.Focus()`、`.Blur()`、`.SetValue()` 等
4. **确认导航键路径都被测试**：Ctrl-P/N vs UP/DOWN 是不同的代码路径
5. **检查清除/退格行为**：输入清空后状态是否正确恢复

## Bubbletea v2 测试模式

### 无 TTY 环境的 Model 驱动

```go
func driveModel(t *testing.T, cfg Config) Model {
    m := New(cfg)
    var model tea.Model = m
    model, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
    sm := model.(Model)

    for _, k := range cfg.TestKeys {
        model, _ = sm.Update(KeyToMsg(k))
        sm = model.(Model)
        if sm.done {
            break
        }
    }
    return sm
}
```

### 按键构造

```go
// 普通字符
tea.KeyPressMsg{Code: 'a', Text: "a"}

// 控制键
tea.KeyPressMsg{Code: 't', Mod: tea.ModCtrl}  // Ctrl-T

// 特殊键
tea.KeyPressMsg{Code: tea.KeyEnter}
tea.KeyPressMsg{Code: tea.KeyEscape}
tea.KeyPressMsg{Code: tea.KeyBackspace}
```
