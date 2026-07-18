# try GUI UI/UX 规范

本文件描述 `try-gui` 原生桌面窗口的布局、配色、快捷键与状态机。仓库 `try-gui/` 中的设计稿用于对齐视觉与交互，不绑定具体实现框架。

## 1. 设计目标与原则

### 1.1 核心目标

- **键盘驱动**：Enter、Ctrl-T、Ctrl-D、Ctrl-R、Ctrl-G、Esc、Tab、Ctrl-P/N 等与 TUI 行为对齐。
- **鼠标增强**：点击、双击、悬停反馈补充键盘路径。
- **零认知迁移**：布局层级、配色 token、术语、分隔结构与 TUI 对齐。
- **原生桌面窗口**：启动后进入应用窗口，尺寸、最小尺寸、居中、托盘和关闭语义由应用控制。
- **跨平台一致**：标题栏与窗口装饰按平台规则呈现；内容区布局、组件、配色、字体层级、快捷键、对话框、Toast 完全一致。

### 1.2 设计原则

| 原则 | 说明 |
|---|---|
| 键盘优先，鼠标增强 | 高频操作均可键盘完成；鼠标用于发现与直接操作 |
| 即时反馈 | 搜索过滤、悬停、标记、操作结果在 100ms 内可见 |
| 安全默认 | 删除、Ship 默认焦点取消；危险按钮使用 danger 色 |
| 同一窗口 | 选择器、文件视图、对话框在同一个原生窗口内完成 |
| 可访问 | 可见焦点环、合理 Tab 顺序、按钮可通过键盘触达 |

### 1.3 关键指标

| 指标 | 目标 |
|---|---|
| 启动到主界面可交互 | < 500ms |
| 搜索输入到列表更新 | < 50ms（约 500 条目内） |
| 删除确认默认焦点 | Cancel / NO |
| 默认窗口尺寸 | 约 900 × 600 px |
| 最小窗口尺寸 | 约 720 × 480 px |

## 2. 应用整体结构

### 2.1 窗口与标题栏

| 项目 | 值 |
|---|---|
| 默认尺寸 | 900 × 600 px |
| 最小尺寸 | 720 × 480 px |
| 初始位置 | 屏幕居中 |
| 可调整 | 支持拖拽调整、最大化与还原 |
| 关闭语义 | 隐藏窗口到系统托盘 |
| 退出语义 | 托盘 Quit 退出进程 |

平台标题栏规则：

| 平台 | 标题栏与控件 |
|---|---|
| macOS | 系统标题栏；关闭 / 最小化 / 最大化控件位于左上角，使用红绿灯风格 |
| Windows | WinUI3 风格标题栏；最小化 / 最大化 / 关闭控件位于右上角 |
| Linux | 与 Windows 一致；标题栏控件位于右上角；X11 下支持系统级最小化/最大化/拖拽；Wayland 下这三项当前为空操作 |

### 2.2 布局网格

- 基础栅格：8px
- 平台标题栏高度：44px（Windows / Linux 内容内标题栏）
- 列表行高：40px
- 列表项水平内边距：16px
- 搜索框高度：36px
- 工具栏高度：44px
- 状态栏高度：28px
- 分隔线：1px line 色
- 对话框圆角：8px
- 对话框宽度：最小 400px，最大 560px

### 2.3 主题系统

- **默认**：来自 `config.DetectTheme()`，未识别时使用 Dark。
- **手动覆盖**：窗口内容区标题栏右侧主题按钮在 Dark / Light 间切换（Selector 与 Files 均提供；macOS 系统标题栏不放置主题控件）。
- **覆盖范围**：Selector、Files、标题栏、对话框、Toast、托盘菜单文案全部同步。

配色 token 与 `internal/selector/styles.go` 及设计稿一致。

### 2.4 色彩规范

**Dark**

| 角色 | 色值 | Token 名 |
|---|---|---|
| 背景 | `#0d1117` | background |
| 表面 | `#161b22` | surface |
| 表面悬停 | `#1f242c` | surfaceHover |
| 表面选中 | `#21262d` | surfaceSelected |
| 前景 | `#e6edf3` | foreground |
| 主色/高亮/标题 | `#5fafff` | highlight / header |
| 搜索匹配 | `#ffaf5f` | match |
| 强调/成功 | `#87d787` | accent |
| 危险 | `#ff3b30` | danger |
| 危险表面 | `#3f1e1c` | dangerSurface |
| 中性灰 | `#8b949e` | muted |
| 边框 | `#30363d` | line |

