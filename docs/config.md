# 配置文件

## 概述

try 支持通过配置文件持久化目录设置，避免每次都通过环境变量或命令行参数指定路径。配置文件为 JSON 格式，使用 Go 标准库 `encoding/json` 解析。

## 配置文件路径

`~/.config/try/config.json`（XDG 配置目录下的配置文件）。

## 文件格式

标准 JSON 格式：

```json
{
  "path": "~/src/tries",
  "ship": "~/src/ship",
  "theme": "dark",
  "locale": "zh"
}
```

解析规则：
- 路径值支持 `~` 前缀（解析为用户 home 目录）
- 未识别的 key 静默忽略（向前兼容）
- 未设置的字段保留默认值
- 文件不存在时静默使用默认值
- 空文件视为无配置（不告警）
- JSON 语法错误时输出 warning 到 stderr 并使用默认值

## 配置项

| key | 说明 | 可选值 | 默认值 |
|-----|------|--------|--------|
| `path` | tries 根目录（实验目录存放位置） | 任意路径 | `~/src/tries` |
| `ship` | ship 目标目录（发布为正式项目的目标位置） | 任意路径 | `~/src/ship` |
| `theme` | 配色主题 | `dark` / `light` / `auto` | `auto` |
| `locale` | 界面语言 | `en` / `zh` / `auto` | `auto` |

## 优先级

tries 路径（`path`）解析优先级：

```
1. --path 命令行参数（显式指定，最高优先）
2. TRY_PATH 环境变量
3. ~/.config/try/config.json 中的 path
4. 默认值 ~/src/tries
```

ship 目标目录（`ship`）解析优先级：

```
1. TRY_PROJECTS 环境变量
2. ~/.config/try/config.json 中的 ship
3. 默认值 ~/src/ship
```

主题（`theme`）解析优先级：

```
1. --theme 命令行参数（最高优先）
2. TRY_THEME 环境变量
3. ~/.config/try/config.json 中的 theme
4. auto（通过 COLORFGBG 推断终端亮暗，默认 dark）
```

语言（`locale`）解析优先级：

```
1. --locale 命令行参数（最高优先）
2. TRY_LOCALE 环境变量
3. ~/.config/try/config.json 中的 locale
4. auto（通过 LC_ALL/LC_MESSAGES/LANG 推断，默认 en）
```

所有路径最终展开为绝对路径（`~` → home 目录）。

## 解析实现

```go
// Config 配置结构体，JSON 字段名为小写
type Config struct {
    Path   string `json:"path"`   // tries 根目录
    Ship   string `json:"ship"`   // ship 目标目录
    Theme  string `json:"theme"`  // 主题：dark / light / auto
    Locale string `json:"locale"` // 语言：en / zh / auto
}

var defaultConfig = Config{
    Path:   "~/src/tries",
    Ship:   "~/src/ship",
    Theme:  "auto",
    Locale: "auto",
}

// LoadConfig 从 ~/.config/try/config.json 读取配置，合并默认值
func LoadConfig() Config {
    home, err := os.UserHomeDir()
    if err != nil {
        return defaultConfig
    }
    data, err := os.ReadFile(filepath.Join(home, ".config", "try", "config.json"))
    if err != nil {
        return defaultConfig
    }
    return parseConfigData(data)
}

// parseConfigData 解析 JSON 格式的配置内容，未设置的字段保留默认值。
// 空内容视为无配置（不告警），JSON 语法错误时输出 warning。
func parseConfigData(data []byte) Config {
    cfg := defaultConfig
    if len(data) == 0 {
        return cfg
    }
    if err := json.Unmarshal(data, &cfg); err != nil {
        fmt.Fprintf(os.Stderr, "try: 配置文件解析失败，使用默认值: %v\n", err)
    }
    return cfg
}
```

### 主题解析

```go
// ResolveTheme 按优先级解析主题
func ResolveTheme(cliTheme string, cfg Config) string {
    theme := cfg.Theme
    if env := os.Getenv("TRY_THEME"); env != "" {
        theme = env
    }
    if cliTheme != "" {
        theme = cliTheme
    }
    switch theme {
    case "light", "dark":
        return theme
    default:
        return detectTheme() // COLORFGBG 推断，默认 dark
    }
}
```

## 路径解析函数

`config` 包导出 `ResolvePaths`，按优先级合并所有来源的路径：

```go
// ResolvePaths 按优先级解析 tries 和 ship 路径
func ResolvePaths(cliPath string, cfg Config) (triesPath, shipPath string) {
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

    triesPath = ExpandPath(triesPath)
    shipPath = ExpandPath(shipPath)
    return
}
```

## 配置文件不存在时的行为

配置文件不存在不报错，静默使用默认值。这是最常见的场景 — 大多数用户直接用默认路径，不需要创建配置文件。

## 模块位置

`internal/config/config.go`
