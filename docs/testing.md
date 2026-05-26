# 测试体系

## 概述

try 有两套测试：Shell 集成测试（主要，语言无关）和 Go 单元测试。Shell 集成测试是规格合规性测试，设计为可以验证任何语言的实现。

## 测试目录结构

```
spec/
  ├── *.md                    # 行为规格文档
  └── tests/
      ├── runner.sh           # 测试执行器
      ├── test_01_basic.sh    # 基础合规（--help, --version）
      ├── test_02_test_parameters.sh  # 测试参数验证
      ├── ...                 # 共 37 个测试文件
      └── tmux/               # 端到端 TUI 测试（需要 tmux）

internal/
  ├── fuzzy/fuzzy_test.go     # 模糊匹配单元测试
  ├── script/script_test.go   # 脚本生成单元测试
  ├── git/uri_test.go         # Git URI 解析单元测试
  └── shell/detect_test.go    # Shell 检测单元测试
```

## Shell 集成测试（spec/tests/）

### 测试执行器

```bash
./spec/tests/runner.sh /path/to/try
./spec/tests/runner.sh "valgrind -q --leak-check=full ./dist/try"
```

支持在可执行文件前添加包装命令（如 valgrind 做内存检查）。

### runner_and_compare.sh

```bash
./spec/tests/runner_and_compare.sh /path/to/try1 /path/to/try2
```

对比两个实现的测试结果差异。用于验证新实现与参考实现的行为一致性。

### 测试环境

runner.sh 创建以下环境：

```bash
export TEST_ROOT=$(mktemp -d)
export TEST_TRIES="$TEST_ROOT/tries"

# 预创建测试目录，设置不同 mtime
mkdir -p "$TEST_TRIES/alpha-2025-11-01"      # touch -t 202511010000
mkdir -p "$TEST_TRIES/beta-2025-11-15"       # touch -t 202511150000
mkdir -p "$TEST_TRIES/gamma-2025-11-20"      # touch -t 202511200000
mkdir -p "$TEST_TRIES/project-with-long-name-2025-11-25"  # touch -t 202511250000
mkdir -p "$TEST_TRIES/no-date-suffix"        # touch（最近）

export TRY_WIDTH=80
export TRY_HEIGHT=24
```

5 个测试目录，不同 mtime 用于验证排序和时间显示。`TRY_WIDTH`/`TRY_HEIGHT` 固定终端大小，确保布局测试确定性。

### 导出的工具函数

| 函数 | 用途 |
|------|------|
| `try_run [args...]` | 执行 try（自动展开 `$TRY_CMD`，捕获内存错误） |
| `pass` | 标记测试通过（输出绿色 `.`） |
| `fail "desc" "expected" "output" "spec"` | 标记失败 |
| `section "name"` | 开始新的测试节 |

## 测试专用参数

### --and-exit

渲染 TUI 一次后立即退出，不进入交互循环。

- 退出码：始终为 1（视为取消）
- stdout：空
- stderr：渲染的 TUI 帧
- 用途：验证显示内容（排序、格式、条目列表）

```bash
output=$(try_run --path="$TEST_TRIES" --and-exit exec 2>&1)
echo "$output" | grep -q "expected text"
```

### --and-keys

注入按键序列，TUI 处理完后退出。

两种格式：
- **Symbolic（推荐）**：`--and-keys='DOWN,DOWN,ENTER'`
- **Raw**：`--and-keys=$'\x1b[B\x1b[B\r'`

关键规则：
- 序列必须以终止键结尾（Enter/ESC/Ctrl-C）
- 序列耗尽后自动发送 ESC
- Enter → exit 0 + stdout 输出 cd 脚本
- ESC → exit 1 + stdout 为空（selected == nil，不输出任何内容）

### --and-type

注入初始搜索文本（不需要逐字符输入）。

```bash
output=$(try_run --path="$TEST_TRIES" --and-type="beta" --and-exit exec 2>&1)
```

### --and-confirm

注入删除确认文本。

```bash
output=$(try_run --path="$TEST_TRIES" --and-keys='CTRL-D,ENTER' --and-confirm='YES' exec 2>/dev/null)
```

## 测试文件总览