**Light**

| 角色 | 色值 | Token 名 |
|---|---|---|
| 背景 | `#ffffff` | background |
| 表面 | `#f6f8fa` | surface |
| 表面悬停 | `#f3f4f6` | surfaceHover |
| 表面选中 | `#eef0f2` | surfaceSelected |
| 前景 | `#1f2328` | foreground |
| 主色/高亮/标题 | `#005fd7` | highlight / header |
| 搜索匹配 | `#af5f00` | match |
| 强调/成功 | `#008700` | accent |
| 危险 | `#d70000` | danger |
| 危险表面 | `#ffe5e5` | dangerSurface |
| 中性灰 | `#6e7781` | muted |
| 边框 | `#d0d7de` | line |

### 2.5 字体与图标

- **正文**：系统无衬线，14px。
- **路径/代码**：系统等宽，13px。
- **图标**：文件夹、文件类型、工具栏动作使用统一图标集。

## 3. 界面流程与状态机

### 3.1 状态定义

```go
type ViewState string // selector | files
type SourceTab string // all | tries | <ship basename>

type AppState struct {
    View         ViewState
    Source       SourceTab
    Query        string
    Entries      []EntryView
    Counts       map[string]int
    Selected     int
    Marked       map[string]bool
    FilesPath    string
    FilesRoot    string
    Files        []FileEntry
    FileSelected int
    FileMarked   map[string]bool
    Inline       *InlineInput
    Modal        *ModalState
    Theme        string
    Toast        *ToastState
}
```

来源 Tab 由配置中的 `path` 与 `ships` 数组派生：`all`、`tries`，以及各 ship 目录 basename。

### 3.2 状态转换

| 当前 | 触发 | 目标 | 说明 |
|---|---|---|---|
| 启动 | 配置、目录列表加载完成 | selector | 读取配置与目录列表 |
| selector | 输入搜索 | selector | 实时过滤，无 debounce |
| selector | Ctrl-T | 行内输入 | 新建目录 |
| selector | Ctrl-D / Space | selector | 标记 / 取消标记当前项 |
| selector | 有标记时 Enter / Delete | 对话框 | 删除确认 |
| selector | Enter（无标记） | files | 进入选中目录；并对该 try 根目录 `Chtimes`（对齐 TUI 访问排序），返回选择器后按 path 重定位光标 |
| selector | Ctrl-R | 行内输入 | 重命名当前项 |
| selector | Ctrl-G | 对话框 | Ship 到某 ships 目录 |
| files | Enter 于文件夹 / 双击 | files | 进入子目录 |
| files | Enter 于文件 / 双击 | files | 系统默认程序打开 |
| files | Delete / 删除按钮 | 对话框 → files | 确认后刷新 |
| files | Esc / 返回 | files 或 selector | 逐级返回 |

## 4. 选择器（Selector）

### 4.1 布局结构

自上而下：

1. 标题区：产品名、当前视图标题、主题切换按钮（图标按钮）。
2. 搜索行：搜索框 + 来源 Tab（all / tries / 各 ship basename）。
3. 目录列表：选中指示、名称（含匹配高亮）、日期、相对修改时间、来源徽章。
4. 行内输入行（Ctrl-T 或空列表下对非空查询按 Enter 时显示）。
5. 底栏：左侧项目计数（DeleteMode 下为删除模式文案）；右侧快捷键键帽（↑↓ / Enter / ⌃T / ⌃D）。

### 4.2 搜索框

- 占位符从 `i18n.Messages` 获取。
- 输入即时过滤，无 debounce。
- `/` 或 Ctrl+F 聚焦搜索框。
- Enter：列表非空则进入选中项；列表为空且查询非空则进入新建流程。

### 4.3 目录列表

