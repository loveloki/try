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

## 类型定义

```go
type Config struct {
    Path   string `json:"path"`   // tries 根目录
    Ship   string `json:"ship"`   // ship 目标目录
    Theme  string `json:"theme"`  // 主题：dark / light / auto
    Locale string `json:"locale"` // 语言：en / zh / auto
}
```

默认值：`Path="~/src/tries"`, `Ship="~/src/ship"`, `Theme="auto"`, `Locale="auto"`。

## 导出函数

```go
func LoadConfig() Config
func ResolvePaths(cliPath string, cfg Config) (triesPath, shipPath string)
func ResolveTheme(cliTheme string, cfg Config) string
func ResolveLocale(cliLocale string, cfg Config) string
func ExpandPath(s string) string
```

### 行为规格

- `LoadConfig`：从 `~/.config/try/config.json` 读取。文件不存在时返回默认值。空文件视为无配置。JSON 语法错误时输出 warning 到 stderr 并返回默认值。
- `ResolvePaths`：按优先级链合并 tries 和 ship 路径，最终展开为绝对路径。
- `ResolveTheme`：auto 模式通过 `COLORFGBG` 环境变量推断终端亮暗（背景色值 0-6 判定为 light），无法推断时默认 dark。
- `ResolveLocale`：auto 模式通过 `LC_ALL` > `LC_MESSAGES` > `LANG` 推断语言（以 `zh` 开头时为中文），默认 en。
- `ExpandPath`：展开 `~` 为用户 home 目录。

## 配置文件不存在时的行为

配置文件不存在不报错，静默使用默认值。这是最常见的场景 — 大多数用户直接用默认路径，不需要创建配置文件。

## 模块位置

`internal/config/config.go`
