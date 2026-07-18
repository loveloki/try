# try GUI 方案设计

## 1. 目标与范围

### 1.1 交付目标

`try-gui` 是 try 的跨平台原生桌面入口。启动后显示一个原生应用窗口，默认尺寸约 900 × 600 px，最小尺寸约 720 × 480 px，初始位置居中。主界面进入目录选择器，列出配置中的 tries 根目录及各 ships 目录下的实验目录，支持实时模糊搜索、来源过滤与键盘驱动操作。

选中目录后切换到文件视图，支持面包屑浏览、进入子目录、删除文件或目录、用系统默认程序打开文件，以及返回选择器。GUI 与 TUI 共享配置、目录扫描、模糊匹配、脚本副作用、路径沙箱和 i18n 文案。

### 1.2 交付形态

交付物是 Go 二进制 `try-gui`。程序启动 Fyne 桌面应用并创建原生窗口，不依赖 Node.js、前端构建产物或本地网络服务。

- macOS 使用系统窗口装饰，关闭、最小化、最大化控件位于左上角，呈系统红绿灯风格。
- Windows 使用无边框 Fyne 窗口和应用内标题栏，标题栏控件按 WinUI3 风格位于右上角。
- Linux 使用无边框 Fyne 窗口和应用内标题栏，标题栏控件与 Windows 一致位于右上角。
- 标题栏与窗口装饰之外，布局、配色、组件、快捷键、对话框、Toast 与文件视图行为跨平台一致。
- 系统托盘提供显示窗口和退出应用的入口；关闭窗口默认隐藏到托盘，托盘 Quit 退出进程。

仓库根目录下的 `try-gui/` 是 UI/UX 设计参考区域，仅用于对齐布局、配色与交互原型。实现以本规格、`spec/gui-uiux.md` 和 TUI 设计 token 为准。

### 1.3 范围外内容

- **Shell 集成与 `cd` 语义**：GUI 是独立进程，副作用通过 `internal/script` 直接执行，不调用 `EmitCd`，也不向父 Shell 输出可 eval 的脚本。
- **clone / worktree / install**：这些操作由 TUI / CLI 承担。
- **多窗口**：同一会话内一个选择器与一个文件视图，切换项目时返回选择器。
- **富文本 docx 编辑**：docx 处理为后续 Phase。

### 1.4 与 TUI 功能对照

| TUI 功能 | GUI 处理方式 | 说明 |
|---|---|---|
| 目录选择器 + 模糊搜索 | 复用 | 评分与排序由 `internal/fuzzy` 提供 |
| 创建新目录 (Ctrl-T) | 复用 | 选择器底部行内输入 |
| 删除目录 (Ctrl-D / Space) | 复用 | 选择器 DeleteMode；文件视图另有删除文件 |
| 重命名目录 (Ctrl-R) | 复用 | 快捷键触发行内输入 |
| Ship 目录 (Ctrl-G) | 复用 | 复用 ship 副作用；ships 目录作来源过滤项 |
| 来源过滤 (Tab) | 复用 | all / tries / 各 ship 目录 basename |
| 进入文件视图 (Enter) | GUI 专属 | 文件视图替代 Shell `cd` |
| clone / worktree / install | 范围外 | 继续由 TUI / CLI 提供 |

## 2. 技术选型

### 2.1 运行时栈

| 层级 | 技术 | 说明 |
|---|---|---|
| 进程入口 | `cmd/try-gui` | 解析 `-path`，调用 `gui.Run` |
| 桌面框架 | Fyne v2 | 创建跨平台原生窗口、绘制内容区、接收键鼠输入 |
| 系统托盘 | Fyne desktop API | 设置托盘图标与 Show / Quit 菜单 |
| 平台标题栏 | Fyne 窗口能力 + build tags | macOS 系统装饰；Windows/Linux 应用内 WinUI3 风格标题栏 |
| 业务逻辑 | 现有 `internal/` 包 | 配置、扫描、模糊、脚本副作用、路径沙箱、i18n |
| 设计参考 | `try-gui/` | 不参与构建、测试或发布 |

### 2.2 候选方案结论

