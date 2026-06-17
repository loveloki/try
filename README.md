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

```bash
curl -fsSL https://raw.githubusercontent.com/loveloki/try/main/install.sh | sh
```

自定义安装路径：

```bash
TRY_INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/loveloki/try/main/install.sh | sh
```

### 从源码安装

```bash
go install github.com/loveloki/try/cmd/try@latest
```

### Shell 集成

> 一键安装脚本已自动执行 `try install`，无需手动操作。

安装后运行一次 `try install` 设置 Shell 包装函数，然后重启终端。支持 Bash、Zsh、Fish。

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
| `Ctrl-T` | 创建新目录 |
| `Ctrl-D` | 标记/取消删除 |
| `Ctrl-R` | 重命名 |
| `Ctrl-G` | Ship（发布为正式项目） |
| `Tab` | 切换来源过滤（all / tries / ship / bug） |
| `Ctrl-P` / `Ctrl-N` | 上下移动 |
| `Esc` | 退出 / 取消 |

## 配置

配置文件位于 `~/.config/try/config.json`，JSON 格式：

```json
{
  "path": "~/src/tries",
  "ships": ["~/src/ship", "~/src/bug"],
  "locale": "auto"
}
```

所有字段均可省略，未设置时使用默认值。各配置项的优先级均为：环境变量 > 配置文件 > 默认值。

| 配置项 | 环境变量 | 默认值 | 说明 |
|--------|----------|--------|------|
| `path` | `TRY_PATH` | `~/src/tries` | tries 根目录 |
| `ships` | `TRY_PROJECTS` | `["~/src/ship", "~/src/bug"]` | ship 目标目录列表 |
| `locale` | `TRY_LOCALE` | `auto` | 界面语言（`en` / `zh` / `auto`） |

## 项目结构

```
cmd/try/main.go              # 入口
internal/
  cli/                       # CLI 参数解析与命令分派
    cli.go                     # 主入口（Run）、全局标志解析
    commands.go                # 子命令实现（clone/worktree/exec/dot）
    flags.go                   # 标志提取工具
  config/                    # 配置文件解析（JSON）、主题/语言检测
  selector/                  # 交互式选择器（Bubbletea TUI）
    model.go                   # 核心状态与生命周期
    view.go                    # 渲染（标题/搜索/来源标签/状态栏）
    delegate.go                # 列表条目自定义渲染
    keyhandler.go              # 按键分派
    keys.go                    # 按键绑定定义
    entry.go                   # 条目类型与选择结果
    dialogs.go                 # 对话框实例接口与工厂模式
    overlay.go                 # 模态弹窗合成
    loader.go                  # 目录扫描、模糊匹配与列表刷新
  dialog/                    # 对话框实现（删除/重命名/Ship）
    dialog.go                    # Dialog 接口定义
    delete.go / delete_styles.go # 删除确认弹窗
    rename.go                    # 重命名输入弹窗
    ship.go                      # Ship 目标选择弹窗
    modal.go                     # 弹窗盒子渲染工具
  fuzzy/                     # 模糊匹配引擎（时间权重 + 子序列评分）
  i18n/                      # 国际化（中英文界面文本，10 组 62 个字段）
  script/                    # 脚本生成（cd 指令）与操作执行
    script.go                    # EmitCd / Quote
    exec.go                      # Execute（cd/mkdir/clone/worktree/delete/rename/ship）
  shell/                     # Shell 检测与集成安装（bash/zsh/fish）
  git/                       # Git URI 解析与目录命名
```

## 技术栈

- Go（单二进制，零运行时依赖）
- [Bubbletea v2](https://charm.land/bubbletea) — TUI 框架（Elm Architecture）
- [Bubbles v2](https://charm.land/bubbles) — TUI 组件库
- [Lipgloss v2](https://charm.land/lipgloss) — 终端样式
- [sahilm/fuzzy](https://github.com/sahilm/fuzzy) — 子序列匹配

## 开发

```bash
go build ./cmd/try  # 构建二进制
go test ./...       # 运行所有测试
go vet ./...        # 官方静态检查
staticcheck ./...   # 第三方静态检查（需安装：go install honnef.co/go/tools/cmd/staticcheck@latest）
```

## 发布

一键发布新版本（自动推断版本号、运行测试、创建 tag、推送触发 CI）：

```bash
./scripts/release.sh         # 自动推断版本（根据 commit 历史）
./scripts/release.sh patch   # 补丁版本 (x.y.Z)
./scripts/release.sh minor   # 次版本 (x.Y.0)
./scripts/release.sh major   # 主版本 (X.0.0)
```

依赖 [svu](https://github.com/caarlos0/svu)（`go install github.com/caarlos0/svu@latest`）。GoReleaser 会构建全平台二进制并创建 GitHub Release。

## 致谢

本项目受 [tobi/try](https://github.com/tobi/try) 启发。
