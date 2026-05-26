# 配置文件

## 概述

try 支持通过配置文件持久化目录设置，避免每次都通过环境变量或命令行参数指定路径。配置文件为简单的 key=value 格式，无需第三方解析库。

## 配置文件路径

`~/.try`（用户 home 目录下的 `.try` 文件）。

## 文件格式

每行一个配置项，`key = value` 格式。`#` 开头为注释行，空行忽略。

```
# try 配置文件
path = ~/src/tries
ship = ~/src/ship
```

解析规则：
- key 和 value 两侧的空白自动去除
- value 支持 `~` 前缀（解析为用户 home 目录）
- 重复 key 以最后一个为准
- 未识别的 key 静默忽略（向前兼容）

## 配置项

| key | 说明 | 默认值 |
|-----|------|--------|
| `path` | tries 根目录（实验目录存放位置） | `~/src/tries` |
| `ship` | ship 目标目录（发布为正式项目的目标位置） | `~/src/ship` |

## 优先级

tries 路径（`path`）解析优先级：

```
1. --path 命令行参数（显式指定，最高优先）
2. TRY_PATH 环境变量
3. ~/.try 配置文件中的 path
4. 默认值 ~/src/tries
```

ship 目标目录（`ship`）解析优先级：

```
1. TRY_PROJECTS 环境变量
2. ~/.try 配置文件中的 ship
3. 默认值 ~/src/ship
```

所有路径最终展开为绝对路径（`~` → home 目录）。

## 解析实现

```go
// 配置结构体，字段对应配置文件中的 key
type Config struct {
    Path string // tries 根目录
    Ship string // ship 目标目录
}

// 默认值
var defaultConfig = Config{
    Path: "~/src/tries",
    Ship: "~/src/ship",
}

// LoadConfig 从 ~/.try 读取配置，合并默认值
func LoadConfig() Config {
    cfg := defaultConfig

    home, err := os.UserHomeDir()
    if err != nil {
        return cfg
    }

    data, err := os.ReadFile(filepath.Join(home, ".try"))
    if err != nil {
        return cfg
    }

    for _, line := range strings.Split(string(data), "\n") {
        line = strings.TrimSpace(line)
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        k, v, ok := strings.Cut(line, "=")
        if !ok {
            continue
        }
        key := strings.TrimSpace(k)
        value := strings.TrimSpace(v)
        switch key {
        case "path":
            cfg.Path = value
        case "ship":
            cfg.Ship = value
        }
    }

    return cfg
}
```

## 路径解析函数

CLI 层负责按优先级合并所有来源的路径：

```go
// resolvePaths 按优先级解析 tries 和 ship 路径
func resolvePaths(cliPath string, cfg Config) (triesPath, shipPath string) {
    // tries 路径：--path > TRY_PATH > config > default
    triesPath = cfg.Path
    if env := os.Getenv("TRY_PATH"); env != "" {
        triesPath = env
    }
    if cliPath != "" {
        triesPath = cliPath
    }

    // ship 路径：TRY_PROJECTS > config > default
    shipPath = cfg.Ship
    if env := os.Getenv("TRY_PROJECTS"); env != "" {
        shipPath = env
    }

    triesPath = expandPath(triesPath)
    shipPath = expandPath(shipPath)
    return
}
```

## 配置文件不存在时的行为

配置文件不存在不报错，静默使用默认值。这是最常见的场景 — 大多数用户直接用默认路径，不需要创建配置文件。

## 模块位置

`internal/config/config.go`
