# try GUI 方案设计

## 1. 目标与范围

### 1.1 交付目标

为 try 增加独立 GUI 入口 `try-gui`：启动后进入目录选择器，列出 `~/.config/try/config.json` 中配置的 tries 根目录及各 ships 目录下的实验目录，支持实时模糊搜索与来源过滤。

选中目录后切换到该目录内部的文件列表视图，支持浏览子目录、删除文件、用系统默认程序打开文件，以及返回选择器。拖拽上传与 docx 解压/压缩为后续 Phase，界面预留入口。

### 1.2 交付形态

交付物是单一 Go 二进制 `try-gui`（`CGO_ENABLED=0` 静态链接）：

- 内嵌静态 Web UI（`internal/gui/web/` 经 `//go:embed` 打包进二进制）。
- 在本机 `127.0.0.1` 上以随机可用端口启动 HTTP 服务。
- 服务就绪后用系统默认浏览器打开 `http://127.0.0.1:<port>/`。
- 收到 `SIGINT` / `SIGTERM` 时优雅关闭 HTTP 并退出。

仓库根目录下的 `try-gui/`（Next.js + React + Tailwind）是 UI/UX 设计参考区域，用于对齐布局、配色与交互原型。它不是运行时交付物，不进入 CI 构建，不随 Release 分发；运行 `try-gui` 二进制不依赖 Node.js。

### 1.3 范围外内容

- **Shell 集成与 `cd` 语义**：GUI 是独立进程，副作用通过 `internal/script` 直接执行，不调用 `EmitCd`，也不向父 Shell 输出可 eval 的脚本。
- **从 GUI 内 clone / worktree / install**：这些操作由 TUI / CLI 承担。
- **全局快捷键 / 系统托盘 / 后台守护**：`try-gui` 是普通进程，启动即用、退出即停。
- **多标签 / 多窗口**：同一会话内一个选择器与一个文件视图，切换项目时返回选择器。
- **富文本 docx 编辑**：docx 仅作 ZIP 容器处理（后续 Phase）。

### 1.4 与 TUI 功能对照

| TUI 功能 | GUI 处理方式 | 说明 |
|---|---|---|
| 目录选择器 + 模糊搜索 | 复用 | 评分与排序由 `internal/fuzzy` 提供，与 TUI 一致 |
| 创建新目录 (Ctrl-T) | 复用 | 选择器底部行内输入 |
| 删除目录 (Ctrl-D) | 复用 | 选择器 DeleteMode；文件视图另有删除文件 |
| 重命名目录 (Ctrl-R) | 复用 | 快捷键触发行内输入 |
| Ship 目录 (Ctrl-G) | 复用 | 复用 ship 副作用；ships 中各目录作来源过滤项 |
| 来源过滤 (Tab) | 复用 | all / tries / 各 ship 目录 basename |
| clone / worktree / install | 范围外 | 继续在 TUI / CLI |
| Shell cd 脚本 / EmitCd | 范围外 | GUI 副作用不 EmitCd |

---

## 2. 技术选型

### 2.1 运行时栈

| 层级 | 技术 | 说明 |
|---|---|---|
| 进程入口 | `cmd/try-gui` | Go 主程序，`CGO_ENABLED=0` |
| HTTP 服务 | `net/http`，监听 `127.0.0.1:0` | 仅本机可访问，端口由系统分配 |
| Web UI | 内嵌静态资源 | `internal/gui/web/` 经 `//go:embed` 打包 |
| 浏览器 | 系统默认浏览器 | 服务就绪后打开本机地址 |
| 业务逻辑 | 现有 `internal/` 包 | 配置、扫描、模糊、脚本副作用、i18n |

### 2.2 设计参考区域（非交付物）

`try-gui/` 目录存放 Next.js 设计稿与交互原型，用于对齐布局、配色与键盘行为。实现以本规格与 `spec/gui-uiux.md` 为准。

### 2.3 构建与分发约束

