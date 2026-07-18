# try

临时实验目录管理工具。快速创建、切换、搜索和清理实验性项目目录。

## 功能

- **模糊搜索** — 输入关键词即时过滤目录，按时间权重和匹配质量排序
- **一键创建** — 自动生成 `name-YYYY-MM-DD` 格式目录并 cd 进入
- **Git 集成** — 支持 clone 仓库和创建 worktree 到 tries 目录
- **批量删除** — 标记多个过期目录后一次性删除，含路径安全检查
- **重命名** — 就地重命名目录
- **Ship** — 将成熟的实验项目移动到正式项目目录
- **Shell 集成** — 通过 `eval` 机制实现跨 Shell 的 cd 操作
- **主题自适应** — 自动检测终端亮暗，适配 dark / light 配色
- **中英文界面** — 自动检测系统语言，可通过配置切换

## 安装

### 一键安装（推荐）

安装 **TUI（`try`）** 与 **GUI（`try-gui`）**（当前平台 Release 含 GUI 时）：

```bash
curl -fsSL https://raw.githubusercontent.com/loveloki/try/main/install.sh | sh
```

自定义安装路径：

```bash
TRY_INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/loveloki/try/main/install.sh | sh
```

仅安装 TUI：

```bash
TRY_INSTALL_GUI=0 curl -fsSL https://raw.githubusercontent.com/loveloki/try/main/install.sh | sh
```

脚本会将二进制放到 `~/.local/bin`（可用 `TRY_INSTALL_DIR` 覆盖），并自动执行 `try install` 完成 Shell 集成。若归档中无 `try-gui`（例如部分 Linux arm64 构建），会跳过 GUI 并提示源码安装方式。

### 从源码安装

```bash
go install github.com/loveloki/try/cmd/try@latest
CGO_ENABLED=1 go install github.com/loveloki/try/cmd/try-gui@latest
```

### Shell 集成

> 一键安装脚本已自动执行 `try install`，无需手动操作。

安装后运行一次 `try install` 设置 Shell 包装函数并自动创建默认配置文件，然后重启终端。支持 Bash、Zsh、Fish。

### 支持的平台

| OS | 架构 |
|----|------|
| Linux | amd64, arm64 |
| macOS | amd64 (Intel), arm64 (Apple Silicon) |
| Windows | amd64 |

## 使用

```bash
try                    # 打开选择器
try redis              # 模糊搜索 "redis"
try clone <url>        # 克隆 Git 仓库到 tries 目录
try clone <url> name   # 克隆并指定目录名
try <git-url>          # 自动识别 Git URL 并 clone
try worktree <dir> name # 从指定仓库创建 worktree
try . name             # 创建 worktree（有 .git）或普通目录
try install            # 安装 Shell 集成
try --help             # 查看完整帮助
try --version          # 查看版本号
```

## 快捷键

| 键 | 功能 |
|---|---|
| `Enter` | 选择目录 / 确认操作 |
| `↑` / `↓` | 上下移动（到边界后循环） |
| `Ctrl-T` | 创建新目录 |
| `Ctrl-D` | 标记/取消删除 |
| `Ctrl-R` | 重命名 |
| `Ctrl-G` | Ship（发布为正式项目） |
| `Tab` / `Shift-Tab` | 切换来源过滤（all / tries / ship / bug） |
| `Space` / `Delete` | 切换当前项删除标记 |
| `Ctrl-A` | 标记当前过滤结果全部条目 |
| `Ctrl-P` / `Ctrl-N` | 上下移动（到边界后循环） |
| `/` / `Ctrl-F` | 清空搜索框 |
| `Esc` | 退出 / 取消 |

## GUI（try-gui）

`try-gui` 是与 TUI 并列的跨平台原生桌面入口，读取同一份配置，复用相同的目录扫描、模糊匹配与文件操作逻辑。启动后显示原生应用窗口，默认尺寸约 900 × 600，最小尺寸约 720 × 480，内容区配色与快捷键与 TUI 对齐。

```bash
try-gui                # 启动 GUI
try-gui -path ~/src/tries  # 临时覆盖 tries 根目录
```

平台窗口规则：

- macOS：系统标题栏，关闭 / 最小化 / 最大化控件在左上角。
- Windows / Linux：WinUI3 风格标题栏，最小化 / 最大化 / 关闭控件在右上角。
- 内容区布局、配色、字体层级、快捷键、Selector、Files、对话框与 Toast 跨平台一致。
- 系统托盘提供 Show 与 Quit；关闭窗口时隐藏到托盘。

两大视图：

- **选择器**：搜索、来源过滤（all / tries / ship / bug）、循环导航、创建（Ctrl-T）、删除（Ctrl-D）、重命名（Ctrl-R）、Ship（Ctrl-G）。
- **文件视图**：进入目录后浏览文件、删除、调用系统默认程序打开文件；支持从 Finder/资源管理器拖入文件或文件夹复制到当前目录；Esc 返回选择器。

GUI 与 TUI 的差异：GUI 用「进入文件视图」替代 TUI 的 `cd` 脚本输出，不提供 clone / worktree / install 与 Shell 集成。所有文件操作限制在配置解析出的 tries 与 ship 目录子树内，且不允许删除或重命名根目录本身。

`try-gui` 使用 Fyne 构建原生桌面窗口。构建 GUI 产物需要 CGO 和目标平台图形依赖；CI 在 macOS、Windows、Linux runner 上分别原生编译（见 `.github/workflows/ci.yml`）。

