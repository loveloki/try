#!/usr/bin/env bash
# ==============================================================================
# 场景 6: 多级 Escape 返回及直接中止退出测试 (scripts/test_06_escape_cancel.sh)
# ==============================================================================

source "$(dirname "$0")/tui_test_common.sh"

log_info "【测试场景 6】测试对话框多级 Escape 取消、主界面 Escape 中止退出逻辑..."

# 初始化沙盒环境
setup_test_env

# 启动 try
agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "./try" --no-wait --json >/dev/null
wait_stable

# 1. 选中第一项，打开重命名对话框
log_info "打开重命名对话框..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Ctrl+r --json >/dev/null
wait_stable

assert_contains "Rename" "对话框已开启"

# 2. 按一次 Escape 取消，退回到主列表
log_info "按一次 Escape 键取消对话框..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Escape --json >/dev/null
wait_stable

assert_contains "🏠 Try Directory Selection" "成功退回到普通主列表"

# 3. 再次按 Escape 键，应直接优雅退出 try 程序
log_info "在主列表再次按 Escape 键优雅退出程序..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Escape --json >/dev/null
wait_stable

assert_not_contains "🏠 Try Directory Selection" "程序已成功优雅中止退出"

log_success "【测试场景 6】通过"