- 默认构建仅编排 API，不含前端资源：`go build ./...`（无 `embed` tag 时 `WebAssets()` 返回 `nil`）。
- 完整二进制通过 `-tags embed` 构建，将 `internal/gui/web/` 打包进产物。
- `try-gui/`（Next.js）不纳入 CI 构建矩阵，不随 GitHub Release 分发。
- 静态 UI 在发布前落入 `internal/gui/web/`，再经 `//go:embed` 打包。

---

## 3. 模块划分

```
try-2026-06-17/
├── cmd/try/                      # 现有 TUI/CLI 入口
├── cmd/try-gui/                  # GUI 入口：解析参数、调用 gui.Run
├── internal/
│   ├── config/                   # 配置加载、路径/locale/主题解析（复用）
│   ├── selector/                 # 扫描、评分、Entry（复用）
│   │   └── catalog.go            # GUI/API 使用的聚合扫描与过滤接口
│   ├── fuzzy/                    # 模糊匹配（复用）
│   ├── script/                   # mkdir/rm/mv/ship 等副作用（复用，不 EmitCd）
│   ├── i18n/                     # 文案（复用）
│   └── gui/
│       ├── app.go                # Run：加载配置、监听、开浏览器、生命周期
│       ├── server.go             # HTTP mux 装配、路由、本机 CORS
│       ├── handlers.go           # 各 API 处理器与路径安全校验
│       ├── dto.go                # 请求/响应 DTO 与前端契约
│       ├── browser.go            # openURL：跨平台打开浏览器/文件
│       ├── paths.go              # IsAllowedPath：根目录前缀校验
│       ├── embed_prod.go         # //go:embed web（-tags embed）
│       ├── embed_dev.go          # 非 embed 构建回退，WebAssets 返回 nil
│       └── web/                  # 内嵌静态 Web UI（index.html / app.js / app.css）
├── try-gui/                      # UI/UX 设计参考（不进 CI/release）
└── ...
```

### 核心模块职责

| 模块 | 职责 |
|---|---|
| `cmd/try-gui` | 解析 `-path` 参数，调用 `gui.Run` |
| `internal/gui/app.go` | 加载配置、`ensureGUIDirs`、监听 `127.0.0.1:0`、启动 server、开浏览器、信号退出 |
| `internal/gui/server.go` | 装配 mux：`/api/*` 路由与静态资源，`corsLocalhost` 限制本机来源 |
| `internal/gui/handlers.go` | bootstrap、entries 列表/搜索、创建/删除/重命名/ship、文件列表/删除/打开 |
| `internal/gui/dto.go` | JSON 契约：`BootstrapDTO`、`EntryDTO`、`FileDTO`、请求体与错误响应 |
| `internal/gui/browser.go` | `openURL` 跨平台打开浏览器与文件（open / rundll32 / xdg-open） |
| `internal/gui/paths.go` | `IsAllowedPath` 校验路径落在允许根目录子树内 |
| `internal/gui/web` | 内嵌前端静态文件 |
| `internal/selector/catalog.go` | `LoadAllEntries` / `MatchEntries` / `SourceOptions` / `SourceCounts`，供 handlers 调用 |
| `internal/config` / `fuzzy` / `script` / `i18n` | 与 TUI 共享同一套业务实现 |

---

## 4. 与现有 Go 代码的关系

### 4.1 业务复用

| 能力 | 包 |
|---|---|
| 配置读取、路径/locale/主题解析 | `internal/config`（`ResolvePaths` / `ResolveLocale` / `DetectTheme`） |
| 目录扫描、来源标签、评分聚合 | `internal/selector`（含 `catalog.go`） |
| 模糊匹配 | `internal/fuzzy` |
| 创建 / 删除 / 重命名 / ship | `internal/script`（`ExecuteSideEffect`，不 EmitCd） |
| 用户可见文案 | `internal/i18n` |

