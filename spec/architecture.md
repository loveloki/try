# 总体架构

## 项目定位

try 是一个临时实验目录管理工具。所有实验集中存放在单一根目录（默认 `~/src/tries`），用日期后缀自动命名（`name-YYYY-MM-DD`），通过模糊搜索和时间权重快速定位。名称在前可提高模糊匹配效率。

## 技术栈

- **语言**：Go
- **CLI / TUI 产物**：`try` 编译为 `CGO_ENABLED=0` 单一二进制，零运行时依赖
- **TUI 框架**：Bubbletea v2（Elm Architecture：Model/Update/View）
- **样式**：Lipgloss v2
- **组件**：Bubbles v2（list、textinput 等）
- **GUI 框架**：Fyne v2（原生桌面窗口、系统托盘、键鼠输入）

```
go get charm.land/bubbletea/v2
go get charm.land/bubbles/v2
go get charm.land/lipgloss/v2
```

## 项目结构

```
cmd/try/main.go        # TUI/CLI 入口：调用 cli.Run(os.Args[1:])
cmd/try-gui/main.go    # GUI 入口：创建原生桌面窗口
cmd/try-gui/FyneApp.toml  # fyne package 元数据（App ID / Name / Icon）
cmd/try-gui/Icon.png   # GUI 打包与桌面图标源
install.sh             # Linux/macOS 安装适配器（CLI + 官方 GUI 包）
install.ps1            # Windows 安装适配器
scripts/package-gui.sh # CI：fyne package → try-gui_* 官方包
internal/
  cli/                 # CLI 解析与命令分派
    cli.go             # Run 主入口、parseGlobalFlags、runSelector、帮助文本
    commands.go        # cmdExec、cmdClone、cmdWorktree、handleDot、worktreePath
    flags.go           # 参数提取工具函数（hasFlag、extractPath、extractValueFlag 等）
  config/              # 配置文件加载（~/.config/try/config.json）
    config.go          # Config 结构、LoadConfig、ResolvePaths、DetectTheme、ResolveLocale
  docx/                # .docx ZIP 打包/解压（无 OOXML 业务逻辑）
    zip.go             # Unpack / Pack、时间戳避重名、zip-slip 防护
  gui/                 # 跨平台原生 GUI（Fyne）
    app.go             # Run：加载配置、创建窗口、托盘、生命周期
    actions.go         # 键鼠动作与对话框流程
    actions_clone.go   # 克隆对话框与异步执行
    actions_docx.go    # 工具栏打包/解压 .docx
    actions_mouse.go   # 单击选中、双击按行索引打开、文件管理器揭示
    view.go            # Selector / Files 视图切换与刷新
    view_selector.go   # 选择器布局
    view_files.go      # 文件视图布局、工具栏、面包屑、drop overlay
    service.go         # entries/files 操作与副作用分派（含 touchDir）
    service_import.go  # 拖拽复制 copyDroppedFiles + 进度回调
    drop.go            # Window.SetOnDropped 注册与 overlay/进度反馈
    watch.go           # FilesPath fsnotify 防抖刷新
    inset.go           # 内容区 16px 水平内边距
    format.go          # 相对时间与文件大小格式化
    browser.go         # 系统默认打开文件 / 文件管理器揭示目录
    context_menu.go   # 文件右键菜单构建与操作分发
    app_discovery.go  # 可用应用发现：内置列表 + 自定义应用解析 + PATH 可执行名扫描
    dto.go             # GUI 视图模型
    paths.go           # 路径沙箱（tries/ships 子树）
    chrome.go          # 窗口 chrome 接口与尺寸常量
    chrome_darwin.go   # macOS 系统标题栏策略
    chrome_desktop.go  # Windows/Linux 无边框 + 自绘标题栏
    chrome_other.go    # 其他平台回退
    titlebar.go        # WinUI3 风格自绘标题栏
    window_root.go     # 最小窗口尺寸约束
    ewmh.go            # Linux EWMH 动作码与 atom 名
    native_windows.go  # Windows Minimize/Maximize/Drag
    native_linux.go    # Linux X11 CGO 原生操作
    native_linux_stub.go # Linux CGO=0 stub
    actions_helpers.go # 动作辅助函数
    actions_mark.go    # 文件复选框 / Ctrl+点击多选
    nav_list.go        # 列表键盘导航
    search_entry.go    # 搜索框 Tab 拦截
    status_bar.go      # 底栏：左计数 / 右键帽快捷键
    settings.go        # 设置页（全屏视图：主题/语言/布局骨架）
    settings_openwith.go # 设置页 Open With 映射编辑
    widgets.go         # 行组件（Hoverable / Mouseable）
    widgets_source_tab.go # 来源 Tab hover 对比度
    widgets_row_state.go # 行背景优先级：marked > selected > hover
    widgets_render.go  # 匹配高亮与文件 meta
    theme.go           # GUI 主题 token 映射
    shortcuts.go       # Ctrl 快捷键包装
    update.go          # 纯函数：来源循环 / Enter 决策
  selector/            # 交互式选择器（Bubbletea Model + Bubbles list）
    model.go           # SelectorModel：结构体定义、New、Init、Update、View
    keyhandler.go      # 按键处理函数（handleKey、handleEnter、toggleDelete、循环导航等）
    layout.go          # Header / Footer / 空状态布局
    styles.go          # 主题 token 与 Lipgloss 样式
    icons.go           # 列表与空状态图标常量
    delegate.go        # 自定义 list.ItemDelegate 渲染
    view.go            # View 入口（主界面 / 对话框叠层）
    keys.go            # 按键绑定
    entry.go           # 目录条目类型定义（实现 list.Item 接口）+ 工具函数
    loader.go          # 目录加载（loadAllTries）和列表刷新（refreshList）
    catalog.go         # 包级扫描/匹配 API（LoadAllEntries、MatchEntries、SourceCounts）
    dialogs.go         # DialogInstance 接口、DialogFactory、对话框路由
    overlay.go         # 模态弹窗合成
    testkeys.go        # 测试按键解析（ParseTestKeys、KeyToMsg）
  dialog/              # 对话框子模型
    dialog.go          # Dialog 接口定义
    delete.go          # 删除确认对话框
    rename.go          # 重命名对话框
    ship.go            # ship 对话框
    styles.go          # 对话框样式（从 selector Styles 派生）
    icons.go           # 对话框图标常量
    modal.go           # 弹窗尺寸与盒子渲染
  fuzzy/               # 模糊匹配引擎
    fuzzy.go           # 子序列匹配 + 多维评分
  shell/               # Shell 集成
    detect.go          # Shell 类型检测
    install.go         # install 命令：写入 Shell 配置文件
    template.go        # 包装函数模板生成
  script/              # 操作执行 + cd 脚本生成
    exec.go            # Go 直接执行副作用操作（mkdir/rm/mv/git）
    actions.go         # ExecuteSideEffect：GUI 用，不写 cd 到 stdout
    script.go          # cd 脚本输出（Quote + EmitCd）
  git/                 # Git 集成
    uri.go             # Git URI 解析、IsGitURI、GenerateCloneDirName、ResolveUniqueName
go.mod
go.sum
```

