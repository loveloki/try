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
  生成由真实浏览器内核（Chromium）渲染的 TUI 截图，非常适合审查弹窗动画、复杂交互样式以及提交给人类开发者或 CI 系统进行终期视觉走查：
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

## ⚠️ 反模式与避坑指南

1. **禁止使用纯硬编码的 `sleep`**。始终应优先使用 `agent-tty wait --screen-stable-ms <ms>` 或者是 `wait --text <text>` 进行自适应、响应式的状态监听。
2. **禁止不带分隔符直接 `type` 以减号开头的输入**（例如 `type "$SID" "-updated"`），必须用 POSIX 标准参数分割符 `--` 对其进行封装。
3. **测试结束时切记销毁会话**。务必在脚本的错误捕捉或最末尾调用 `destroy`。
4. **确保在启动前强制锁定 TUI 的显示宽高**。当 TUI 对视口（Viewport）大小敏感时，在运行前先修改 `TRY_WIDTH` / `TRY_HEIGHT` 变量或调用 `agent-tty resize`，保证截图与快照结构的高保真度。
