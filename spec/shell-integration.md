# Shell 集成

## 概述

try 通过生成 Shell 包装函数实现 `cd` 等需要改变父进程状态的操作。包装函数捕获 try 的 stdout 输出并 `eval` 执行。TUI 渲染通过 stderr（重定向到 `/dev/tty`）直接呈现给用户。

## stdout/stderr 分离方案

```
┌─ Shell 包装函数 ─────────────────────────┐
│  out=$(try exec "$@" 2>/dev/tty)         │
│  eval "$out"                              │
└───────────────────────────────────────────┘
         │                    │
    stdout（管道捕获）    stderr → /dev/tty
    cd 脚本              TUI 渲染（用户可见）
```

关键点：
- `2>/dev/tty` 将 stderr 直接连接到终端，用户看到 TUI
- stdout 被 `$()` 捕获为字符串
- `eval "$out"` 在父 Shell 中执行捕获的 cd 命令
- 退出码 0 才执行，非零（取消/错误）不 eval

## 可扩展架构

Shell 类型和操作封装为注册表模式，新增 Shell 只需添加一项配置：

```go
type ShellConfig struct {
    Name     string
    RCFile   func() string
    InitFunc func(binaryPath string) string
}

var Shells = map[string]ShellConfig{
    "bash": {Name: "bash", RCFile: bashRCFile, InitFunc: posixInit},
    "zsh":  {Name: "zsh",  RCFile: zshRCFile,  InitFunc: posixInit},
    "fish": {Name: "fish", RCFile: fishRCFile,  InitFunc: fishInit},
}
```

bash 和 zsh 共用 `posixInit`（POSIX 兼容语法），fish 单独实现。未来扩展新 Shell（如 nushell、PowerShell）只需在注册表中添加一项。

## Shell 检测

```go
func DetectShell() string
```

检测优先级：`$SHELL` 环境变量 → 父进程名（`/proc/$PPID/comm`）。按 fish → zsh → bash 顺序匹配（`strings.Contains`）。均未匹配则返回空字符串。

## 包装函数模板

`try install` 使用以下模板生成包装函数并写入配置文件。

**Bash/Zsh（posixInit）：**

```bash
# try shell integration
try() {
  local out
  out=$('/path/to/try' exec "$@" 2>/dev/tty)
  if [ $? -eq 0 ]; then
    eval "$out"
  else
    echo "$out"
  fi
}
```

**Fish（fishInit）：**

```fish
# try shell integration
function try
  set -l out ('/path/to/try' exec $argv 2>/dev/tty | string collect)
  if test $pipestatus[1] -eq 0
    eval $out
  else
    echo $out
  end
end
```

### 路径嵌入逻辑

- `binaryPath`：try 二进制的绝对路径（`os.Executable()` 解析符号链接后的路径），用 `quote()` 包裹

包装函数不嵌入 `--path`。路径解析由 try 内部按优先级处理：`--path` 参数 > 环境变量 > `~/.config/try/config.json` > 默认值（详见 `config.md`）。

## try install

`cmdInstall` 自动将包装函数追加到 Shell 配置文件。

### Shell 配置文件映射

| Shell | 配置文件 |
|-------|---------|
| fish | `~/.config/fish/config.fish` |
| zsh | `~/.zshrc` |
| bash | `~/.bashrc`（不存在则回退 `~/.bash_profile`） |

### 安装流程

```
1. detectShell → 确定 Shell 类型
2. 从注册表获取 ShellConfig → RCFile()
3. 检查是否已安装（搜索 "# try shell integration" 标记）
4. 检查文件写权限
5. 追加包装函数（带 "# try shell integration" 标记）
6. 提示用户重启 Shell 或 source 配置文件
```

### 安全检查

- 已安装时提示用户先手动移除旧版再重装
- 文件只读时提示手动操作
- 配置文件父目录不存在时自动创建（`os.MkdirAll`）
