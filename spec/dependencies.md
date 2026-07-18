# Go 依赖清单

## 直接依赖（go.mod require）

| 模块 | 版本 | 用途 |
|------|------|------|
| `charm.land/bubbletea/v2` | v2.0.6 | TUI 框架核心。提供 Elm Architecture 运行时（Model/Update/View）、终端管理（alt screen、鼠标、光标）、消息循环、Program 生命周期。所有交互式界面的基础。 |
| `charm.land/bubbles/v2` | v2.1.0 | TUI 组件库。使用的子包：`list`（列表管理：光标追踪、滚动、分页、ItemDelegate）、`textinput`（搜索输入框和对话框输入：光标位置、Emacs 快捷键、Placeholder）、`key`（按键绑定定义与匹配）。 |
| `charm.land/lipgloss/v2` | v2.0.2 | 声明式终端样式。用于所有 TUI 渲染：文字颜色/背景色/粗体、ANSI-aware 宽度计算（`Width()`）、文本截断（`Truncate()`/`TruncateLeft()`）、水平对齐（`PlaceHorizontal()`）。NO_COLOR 支持通过 color profile 实现。 |
| `fyne.io/fyne/v2` | v2.8.0 | GUI 桌面框架。用于 `try-gui` 原生窗口、系统托盘、窗口内容绘制、键鼠输入、对话框与主题。 |
| `github.com/charmbracelet/x/ansi` | v0.11.7 | ANSI 文本处理。由 selector 渲染测试直接使用，并由 Charm 套件复用。 |
| `github.com/sahilm/fuzzy` | v0.1.2 | 子序列匹配库。仅使用匹配位置与基础结果，自定义时间权重、日期后缀加成和排序。 |
| `github.com/jeandeaual/go-locale` | v0.0.0-20250612000132-0ef82f21eade | OS 语言检测。`locale: auto` 且无 `LANG`/`LC_*` 时（GUI 从 Dock/开始菜单启动）回退系统语言。 |
| `github.com/fsnotify/fsnotify` | v1.9.0 | Files 视图监听当前目录变更并防抖刷新列表。 |

## 使用的 Bubbles 子包

| 子包 | import 路径 | 用途 |
|------|------------|------|
| list | `charm.land/bubbles/v2/list` | 选择器主列表组件。管理光标位置（`Index()`/`Select()`/`CursorUp()`/`CursorDown()`）、可见区域滚动、item 渲染委托（`ItemDelegate`）。禁用内置 UI（标题、状态栏、帮助、分页、过滤），仅用其列表管理能力。 |
| textinput | `charm.land/bubbles/v2/textinput` | 搜索框和对话框输入。内置光标闪烁、Emacs 快捷键（Ctrl-A/E/B/F/K/W）、Placeholder、字符过滤（`Validate` 回调）。选择器搜索框 + 删除确认 + 重命名 + ship 对话框各一个实例。 |
| key | `charm.land/bubbles/v2/key` | 按键绑定工具。`key.NewBinding()` 定义自定义快捷键，`key.Matches()` 在 Update 中匹配按键消息。用于 Enter、Ctrl-P/N/D/T/R/G/C、ESC 等自定义按键的声明和匹配。 |

## 标准库使用

| 包 | 用途 |
|----|------|
| `os` | 文件系统操作（`ReadDir`、`Stat`、`MkdirAll`）、环境变量（`TRY_PATH`、`TRY_PROJECTS`、`NO_COLOR`）、进程信息（`Executable()`、`Getwd()`） |
| `path/filepath` | 路径操作（`Join`、`Dir`、`Base`、`EvalSymlinks`、`Abs`） |
| `strings` | 字符串处理（`HasPrefix`、`Contains`、`TrimSpace`、`ReplaceAll`、`Repeat`） |
| `math` | 评分公式中的 `Sqrt()`（时间衰减、邻近加成） |
| `regexp` | 日期后缀匹配（`/-\d{4}-\d{2}-\d{2}$/`）、Git URI 解析（4种 URL 格式）、字符过滤 |
| `container/heap` | 模糊匹配 top-k 部分排序（O(n log k)，优于全排序） |
| `time` | 时间戳处理（目录 mtime 计算、日期后缀生成 `2006-01-02` 格式） |
| `fmt` | 脚本输出到 stdout（`Println`/`Print`）、错误消息到 stderr |
| `io` | `ItemDelegate.Render()` 的 `io.Writer` 参数 |
| `strconv` | 目录名版本化中的数字递增（`Atoi`/`Itoa`） |
| `sort` | 备用排序（当结果数 <= limit 时用 `sort.Slice` 替代 heap） |

