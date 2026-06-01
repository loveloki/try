#!/usr/bin/env bash
# ==============================================================================
# try 交互式 TUI 全测试用例套件一键集成运行器 (scripts/run_all_tui_tests.sh)
# ==============================================================================

set -uo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_suite() { echo -e "\n${BLUE}======================================================================${NC}\n${BLUE}[SUITE]${NC} $*"; }
log_stat() { echo -e "${GREEN}[STATS]${NC} $*"; }

# 1. 强制重新编译，确保使用的是最新代码二进制
echo -e "${BLUE}[INFO] 正在编译二进制以进行干净的全套件测试...${NC}"
go build -o try ./cmd/try/main.go
echo -e "${GREEN}[SUCCESS] 编译成功。${NC}"

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 测试列表
TESTS=(
    "test_01_search_navigation.sh"
    "test_02_create_clash.sh"
    "test_03_rename.sh"
    "test_04_delete.sh"
    "test_05_ship.sh"
    "test_06_escape_cancel.sh"
)

TOTAL=${#TESTS[@]}
PASSED=0
FAILED=0
FAILED_LIST=()

# 一步步执行场景测试
for test_file in "${TESTS[@]}"; do
    log_suite "正在运行测试用例: ${test_file}"
    
    if [ ! -f "${SCRIPT_DIR}/${test_file}" ]; then
        echo -e "${RED}[ERROR] 测试脚本 ${test_file} 不存在！${NC}"
        FAILED=$((FAILED + 1))
        FAILED_LIST+=("${test_file} (Not Found)")
        continue
    fi
    
    chmod +x "${SCRIPT_DIR}/${test_file}"
    
    # 物理执行
    if "${SCRIPT_DIR}/${test_file}"; then
        PASSED=$((PASSED + 1))
        echo -e "${GREEN}[PASS] ${test_file} 通过。${NC}"
    else
        FAILED=$((FAILED + 1))
        FAILED_LIST+=("${test_file}")
        echo -e "${RED}[FAIL] ${test_file} 失败！${NC}"
    fi
done

# ==============================================================================
# 📊 终期测试套件汇总汇报
# ==============================================================================
echo -e "\n${BLUE}======================================================================${NC}"
echo -e "                    TUI 自动化测试套件汇总报告"
echo -e "${BLUE}======================================================================${NC}"
log_stat "总测试数: ${TOTAL}"
log_stat "成功通过: ${GREEN}${PASSED}${NC}"

if [ ${FAILED} -gt 0 ]; then
    log_stat "失败用例: ${RED}${FAILED}${NC}"
    for failed_case in "${FAILED_LIST[@]}"; do
        echo -e "  - ${RED}${failed_case}${NC}"
    done
    echo -e "${BLUE}======================================================================${NC}\n"
    exit 1
else
    log_stat "失败用例: ${GREEN}0${NC}"
    echo -e "\n${GREEN}try TUI 测试用例套件全部通过 (ALL TESTS PASSED)${NC}"
    echo -e "${BLUE}======================================================================${NC}\n"
    exit 0
fi
