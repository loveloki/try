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
  "ships": ["~/src/ship", "~/src/bug"],
  "locale": "zh",
  "theme": "dark",
  "openWith": {
    ".go": "Visual Studio Code",
    ".md": "Typora"
  }
}
```

解析规则：
- 路径值支持 `~` 前缀（解析为用户 home 目录）
- 未识别的 key 静默忽略（向前兼容）
- 未设置的字段保留默认值
- 文件不存在、空文件或 JSON 语法错误时 `LoadConfig` 返回 error
- 配置文件由 `try install` 自动创建，正常情况下应当存在

## 配置项

| key | 说明 | 可选值 | 默认值 |
|-----|------|--------|--------|
| `path` | tries 根目录（实验目录存放位置） | 任意路径 | `~/src/tries` |
| `ships` | ship 目标目录列表（发布为正式项目的目标位置） | 路径数组 | `["~/src/ship", "~/src/bug"]` |
| `locale` | 界面语言 | `en` / `zh` / `auto` | `auto` |
| `theme` | GUI 主题（仅 GUI 使用） | `dark` / `light` / `auto` | `auto` |
| `openWith` | 文件扩展名→应用名映射（GUI 右键菜单；可在设置对话框中编辑）。应用名支持内置应用名、macOS 应用名（自动补 `.app`）、PATH 可执行名或绝对路径。特殊键 `*` 为通用映射：无精确扩展名命中时生效（含无扩展名文件与目录） | 扩展名→应用名 JSON 对象 | `{}` |

## 优先级

tries 路径（`path`）解析优先级：

```
1. TRY_PATH 环境变量
2. ~/.config/try/config.json 中的 path
3. 默认值 ~/src/tries
```

ship 目标目录列表（`ships`）解析优先级：

```
1. TRY_PROJECTS 环境变量（单路径，转为单元素数组）
2. `~/.config/try/config.json` 中的 ships
3. 默认值 ["~/src/ship", "~/src/bug"]
```

语言（`locale`）解析优先级：

```
1. TRY_LOCALE 环境变量
2. ~/.config/try/config.json 中的 locale
3. auto：LC_ALL > LC_MESSAGES > LANG；均空时回退操作系统语言（默认 en）
```

TUI 与 GUI 读取同一 `config.json` 与同一套 `ResolveLocale`。`locale: "auto"` 时，从终端启动通常有 `LANG`；从 Dock / `.app` / 开始菜单启动时常无 locale 环境变量，此时回退 OS 语言，使两端在显式未配置时尽量一致。需要强制一致时可设 `"locale": "zh"` 或 `"en"`。

所有路径最终展开为绝对路径（`~` → home 目录）。

## 类型定义

```go
type Config struct {
    Path     string            `json:"path"`     // tries 根目录
    Ships    []string          `json:"ships"`    // ship 目标目录列表
    Locale   string            `json:"locale"`   // 语言：en / zh / auto
    Theme    string            `json:"theme"`    // GUI 主题：dark / light / auto
    OpenWith map[string]string `json:"openWith"` // 文件扩展名→应用名映射（GUI）
}
```

默认值：`Path="~/src/tries"`, `Ships=["~/src/ship", "~/src/bug"]`, `Locale="auto"`, `Theme="auto"`, `OpenWith=nil`。

## 导出函数

```go
func LoadConfig() (Config, error)
func InitConfigFile() (bool, error)
func ResolvePaths(cliPath string, cfg Config) (triesPath string, shipPaths []string)
func DetectTheme() string
func SaveTheme(themeName string) error
func SaveLocale(locale string) error
func SaveOpenWith(openWith map[string]string) error
func ResolveLocale(cliLocale string, cfg Config) string
func ExpandPath(s string) string
```

### 行为规格

- `LoadConfig`：从 `~/.config/try/config.json` 读取。文件不存在、空文件或 JSON 语法错误时返回 error。配置文件由 `try install` 自动初始化，正常情况下应当存在。
- `InitConfigFile`：在 `~/.config/try/config.json` 创建默认配置文件（如果不存在）。文件已存在时返回 `(false, nil)`，新创建时返回 `(true, nil)`。在 `try install` 命令中自动调用。
- `ResolvePaths`：按优先级链合并 tries 和 ships 路径，最终展开为绝对路径。启动时自动创建所有 ship 目录。
- `DetectTheme`：通过 `COLORFGBG` 环境变量推断终端亮暗（背景色 `7`/`15` 判定为 light），无法推断时默认 dark。
- `SaveTheme`：将主题偏好写入 `~/.config/try/config.json`，保留其他字段不变。
- `SaveLocale`：将语言设置（`en` / `zh` / `auto`）写入 `~/.config/try/config.json`，保留其他字段不变。
- `SaveOpenWith`：将文件扩展名→应用名映射写入 `~/.config/try/config.json`，保留其他字段不变。

`SaveTheme`、`SaveLocale`、`SaveOpenWith` 统一经由私有 `updateConfig(mutate func(*Config))` 完成：先 `LoadConfig` 加载现有配置，应用 `mutate` 修改后由 `writeConfig` 写回，未涉及的字段原样保留。

- `ResolveLocale`：auto 模式通过 `LC_ALL` > `LC_MESSAGES` > `LANG` 推断语言（以 `zh` 开头时为中文）；环境变量均空时回退操作系统语言，默认 en。
- `ExpandPath`：展开 `~` 为用户 home 目录。

## 配置文件初始化

`try install` 命令会自动在 `~/.config/try/config.json` 创建默认配置文件。所有其他命令（`clone`, `worktree`, `exec`, 搜索选择器）在执行前要求配置文件必须存在且内容合法。

## 模块位置

`internal/config/config.go`
