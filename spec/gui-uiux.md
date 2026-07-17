# try GUI UI/UX 规范

本文件是实现无关的 UI/UX 规格，描述 `try-gui` 前端的布局、配色、快捷键与状态机。文中出现的「设计参考区域」指仓库 `try-gui/` 中的设计稿标注，仅用于对齐视觉与交互，不绑定任何具体框架文件。

## 1. 设计目标与原则

### 1.1 核心目标

- **键盘驱动**：Enter、Ctrl-T、Ctrl-D、Ctrl-R、Ctrl-G、Esc、Tab、Ctrl-P/N 等与 TUI 行为对齐。
- **鼠标增强**：点击、双击、悬停反馈；拖拽上传为后续能力，界面预留投放区。
- **零认知迁移**：布局层级、配色 token、术语、分隔结构与 TUI 对齐。
- **跨平台一致**：macOS / Windows / Linux 交互一致；窗口装饰跟随系统浏览器壳。

### 1.2 设计原则

| 原则 | 说明 |
|---|---|
| 键盘优先，鼠标增强 | 高频操作均可键盘完成；鼠标用于发现与直接操作 |
| 即时反馈 | 搜索过滤、悬停、标记、操作结果在 100ms 内可见 |
| 安全默认 | 删除、Ship 默认焦点取消；危险按钮使用 danger 色 |
| 同一会话 | 选择器、文件视图、对话框在同一浏览器标签页内完成 |
| 可访问 | 可见焦点环、合理 Tab 顺序、辅助技术标签 |

### 1.3 关键指标

| 指标 | 目标 |
|---|---|
| 启动到主界面可交互 | < 500ms（本机服务就绪后） |
| 搜索输入到列表更新 | < 50ms（约 500 条目内） |
| 拖拽反馈延迟 | < 100ms（启用拖拽后） |
| 删除确认默认焦点 | Cancel / NO |

---

## 2. 应用整体结构

### 2.1 视口与窗口

| 项目 | 值 |
|---|---|
| 设计目标视口 | 900 × 600 px |
| 内容区最小尺寸 | 720 × 480 px |
| 启动位置 | 设计目标为屏幕居中 |
| 可调整 | 浏览器允许缩放、最大化、全屏 |

交付形态为本机 HTTP + 系统浏览器。浏览器壳无法保证原生窗口尺寸、居中或标题栏控件；实现以 CSS 最小宽度与布局目标约束内容区，不依赖原生窗口 API。应用壳顶部提供一条模拟标题栏（含主题切换按钮），用于视觉一致，不代表可控的原生窗口。

### 2.2 布局网格

- 基础栅格：8px
- 列表行高：40px
- 列表项水平内边距：16px
- 搜索框高度：36px
- 工具栏高度：44px
- 状态栏高度：28px
- 分隔线：1px line 色
- 对话框圆角：8px
- 对话框宽度：最小 400px，最大 560px

### 2.3 主题系统

- **默认**：Dark（对应 TUI `dark` 调色板）。
- **手动覆盖**：顶部主题按钮在 Dark / Light 间切换（切换 `<html>` 的 `dark` / `light` 类）。
- **自动跟随系统**：`prefers-color-scheme` 为后续完善项。
- **初始主题**：来自 `/api/bootstrap` 的 `theme`（由后端 `config.DetectTheme` 提供）。

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
- **图标**：文件夹、文件类型、工具栏动作使用统一图标集；设计参考区域对应「文件图标 / 工具栏图标」标注。

---

## 3. 界面流程与状态机

### 3.1 状态定义

```
ViewState = selector | files
SourceTab = all | tries | <ship basename> ...

AppState:
  view              ViewState
  source            SourceTab
  query             string
  entries           []Entry
  counts            map[source]int
  selected          int       # 当前高亮索引
  marked            set        # 选择器删除标记
  filesPath         string     # 文件视图当前目录
  filesRoot         string     # 进入文件视图的根目录
  files             []File
  fileSelected      int
  fileMarked        set        # 文件视图删除标记
  inline            {mode, value} | null   # 行内输入（create / rename）
  modal             对话框描述 | null
```

来源 Tab 由配置中的 `path`（tries）与 `ships` 数组派生：`all`、`tries`，以及各 ship 目录 basename。设计稿中的 `ship`、`bug` 即典型的两个 ships 目录名。

### 3.2 状态转换