GUI 通过 `catalog.go` 与 handlers 暴露操作接口，不驱动 Bubbletea Model。

### 4.2 配置格式

GUI 读取与 TUI 相同的 `~/.config/try/config.json`：

```json
{
  "path": "~/src/tries",
  "ships": ["~/src/ship", "~/src/bug"],
  "locale": "auto"
}
```

`ships` 为数组，其中每个目录以其 basename 作为一个来源过滤项。启动时若配置缺失，`gui.Run` 会初始化默认配置文件后再加载。

路径优先级与 Go 端一致：

1. 命令行参数（`-path`）
2. 环境变量
3. 配置文件
4. 默认值

### 4.3 TUI 并存

TUI 继续作为 Shell `eval` / `cd` 工作流入口，GUI 为附加入口，两者共享配置与 `internal/` 逻辑。

---

## 5. 核心数据流

```
启动 try-gui
    │
    ▼
LoadConfig / ResolvePaths / ResolveLocale / DetectTheme / ensureGUIDirs
    │
    ▼
监听 127.0.0.1:0，装配 embed 静态资源 + /api 路由
    │
    ▼
打开系统浏览器 → 加载 Web UI
    │
    ▼
GET /api/bootstrap（locale / theme / messages / paths）
GET /api/entries（q + source）→ 渲染 Selector
    ├── 搜索、来源 Tab、循环导航
    ├── Ctrl-T / Ctrl-R / Ctrl-G / Ctrl-D（DeleteMode）
    └── Enter → 进入 files 视图
    │
    ▼
Files 视图
    ├── GET /api/files 列出目录条目
    ├── POST /api/files/delete 删除（确认后）
    ├── POST /api/files/open 外部打开
    └── 返回 Selector（刷新列表）
    │
    ▼
收到退出信号：Shutdown HTTP，释放端口
```

副作用路径：handlers → `internal/script.ExecuteSideEffect`（创建/删除/重命名/ship）或 `os` 调用（文件删除/打开）。结果以 JSON 响应与前端 Toast 反馈；stdout 不输出 cd 脚本。

---

## 6. HTTP API

所有接口同源挂载于 `127.0.0.1`，`corsLocalhost` 仅放行 `127.0.0.1` / `localhost` / `::1` 来源。

| 方法 | 路径 | 作用 |
|---|---|---|
| GET | `/api/bootstrap` | 返回 locale、theme、i18n messages、paths |
| GET | `/api/entries?q=&source=` | 返回过滤+模糊匹配后的条目与来源计数 |
| POST | `/api/entries/create` | 在 tries 根下创建 `name-YYYY-MM-DD` 目录 |
| POST | `/api/entries/delete` | 批量删除条目 |
| POST | `/api/entries/rename` | 重命名条目 |
| POST | `/api/entries/ship` | 将条目 ship 到指定 ships 目录 |
| GET | `/api/files?path=` | 列出目录下条目（隐藏 `.` 开头项） |
| POST | `/api/files/delete` | 批量删除文件/子目录 |
| POST | `/api/files/open` | 用系统默认程序打开路径 |

### 6.1 DTO 契约

- `EntryDTO`：`id`、`name`、`baseName`、`date`、`source`、`score`、`lastModified`、`path`、`highlights`。
- `FileDTO`：`id`、`name`、`type`、`sizeKB`、`modified`、`isDir`、`path`。
- `BootstrapDTO`：`locale`、`theme`、`messages`、`paths{tries, ships}`。

### 6.2 路径安全

所有涉及文件系统的接口先经 `IsAllowedPath`：对入参做 `filepath.Abs` + `filepath.Clean`，再校验其落在 tries 根或任一 ships 目录子树内，越界返回 `403`，防止路径遍历。

---

## 7. 关键实现点

### 7.1 目录聚合与评分

`catalog.go` 的 `LoadAllEntries` 扫描 tries 根与各 ships 目录，`MatchEntries` 按来源过滤后交由 `internal/fuzzy` 评分排序。DTO 字段与前端展示对齐：`baseName`、`date`、`source`、`score`、`lastModified` 等。

