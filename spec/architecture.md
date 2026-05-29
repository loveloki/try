# 总体架构

## 项目定位

try 是一个临时实验目录管理工具。所有实验集中存放在单一根目录（默认 `~/src/tries`），用日期后缀自动命名（`name-YYYY-MM-DD`），通过模糊搜索和时间权重快速定位。名称在前可提高模糊匹配效率。

## 技术栈

- **语言**：Go（编译为单一二进制，零运行时依赖）
- **TUI 框架**：Bubbletea v2（Elm Architecture：Model/Update/View）
- **样式**：Lipgloss v2
- **组件**：Bubbles v2（list、textinput 等）

```
go get charm.land/bubbletea/v2
go get charm.land/bubbles/v2
go get charm.land/lipgloss/v2
```

## 项目结构

```
cmd/try/main.go        # 入口：调用 cli.Run(os.Args[1:])
internal/
  cli/                 # CLI 解析与命令分派
    cli.go             # Run 主入口、parseGlobalFlags、runSelector、帮助文本
    commands.go        # cmdExec、cmdClone、cmdWorktree、handleDot、worktreePath
    flags.go           # 参数提取工具函数（hasFlag、extractPath、extractValueFlag 等）
  config/              # 配置文件加载（~/.config/try/config.json）
    config.go          # Config 结构、LoadConfig、ResolvePaths、ResolveTheme、ResolveLocale
  selector/            # 交互式选择器（Bubbletea Model + Bubbles list）
    model.go           # SelectorModel：主界面状态与逻辑
    delegate.go        # 自定义 list.ItemDelegate 渲染
    view.go            # View 渲染（Header + list.View() + Footer）+ styles 定义
    keys.go            # 按键绑定
    entry.go           # 目录条目类型定义（实现 list.Item 接口）+ 工具函数
    loader.go          # 目录加载（loadAllTries）和列表刷新（refreshList）
    dialogs.go         # DialogInstance 接口、DialogFactory、对话框路由
    testkeys.go        # 测试按键解析（ParseTestKeys、KeyToMsg）
  dialog/              # 对话框子模型
    dialog.go          # Dialog 接口定义
    delete.go          # 删除确认对话框
    rename.go          # 重命名对话框
    ship.go            # ship 对话框
  fuzzy/               # 模糊匹配引擎
    fuzzy.go           # 子序列匹配 + 多维评分
  shell/               # Shell 集成
    detect.go          # Shell 类型检测
    install.go         # install 命令：写入 Shell 配置文件
    template.go        # 包装函数模板生成
  script/              # 操作执行 + cd 脚本生成
    exec.go            # Go 直接执行副作用操作（mkdir/rm/mv/git）
    script.go          # cd 脚本输出（Quote + EmitCd）
  git/                 # Git 集成
    uri.go             # Git URI 解析、IsGitURI、GenerateCloneDirName、ResolveUniqueName
go.mod
go.sum
```

## 设计原则

### 单二进制零依赖

编译为静态链接的单一二进制文件，用户无需安装 Go 运行时。通过 `go build` 或 Nix 分发。

### stdout/stderr 分离

这是核心架构决策。Shell 子进程无法改变父进程的工作目录，try 通过如下方式解决：

- **stderr** → TUI 渲染（连接 TTY，用户可见）
- **stdout** → Shell 命令脚本（被包装函数捕获后 `eval` 执行）

Bubbletea 程序创建时指定输出到 stderr：

```go
p := tea.NewProgram(model, tea.WithOutput(os.Stderr))
```
Shell 包装函数（由 `try install` 写入配置文件）负责 `eval` stdout 输出的脚本，实现 `cd` 等父进程操作。

### Elm Architecture（Model/Update/View）

Bubbletea 基于 Elm Architecture，数据单向流动：

```
用户输入 → Msg → Update(model, msg) → 新 model → View(model) → 渲染
```

- **Model**：所有应用状态集中在一个结构体中
- **Update**：处理消息（按键、窗口大小变化等），返回新状态和可选命令
- **View**：纯函数，根据当前状态返回 `tea.View`（包含渲染内容和终端选项）

## 模块关系

```
CLI 解析 (internal/cli/)
    │
    ├──→ Shell 集成 (internal/shell/)
    │       输出包装函数到 stdout
    │
    ├──→ Git 集成 (internal/git/)
    │       URI 解析 + 目录命名
    │
    └──→ Bubbletea Program
            │
            ├──→ SelectorModel (internal/selector/)
            │       │
            │       ├──→ 目录加载 + 评分
            │       │       基础分 = 时间权重 + 日期后缀加成
            │       │
            │       ├──→ 模糊匹配 (internal/fuzzy/)
            │       │       子序列匹配 + 多维评分 + top-k 排序
            │       │
            │       ├──→ Lipgloss 样式渲染
            │       │       View() 返回 tea.View
            │       │
            │       └──→ 对话框子模型 (internal/dialog/)
            │               删除/重命名/ship
            │
            └──→ 结果 → 操作执行 (internal/script/)
                    Go 直接执行副作用 + cd 脚本 → stdout
```

