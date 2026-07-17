---
name: tui-testing-with-agent-tty
description: 使用 agent-tty 测试 Go TUI 应用程序，模拟键盘输入、文本输入，捕获屏幕稳定性，并输出用于审查的像素级截图产物。在测试 TUI 应用程序、执行交互式终端命令、验证终端状态，或被要求编写自动化终端测试场景时使用。
disable-model-invocation: true
---

# Go TUI 终端自动化测试（基于 agent-tty）

本技能旨在指导 Agent 使用 `agent-tty`（终端自动化测试工具）对基于 PTY 运行的 Go 语言终端交互式程序（TUI，如 Bubbletea 应用）进行全自动的模拟交互与像素级验证。

## 核心概念

基于 Bubbletea 等框架编写的 Go TUI 程序必须运行在真实的 PTY（伪终端）会话中，才能正常捕获键盘事件、渲染交替屏幕（`AltScreen`）并精确计算终端的行列尺寸。`agent-tty` 能够在不同的 CLI 命令调用间保持一个长活（Long-lived）的、状态保留的 PTY 会话，这使得 TUI 自动化测试成为可能。

## 典型测试工作流

当需要测试或验证某个 TUI 程序时，应严格遵循以下步骤进行：

### 1. 准备阶段

- 在工作区内编译并生成 Go TUI 的可执行文件（例如：`go build -o try ./cmd/try/main.go`）。
- 准备一个干净、隔离的 mock 数据目录（例如：创建 `test-tries/` 模拟真实的文件夹树），确保测试运行的可重复性。
- 使用 `mktemp -d` 创建临时且隔离的 `AGENT_TTY_HOME`，避免污染或读取用户的默认全局配置。

### 2. PTY 会话生命周期管理

- **创建终端会话**：
  ```bash
  export AGENT_TTY_HOME="$(mktemp -d)"
  JSON_OUTPUT=$(agent-tty --home "$AGENT_TTY_HOME" create --json -- /bin/bash)
  SESSION_ID=$(echo "$JSON_OUTPUT" | node -e "console.log(JSON.parse(require('fs').readFileSync(0, 'utf8')).result.sessionId)")
  ```
- **注入测试环境变量**：
  在 PTY 中先执行一些关键的环境变量初始化命令（例如：指定测试路径，以及强制限制终端的尺寸，这对于 TUI 的界面断言至关重要）：
  ```bash
  agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "export TRY_PATH=\"$(pwd)/test-tries\"" --json
  agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "export TRY_WIDTH=80" --json
  agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "export TRY_HEIGHT=24" --json
  # 强制使 TUI 程序激活色彩系统与高级转义属性（避免因宿主环境 NO_COLOR 污染导致黑白降级）
  agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "export COLORTERM=truecolor" --json
  agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "export TERM=xterm-256color" --json
  agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "unset NO_COLOR" --json
  ```
- **非阻塞启动 TUI**：
  在 `run` 命令中使用 `--no-wait` 选项，以便让 TUI 程序在 PTY 后台自主持续运行，而不会导致测试脚本被无限阻塞：
  ```bash
  agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "./try" --no-wait --json
  ```
- **智能等待渲染稳定**：
  绝不进行盲目的 `sleep`！调用 `agent-tty` 监听屏幕渲染，当画面在一定时间内无任何新字符变化时，即可认定加载完成：
  ```bash
  agent-tty --home "$AGENT_TTY_HOME" wait "$SESSION_ID" --screen-stable-ms 1500 --json
  ```
- **强制销毁会话**：
  测试结束时，无论成功与否，都必须通过 `destroy` 命令强制清理并销毁该 PTY 会话，以防后台子进程泄漏和文件锁冲突：
  ```bash
  agent-tty --home "$AGENT_TTY_HOME" destroy "$SESSION_ID" --json
  ```

### 3. 验证与产物生成

- **提取纯文本快照**：
  适合用于快速、非阻塞的文本结构断言：
  ```bash
  agent-tty --home "$AGENT_TTY_HOME" snapshot "$SESSION_ID" --format text --json
  ```
- **捕获像素级 TUI 截图**：
  生成由真实浏览器内核（Chromium）渲染的 TUI 截图，非常适合审查弹窗动画、复杂交互样式以及提交给人类开发者或 CI 系统进行终期视觉走查。
  **自定义弹窗/叠层**若出现边框断裂、竖线错位，多为手写字符串叠层导致；实现规范见 `tui-testing` skill 中「自定义 TUI 组件实现」一节。
  ```bash
  IMG_JSON=$(agent-tty --home "$AGENT_TTY_HOME" screenshot "$SESSION_ID" --json)
  # 解析路径（在未安装 jq 的环境中使用 node.js 稳健读取）
  IMG_PATH=$(echo "$IMG_JSON" | node -e "console.log(JSON.parse(require('fs').readFileSync(0, 'utf8')).result.artifactPath)")
  cp "$IMG_PATH" ./screenshot.png
  ```