### 7.2 静态 UI 与主题

Web UI 使用与 `internal/selector/styles.go` 一致的 dark/light hex token（详见 `spec/gui-uiux.md`）。主题默认 Dark，前端支持手动切换；自动跟随系统为后续完善项。

### 7.3 窗口尺寸

设计目标默认视口 **900 × 600**。交付形态为本机 HTTP + 系统浏览器，浏览器壳无法保证原生窗口尺寸与居中。规格上以 CSS 最小宽度与布局目标约束内容区，不依赖原生窗口 API。

### 7.4 文件视图

- 列表浏览、进入子目录、面包屑跳转、返回上级 / 返回选择器。
- 删除文件或目录（确认对话框，默认焦点 Cancel）。
- 外部打开（系统默认程序）。

拖拽上传、docx 解压/压缩为后续 Phase，界面预留入口。

### 7.5 docx（后续 Phase）

规划新增 `internal/docx`：`Extract` / `Pack`，基于 `archive/zip` 与最小 OOXML 结构。GUI 仅调用该包。

---

## 8. 风险与折中

| 风险 | 影响 | 缓解 |
|---|---|---|
| 浏览器无法锁定窗口尺寸 | 与 900×600 设计目标有偏差 | 文档标明；内容区用 min-width / 布局约束 |
| GUI 与 TUI 行为漂移 | 体验不一致 | 共享 fuzzy/selector/script；快捷键对照 `gui-uiux.md` |
| 静态资源与 API 契约不同步 | 前端解析错误 | dto 单一来源；发布前固定 embed 产物 |
| 拖拽/docx 延后 | v1 能力不完整 | UI 预留入口，里程碑单独跟踪 |

---

## 9. 实施里程碑

状态标记：`[x]` 已完成，`[~]` 进行中，`[ ]` 待实现。

### Phase 1：Go 骨架 + Selector API + 内嵌 UI

目标：`try-gui` 启动本机服务并打开浏览器，选择器列表与搜索可用。

- [x] `cmd/try-gui` 入口与 `gui.Run` 进程生命周期
- [x] `internal/gui`：app / server / handlers / dto / browser / paths / embed
- [x] `internal/selector/catalog.go` 聚合与过滤接口
- [x] 接入 `config` + `fuzzy`，搜索排序与 TUI 一致
- [x] 内嵌静态 UI：搜索、来源 Tab、循环导航、状态栏
- [x] Ctrl-T / Ctrl-R / Ctrl-G / Ctrl-D（DeleteMode）与确认流
- [x] Enter 进入 files 视图
- [x] 路径安全测试（`paths_test.go`）

### Phase 2：Files 视图 + 基本文件操作

目标：选中目录后完成文件浏览与基础管理。

- [x] 文件列表、子目录进入、面包屑 / 返回
- [x] 删除确认与执行
- [x] 外部程序打开文件
- [x] 返回 selector 并刷新
- [ ] catalog 与关键 handlers 的单元/集成测试补齐

### Phase 3：拖拽上传 + docx

目标：补齐拖拽与 docx（UI 已预留入口）。

- [ ] 浏览器拖拽上传 → 复制到当前或目标子目录
- [ ] `internal/docx.Extract` / `Pack`
- [ ] 进度与覆盖确认对话框
- [ ] 跨平台拖拽与 ZIP 一致性测试

### Phase 4：完善与发布

目标：与 TUI 一并作为正式产物发布。

- [ ] 主题自动跟随系统（`prefers-color-scheme`）
- [ ] GUI 相关配置字段文档化（`spec/config.md` / README）
- [ ] CI：以 `-tags embed` 构建 `try-gui`
- [ ] Release 提供 `try-gui` 二进制（`CGO_ENABLED=0`）
- [ ] README 增加 GUI 启动说明
