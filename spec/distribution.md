# 分发与安装

本文描述 `try` / `try-gui` 的 Release 资产形态与安装路径。行为以当前实现为准。

## 双轨资产

每个 GitHub Release（`v*` tag）同时提供两类资产，校验和写入 `checksums.txt`。

### CLI 轨（裸二进制）

| 文件 | 内容 |
|------|------|
| `try_linux_amd64.tar.gz` | `try` + `try-gui` |
| `try_linux_arm64.tar.gz` | 仅 `try` |
| `try_darwin_amd64.tar.gz` | `try` + `try-gui` |
| `try_darwin_arm64.tar.gz` | `try` + `try-gui` |
| `try_windows_amd64.zip` | `try.exe` + `try-gui.exe` |

用途：终端 TUI、`install.sh` / `install.ps1` 安装 CLI、高级用户裸跑。Windows 裸 `try-gui.exe` 以 `-H=windowsgui` 构建。

### GUI 轨（`fyne package` 官方包）

由 `scripts/package-gui.sh` 在 CI 中对已构建的 `try-gui` 调用 `fyne package` 生成。元数据真源为 `cmd/try-gui/FyneApp.toml`（`ID = com.loveloki.try.gui`，`Name = Try`）与 `cmd/try-gui/Icon.png`。

| 文件 | 形态 |
|------|------|
| `try-gui_darwin_amd64.app.zip` | `Try.app` bundle |
| `try-gui_darwin_arm64.app.zip` | `Try.app` bundle |
| `try-gui_windows_amd64.zip` | 带图标元数据的 `try-gui.exe` |
| `try-gui_linux_amd64.tar.gz` | Fyne `usr/local/` 安装树（bin + `.desktop` + 图标） |

不产出 `try-gui_linux_arm64.*`（与 CI `gui: false` 一致）。

应用标识符：`com.loveloki.try.gui`（与 `internal/gui.AppID` / `app.NewWithID` 一致）。显示名为 **Try**。

## 手动安装（不依赖脚本）

### macOS

1. 下载 `try-gui_darwin_<arch>.app.zip`
2. 解压得到 `Try.app`
3. 拖入「应用程序」或 `~/Applications`
4. 若 Gatekeeper 拦截：右键 → 打开

CLI：下载 `try_darwin_<arch>.tar.gz`，将 `try` 放入 PATH，运行 `try install`。

### Windows

1. 下载 `try-gui_windows_amd64.zip`，解压后运行 `try-gui.exe`
2. 可选：自行创建开始菜单快捷方式

CLI：下载 `try_windows_amd64.zip`，将目录加入 PATH，运行 `try install`（Shell 包装在 Windows 上可能不可用，配置文件仍会初始化）。

### Linux amd64

1. 下载 `try-gui_linux_amd64.tar.gz`
2. 解压到 `/usr/local`（通常需 sudo）：`sudo tar -xzf try-gui_linux_amd64.tar.gz -C /`
3. 或按需安装到自定义前缀（保持 `usr/local` 相对布局）

CLI：下载 `try_linux_<arch>.tar.gz`，将 `try` 放入 PATH，运行 `try install`。

## 脚本安装（适配器）

脚本消费上述 Release 资产，**不是**唯一交付物。

### `install.sh`（Linux / macOS）

| 步骤 | 行为 |
|------|------|
| CLI | 下载 `try_<os>_<arch>.tar.gz` → `$TRY_INSTALL_DIR`（默认 `~/.local/bin`） |
| GUI（默认开，`TRY_INSTALL_GUI=0` 跳过） | 下载对应 `try-gui_*` 官方包 |
| Linux GUI | 适配到 `~/.local/bin/try-gui` + `~/.local/share/applications/try-gui.desktop` + hicolor 图标 |
| macOS GUI | 解压到 `$TRY_APPS_DIR/Try.app`（默认 `~/Applications`），并 symlink `try-gui` 到 `.app` 内二进制 |
| Shell | 调用 `try install`（仅 Shell 包装 + 配置；**不**写桌面入口） |

### `install.ps1`（Windows）

| 步骤 | 行为 |
|------|------|
| CLI | 下载 `try_windows_amd64.zip` → `%LOCALAPPDATA%\Programs\try\` |
| GUI | 下载 `try-gui_windows_amd64.zip` → 同目录，创建开始菜单 `Try.lnk` |
| PATH | 将安装目录加入用户 PATH |
| 配置 | 调用 `try install` |

## `try install` 职责

`try install` 只负责：

1. 初始化 `~/.config/try/config.json`（若不存在）
2. 将 Shell 包装函数写入 bash/zsh/fish 配置

不负责桌面入口、`.app`、开始菜单快捷方式。桌面集成由 `install.sh` / `install.ps1` 或用户手动安装官方包完成。

## 构建链

1. `release.yml` 分平台 `go build`：`try`（`CGO_ENABLED=0`）、`try-gui`（`CGO_ENABLED=1`）
2. `gui: true` 的矩阵行安装 `fyne` CLI，运行 `scripts/package-gui.sh`
3. 上传 CLI 归档与 GUI 官方包，汇总 checksums 后发布

本地复现 GUI 官方包：

```bash
CGO_ENABLED=1 go build -o dist/try-gui ./cmd/try-gui
go install fyne.io/tools/cmd/fyne@latest
./scripts/package-gui.sh darwin arm64 3.2.0
```
