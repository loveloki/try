#!/usr/bin/env bash
# ==============================================================================
# 场景 4: 批量删除、错误确认拦截与 YES 物理执行测试 (scripts/test_04_delete.sh)
# ==============================================================================

source "$(dirname "$0")/tui_test_common.sh"

log_info "【测试场景 4】测试 Ctrl+D 批量标记删除、默认 NO 取消与切换 YES 删除..."

# 初始化沙盒环境
setup_test_env

# 启动 try
agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "./try" --no-wait --json >/dev/null
wait_stable

# 1. 过滤并标记第 1 个待删项目
log_info "检索 'projB'..."
agent-tty --home "$AGENT_TTY_HOME" type "$SESSION_ID" --json -- "projB"
wait_stable

log_info "按 Ctrl+D 标记..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Ctrl+d --json >/dev/null
wait_stable

# 回删 5 字符清空搜索
log_info "退格清空搜索词..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Backspace Backspace Backspace Backspace Backspace --json >/dev/null
wait_stable

# 2. 过滤并标记第 2 个待删项目
log_info "检索 'projC'..."
agent-tty --home "$AGENT_TTY_HOME" type "$SESSION_ID" --json -- "projC"
wait_stable

log_info "按 Ctrl+D 标记..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Ctrl+d --json >/dev/null
wait_stable

# 退格清空搜索词，以便呈现完整主视图
log_info "清空检索，返回完整主列表..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Backspace Backspace Backspace Backspace Backspace --json >/dev/null
wait_stable

assert_contains "DELETE MODE" "底部栏正确展示 DELETE 状态及已标记数"

# 3. 按 Enter 唤起删除对话框
log_info "按 Enter 打开删除确认框..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Enter --json >/dev/null
wait_stable

assert_contains "Delete" "删除对话框标题渲染"
assert_contains "projB-2026-05-02" "待删除列表正确高亮 projB"
assert_contains "projC-2026-05-03" "待删除列表正确高亮 projC"

# 4. 安全校验：默认 NO 直接 Enter (应取消)
log_info "默认选中 NO，直接按 Enter..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Enter --json >/dev/null
wait_stable

# 验证是否返回主列表，且目录没有被物理删除
assert_contains "🏠 Try Directory Selection" "安全退回主界面"
if [ -d "$TEST_TRIES_DIR/projB-2026-05-02" ] && [ -d "$TEST_TRIES_DIR/projC-2026-05-03" ]; then
    log_success "物理验证成功：默认 NO 下 Enter 取消，测试文件夹未被物理删除。"
else
    log_error "安全删除检验失败：文件夹在取消删除时仍被破坏！"
    exit 1
fi

# 5. 再次按 Enter 打开删除，输入 YES 执行真实物理删除
log_info "重新按 Enter 打开删除对话框..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Enter --json >/dev/null
wait_stable

log_info "按 Tab 切换到 YES 并 Enter 执行彻底删除..."
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Tab --json >/dev/null
wait_stable
agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" Enter --json >/dev/null
wait_stable

# 磁盘最终物理验证
if [ ! -d "$TEST_TRIES_DIR/projB-2026-05-02" ] && [ ! -d "$TEST_TRIES_DIR/projC-2026-05-03" ]; then
    log_success "磁盘物理验证成功：选中 YES 后批量物理删除文件夹"
else
    log_error "批量物理删除执行失败：目录依然存在！"
    exit 1
fi

log_success "【测试场景 4】通过"