| 方案 | 结论 | 原因 |
|---|---|---|
| Fyne v2 | 采用 | 原生窗口、成熟桌面组件、内置系统托盘、Go 业务层可直接复用 |
| Gio | 备选 | 适合自绘复杂 UI，但系统托盘和完整组件需要更多自研 |
| Wails / WebView | 排除 | 主界面依赖 WebView |
| Qt Go binding | 排除 | Go 绑定维护和发布链复杂度高 |

### 2.3 构建与分发约束

`try` CLI 保持纯 Go 构建。`try-gui` 使用 Fyne 后需要 CGO 与平台图形依赖，发布链按平台原生构建：

1. CI 为 `try-gui` 设置 `CGO_ENABLED=1`。
2. macOS 产物在 macOS runner 构建，满足 AppKit / Metal / OpenGL 相关依赖。
3. Windows 产物在 Windows runner 构建，并使用 `-ldflags -H=windowsgui` 避免启动控制台窗口。
4. Linux 产物在 Linux runner 构建，安装 Fyne 所需图形开发包。
5. `.goreleaser.yaml` 为 `try` 与 `try-gui` 分开声明构建策略；`try-gui` 构建 flags 保留 `-trimpath` 和平台链接参数。

## 3. 模块划分

```
cmd/try-gui/                 # GUI 入口：解析参数、调用 gui.Run
internal/
  config/                    # 配置加载、路径/locale/主题解析
  selector/
    catalog.go               # GUI 使用的聚合扫描与过滤接口
  fuzzy/                     # 模糊匹配
  script/                    # mkdir/rm/mv/ship 等副作用，不 EmitCd
  i18n/                      # 用户可见文案
  gui/
    app.go                   # Run：加载配置、创建窗口、托盘、生命周期
    state.go                 # GUI 状态模型
    update.go                # 输入消息到状态变更
    service.go               # 业务操作：entries/files/副作用/路径校验
    service_import.go        # 拖拽复制：copyDroppedFiles
    drop.go                  # Window.SetOnDropped 注册与 overlay
    format.go                # 相对时间与文件大小格式化
    view_files.go            # 文件视图布局、工具栏、面包屑
    view_selector.go         # 选择器布局
    nav_list.go / search_entry.go  # Tab/Enter 键盘拦截
    status_bar.go            # 底栏左计数 / 右键帽
    widgets.go / widgets_source_tab.go / widgets_render.go / widgets_row_state.go # 行组件、来源 Tab、悬停/选中
    inset.go / actions_mark.go # 水平留白；文件复选框与 Ctrl+点击多选
    chrome.go               # 窗口 chrome 接口与尺寸常量
    chrome_darwin.go        # macOS 系统标题栏策略
    chrome_desktop.go       # Windows/Linux 无边框 + 自绘标题栏
    chrome_other.go         # 其他平台回退系统装饰
    titlebar.go             # WinUI3 风格自绘标题栏 widget
    window_root.go          # 最小窗口尺寸约束
    ewmh.go                 # Linux EWMH 动作码与 atom 名常量
    native_windows.go       # Windows Minimize/Maximize/Drag
    native_linux.go         # Linux X11 CGO 原生操作
    native_linux_stub.go    # Linux CGO=0 编译 stub
    actions.go              # 键鼠动作与对话框流程
    actions_mouse.go        # 按行索引打开与文件管理器揭示
    actions_helpers.go      # 动作辅助函数
    browser.go              # 打开文件 / 揭示目录
    theme.go                # GUI palette，与 TUI token 对齐
try-gui/                     # UI/UX 设计参考
```

### 核心模块职责

| 模块 | 职责 |
|---|---|
| `cmd/try-gui` | 解析命令行参数，调用 `gui.Run` |
| `internal/gui/app.go` | 加载配置、初始化 i18n、创建窗口、设置托盘、协调退出 |
| `internal/gui/state.go` | 持有 Selector、Files、对话框、Toast、主题和选择状态 |
| `internal/gui/update.go` | 处理键盘、鼠标和异步操作结果，返回新状态 |
| `internal/gui/service.go` | 调用 `selector.catalog`、`script.ExecuteSideEffect`、文件系统和 opener |
| `internal/gui/paths.go` | 校验读路径与可变更目标路径 |
| `internal/gui/view_*.go` | 用 Fyne 组件绘制跨平台一致的内容区 |
| `internal/gui/chrome_*.go` | 实现平台标题栏与窗口控制策略 |
| `internal/gui/theme.go` | 将 dark/light hex token 映射为 Fyne 主题 |