```bash
# macOS / Linux（需本机图形开发头文件，Linux 见 spec/dependencies.md）
CGO_ENABLED=1 go build ./cmd/try-gui

# Windows（需 MinGW，且通常加 -H=windowsgui 避免控制台窗）
CGO_ENABLED=1 go build ./cmd/try-gui
```

## 配置

配置文件位于 `~/.config/try/config.json`，运行 `try install` 时自动生成，JSON 格式：

```json
{
  "path": "~/src/tries",
  "ships": ["~/src/ship", "~/src/bug"],
  "locale": "auto"
}
```

所有字段均可省略，未设置时使用默认值。各配置项的优先级均为：环境变量 > 配置文件 > 默认值。

配置文件解析失败 `try` 会报错退出。

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|----------|--------|------|
| `path` | `TRY_PATH` | `~/src/tries` | tries 根目录 |
| `ships` | `TRY_PROJECTS` | `["~/src/ship", "~/src/bug"]` | ship 目标目录列表 |
| `locale` | `TRY_LOCALE` | `auto` | 界面语言（`en` / `zh` / `auto`） |

## 项目结构

```
cmd/try/main.go              # TUI/CLI 入口
cmd/try-gui/main.go          # GUI 入口
internal/
  cli/                       # CLI 参数解析与命令分派
    cli.go                     # 主入口（Run）、全局标志解析
    commands.go                # 子命令实现（clone/worktree/exec/dot）
    flags.go                   # 标志提取工具
  config/                    # 配置文件解析（JSON）、主题/语言检测
  selector/                  # 交互式选择器（Bubbletea TUI）
    model.go                   # 核心状态与生命周期
    view.go                    # View 入口（主界面 / 对话框叠层）
    layout.go                  # Header / Footer / 空状态布局
    styles.go                  # 主题 token 与 Lipgloss 样式
    icons.go                   # 列表与空状态图标常量
    delegate.go                # 列表条目自定义渲染
    keyhandler.go              # 按键分派（含循环导航）
    keys.go                    # 按键绑定定义
    entry.go                   # 条目类型与选择结果
    dialogs.go                 # 对话框实例接口与工厂模式
    overlay.go                 # 模态弹窗合成
    loader.go                  # 目录扫描、模糊匹配与列表刷新
  dialog/                    # 对话框实现（删除/重命名/Ship）
    dialog.go                    # Dialog 接口定义
    delete.go / rename.go / ship.go
    styles.go / icons.go / modal.go
  fuzzy/                     # 模糊匹配引擎（时间权重 + 子序列评分）
  i18n/                      # 国际化（中英文界面文本）
  script/                    # 脚本生成（cd 指令）与操作执行
    script.go                    # EmitCd / Quote
    exec.go                      # Execute（cd/mkdir/clone/worktree/delete/rename/ship）
  shell/                     # Shell 检测与集成安装（bash/zsh/fish）
  git/                       # Git URI 解析与目录命名
  gui/                       # GUI：Fyne 原生窗口 + 系统托盘
    app.go                     # 生命周期：加载配置、创建窗口、托盘
    state.go / update.go       # GUI 状态机与键鼠输入处理
    service.go                 # entries / files / 副作用操作
    paths.go                   # 路径沙箱（拒绝越界与根目录本身）
    opener.go                  # 跨平台打开文件
    chrome_*.go                # 平台标题栏策略
    view_*.go                  # Selector / Files / Dialog / Toast
    theme.go                   # GUI 主题 token 映射
```

## 技术栈

- Go（单二进制，零运行时依赖）
- [Bubbletea v2](https://charm.land/bubbletea) — TUI 框架（Elm Architecture）
- [Bubbles v2](https://charm.land/bubbles) — TUI 组件库
- [Lipgloss v2](https://charm.land/lipgloss) — 终端样式
- [Fyne v2](https://fyne.io/) — 原生桌面 GUI 框架
- [sahilm/fuzzy](https://github.com/sahilm/fuzzy) — 子序列匹配

## 开发

```bash
go build ./cmd/try  # 构建二进制
CGO_ENABLED=1 go build ./cmd/try-gui # 构建 GUI
go test ./...       # 运行所有测试
go vet ./...        # 官方静态检查
staticcheck ./...   # 第三方静态检查（需安装：go install honnef.co/go/tools/cmd/staticcheck@latest）
```

## 发布

一键发布新版本（自动推断版本号、运行测试、创建 tag、推送 tag 触发 Actions）：

```bash
./scripts/release.sh         # 自动推断版本（根据 commit 历史）
./scripts/release.sh patch   # 补丁版本 (x.y.Z)
./scripts/release.sh minor   # 次版本 (x.Y.0)
./scripts/release.sh major   # 主版本 (X.0.0)
```

依赖 [svu](https://github.com/caarlos0/svu)（`go install github.com/caarlos0/svu@latest`）。推送 `v*` tag 时运行 CI（lint/test/build）与 Release（分平台归档）；推送 `main` 或 PR 不触发 Actions。

Release（`.github/workflows/release.yml`）在 Linux / macOS / Windows runner 上分平台原生构建：`try` 使用 `CGO_ENABLED=0`，`try-gui` 使用 `CGO_ENABLED=1`，打入同一归档 `try_<os>_<arch>.tar.gz`（Windows 为 `.zip`）供 `install.sh` 下载。Linux arm64 仅打包 `try`（无可靠 CGO 交叉编 GUI）。

## 致谢

本项目受 [tobi/try](https://github.com/tobi/try) 启发。