## 设计原则

### 单二进制零依赖

`try` 编译为静态链接的单一二进制文件，用户无需安装 Go 运行时。`try-gui` 编译为原生桌面二进制，使用 Fyne 和系统图形栈，发布构建启用 CGO 并在目标平台 runner 上完成。

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

GUI 数据流：

```
GUI 入口 (cmd/try-gui)
    │
    └──→ Fyne App / 主窗口 / 系统托盘 (internal/gui/)
            │
            ├──→ Selector / Files 状态机
            │       │
            │       ├──→ 目录加载与来源过滤 (internal/selector/catalog.go)
            │       ├──→ 模糊匹配 (internal/fuzzy/)
            │       ├──→ 副作用操作 (internal/script.ExecuteSideEffect)
            │       ├──→ 路径沙箱 (internal/gui/paths.go)
            │       └──→ 用户可见文案 (internal/i18n/)
            │
            └──→ 关闭窗口隐藏到托盘，托盘 Quit 退出进程
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
| `locale` | 界面语言 | `auto` |

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `TRY_PATH` | tries 根目录（优先于配置文件） | `~/src/tries` |
| `TRY_PROJECTS` | ship 目标目录（优先于配置文件） | `~/src/ship` |
| `TRY_LOCALE` | 界面语言（`en` / `zh`，优先于配置文件） | `auto` |
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

两处同步：`cli.version` 变量（通过 `go build -ldflags "-X github.com/loveloki/try/internal/cli.version=..."` 注入，未注入时回退到硬编码的 `"dev"`）、Git tag。

## 分发渠道

| 渠道 | 方式 |
|------|------|
| GitHub Releases | 预编译二进制（linux/darwin/windows × amd64/arm64） |
| Nix | `flake.nix`（packages.default + Home Manager module） |
| Homebrew | `Formula/try.rb` |
| `go install` | `go install github.com/loveloki/try/cmd/try@latest` |

`try-gui` 使用 Fyne 原生窗口。GitHub Actions（`.github/workflows/ci.yml` 与 `release.yml`）仅在推送 `v*` tag 时运行：CI 做 lint/test/分平台构建，Release 打包归档；推送 `main` 或 PR 不触发。Release 同时发布 CLI 裸归档 `try_*` 与 `fyne package` 官方 GUI 包 `try-gui_*`（见 `spec/distribution.md`）。`install.sh` / `install.ps1` 作为适配器默认安装二者（`TRY_INSTALL_GUI=0` 可跳过 GUI）。Linux arm64 仅含 CLI。`.goreleaser.yaml` 保留供本地 CLI snapshot。