## 4. 业务复用

| 能力 | 包 |
|---|---|
| 配置读取、路径/locale/主题解析 | `internal/config` |
| 目录扫描、来源标签、评分聚合 | `internal/selector/catalog.go` |
| 模糊匹配 | `internal/fuzzy` |
| 创建 / 删除 / 重命名 / ship | `internal/script.ExecuteSideEffect` |
| 路径沙箱 | `internal/gui.IsAllowedPath` / `internal/gui.IsAllowedTarget` |
| 用户可见文案 | `internal/i18n.Messages` |

GUI 不驱动 Bubbletea Model。Selector 与 Files 视图通过进程内 service 调用共享业务包。所有用户可见文本从 `i18n.Messages` 获取，新增 GUI 文案同时补齐 EN / ZH。

## 5. 核心数据流

```
启动 try-gui
  │
  ▼
LoadConfig / ResolvePaths / ResolveLocale / DetectTheme / ensureGUIDirs
  │
  ▼
初始化 Fyne App、主窗口、托盘、主题和 AppState
  │
  ▼
Selector 视图
  ├── 搜索：LoadAllEntries → MatchEntries
  ├── 来源 Tab：SourceOptions → SourceCounts
  ├── Ctrl-T / Ctrl-R / Ctrl-G / Ctrl-D：service → ExecuteSideEffect
  └── Enter：切换到 Files 视图
  │
  ▼
Files 视图
  ├── 面包屑与列表：service.ListFiles
  ├── 删除确认：IsAllowedTarget → os.RemoveAll
  ├── 打开文件：opener.Open
  └── Esc：返回上级或 Selector
  │
  ▼
关闭窗口隐藏到托盘；托盘 Quit 退出进程
```

## 6. 进程内接口

### 6.1 Service 方法

```go
type Service struct {
    triesPath string
    shipPaths []string
}

func (s *Service) ListEntries(query, source string) EntriesResult
func (s *Service) CreateEntry(name string) (string, error)
func (s *Service) DeleteEntries(paths []string) error
func (s *Service) RenameEntry(path, newName string) (string, error)
func (s *Service) ShipEntry(path string, destIndex int) (string, error)
func (s *Service) ListFiles(path string) ([]FileEntry, error)
func (s *Service) DeleteFiles(paths []string) error
func (s *Service) OpenFile(path string) error
```

### 6.2 View Model

- `EntryView`：`ID`、`Name`、`BaseName`、`Date`、`Source`、`Score`、`LastModified`、`Path`、`Highlights`。
- `FileEntry`：`ID`、`Name`、`Type`、`SizeKB`、`Modified`、`IsDir`、`Path`。
- `EntriesResult`：`Entries` 与 `Counts`。

### 6.3 路径安全

只读文件列表允许读取 tries 根与 ships 根自身。删除、重命名、ship 和文件删除必须严格位于 tries 或 ships 子树内，根目录自身不可作为可变更目标。路径入参统一执行 `filepath.Clean`、`filepath.Abs` 和根目录前缀校验。

## 7. 关键实现点

### 7.1 标题栏与窗口控制

macOS 走系统窗口装饰（`app.NewWindow`），内容区不经额外包装。

Windows 与 Linux 使用 `desktop.Driver.CreateSplashWindow()` 创建无边框窗口（创建后立即 `SetTitle`，避免空标题显示为 "Fyne Application"），内容顶部自绘 44px WinUI3 风格标题栏（不使用 `container.InnerWindow`）：左侧产品标题（`i18n.GUITitle`），右侧控件顺序为最小化、最大化/还原、关闭。关闭按钮触发 `SetCloseIntercept` 隐藏到托盘；最小化通过 `RunNative` 调用系统 Iconify（不使用 `Hide`）；最大化/还原通过 `RunNative` 调用系统 Maximize/Restore（不使用 `SetFullScreen`）。标题栏空白区 Primary MouseDown 触发系统级拖拽（Windows：`SC_MOVE|HTCAPTION`；Linux X11：`_NET_WM_MOVERESIZE`；Wayland：当前为空操作，待补）。

