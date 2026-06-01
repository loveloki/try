#!/usr/bin/env bash
# ==============================================================================
# 场景 3: 就地重命名、边界与物理校验测试 (scripts/test_03_rename.sh)
# ==============================================================================

source "$(dirname "$0")/tui_test_common.sh"

log_info "【测试场景 3】测试 Ctrl+R 重命名对话框与输入限制校验..."

# 初始化沙盒环境
setup_test_env

# 为了确保列表顶端是指定的项目，我们首先注入一个新的用于重命名的文件夹，确保它是最新的
mkdir -p "$TEST_TRIES_DIR/new-experiment-2-2026-06-01"

# 启动 try
agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "./try" --no-wait --json >/dev/null
wait_stable

# 1. 重命名非法斜杠拦截校验
log_info "按 Ctrl+R 打开重命名对话框..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Ctrl+r --json >/dev/null
wait_stable

assert_contains "Rename" "重命名对话框标题成功呈现"
assert_contains "new-experiment-2-2026-06-01" "正确的待改名项目"

clear_input

log_info "键入非法字符 'invalid/name' 校验拦截..."
agent-tty --home "$AGENT_TTY_HOME" type "$SESSION_ID" --json -- "invalid/name"
wait_stable
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Enter --json >/dev/null
wait_stable

assert_contains "Name cannot contain /" "非法 '/' 被拦截报错"

# 2. 取消并重新进行合法更名
log_info "按 Escape 取消当前报错对话框并退回主列表..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Escape --json >/dev/null
wait_stable

assert_contains "🏠 Try Directory Selection" "成功退回主列表"

log_info "再次按 Ctrl+R 唤起对话框..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Ctrl+r --json >/dev/null
wait_stable

clear_input

log_info "键入合法新名称 'new-experiment-renamed'..."
agent-tty --home "$AGENT_TTY_HOME" type "$SESSION_ID" --json -- "new-experiment-renamed"
wait_stable
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Enter --json >/dev/null
wait_stable

# 物理磁盘验证
if [ -d "$TEST_TRIES_DIR/new-experiment-renamed" ] && [ ! -d "$TEST_TRIES_DIR/new-experiment-2-2026-06-01" ]; then
    log_success "磁盘物理重命名成功：new-experiment-2-2026-06-01 -> new-experiment-renamed"
else
    log_error "物理验证失败，重命名未成功执行！"
    exit 1
fi

log_success "【测试场景 3】通过"
