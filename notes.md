---

## [2026-07-18] GitHub CI Release 缺 Wayland 头文件

- **本会话**：用户 `/architect` `/planner` `/multi-agent-breakthrough` 解决 GitHub CI 错误；截图为 release `build (ubuntu-latest, linux, amd64, true)` 编 glfw 时 `wayland-client-core.h: No such file or directory`，另有 `upload-artifact@v4` Node 20 deprecation 警告。
- **并行**：
  - [Architect](b1e5598b-0a79-4d1d-8ea9-c1a9c2934810)：Fyne→GLFW hybrid 依赖面；最小集 +`libwayland-dev` +`libxkbcommon-dev`。
  - [Planner](1c3d299d-d26d-421f-ba04-c107219596a1)：三处 workflow + `spec/dependencies.md`；artifact 升级单独 PR。
  - explore：`-tags x11` 可官方跳过 Wayland 编译（备选）；包映射确认；artifact 应升 v7/v8 而非 v5。
  - 对抗：装包方案通过；「一定编过」需 CI 绿构建闭环确认。
- **结论 / 决定**：
  - **主路径**：CI/Release Linux apt 追加 `libwayland-dev libxkbcommon-dev`，保持默认 X11+Wayland hybrid；不采用 `-tags x11` 作为发布默认。
  - Node/`upload-artifact` 警告**不阻塞**；另开变更升级，不与本修复混 commit。
- **澄清（用户追问）**：
  - **没有**关掉 Wayland：主路径是装包启用 hybrid；`-tags x11` 只是未采用的备选。新版 KDE（默认 Wayland）可用 GLFW Wayland 后端跑 try-gui；自绘标题栏拖拽/最小化/最大化在 Wayland 下仍为空操作（既有限制）。
- **本会话续**：用户要求清理无用 Node、修版本警告、提交并发版。
  - **【事实】** 产品/CI 不依赖 Node；警告来自 Actions 的 `upload-artifact@v4` 等仍声明 Node 20。`scripts/tui_test_common.sh` 仅用 `node -e` 作可选 JSON 解析（agent-tty 测试），非构建依赖。Next.js `try-gui/` 已在此前 commit 移除，文档残留已清。
  - 已升：`upload-artifact@v7`、`download-artifact@v8`、`softprops/action-gh-release@v3`。
- **相关**：`.github/workflows/ci.yml`、`release.yml`、`spec/dependencies.md`、`README.md`、`spec/gui-*.md`

---

## [2026-07-18] GUI 改 Rust vs 拖拽响应（多 agent 攻坚）

- **本会话**：用户 `/architect` `/planner` `/multi-agent-breakthrough` 问：GUI 如何改用 Rust；拖拽响应类问题能否因此解决。后续追问：Zed GPUI 能否「完美」实现。
- **并行**：
  - [Architect](9ea9b503-7b9e-4516-b550-be58b7ca68a5)：四架构（全量 Rust / Rust+Go IPC / FFI / Go 换框架）+ DnD 根因归类。
  - [Planner](f35d9846-5dd3-490d-a25a-d639c9079f8b)：决策树、验收条目、人周、充分/必要条件。
  - explore×4：egui/iced/slint DnD、Tauri/WebView、混合 IPC、留 Go 原生 hook。
  - 对抗：戳穿「0.5–1.5 人周 CGO」「语言级非充分/非必要」「gotk4 默认优先」等跳跃。
  - GPUI 专项：对照 `FileDropEvent` / 托盘 / 第三方成熟度（docs.rs + Zed discussions）。
- **结论 / 决定**（架构定案，未改代码）：
  - **【事实】** drag-enter overlay 缺口在 Fyne→GLFW 只暴露 drop，不在 Go 语言；松手后反馈已实现（§5.6）。
  - **【事实】** 换 Rust **不必然**解决：egui/iced API 有 hover，但 Wayland/winit 仍常缺；Tauri 技术上能做 enter 但与「无 WebView 主界面」硬冲突。
  - **【共识】** 仅为 drag-enter 全量迁 Rust / 同进程 Go FFI：**否决**。Rust UI+Go IPC 仅当「壳层战略迁 Rust」时合理。
  - **GPUI**：**不能称「完美」**。DnD 模型（`FileDropEvent::{Entered,Pending,Submit,Exited}` + `ExternalPaths`）在 API 面上优于 Fyne，是强候选；但上游托盘/关窗后台仍弱、为 Zed 优先演进、业务仍需 Go IPC、组件与发布链成本高。不因 DnD 单独选 GPUI。
  - **决策单位**是 toolkit+窗口后端，不是语言；「Rust 非必要」随 spike 结果可变为必要。
  - **下一步（需产品确认）**：D-10（悬停 overlay）是否 P0、Wayland 是否必测 → 再开单平台原生 DnD spike 或换栈评估；不在未确认前启动 Rust/GPUI 重写。
