#!/usr/bin/env bash
# ==============================================================================
# 场景 2: 一键创建新项目与重名碰撞测试 (scripts/test_02_create_clash.sh)
# ==============================================================================

source "$(dirname "$0")/tui_test_common.sh"

log_info "【测试场景 2】测试新建实验项目及重名碰撞处理..."

# 初始化沙盒环境
setup_test_env

# 1. 运行并输入新项目名 "new-experiment"
agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "./try" --no-wait --json >/dev/null
wait_stable

log_info "键入新项目名称 'new-experiment'..."
agent-tty --home "$AGENT_TTY_HOME" type "$SESSION_ID" --json -- "new-experiment"
wait_stable

assert_contains "Create new: new-experiment" "底部显示创建提示"

# 按 Ctrl+T 触发一键新建并退出
log_info "按 Ctrl+T 一键创建..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Ctrl+t --json >/dev/null
wait_stable

EXPECTED_DIR="$TEST_TRIES_DIR/new-experiment-2026-06-01"
if [ -d "$EXPECTED_DIR" ]; then
    log_success "磁盘上物理成功创建新项目: $EXPECTED_DIR"
else
    log_error "磁盘物理验证失败: $EXPECTED_DIR"
    exit 1
fi

# 2. 验证同名冲突自动加 "-2"
log_info "再次启动 try 以验证重名碰撞..."
agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "./try" --no-wait --json >/dev/null
wait_stable

log_info "输入相同的项目名 'new-experiment'..."
agent-tty --home "$AGENT_TTY_HOME" type "$SESSION_ID" --json -- "new-experiment"
wait_stable

log_info "再次按 Ctrl+T 创建..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Ctrl+t --json >/dev/null
wait_stable

EXPECTED_DIR_2="$TEST_TRIES_DIR/new-experiment-2-2026-06-01"
if [ -d "$EXPECTED_DIR_2" ]; then
    log_success "同名碰撞成功自动追加后缀: $EXPECTED_DIR_2"
else
    log_error "同名碰撞去重失败: $EXPECTED_DIR_2"
    exit 1
fi

log_success "【测试场景 2】通过"
