# Git 集成

## 概述

try 提供三种 Git 操作：仓库克隆（clone）、创建 worktree、URL 快捷识别。相关实现位于 `internal/git/`。

## 函数接口

| 函数 | 返回值 |
|------|--------|
| `ParseGitURI(uri)` | `*GitURIInfo` 或 `nil` |
| `IsGitURI(arg)` | `bool`（不做完整解析，仅判断"看起来像"Git URL） |
| `GenerateCloneDirName(uri, customName)` | `string`（目录名）或 `""` |

```go
type GitURIInfo struct {
    User string
    Repo string
    Host string
}
```

## Git URI 解析

### ParseGitURI

两组通用正则覆盖所有 Git 托管平台（GitHub、GitLab、自建等）：

```go
var (
    httpsRe = regexp.MustCompile(`^https?://([^/]+)/([^/]+)/([^/]+)`)
    sshRe   = regexp.MustCompile(`^git@([^:]+):([^/]+)/([^/]+)`)
)

func ParseGitURI(uri string) *GitURIInfo {
    uri = strings.TrimSuffix(uri, ".git")

    if m := httpsRe.FindStringSubmatch(uri); m != nil {
        return &GitURIInfo{Host: m[1], User: m[2], Repo: m[3]}
    }
    if m := sshRe.FindStringSubmatch(uri); m != nil {
        return &GitURIInfo{Host: m[1], User: m[2], Repo: m[3]}
    }
    return nil
}
```

| 格式 | 正则 | 匹配示例 |
|------|------|---------|
| HTTPS | `^https?://([^/]+)/([^/]+)/([^/]+)` | `https://github.com/user/repo`、`https://gitlab.company.com/team/project` |
| SSH | `^git@([^:]+):([^/]+)/([^/]+)` | `git@github.com:user/repo`、`git@gitlab.company.com:team/project` |

### IsGitURI

```go
func IsGitURI(arg string) bool {
    if arg == "" { return false }
    return strings.HasPrefix(arg, "https://") ||
           strings.HasPrefix(arg, "http://") ||
           strings.HasPrefix(arg, "git@") ||
           strings.Contains(arg, "github.com") ||
           strings.Contains(arg, "gitlab.com") ||
           strings.HasSuffix(arg, ".git")
}
```

满足任一条件即识别为 Git URL。

## Clone 命令

### 入口

```
try clone <url> [name]      # 显式 clone 命令
try <url> [name]            # URL 快捷方式（自动识别）
try exec clone <url> [name] # 包装函数内部调用
```

三种入口最终都调用 `cmdClone`。

### 目录命名

```go
func GenerateCloneDirName(gitURI, customName string) string {
    if customName != "" { return customName }

    parsed := ParseGitURI(gitURI)
    if parsed == nil { return "" }

    dateSuffix := time.Now().Format("2006-01-02")
    return parsed.User + "-" + parsed.Repo + "-" + dateSuffix
}
```

自定义名称优先。自动命名格式：`user-repo-YYYY-MM-DD`。名称在前，提高模糊匹配命中效率。

示例：
- `try clone https://github.com/tobi/try.git` → `tobi-try-2025-08-17`
- `try clone https://github.com/tobi/try.git my-fork` → `my-fork`

### 执行流程

最终目录路径为 `triesPath/dirName`。Go 直接执行副作用操作（mkdir + git clone），仅输出 `cd` 脚本到 stdout。详见 `script-generation.md` 中的 `execClone`。

## Worktree 命令

### 入口

```
try worktree <dir> [name]   # 显式 worktree 命令
try . <name>                # 点号快捷方式（name 必需）
try ./path [name]           # 指定仓库路径
```

### worktreePath 计算

```go
func worktreePath(triesPath, repoDir, customName string) string {
    var base string
    if customName != "" {
        base = strings.ReplaceAll(customName, " ", "-")
    } else {
        base = filepath.Base(repoDir)
    }
    dateSuffix := time.Now().Format("2006-01-02")
    base = git.ResolveUniqueName(triesPath, base, dateSuffix)
    return filepath.Join(triesPath, base + "-" + dateSuffix)
}
```

worktree 总是创建在 tries 根目录下，以日期后缀命名。

### "." 命令处理

```
try . name     → 使用当前目录作为仓库
try ./path name → 使用指定路径作为仓库
try .          → 报错退出（防止误触）
```

关键逻辑：
1. `try .` 无名称参数 → 报错退出
2. 检查仓库路径下是否有 `.git`（文件或目录）
3. 有 `.git` → 调用 `execWorktree`（Go 直接创建 detached HEAD worktree）
4. 无 `.git` → 调用 `execMkdir`（Go 直接创建普通目录）

### worktree 执行

Go 直接执行 worktree 创建，不生成 Shell 脚本。详见 `script-generation.md` 中的 `execWorktree`：

- `git rev-parse --show-toplevel` 获取仓库根目录（处理在子目录中执行的情况）
- `--detach` 创建 detached HEAD worktree（不关联分支）
- worktree 创建失败不阻塞，仍然 cd 到目标目录

## 目录名版本化

当目标目录已存在时，自动递增后缀：

```go
func ResolveUniqueName(triesPath, base, dateSuffix string) string {
    initial := base + "-" + dateSuffix
    if !dirExists(filepath.Join(triesPath, initial)) { return base }

    // 尾部是数字（日期前的部分）：递增数字
    if m := regexp.MustCompile(`^(.*?)(\d+)$`).FindStringSubmatch(base); m != nil {
        stem := m[1]
        n, _ := strconv.Atoi(m[2])
        for {
            n++
            candidate := stem + strconv.Itoa(n)
            if !dirExists(filepath.Join(triesPath, candidate+"-"+dateSuffix)) {
                return candidate
            }
        }
    }

    // 无数字后缀：追加 -2, -3, ...
    for i := 2; ; i++ {
        candidate := base + "-" + strconv.Itoa(i)
        if !dirExists(filepath.Join(triesPath, candidate+"-"+dateSuffix)) {
            return candidate
        }
    }
}
```

两种策略：
1. 原名以数字结尾（如 `v2`）→ 递增数字（`v3`、`v4`...）
2. 原名不以数字结尾 → 追加 `-2`、`-3`...

## Git 检测逻辑

判断一个目录是否是 Git 仓库：

```go
_, err := os.Stat(filepath.Join(repoDir, ".git"))
```

`.git` 可以是目录（普通 Git 仓库）或文件（Git worktree）。ship 命令进一步区分：

```go
info, _ := os.Stat(gitFile)
isWorktree := !info.IsDir()  // .git 是文件 → worktree
```

worktree 使用 `git worktree move`，普通仓库使用 `mv`。