- 行高约 40px。
- 悬停：未选中行使用 `surfaceHover` 底色（行级 `desktop.Hoverable`）。
- 选中：`surfaceSelected` 底色 + 左侧 `›` 指示符；按下即选中（`MouseDown`），对齐 VS Code 文件列表。
- 双击：打开指针下的行（闭包绑定行索引），不依赖全局选中是否已更新。
- 标记删除：危险色 + 删除线，指示符切换为删除标记 `x`。
- 名称中的搜索匹配字符使用 match 色高亮。
- 导航：↑/↓ 与 Ctrl-P/N 在首尾循环。

### 4.4 来源过滤

- Tab / Shift+Tab 在来源间循环切换。
- 鼠标点击 Tab 切换；活动 Tab 使用主色（header）。
- 每个 Tab 显示来源计数。
- 非激活 Tab hover：背景 `surfaceHover`，文字与计数为 `foreground`。
- 切换来源时清空先前选中 path，光标回到过滤结果列表首项。

### 4.5 操作入口

| 操作 | 快捷键 | 鼠标 |
|---|---|---|
| 进入目录 | Enter | 双击行 |
| 新建 | Ctrl-T | 工具区按钮 |
| 标记删除 | Ctrl-D / Space | 点击选中后按键 |
| 有标记删除确认 | Enter / Delete | 删除按钮 |
| 重命名 | Ctrl-R | 行操作按钮 |
| Ship | Ctrl-G | 行操作按钮 |
| 切换来源 | Tab / Shift+Tab | 点击 Tab |

## 5. 文件视图（Files）

### 5.1 布局结构

自上而下：

1. 标题区：当前目录名、主题切换按钮（与选择器共用内容区 header）。
2. 工具栏：返回、编辑、打包 .docx、解压 .docx、在文件管理器中打开、删除；其后为可点击面包屑。
3. 列头：名称 / 大小 / 修改时间。
4. 文件列表：复选框、图标、名称、大小（文件）、相对修改时间。
5. 底栏：左侧项目计数；右侧快捷键键帽（↑↓ 导航、Esc 返回）与「拖拽上传」提示。

工具栏「编辑」打开选中项（目录进入、文件用系统默认程序）。「打包 .docx」「解压 .docx」为本阶段 UI 占位，点击后 Toast 提示尚未实现。

「在文件管理器中打开」在系统文件管理器中揭示当前 `FilesPath`（macOS `open`、Windows `explorer`、Linux `xdg-open`）；路径须通过 `IsAllowedPath`。

### 5.2 面包屑

- 路径分段可点击，跳转到对应层级。
- 最左段为进入文件视图的根目录。
- 返回按钮先返回上级，已在根目录时返回选择器。

### 5.3 文件列表

- 默认列表视图，隐藏 `.` 开头的条目。
- 行首复选框：勾选/取消勾选将该项加入多选集合（`fileMarked`）；删除等批量操作优先使用已勾选项。
- Ctrl+点击（macOS 亦认 ⌘）：切换该项勾选，不清除其余已选项。
- 普通单击：移动光标（`fileSelected`），不改变勾选集合。
- 进入子目录：Enter / 双击文件夹。
- 打开文件：Enter / 双击。
- 导航：↑/↓ 与 Ctrl-P/N 在首尾循环。

### 5.4 文件图标类型

后端按扩展名归类，GUI 据此选择图标。

| 类别 | 扩展名或条件 |
|---|---|
| `dir` | 目录 |
| `go` | `.go` |
| `ts` | `.ts` / `.tsx` |
| `js` | `.js` / `.jsx` |
| `md` | `.md` / `.markdown` |
| `json` | `.json` |
| `txt` | `.txt` |
| `docx` | `.docx` |
| `zip` | `.zip` |
| `image` | `.png` / `.jpg` / `.jpeg` / `.gif` / `.webp` |
| `unknown` | 其他文件 |

### 5.5 删除确认

- 模态对话框；默认焦点 Cancel。
- 危险按钮使用 danger 色。
- 列出前 5 个路径，超出显示 i18n 文案描述的剩余数量。

### 5.6 拖拽复制（上传）

Files 视图支持从系统文件管理器将文件或文件夹拖入应用窗口，复制到当前 `FilesPath`。

