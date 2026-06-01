#!/usr/bin/env bash
# ==============================================================================
# 场景 5: Ship 一键正式发布与目录物理迁移测试 (scripts/test_05_ship.sh)
# ==============================================================================

source "$(dirname "$0")/tui_test_common.sh"

log_info "【测试场景 5】测试 Ctrl+G 一键发布、目标路径推导及物理迁移逻辑..."

# 初始化沙盒环境
setup_test_env

# 为该场景注入一个明确的待搬运实验：new-experiment-renamed
mkdir -p "$TEST_TRIES_DIR/new-experiment-renamed"

# 启动 try
agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "./try" --no-wait --json >/dev/null
wait_stable

# 1. 过滤到 renamed 实验文件夹
log_info "模糊匹配 'renamed'..."
agent-tty --home "$AGENT_TTY_HOME" type "$SESSION_ID" --json -- "renamed"
wait_stable

# 2. 按 Ctrl+G 打开发布框
log_info "按 Ctrl+G 唤起发布对话框..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Ctrl+g --json >/dev/null
wait_stable

assert_contains "Ship" "成功打开发布对话框"
assert_contains "new-experiment-renamed" "正确的源项目"

# 3. 按 Enter 提交，直接使用推导出的默认路径
log_info "按 Enter 提交发布迁移..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Enter --json >/dev/null
wait_stable

# 4. 磁盘物理校验
if [ -d "$TEST_SHIP_DIR/new-experiment-renamed" ] && [ ! -d "$TEST_TRIES_DIR/new-experiment-renamed" ]; then
    log_success "磁盘物理迁移成功：$TEST_TRIES_DIR/new-experiment-renamed -> $TEST_SHIP_DIR/new-experiment-renamed"
else
    log_error "物理发布失败，源目录未搬移，或者目标目录不存在！"
    exit 1
fi

log_success "【测试场景 5】通过"