| 当前 | 触发 | 目标 | 说明 |
|---|---|---|---|
| 启动 | bootstrap + entries 加载完成 | selector | 读取配置与目录列表 |
| selector | 输入搜索 | selector | 实时过滤，无 debounce |
| selector | Ctrl-T | 行内输入 | 新建目录 |
| selector | Ctrl-D / Space | selector | 标记 / 取消标记当前项 |
| selector | 有标记时 Enter / Delete | 对话框 | 删除确认 |
| selector | Enter（无标记） | files | 进入选中目录 |
| selector | Ctrl-R | 行内输入 | 重命名当前项 |
| selector | Ctrl-G | 对话框 | Ship 到某 ships 目录 |
| files | Enter 于文件夹 / 双击 | files | 进入子目录 |
| files | Enter 于文件 / 双击 | files | 系统默认程序打开 |
| files | Delete / 删除按钮 | 对话框 → files | 确认后刷新 |
| files | Esc / 返回 | files 或 selector | 逐级返回（见 §8.1） |

---

## 4. 选择器（Selector）

设计参考区域：应用壳、标题栏、搜索行、来源 Tab、目录行、底栏/行内输入、删除确认。

### 4.1 布局结构

自上而下：

1. 标题栏：产品名 / 「Directory Selection」+ 主题切换按钮。
2. 搜索行：搜索框 + 来源 Tab（all / tries / 各 ship basename）。
3. 目录列表：选中指示、名称（含匹配高亮）、日期、相对修改时间、来源徽章。
4. 行内输入行（Ctrl-T 或空列表下对非空查询按 Enter 时显示）。
5. 底栏：快捷键提示；DeleteMode 下显示已标记数量。

### 4.2 搜索框

- **占位符**：EN `Type to filter or create...`；ZH `输入以过滤或创建...`（经 i18n）。
- **行为**：
  - 输入即时过滤（`GET /api/entries`），无 debounce。
  - `/` 或 Ctrl+F 聚焦搜索框。
  - Esc：见 §8.1。
  - Enter：列表非空则进入选中项；列表为空且查询非空则进入新建流程。

### 4.3 目录列表

- 行高约 40px。
- 选中：surfaceSelected 底色 + 左侧 `›` 指示符。
- 标记删除：危险色 + 删除线，指示符切换为删除标记 `✕`。
- 名称中的搜索匹配字符使用 match 色高亮。
- 导航：↑/↓ 与 Ctrl-P/N 在首尾循环。

### 4.4 来源过滤

- Tab / Shift+Tab 在来源间循环切换（正向 / 反向）。
- 鼠标点击 Tab 切换；活动 Tab 使用主色。
- 每个 Tab 显示来源计数（来自 `/api/entries` 的 counts）。

### 4.5 操作入口

| 操作 | 快捷键 | 鼠标 |
|---|---|---|
| 进入目录 | Enter | 双击行 |
| 新建 | Ctrl-T | 空查询状态由底栏行内输入 |
| 标记删除 | Ctrl-D / Space | 点击选中后按键 |
| 有标记删除确认 | Enter / Delete | — |
| 重命名 | Ctrl-R | — |
| Ship | Ctrl-G | — |
| 切换来源 | Tab / Shift+Tab | 点击 Tab |

### 4.6 底栏

- 默认：快捷键提示条。
- DeleteMode（存在标记）：「DELETE MODE」+ 已标记数量，使用 danger 呈现。
- 操作结果：底部居中 Toast。

---

## 5. 文件视图（Files）

设计参考区域：面包屑、工具栏、文件行、拖拽遮罩、删除确认、空目录提示。

### 5.1 布局结构

自上而下：

1. 工具栏：返回按钮、面包屑路径、删除按钮（有标记时显示）。
2. 文件列表：图标、名称、大小（文件）、相对修改时间。
3. 底栏：快捷键提示。
4. 拖拽遮罩（拖拽能力启用后、悬停时显示）。

### 5.2 面包屑

- 路径分段可点击，跳转到对应层级。
- 最左段为根目录；返回按钮先返回上级，已在根目录时返回选择器。

### 5.3 文件列表

- 默认列表视图，隐藏 `.` 开头的条目。
- 进入子目录：Enter / 双击文件夹。
- 打开文件：Enter / 双击 → 系统默认程序（`POST /api/files/open`）。
- 导航：↑/↓ 与 Ctrl-P/N 在首尾循环。

### 5.4 文件图标类型

后端按扩展名归类（`dir`、`go`、`ts`、`js`、`md`、`json`、`txt`、`docx`、`zip`、`image`、`unknown`），前端据此选择图标。