## 间接依赖（由 Charm 套件自动引入）

以下依赖由 `go mod tidy` 自动管理，不需要手动声明。列出以便了解依赖树：

| 模块 | 来源 | 用途 |
|------|------|------|
| `github.com/charmbracelet/x/ansi` | lipgloss, bubbles | ANSI 转义码解析和处理 |
| `github.com/charmbracelet/x/term` | bubbletea | 终端能力检测（尺寸、颜色支持） |
| `github.com/charmbracelet/x/termios` | bubbletea | 终端 I/O 控制 |
| `github.com/charmbracelet/x/windows` | bubbletea | Windows 终端兼容 |
| `github.com/charmbracelet/colorprofile` | bubbletea | 终端颜色降采样（TrueColor → 256 → 16 → 无色） |
| `github.com/charmbracelet/ultraviolet` | bubbletea | 终端渲染引擎 |
| `github.com/mattn/go-runewidth` | lipgloss, bubbles | Unicode 字符显示宽度计算（CJK、emoji 等） |
| `github.com/muesli/cancelreader` | bubbletea | 可取消的 stdin 读取（Program 生命周期管理） |
| `github.com/atotto/clipboard` | bubbles | 剪贴板操作支持 |
| `github.com/clipperhouse/displaywidth` | bubbles | 显示宽度计算 |
| `github.com/clipperhouse/uax29/v2` | bubbles | Unicode 分词算法 |
| `github.com/lucasb-eyer/go-colorful` | lipgloss | 颜色空间转换 |
| `github.com/rivo/uniseg` | bubbles | Unicode 分词（文本光标定位） |
| `github.com/xo/terminfo` | bubbletea | 终端信息查询 |
| `golang.org/x/sync` | bubbletea | 并发同步原语 |
| `golang.org/x/sys` | bubbletea, term | 系统调用接口（终端控制等） |

## go.mod 示例

```go
module github.com/loveloki/try

go 1.26.3

require (
    charm.land/bubbles/v2   v2.1.0
    charm.land/bubbletea/v2 v2.0.6
    charm.land/lipgloss/v2  v2.0.2
    fyne.io/fyne/v2           v2.8.0
    github.com/charmbracelet/x/ansi v0.11.7
    github.com/sahilm/fuzzy   v0.1.2
)
```

> 实际 `go.sum` 和间接依赖由 `go mod tidy` 生成。Fyne 引入桌面图形、字体、SVG、托盘与平台窗口相关间接依赖，`try-gui` 发布构建使用 CGO。

## try-gui 平台构建依赖

CI（`.github/workflows/ci.yml`，仅 `v*` tag）在 ubuntu / macos / windows 三平台 runner 上原生编译，不交叉编译。

| 平台 | CGO | 额外依赖 |
|------|-----|----------|
| Linux | `CGO_ENABLED=1` | `gcc`、`libgl1-mesa-dev`、`xorg-dev`、`libxcursor-dev`、`libxrandr-dev`、`libxinerama-dev`、`libxi-dev`、`libxxf86vm-dev`、`libwayland-dev`、`libxkbcommon-dev`。go-gl/glfw v3.4 默认同时编译 X11 与 Wayland 后端；X11 头文件亦供 `native_linux.go` 自绘标题栏拖拽/最大化 |
| macOS | `CGO_ENABLED=1` | Xcode Command Line Tools（GitHub `macos-latest` 已具备） |
| Windows | `CGO_ENABLED=1` | MinGW（CI 用 `choco install mingw`） |

`try` CLI 始终 `CGO_ENABLED=0`。

Release（`release.yml`）在 `gui: true` 的矩阵行额外安装 `fyne.io/tools/cmd/fyne`，运行 `scripts/package-gui.sh` 产出官方 GUI 包（见 `spec/distribution.md`）。该 CLI 为构建时工具，不进入 `go.mod` 运行时依赖。

## 不使用的库（设计决策）

| 类别 | 未选方案 | 原因 |
|------|---------|------|
| CLI 框架 | cobra, urfave/cli | 参数结构简单，手工解析 `os.Args` 更轻量且可控 |
| 模糊匹配评分 | sahilm/fuzzy 内置评分 | 库的内置评分不支持时间权重和日期后缀加成，仅使用其匹配功能，评分自定义 |
| 日志 | slog, zerolog | TUI 程序 stderr 用于渲染，不适合混合日志输出 |
| 测试框架 | testify, gomega | 标准 `testing` 包已足够，减少依赖 |
| 颜色 | fatih/color, aurora | Lipgloss 已提供完整的样式能力 |
| WebView 框架 | Wails, webview | GUI 主界面采用 Fyne 原生窗口与 Go 组件 |
