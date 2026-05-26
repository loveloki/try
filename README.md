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
- **主题支持** — dark / light 配色主题，支持自动检测终端亮暗
- **中英文界面** — 自动检测系统语言，可通过配置切换

## 安装

### 一键安装（推荐）

```bash
curl -fsSL https://raw.githubusercontent.com/xleine/try/main/install.sh | sh
```

自动检测 OS 和架构，下载对应预编译二进制。自定义安装路径：

```bash
TRY_INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/xleine/try/main/install.sh | sh
```

### 从源码安装

```bash
go install github.com/xleine/try/cmd/try@latest
```

### Shell 集成

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
try worktree . name    # 从当前仓库创建 worktree
try . name             # 创建 worktree 或目录（简写）
try install            # 安装 Shell 集成
try --help             # 查看完整帮助
```

## 快捷键

| 键 | 功能 |
|---|---|
| `Enter` | 选择目录 / 确认操作 |
| `Ctrl-T` | 创建新目录 |
| `Ctrl-D` | 标记/取消删除 |
| `Ctrl-R` | 重命名 |
| `Ctrl-G` | Ship（发布为正式项目） |
| `Ctrl-P` / `Ctrl-N` | 上下移动 |
| `Esc` | 退出 / 取消 |

## 配置

配置文件位于 `~/.config/try/config.json`，JSON 格式：

```json
{
  "path": "~/src/tries",
  "ship": "~/src/ship",
  "theme": "auto",
  "locale": "auto"
}
```

所有字段均可省略，未设置时使用默认值。各配置项的优先级均为：命令行参数 > 环境变量 > 配置文件 > 默认值。

| 配置项 | 环境变量 | 命令行参数 | 默认值 | 说明 |
|--------|----------|-----------|--------|------|
| `path` | `TRY_PATH` | `--path` | `~/src/tries` | tries 根目录 |
| `ship` | `TRY_PROJECTS` | — | `~/src/ship` | ship 目标目录 |
| `theme` | `TRY_THEME` | `--theme` | `auto` | 配色主题（`dark` / `light` / `auto`） |
| `locale` | `TRY_LOCALE` | `--locale` | `auto` | 界面语言（`en` / `zh` / `auto`） |

## 项目结构

```
cmd/try/main.go          # 入口
internal/
  cli/                   # CLI 参数解析与命令分派
  config/                # 配置文件解析（JSON）
  selector/              # 交互式选择器（Bubbletea TUI）
  dialog/                # 对话框（删除/重命名/Ship）
  fuzzy/                 # 模糊匹配引擎
  i18n/                  # 国际化（中英文界面文本）
  script/                # 脚本生成与文件操作
  shell/                 # Shell 检测与集成安装
  git/                   # Git URI 解析与目录命名
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

打 tag 触发 GitHub Actions 自动构建发布：

```bash
git tag v0.1.0
git push origin v0.1.0
```

GoReleaser 会构建全平台二进制并创建 GitHub Release。