- **相关**：`spec/gui-uiux.md` §5.6、`internal/gui/drop.go`、`spec/gui-plan.md` §2.2；[gpui FileDropEvent](https://docs.rs/gpui/latest/gpui/enum.FileDropEvent.html)

---

## [2026-07-18] 一键安装支持 TUI + GUI

- **本会话**：用户要求 `install.sh` 同时安装 TUI 与 GUI。
- **结论 / 决定**：
  - `install.sh` 默认安装 `try` + `try-gui`；`TRY_INSTALL_GUI=0` 可跳过 GUI；归档缺 GUI 时警告并继续。
  - Release 改为分平台原生构建（非单 ubuntu goreleaser），归档名仍为 `try_<os>_<arch>.tar.gz`。
  - Linux arm64 仅发布 `try`（无 CGO 交叉编 GUI）。
- **相关**：`install.sh`、`.github/workflows/release.yml`、`README.md`、`spec/architecture.md`

---

## [2026-07-18] GUI 六项 UX（hover/排序/多选/上传/拖拽/留白）

- **本会话**：
  - 用户六项：tag hover 看不清、进目录后排序不更新、无多选、去无用上传、拖拽无反馈/无进度、四周无留白。
  - 并行 [Planner](b2299d79-2248-4feb-b3fb-5ce7aaab7729) / [Architect](1c65b3d0-0ce4-48a9-81f7-39ccc149489e) / explore×3 / 对抗×3 / [Loop-op](41079d37-a055-493a-afbb-a803c12425f4)。
  - 用户确认：tag hover 修背景/字色；文件列表复选框 + Ctrl+点击多选；接受无 drag-over，松手后 overlay+进度+完成 Toast。
- **结论 / 决定**（已落地，门禁绿）：
  - **#1**：自定义 `sourceTab` Hoverable；inactive hover → `surfaceHover` + `foreground`。
  - **#2**：`touchDir`（Chtimes）于进入 try 根 + `selectedPath` 重定位。
  - **#3**：行首 `widget.Check` + Ctrl/⌘+点击 toggle `fileMarked`。
  - **#4**：去掉工具栏上传；i18n 改为 `GUIDropImporting` / `GUIDropProgress`。
  - **#5**：松手后 overlay + 进度条/文案 + persistent Toast；完成后结果 Toast。
  - **#6**：`withHInset` 16px（header/tabs/toolbar/status/行）。
  - 切换来源 Tab 时清空 `selectedPath`，光标回到列表首项（不保留跨 Tab 选中）。
- **相关**：`widgets_source_tab.go`、`actions_mark.go`、`inset.go`、`drop.go`、`view_files.go`、`spec/gui-uiux.md`

---

## [2026-07-18] GUI 九项 UX 对齐（多 agent 攻坚）

- **本会话**：
  - 用户对照 `localhost:3456` 反馈 hover/选中、双击进错目录、单击不选中、拖拽无 UI、缺文件管理器打开、状态栏、主题钮、Files 工具栏（含 docx）等。
  - 并行 [Architect](3999652f-2e6f-46f8-9ca4-0ef9f6425631) / [Planner](2944ed5c-95f9-456e-9d29-34e1242af057) / explore×2 / 对抗审查 / [Loop-op](0e7ece7e-ea71-4568-9444-dd6cc2990d3a)。
  - **根因（对抗通过后落地）**：Fyne `DoubleTappable` 双击不调 `Tapped`；`onOpen` 读全局 selected → 改为闭包绑 idx + `MouseDown` 即时选中；行级 `Hoverable` + `surfaceHover`；独立 `statusBar` 左右分栏键帽；Files 共用 header 主题钮 + 完整工具栏；drop 松手时 overlay；`revealInFileManager`。
  - docx 按钮：UI 占位 + Toast「尚未实现」（spec §5.6 同步）。
  - Fyne 无 drag-enter → drag-over overlay 仍不可用；松手/复制中有 overlay。
- **结论 / 决定**：
  - 目录交互对齐 VS Code：按下选中、双击打开 hit-test 行。
  - 主题按钮在内容区 header（macOS 不塞系统标题栏）；主题切换原先未改 `themeName`，已修。
  - 门禁全绿。
- **相关**：`actions_mouse.go`、`status_bar.go`、`widgets_row_state.go`、`view_files.go`、`spec/gui-uiux.md`

---

## [2026-07-18] GUI UX 对齐 + 拖拽复制（多 agent 攻坚）

- **本会话**：
  - 用户反馈：能进目录，但功能/UI/UX 不对，且不能拖拽复制；要求启动 `try-gui/` 设计稿对照并修复 Go Fyne GUI。
  - 并行 [Architect](3cba8f48-9659-4e28-95e0-b9c12c5a8447) / [Planner](58536406-4bce-4040-ad86-2504e5a95bcf) / explore / [Loop-op](77670e1d-667a-4da6-bcd7-a6208c933652)：以 spec 为准、设计稿补视觉；docx/评分条范围外；拖拽写入 §5.6。
  - **P0 落地**：面包屑 `refreshFilesUI` 重建；相对时间（`FormatTimeAgo`）；Files 目录优先排序；工具栏返回/上传/打开/删除；`SetOnDropped` + `copyDroppedFiles`（复制非 move、重名 skip、symlink skip、失败回滚）。
  - 对抗审计曾否决 symlink 跟随与半截目录残留；已用 `Lstat` + `RemoveAll` 回滚修补。
  - 门禁全绿；Next.js 设计稿 `http://127.0.0.1:3456` 可起。
- **结论 / 决定**：
  - Files 拖拽必须 `Window.SetOnDropped`；Selector 忽略；目标文件走 `IsAllowedTarget`。
  - 仍 defer：行内创建（Ctrl-T 仍模态）、score bar、drop overlay hover、docx。
- **相关**：`drop.go`、`service_import.go`、`view_files.go`、`format.go`、`spec/gui-uiux.md` §5.6

---

## [2026-07-18] GUI 交互修复 + try-gui 视觉对齐（P0/P1）

- **本会话**：
  - 用户截图确认：原生窗口已起，但 UI 像默认 Fyne，Enter 不进 Files，Tab 不切来源。
  - 对抗审查否决「TypedKey/AddShortcut(Tab) 即可」：Fyne 在 `AcceptsTab()==false` 时 `FocusNext` 吃掉 Tab；`List.TypedKey` 不处理 Enter；selector/files 曾共用 `g.list`。
  - P0：持久 `selectorBody`/`filesBody`；`searchEntry.AcceptsTab()==true` + `CurrentKeyModifiers` 做 Shift+Tab；`navList` 处理列表获焦键；分离 `entryList`/`fileList`；双击进 Files。
  - P1：Header / SourceTabs / EntryRow（Highlights）/ FileRow 最小对齐 `try-gui/`；theme token 补 match/accent/dangerSurface。
- **结论 / 决定**：
  - Tab 切换来源必须以 `AcceptsTab` 进入 TypedKey；禁止无效的 `AddShortcut(Tab)`。
  - 门禁全绿；视觉仍不及设计稿完整（无 score bar、行内创建、完整 Files 工具栏）。
- **相关**：`search_entry.go`、`nav_list.go`、`view_*.go`、`widgets.go`、`update_test.go`

---

## [2026-07-18] 平台标题栏差异 + 内容区一致（chrome 落地）

- **本会话**：
  - 多 agent 定案窗口壳层后落地；随后用户反馈 UI 未对齐 `try-gui/`、Enter/Tab 失效。
  - **交互根因（对抗审查）**：Fyne 在 `AcceptsTab()==false` 时用 `FocusNext` 吃掉 Tab，TypedKey/AddShortcut(Tab) 均接不到；`List.TypedKey` 不处理 Enter；曾误把 selector/files 共用一个 `g.list`。
  - **修复**：`searchEntry.AcceptsTab()==true` + `CurrentKeyModifiers` 区分 Shift+Tab；`navList` 覆写 TypedKey/AcceptsTab；分离 `entryList`/`fileList`；停止每次搜索 `SetContent`；双击进 Files；Header/SourceTabs/EntryRow 最小视觉对齐。
  - CI：三平台原生编译 `try`/`try-gui`（`.github/workflows/ci.yml`）。
  - 门禁本机全绿；Win/Linux 标题栏与完整设计稿对齐仍需人工验收。
- **结论 / 决定**：
  - `chrome*.go` 中 chrome = window chrome（窗口装饰），非浏览器。
  - Tab 切换来源必须 `AcceptsTab`；禁止用无效的 `AddShortcut(Tab)` 假装修好。
  - Linux Wayland min/max/drag 仍为空操作。
- **相关**：`search_entry.go`、`nav_list.go`、`view_*.go`、`widgets.go`、`.github/workflows/ci.yml`

---

## [2026-07-18] 否决 HTTP+浏览器 GUI，改为原生窗口

- **本会话**：
  - 用户明确否决「Go HTTP + 系统浏览器」方案：选择 Go 的目的是**原生 GUI**，不是本地 Web 壳。
  - 新约束：macOS 用 macOS 风格标题栏（关闭按钮在左上角）；Windows / Linux 用 WinUI 风格（关闭等按钮在右上角）；除标题栏外，其余 UI 跨平台一致。
  - 本轮使用 planner / architect / loop-operator 多 agent 方式审查路线，确认 Wails/WebView/HTTP 主界面不满足硬约束；Fyne 原生窗口方案更适合作为 MVP 落地路线。
  - 已先改规格文档：`spec/gui-plan.md`、`spec/gui-uiux.md`、`spec/architecture.md`、`README.md` 均改为原生桌面窗口与 Fyne/CGO 发布模型。
  - 已实现首版 Fyne 原生窗口：`cmd/try-gui` 启动 Fyne app；`internal/gui` 拆除 HTTP/server/embed/web 主界面，改为窗口、托盘、Selector、Files、service、opener 与主题 wrapper。
  - 已运行完整门禁：`go build ./...` → `go vet ./...` → `staticcheck ./...` → `go test ./... -count=1` 全部通过；macOS 构建有重复 `-lobjc` linker warning。
- **结论 / 决定**：
  - GUI 必须是原生桌面窗口，不得以浏览器标签页作为产品形态。
  - 平台差异仅限窗口装饰/标题栏；内容区（选择器、文件视图、配色、快捷键）保持统一。
  - 规格采用 Fyne v2：macOS 使用系统标题栏；Windows / Linux 使用应用内 WinUI3 风格标题栏；系统托盘提供 Show / Quit；关闭窗口隐藏到托盘。
  - `try` CLI 保持 `CGO_ENABLED=0`；`try-gui` 使用 `CGO_ENABLED=1` 并按 macOS / Windows / Linux runner 原生构建。
  - `cmd/try-gui` + `internal/gui` 中的 HTTP/embed 主界面代码已替换；业务层继续复用 `selector.catalog`、`script.ExecuteSideEffect`、路径沙箱与 i18n。
- **相关**：`spec/gui-plan.md`、`spec/gui-uiux.md`、`spec/architecture.md`、`README.md`、`.goreleaser.yaml`、`cmd/try-gui`、`internal/gui`、`internal/i18n`

---

## [2026-07-17] 落地跨平台 GUI（try-gui 二进制）并发版

- **本会话**：
  - 多 agent 分析后曾采纳 Go `net/http` + embed + 系统浏览器（CGO=0）。**已被 2026-07-18 条目否决。**
  - 仍可复用的产出：`selector.catalog`、`script.ExecuteSideEffect`、路径沙箱、`IsAllowedTarget`、i18n/API 字段约定等业务层。
- **结论 / 决定**（历史，已被否决）：
  - ~~交付物 = HTTP + 浏览器壳~~ → 改为原生窗口（见上条）。
- **相关**：`cmd/try-gui/`、`internal/gui/`、`spec/gui-plan.md`

---

## [2026-07-17] try-gui 对齐 TUI 首轮改造

- **本会话**：
  - 使用多 agent 方式分析 `try-gui/` 与现有 Bubbletea TUI 的差异，确认 `try-gui` 当前主要提供状态模型、导航 hook、设计 token 与规格，React `TryApp` 组件尚未落地。
  - 首轮实现聚焦 TUI selector 可验证对齐：i18n footer 文案、标记删除行危险背景、目录图标、循环导航、相对时间格式。
  - 修复阻断 `staticcheck` 的 dialog 小问题：删除未使用 modal helper，按 Lipgloss v2 用样式值派生替代 `Copy()`。
  - 用户截图反馈列表目录名几乎不可读：修复选中行前景误用 background、普通行错误套用 `SurfaceSelected`，并纠正 `COLORFGBG` 亮暗判定。
- **结论 / 决定**：
  - TUI 保持 `Enter` 选中后输出 cd 脚本的既有契约，不实现 GUI 的 FilesView。
  - selector 导航与 `try-gui/lib/use-list-navigation.ts` 对齐：`↑/↓` 与 `Ctrl-P/N` 在首尾循环。
  - 标记删除行采用 `✕` 图标、`dangerSurface` 背景、`onDanger` 前景和删除线；普通目录图标采用 `📁`。
  - 相对时间格式对齐 `try-gui/lib/try-types.ts`：`just now`、`Xm`、`Xh`、`Xd`、`Xmo`、`Xy`。
  - 列表文字使用 `foreground`；仅选中/标记行套状态背景。`COLORFGBG` 背景为 `7`/`15` 判为 light。
- **相关**：`internal/selector/`、`internal/i18n/messages.go`、`internal/dialog/`、`internal/config/config.go`、`spec/selector.md`、`spec/tui-framework.md`、`spec/i18n.md`、`spec/config.md`、`spec/architecture.md`、`README.md`

---
