#!/usr/bin/env bash
# ==============================================================================
# 场景 1: 多维搜索与列表导航测试 (scripts/test_01_search_navigation.sh)
# ==============================================================================

source "$(dirname "$0")/tui_test_common.sh"

log_info "【测试场景 1】启动 TUI 主界面，测试搜索、列表导航与 Esc 退出..."

# 初始化沙盒环境
setup_test_env

# 启动 try 交互式选择器
agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "./try" --no-wait --json >/dev/null
wait_stable

# 1. 验证 3 个初始项目是否渲染
assert_contains "projA-2026-05-01" "主列表显示 projA"
assert_contains "projB-2026-05-02" "主列表显示 projB"
assert_contains "projC-2026-05-03" "主列表显示 projC"

# 2. 测试列表向下导航
log_info "发送 Down 按键，向下移动焦点..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Down --json >/dev/null
wait_stable

# 3. 测试搜索过滤逻辑
log_info "输入搜索词 'projB'..."
agent-tty --home "$AGENT_TTY_HOME" type "$SESSION_ID" --json -- "projB"
wait_stable

assert_contains "projB-2026-05-02" "搜索过滤包含 projB"
assert_not_contains "projA-2026-05-01" "搜索过滤排除 projA"
assert_not_contains "projC-2026-05-03" "搜索过滤排除 projC"

# 4. 测试 Backspace 逐字删除逻辑
log_info "回删 5 个字符以复原列表..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Backspace Backspace Backspace Backspace Backspace --json >/dev/null
wait_stable

assert_contains "projA-2026-05-01" "回删搜索词重新出现 projA"
assert_contains "projB-2026-05-02" "回删搜索词重新出现 projB"
assert_contains "projC-2026-05-03" "回删搜索词重新出现 projC"

# 5. 测试 Esc 退出
log_info "按 Escape 退出程序..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Escape --json >/dev/null
wait_stable

assert_not_contains "🏠 Try Directory Selection" "程序已成功退出"
log_success "【测试场景 1】通过"