## 数据流

1. 用户在 Shell 中执行 `try query`
2. 包装函数调用 `try exec query 2>/dev/tty`
3. CLI 解析参数，判断命令类型
4. 交互式命令启动 Bubbletea Program（输出到 stderr）
5. SelectorModel 加载目录、计算评分、通过 View() 渲染
6. 用户交互后产生 SelectionResult（`:cd`、`:mkdir`、`:delete`、`:rename`、`:ship`）
7. Bubbletea 退出，Go 直接执行副作用操作（mkdir/rm/mv/git clone）
8. 仅输出 `cd '/path'` 脚本到 stdout
9. 包装函数 `eval` 执行 cd 脚本（切换父 Shell 工作目录）

## 状态管理

SelectorModel 维护以下状态：

| 状态 | 类型 | 说明 |
|------|------|------|
| `textInput` | textinput.Model | 搜索输入（Bubbles 组件，内部管理 buffer、光标位置、闪烁等） |
| `list` | list.Model | 列表组件（Bubbles 组件，内部管理光标位置、滚动、分页等） |
| `deleteMode` | bool | 是否处于删除模式 |
| `deleteStatus` | string | 删除操作后的状态消息（"Deleted: xxx" 等），显示一次后清除 |
| `markedForDeletion` | map[string]bool | 标记删除的路径集合 |
| `allTries` | []Entry | 目录缓存（懒加载） |
| `cachedResults` | []MatchedEntry | 匹配结果缓存 |
| `lastQuery` | string | 上次查询（缓存键） |
| `selected` | *SelectionResult | 最终选择结果 |
| `width`, `height` | int | 终端尺寸 |
| `activeDialog` | Dialog | 当前活跃的对话框子模型（nil 表示主界面） |

## 终端控制（Bubbletea 内置，无需手动实现）

以下均由 Bubbletea 框架自动处理：
- Alt screen buffer 管理：在 `View()` 返回的 `tea.View` 中设置 `AltScreen = true`
- 终端大小变化检测：自动发送 `tea.WindowSizeMsg`
- 单次 flush 渲染（无闪烁）
- 颜色降采样
- 输出到 stderr：`tea.WithOutput(os.Stderr)`（Program 选项）

## 配置文件

`~/.config/try/config.json`，JSON 格式。详见 `config.md`。

| key | 说明 | 默认值 |
|-----|------|--------|
| `path` | tries 根目录 | `~/src/tries` |
| `ship` | ship 目标目录 | `~/src/ship` |
| `theme` | 配色主题 | `auto` |
| `locale` | 界面语言 | `auto` |

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `TRY_PATH` | tries 根目录（优先于配置文件） | `~/src/tries` |
| `TRY_PROJECTS` | ship 目标目录（优先于配置文件） | `~/src/ship` |
| `TRY_THEME` | 配色主题（`dark` / `light`，优先于配置文件） | `auto` |
| `TRY_LOCALE` | 界面语言（`en` / `zh`，优先于配置文件） | `auto` |
| `NO_COLOR` | 非空时禁用颜色 | — |
| `TRY_WIDTH` | 覆盖终端宽度（测试用） | — |
| `TRY_HEIGHT` | 覆盖终端高度（测试用） | — |

## 工具函数

共用工具函数分布在各自包内：

`internal/selector/entry.go`（导出，供其他包使用）：

```go
func DirExists(path string) bool
func FileExists(path string) bool
func IsFile(path string) bool
func EnvInt(key string) int
func FormatTimeAgo(d time.Duration) string
```

`internal/config/config.go`（导出）：

```go
func ExpandPath(s string) string  // 展开 ~ 为用户 home 目录
```

`internal/git/uri.go`（内部）：

```go
func dirExists(path string) bool  // git 包内部使用
```

### getParentProcessName（Shell 检测用）

`internal/shell/detect.go` 中的内部函数。优先读取 `/proc/<ppid>/comm`（Linux），回退到 `ps` 命令（macOS/BSD）。

## 版本号

两处同步：`cli.version` 变量（通过 `go build -ldflags "-X github.com/xleine/try/internal/cli.version=..."` 注入，未注入时回退到硬编码的 `"dev"`）、Git tag。

## 分发渠道

| 渠道 | 方式 |
|------|------|
| GitHub Releases | 预编译二进制（linux/darwin/windows × amd64/arm64） |
| Nix | `flake.nix`（packages.default + Home Manager module） |
| Homebrew | `Formula/try.rb` |
| `go install` | `go install github.com/xleine/try/cmd/try@latest` |