| 类别 | 表现 |
|---|---|
| 文件夹 | 文件夹图标 |
| 文本/代码 | 文本文件图标 |
| docx | 文档图标（可区分色） |
| 图片 | 图片图标 |
| 其他 | 通用文件图标 |

### 5.5 拖拽上传（后续，UI 预留）

- 拖入列表区显示半透明遮罩与提示文案。
- 落到空白处上传至当前目录；落到文件夹行上传至该子目录。

### 5.6 删除确认

- 模态对话框；默认焦点 Cancel。
- 危险按钮使用 danger 色。
- 列出前 5 个路径，超出显示 `+N more`。

### 5.7 空目录

居中提示：目录为空。后续启用拖拽/新建后在此提供入口。

---

## 6. docx 操作（后续，UI 预留）

### 6.1 解压

- 入口：右键 `.docx` → Extract。
- 对话框：源、目标、覆盖选项、进度。
- 结果：Toast。

### 6.2 压缩

- 入口：右键目录或工具栏 → Pack as docx。
- 对话框：源目录、输出名、进度。
- 结果：Toast。

### 6.3 错误呈现

| 情况 | 呈现 |
|---|---|
| 无效 docx | Toast：Invalid docx file（i18n） |
| 输出已存在 | 确认是否覆盖 |
| 权限不足 | Toast：Permission denied（i18n） |

---

## 7. 对话框与反馈

### 7.1 确认对话框

模态；默认焦点 Cancel；危险操作确认按钮使用 danger 色。← / → 在 Cancel 与确认按钮间移动焦点，Enter 触发当前焦点，Esc 取消。

### 7.2 行内输入

选择器底栏的行内输入用于创建与重命名。实时校验：空名视为取消；名称含 `/` 视为非法。

### 7.3 Toast

窗口底部居中；约 2 秒后消失；`role="status"`。

### 7.4 加载

长操作显示进度或不确定动画；可取消的操作提供 Cancel。

---

## 8. 键盘快捷键

| 快捷键 | 功能 |
|---|---|
| `/` | 聚焦搜索框 |
| Ctrl+F | 聚焦搜索框 |
| Ctrl+T | 新建目录 |
| Ctrl+D | 标记 / 取消标记当前项 |
| Space | 标记 / 取消标记当前项 |
| Ctrl+R | 重命名当前项 |
| Ctrl+G | Ship |
| Esc | 见 §8.1 |
| Ctrl+P / ↑ | 上移（循环） |
| Ctrl+N / ↓ | 下移（循环） |
| Enter | 进入 / 打开 / 有标记时打开删除确认 |
| Tab / Shift+Tab | 切换来源（选择器内） |
| Delete | 文件视图删除标记项 |

### 8.1 Esc 语义（桌面 / 浏览器壳）

Esc 在选择器中按优先级从高到低处理：

1. 若行内输入打开：关闭行内输入。
2. 若对话框打开：取消对话框。
3. 若存在删除标记：清空标记，退出 DeleteMode。
4. 若搜索框非空：清空搜索并刷新列表。
5. 若搜索框为空：无进一步页面内动作。此时应用的关闭语义为关闭浏览器标签页 —— 浏览器壳无法通过页面内快捷键可靠地结束本机 `try-gui` 进程，用户关闭标签后由服务进程按自身生命周期退出（收到系统信号时优雅关闭）。

Esc 在文件视图按优先级：

1. 若存在文件删除标记：清空标记。
2. 若当前目录非根目录：返回上级目录。
3. 若已在根目录：返回选择器。

### 8.2 DeleteMode

- Ctrl-D 或 Space 标记 / 取消标记当前项。
- 存在标记即进入 DeleteMode；标记清空即退出。
- 有标记时 Enter（选择器）或 Delete（文件视图）打开删除确认。
- Esc 清空标记并退出 DeleteMode。

---

## 9. 响应式与可访问性

### 9.1 断点

| 宽度 | 调整 |
|---|---|
| < 800px | 隐藏评分/次要元数据列 |
| < 640px | 隐藏修改时间列 |
| < 560px | 来源 Tab 简化为短标签或图标 |

### 9.2 焦点

- 对话框打开时焦点陷阱在对话框内。
- Tab 顺序遵循视觉顺序。
- 可见焦点环。

### 9.3 辅助技术

- 列表：`listbox` / `option`。
- 图标按钮：`aria-label`。
- Toast：`role="status"`。