| 文件 | 测试内容 |
|------|---------|
| test_01_basic.sh | --help, --version, 未知参数 |
| test_02_test_parameters.sh | --and-exit, --and-keys, TRY_WIDTH |
| test_03_commands.sh | 子命令分派（install, clone, worktree, exec） |
| test_04_tui_compliance.sh | TUI 布局（header, footer, 搜索栏） |
| test_05_script_format.sh | 脚本格式（cd 脚本输出, 路径引用, 注释） |
| test_06_path_option.sh | --path 选项和 TRY_PATH 环境变量 |
| test_07_clone_naming.sh | clone 自动命名和自定义命名 |
| test_08_keyboard.sh | 键盘导航（↑↓, Enter, ESC） |
| test_09_new_entry.sh | 创建新目录 |
| test_10_fuzzy.sh | 模糊匹配（大小写、部分匹配、评分） |
| test_11_display.sh | 条目显示格式 |
| test_12_worktree.sh | worktree 命令 |
| test_13_vim_nav.sh | Vim 风格导航 |
| test_14_install_shells.sh | install 多 Shell 支持 |
| test_15_url_shorthand.sh | URL 快捷识别 |
| test_16_delete.sh | 删除模式 |
| test_17_line_layout.sh | 行布局（left/right） |
| test_18_rwrite_backgrounds.sh | 背景色 |
| test_19_input_field.sh | 输入框 |
| test_20_scroll_behavior.sh | 滚动 |
| test_21_create_new.sh | 创建新目录 UI |
| test_22_styles_unicode.sh | 样式和 Unicode |
| test_23_terminal_sizes.sh | 不同终端大小 |
| test_24_navigation_edge.sh | 导航边界情况 |
| test_25_header_footer.sh | 页头页脚 |
| test_26_fuzzy_highlight.sh | 匹配高亮 |
| test_27_metadata_format.sh | 元数据格式 |
| test_28_exit_behavior.sh | 退出行为 |
| test_29_error_edge.sh | 错误边界情况 |
| test_30_ansi_sequences.sh | ANSI 转义序列 |
| test_31_rename.sh | 重命名 |
| test_32_git_uri.sh | Git URI 解析 |
| test_33_versioning.sh | 目录名版本化 |
| test_34_shell_init.sh | Shell 初始化 |
| test_35_rename_validation.sh | 重命名验证 |
| test_36_shell_eval.sh | Shell eval 安全性 |
| test_37_ship.sh | ship 功能 |

## 常见测试模式

### 模式 1：验证 TUI 渲染内容

```bash
output=$(try_run --path="$TEST_TRIES" --and-exit exec 2>&1)
if echo "$output" | grep -q "expected text"; then
    pass
else
    fail "description" "expected text" "$output" "spec.md"
fi
```

`2>&1` 将 stderr（TUI 渲染）重定向到 stdout 以便捕获。

### 模式 2：验证选择后的脚本输出

```bash
output=$(try_run --path="$TEST_TRIES" --and-keys="query"$'\r' exec 2>/dev/null)
if echo "$output" | grep -q "cd '"; then
    pass
fi
```

`2>/dev/null` 丢弃 TUI 渲染，只捕获 stdout 脚本。

### 模式 3：验证退出码

```bash
try_run --path="$TEST_TRIES" --and-keys=$'\x1b' exec >/dev/null 2>&1
if [ $? -eq 1 ]; then pass; fi
```

### 模式 4：临时测试目录

```bash
DEL_TEST_DIR=$(mktemp -d)
mkdir -p "$DEL_TEST_DIR/first-2025-11-01"
# ... 测试 ...
rm -rf "$DEL_TEST_DIR"
```

需要独立状态的测试自建临时目录，测试后清理。

## Go 单元测试

使用 `go test` 运行：

```bash
go test ./internal/...
```

测试覆盖：

- `internal/fuzzy/`：匹配算法、评分公式、空 query、不匹配、top-k 排序
- `internal/script/`：路径引用（含单引号转义）、各种脚本类型格式
- `internal/git/`：URI 解析（HTTPS/SSH/GitHub/其他）、目录命名、版本化
- `internal/shell/`：Shell 检测、install 功能
- `internal/selector/`：集成测试（使用 --and-keys 注入）

## CI 集成

GitHub Actions 中同时运行两套测试：

```yaml
- name: Go unit tests
  run: go test ./...

- name: Spec compliance tests
  run: |
    go build -o ./try ./cmd/try
    ./spec/tests/runner.sh ./try
```
