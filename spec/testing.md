# 测试体系

## 概述

try 使用 Go 标准 `testing` 包编写测试，覆盖各模块的核心功能。

## 测试目录结构

```
internal/
  ├── cli/cli_test.go             # CLI 参数解析与命令分派测试
  ├── config/config_test.go       # 配置文件解析测试
  ├── selector/selector_test.go   # 选择器集成测试（使用 --and-keys 注入）
  ├── selector/entry_test.go      # 目录条目类型与工具函数测试
  ├── dialog/dialog_test.go       # 对话框测试
  ├── fuzzy/fuzzy_test.go         # 模糊匹配单元测试
  ├── script/script_test.go       # 脚本格式测试
  ├── script/exec_test.go         # 操作执行测试
  ├── git/uri_test.go             # Git URI 解析单元测试
  ├── shell/shell_test.go         # Shell 检测与安装测试
  └── i18n/messages_test.go       # 国际化消息测试
```

## 运行测试

```bash
go test ./...              # 运行所有测试
go test -race ./...        # 带竞态检测
go test -cover ./...       # 带覆盖率
go test -race -cover ./... # 完整检查
```

## 测试专用参数

### --and-exit

渲染 TUI 一次后立即退出，不进入交互循环。

- 退出码：始终为 1（视为取消）
- stdout：空
- stderr：渲染的 TUI 帧
- 用途：验证显示内容（排序、格式、条目列表）

```bash
output=$(TRY_PATH="$TEST_TRIES" try_run --and-exit exec 2>&1)
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
output=$(TRY_PATH="$TEST_TRIES" try_run --and-type="beta" --and-exit exec 2>&1)
```

### --and-confirm

注入删除确认文本。

```bash
output=$(TRY_PATH="$TEST_TRIES" try_run --and-keys='CTRL-D,ENTER' --and-confirm='YES' exec 2>/dev/null)
```

## 测试覆盖范围

| 测试文件 | 覆盖内容 |
|----------|---------|
| `cli/cli_test.go` | CLI 参数解析、命令分派、全局选项提取 |
| `config/config_test.go` | 配置文件加载、JSON 解析、路径/主题/语言优先级解析 |
| `selector/selector_test.go` | 选择器完整交互流程（通过 --and-keys 注入按键） |
| `selector/entry_test.go` | 条目类型、工具函数、时间格式化 |
| `dialog/dialog_test.go` | 删除/重命名/ship 对话框逻辑 |
| `fuzzy/fuzzy_test.go` | 匹配算法、评分公式、空 query、不匹配、top-k 排序 |
| `script/script_test.go` | 路径引用（含单引号转义）、cd 脚本格式 |
| `script/exec_test.go` | 操作执行（mkdir、rename、delete 等） |
| `git/uri_test.go` | URI 解析（HTTPS/SSH）、目录命名、版本化 |
| `shell/shell_test.go` | Shell 检测、包装函数模板生成、install 流程 |
| `i18n/messages_test.go` | 语言包完整性、ForLocale 选择 |

## 常见测试模式

### 模式 1：验证 TUI 渲染内容

```bash
output=$(TRY_PATH="$TEST_TRIES" try_run --and-exit exec 2>&1)
if echo "$output" | grep -q "expected text"; then
    pass
else
    fail "description" "expected text" "$output" "spec.md"
fi
```

`2>&1` 将 stderr（TUI 渲染）重定向到 stdout 以便捕获。

### 模式 2：验证选择后的脚本输出

```bash
output=$(TRY_PATH="$TEST_TRIES" try_run --and-keys="query"$'\r' exec 2>/dev/null)
if echo "$output" | grep -q "cd '"; then
    pass
fi
```

`2>/dev/null` 丢弃 TUI 渲染，只捕获 stdout 脚本。

### 模式 3：验证退出码

```bash
TRY_PATH="$TEST_TRIES" try_run --and-keys=$'\x1b' exec >/dev/null 2>&1
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

## CI 集成

GitHub Actions 中运行测试：

```yaml
- name: Go tests
  run: go test -race ./...
```