托盘初始化顺序：先 `SetSystemTrayWindow`，再 `SetCloseIntercept(hideToTray)`（托盘会覆盖 CloseIntercept）。托盘 Quit 调用 `app.Quit()`；Show 调用 `Show` + `RequestFocus`。无边框主窗口托盘左击 toggle 为 Fyne 默认行为。

### 7.2 主题与 token

GUI 使用与 `internal/selector/styles.go` 一致的 dark/light hex token。默认主题来自 `config.DetectTheme()`，窗口内提供 Dark / Light 手动切换。主题切换立即刷新 Selector、Files、对话框和 Toast。

### 7.3 Selector

Selector 支持实时搜索、来源 Tab、循环导航、Ctrl-T 新建、Ctrl-D/Space 删除标记、Ctrl-R 重命名、Ctrl-G Ship、Enter 进入 Files。匹配高亮使用 `selector.MatchEntries` 返回的位置。

### 7.4 Files

Files 支持面包屑、列表、进入子目录、打开文件、删除确认、Esc 返回，以及从系统文件管理器拖入复制到当前目录（见 `spec/gui-uiux.md` §5.6）。删除确认默认焦点为 Cancel。隐藏 `.` 开头的文件和目录。列表按目录优先、名称不区分大小写排序；修改时间显示为相对时间。

### 7.5 托盘

托盘菜单包含 Show 与 Quit。Show 显示并聚焦主窗口；Quit 释放托盘资源并退出进程。关闭窗口时窗口隐藏，进程保留在托盘。

## 8. 风险与验证

| 风险 | 影响 | 验证 |
|---|---|---|
| Fyne CGO 发布链复杂 | Release 需要平台原生 runner | CI 分平台构建 `try-gui` |
| Linux 标题栏与窗口管理器差异 | 无边框拖拽、最大化行为需适配 | Linux 桌面人工验收 |
| 系统托盘在不同 Linux 桌面环境表现不同 | 托盘入口不可见或菜单不完整 | GNOME/KDE 至少各一次 smoke |
| 快捷键被系统或输入法拦截 | 操作不可达 | 三平台键盘验收 |
| GUI 与 TUI 行为漂移 | 搜索、来源、删除语义不一致 | 共享 catalog/fuzzy/script 并补状态机测试 |

## 9. 实施里程碑

### Phase 1：规格与发布链准备

- [x] 修订 `spec/gui-plan.md`、`spec/gui-uiux.md`、`spec/architecture.md`、`README.md`
- [x] 引入 Fyne 依赖并更新 `.goreleaser.yaml` / CI 构建矩阵（CI 在 ubuntu / macos / windows 三 runner 原生编译 `try` 与 `try-gui`）
- [x] 明确 Linux 图形依赖安装步骤（见 `.github/workflows/ci.yml` 与 `spec/dependencies.md`）

### Phase 2：原生窗口骨架

- [x] 重写 `cmd/try-gui` 与 `internal/gui/app.go`
- [x] 实现 900 × 600 默认窗口、720 × 480 最小窗口、居中
- [x] 实现 macOS 系统标题栏与 Windows/Linux WinUI3 风格自绘标题栏（`RunNative` 窗口控制）
- [x] 实现系统托盘 Show / Quit

### Phase 3：Selector MVP

- [x] 实现搜索、来源 Tab、列表、匹配高亮、循环导航
- [x] 实现 Ctrl-T、Ctrl-D/Space、Ctrl-R、Ctrl-G 与确认对话框
- [x] 实现 Enter 进入 Files

### Phase 4：Files MVP

- [ ] 实现面包屑、文件列表、进入子目录、打开文件
- [ ] 实现文件删除确认与 Esc 返回
- [ ] 补齐路径沙箱、状态机、i18n、主题 token 测试

### Phase 5：验收与发布

- [ ] `go build ./...`
- [ ] `go vet ./...`
- [ ] `staticcheck ./...`
- [ ] `go test ./... -count=1`
- [ ] macOS / Windows / Linux 标题栏、托盘、快捷键、配色与路径沙箱人工验收