### 4. 模拟高级人机交互

- **模拟方向键与确认**：
  向终端中发送 `Down`（下）、`Up`（上）、`Enter`（回车）或 `Escape`（退出键）：
  ```bash
  agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Down Enter --json
  ```
- **模拟控制字符与快捷键**：
  向 Bubbletea 发送 `Ctrl+r`、`Ctrl+t`、`Ctrl+d` 等复合快捷键组合：
  ```bash
  agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Ctrl+r --json
  ```
- **模拟连续文本输入**：
  在各种 TUI 输入弹窗或搜索框中，向终端输入连续的文本字符串。**特别注意：**如果需要输入带有减号开头的文本（如 `-updated`），必须前置 `--` 符号来防止命令行解析器将参数误认为 CLI 命令行选项：
  ```bash
  agent-tty --home "$AGENT_TTY_HOME" type "$SESSION_ID" --json -- "-updated"
  ```

### 5. 完整功能与操作覆盖规范

在编写基于 `agent-tty` 的 TUI 测试脚本时，**必须确保进行完整的功能及操作覆盖，绝对不能遗漏任何可能的用户操作路径**。一个完备的 TUI 自动化测试应覆盖以下所有原子交互和逻辑场景：

1. **多维搜索与编辑覆盖**：
   - 键入搜索关键字进行列表实时刷新和排序验证。
   - 验证输入框的基础编辑功能（`Backspace` 逐字删除）以及 Emacs 快捷键（`Ctrl-A`/`Ctrl-E`/`Ctrl-K` 等）对搜索框的影响。
2. **列表导航覆盖**：
   - 包含普通方向键（`Up`/`Down`）和备用导航键（`Ctrl-P`/`Ctrl-N`）对列表焦点的双向移动控制。
3. **创建生命周期覆盖**：
   - 使用 `Ctrl-T` 触发一键创建。
   - 使用搜索词不为空时的 `Enter` 触发创建。
   - 对生成的 `name-YYYY-MM-DD` 格式文件夹进行磁盘物理存在性断言，以及对同名碰撞自动加后缀（如 `-2`）等特性的覆盖验证。
4. **批量标记与安全删除覆盖**：
   - 使用 `Ctrl-D` 在不同检索词下标记一个或多个不连续的实验目录。
   - 按 `Enter` 进入删除确认对话框。
   - 模拟在确认对话框中故意输入错误的确认词（如 `no`, `yes` 小写等），断言其被安全取消。
   - 默认 NO 直接 Enter 断言未删除；`Right` + Enter 选中 YES 后断言物理删除成功。
5. **就地重命名覆盖**：
   - 高亮某项，按 `Ctrl-R` 拉起重命名对话框。
   - 模拟同名碰撞、空格替换连字符、输入为空或包含 `/` 等边界输入，断言界面提示错误，并在磁盘上验证物理目录名确实同步改变。
6. **Ship (发布) 搬运覆盖**：
   - 高亮某项，按 `Ctrl-G` 拉起 Ship 对话框。
   - 验证其自动去日期后缀并推导 ship 路径的行为。
   - 模拟搬移普通目录（`os.Rename`）和 git worktree 目录（自动调用 `git worktree move`），并在磁盘上物理断言发布成功。
7. **取消与强制中止覆盖**：
   - 处于删除模式或任意对话框（删除/重命名/Ship）时，按 `Escape` (或 `Ctrl-C`) 取消当前操作或退回普通主列表。
   - 处于空标记的主界面时，按 `Escape` 直接退出的边界场景。

## ⚠️ 反模式与避坑指南

1. **禁止使用纯硬编码的 `sleep`**。始终应优先使用 `agent-tty wait --screen-stable-ms <ms>` 或者是 `wait --text <text>` 进行自适应、响应式的状态监听。
2. **禁止不带分隔符直接 `type` 以减号开头的输入**（例如 `type "$SID" "-updated"`），必须用 POSIX 标准参数分割符 `--` 对其进行封装。
3. **测试结束时切记销毁会话**。务必在脚本的错误捕捉或最末尾调用 `destroy`。
4. **确保在启动前强制锁定 TUI 的显示宽高**。当 TUI 对视口（Viewport）大小敏感时，在运行前先修改 `TRY_WIDTH` / `TRY_HEIGHT` 变量或调用 `agent-tty resize`，保证截图与快照结构的高保真度。
5. **解除色彩限制，防止“黑白降级”**：由于运行环境（如 IDE、CI）可能在全局环境变量中强行设置了 `NO_COLOR=1`，会诱导 TUI 自动关闭色彩。在 PTY 初始化时，**务必在会话内部执行 `unset NO_COLOR`**，并显式宣告彩色属性（如 `export COLORTERM=truecolor`），才能保证截图能够高保真捕获彩色和特殊修饰（如删除线、闪烁）。