| 项 | 行为 |
|---|---|
| 触发 | 仅 `view == files` 时处理；Selector 视图忽略拖入 |
| 语义 | **复制**（非移动）；源文件保留在原位置 |
| 目标 | 当前浏览目录 `FilesPath`（含子目录） |
| 目录 | 递归复制子树；跳过名称以 `.` 开头的条目 |
| 重名 | 目标已存在同名项则 **跳过**，继续处理其余项 |
| 沙箱 | 目标文件路径必须通过 `IsAllowedTarget`；当前浏览目录须在允许子树内（`IsAllowedPath`） |
| 符号链接 | 跳过源端符号链接，不跟随复制 |
| 失败回滚 | 目录复制中途失败时删除不完整目标，避免重名 skip 吞掉残缺目录 |
| 反馈 | 松手后显示半透明 overlay、「正在复制」与项进度（`done/total`）；复制期间 Toast 不自动清除；完成后 Toast 报告已复制 / 已跳过数量并刷新列表 |
| 限制 | Fyne 无 drag-enter API，悬停窗口时不显示 overlay；仅松手后进入 importing 反馈 |
| 提示 | 底栏右侧含「拖拽上传」accent 文案 |

docx 打包/解压按钮在工具栏可见，本阶段点击后 Toast「尚未实现」，不执行打包逻辑。

## 6. 对话框、Toast 与托盘

### 6.1 确认对话框

模态；默认焦点 Cancel；危险操作确认按钮使用 danger 色。← / → 在 Cancel 与确认按钮间移动焦点，Enter 触发当前焦点，Esc 取消。

### 6.2 行内输入

选择器底栏的行内输入用于创建与重命名。实时校验：空名视为取消；名称含 `/` 视为非法。

### 6.3 Toast

窗口底部居中；约 2 秒后消失。Toast 文案从 `i18n.Messages` 获取。

### 6.4 托盘

托盘菜单包含 Show 与 Quit。Show 显示并聚焦窗口；Quit 退出进程。窗口关闭按钮隐藏窗口到托盘。

## 7. 键盘快捷键

| 快捷键 | 功能 |
|---|---|
| `/` | 聚焦搜索框 |
| Ctrl+F | 聚焦搜索框 |
| Ctrl+T | 新建目录 |
| Ctrl+D | 标记 / 取消标记当前项 |
| Space | 标记 / 取消标记当前项 |
| Ctrl+R | 重命名当前项 |
| Ctrl+G | Ship |
| Esc | 见 §7.1 |
| Ctrl+P / ↑ | 上移（循环） |
| Ctrl+N / ↓ | 下移（循环） |
| Enter | 进入 / 打开 / 有标记时打开删除确认 |
| Tab / Shift+Tab | 切换来源（选择器内） |
| Delete | 删除标记项或打开删除确认 |

### 7.1 Esc 语义

Esc 在选择器中按优先级从高到低处理：

1. 若行内输入打开：关闭行内输入。
2. 若对话框打开：取消对话框。
3. 若存在删除标记：清空标记，退出 DeleteMode。
4. 若搜索框非空：清空搜索并刷新列表。
5. 若搜索框为空：隐藏窗口到托盘。

Esc 在文件视图按优先级：

1. 若对话框打开：取消对话框。
2. 若存在文件删除标记：清空标记。
3. 若当前目录非根目录：返回上级目录。
4. 若已在根目录：返回选择器。

### 7.2 DeleteMode

- Ctrl-D 或 Space 标记 / 取消标记当前项。
- 存在标记即进入 DeleteMode；标记清空即退出。
- 有标记时 Enter（选择器）或 Delete（文件视图）打开删除确认。
- Esc 清空标记并退出 DeleteMode。

## 8. 响应式与可访问性

### 8.1 断点

| 宽度 | 调整 |
|---|---|
| < 800px | 隐藏评分/次要元数据列 |
| < 640px | 隐藏修改时间列 |
| < 560px | 来源 Tab 简化为短标签或图标 |

### 8.2 焦点

- 对话框打开时焦点陷阱在对话框内。
- Tab 顺序遵循视觉顺序。
- 可见焦点环。

### 8.3 辅助技术

- 列表项提供可读名称、来源、修改时间和选中状态。
- 图标按钮提供可读标签。
- Toast 使用状态语义通知操作结果。
