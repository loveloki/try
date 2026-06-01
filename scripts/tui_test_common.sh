#!/usr/bin/env bash
# ==============================================================================
# try TUI 自动化测试通用公共底层库 (scripts/tui_test_common.sh)
# ==============================================================================
#
# 提供隔离沙箱、PTY 会话生命周期管理、常用断言、输入框擦除及视觉快照提取等底层函数。
# 场景脚本通过 `source "$(dirname "$0")/tui_test_common.sh"` 引入并调用即可。
#

set -euo pipefail

# 终端输出着色
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # 无颜色

log_info() { echo -e "${BLUE}[INFO]${NC} $*"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $*"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
log_error() { echo -e "${RED}[ERROR]${NC} $*"; }

# 路径定义
export TEST_TRIES_DIR="$(pwd)/test-tries"
export TEST_SHIP_DIR="$(pwd)/test-ship"
export AGENT_TTY_HOME=""
export SESSION_ID=""

# 1. 初始化测试数据项目
init_test_data() {
    log_info "初始化隔离的测试目录与模拟项目..."
    rm -rf "$TEST_TRIES_DIR" "$TEST_SHIP_DIR"
    mkdir -p "$TEST_TRIES_DIR" "$TEST_SHIP_DIR"
    
    # 创建模拟的实验目录 (IsDir=true 才能被 try 读取)
    mkdir -p "$TEST_TRIES_DIR/projA-2026-05-01"
    mkdir -p "$TEST_TRIES_DIR/projB-2026-05-02"
    mkdir -p "$TEST_TRIES_DIR/projC-2026-05-03"
    log_success "模拟项目初始化完成。"
}

# 2. 编译可执行程序
compile_try() {
    if [ ! -f "./try" ]; then
        log_info "正在编译 try 可执行程序..."
        go build -o try ./cmd/try/main.go
        log_success "try 编译成功。"
    fi
}

# 3. 统一清理函数
cleanup_test_env() {
    if [ -n "${SESSION_ID:-}" ]; then
        log_info "正在销毁 PTY 会话: $SESSION_ID"
        agent-tty --home "$AGENT_TTY_HOME" destroy "$SESSION_ID" --json >/dev/null 2>&1 || true
    fi
    log_info "正在清理临时测试数据..."
    rm -rf "$TEST_TRIES_DIR" "$TEST_SHIP_DIR" "${AGENT_TTY_HOME:-}"
    log_success "清理完毕。"
}

# 4. 创建沙盒 PTY 环境
setup_test_env() {
    compile_try
    init_test_data

    AGENT_TTY_HOME="$(mktemp -d -t agent-tty-try-XXXXXX)"
    
    log_info "正在创建长活的 PTY 会话..."
    local JSON_OUTPUT
    JSON_OUTPUT=$(agent-tty --home "$AGENT_TTY_HOME" create --json -- /bin/bash)
    SESSION_ID=$(echo "$JSON_OUTPUT" | node -e "console.log(JSON.parse(require('fs').readFileSync(0, 'utf8')).result.sessionId)")
    log_success "PTY 会话创建成功, ID: $SESSION_ID"

    # 设置 EXIT 陷阱处理器，让调用方脚本能自动清理
    trap cleanup_test_env EXIT

    # 注入测试环境变量并锁定尺寸
    log_info "初始化 PTY 环境变量..."
    agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "export TRY_PATH=\"$TEST_TRIES_DIR\"" --json >/dev/null
    agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "export TRY_PROJECTS=\"$TEST_SHIP_DIR\"" --json >/dev/null
    agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "export TRY_LOCALE=en" --json >/dev/null
    agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "export TRY_WIDTH=80" --json >/dev/null
    agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "export TRY_HEIGHT=24" --json >/dev/null
    agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "export COLORTERM=truecolor" --json >/dev/null
    agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "export TERM=xterm-256color" --json >/dev/null
    agent-tty --home "$AGENT_TTY_HOME" run "$SESSION_ID" "unset NO_COLOR" --json >/dev/null
    log_success "环境变量配置完毕。"
}

# 5. 等待 TUI 画面渲染稳定
wait_stable() {
    agent-tty --home "$AGENT_TTY_HOME" wait "$SESSION_ID" --screen-stable-ms 450 --json >/dev/null
}

# 6. 获取 PTY 快照文本
get_snapshot() {
    agent-tty --home "$AGENT_TTY_HOME" snapshot "$SESSION_ID" --format text --json | \
        node -e "console.log(JSON.parse(require('fs').readFileSync(0, 'utf8')).result.text)"
}

# 7. 包含断言
assert_contains() {
    local pattern="$1"
    local desc="${2:-内容包含断言}"
    local snap
    snap=$(get_snapshot)
    if echo "$snap" | grep -Fq "$pattern"; then
        log_success "断言通过 [包含]: '$pattern' ($desc)"
    else
        log_error "断言失败 [未包含]: '$pattern' ($desc)"
        echo -e "${YELLOW}=== 当前终端快照开始 ===${NC}"
        echo "$snap"
        echo -e "${YELLOW}=== 当前终端快照结束 ===${NC}"
        
        # 保存失败截图
        local IMG_JSON IMG_PATH
        IMG_JSON=$(agent-tty --home "$AGENT_TTY_HOME" screenshot "$SESSION_ID" --json || true)
        if [ -n "$IMG_JSON" ]; then
            IMG_PATH=$(echo "$IMG_JSON" | node -e "try { console.log(JSON.parse(require('fs').readFileSync(0, 'utf8')).result.artifactPath); } catch(e) {}")
            if [ -n "$IMG_PATH" ] && [ -f "$IMG_PATH" ]; then
                cp "$IMG_PATH" "./failed_screenshot.png"
                log_warn "已捕获失败截图到 ./failed_screenshot.png"
            fi
        fi
        exit 1
    fi
}

# 8. 排除断言
assert_not_contains() {
    local pattern="$1"
    local desc="${2:-排除断言}"
    local snap
    snap=$(get_snapshot)
    if ! echo "$snap" | grep -Fq "$pattern"; then
        log_success "断言通过 [排除]: '$pattern' ($desc)"
    else
        log_error "断言失败 [包含非法内容]: '$pattern' ($desc)"
        echo -e "${YELLOW}=== 当前终端快照开始 ===${NC}"
        echo "$snap"
        echo -e "${YELLOW}=== 当前终端快照结束 ===${NC}"
        exit 1
    fi
}

# 9. 擦除输入框内容
clear_input() {
    log_info "清空输入框..."
    agent-tty --home "$AGENT_TTY_HOME" send-keys "$SESSION_ID" \
        Backspace Backspace Backspace Backspace Backspace Backspace Backspace Backspace Backspace Backspace \
        Backspace Backspace Backspace Backspace Backspace Backspace Backspace Backspace Backspace Backspace \
        Backspace Backspace Backspace Backspace Backspace Backspace Backspace Backspace Backspace Backspace \
        Backspace Backspace Backspace Backspace Backspace Backspace Backspace Backspace Backspace Backspace --json >/dev/null
    wait_stable
}
