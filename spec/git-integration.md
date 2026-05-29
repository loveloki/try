# Git 集成

## 概述

try 提供三种 Git 操作：仓库克隆（clone）、创建 worktree、URL 快捷识别。相关实现位于 `internal/git/`。

## 函数接口

```go
type GitURIInfo struct {
    User string
    Repo string
    Host string
}

func ParseGitURI(uri string) *GitURIInfo
func IsGitURI(arg string) bool
func GenerateCloneDirName(uri, customName string) string
func ResolveUniqueName(triesPath, base, dateSuffix string) string
```

## Git URI 解析

### ParseGitURI

两组通用正则覆盖所有 Git 托管平台（GitHub、GitLab、自建等）：

| 格式 | 正则 | 匹配示例 |
|------|------|---------|
| HTTPS | `^https?://([^/]+)/([^/]+)/([^/]+)` | `https://github.com/user/repo`、`https://gitlab.company.com/team/project` |
| SSH | `^git@([^:]+):([^/]+)/([^/]+)` | `git@github.com:user/repo`、`git@gitlab.company.com:team/project` |

解析前先去掉 `.git` 后缀（`strings.TrimSuffix`）。无匹配返回 nil。

### IsGitURI

满足任一条件即返回 true：
- 以 `https://` / `http://` / `git@` 开头
- 包含 `github.com` / `gitlab.com`
- 以 `.git` 结尾

不做完整解析，仅做快速判断。

## Clone 命令

### 入口

```
try clone <url> [name]      # 显式 clone 命令
try <url> [name]            # URL 快捷方式（自动识别）
try exec clone <url> [name] # 包装函数内部调用
```

三种入口最终都调用 `cmdClone`。

### 目录命名

`GenerateCloneDirName` 逻辑：
- `customName` 非空 → 直接返回
- 否则通过 `ParseGitURI` 解析，格式为 `user-repo-YYYY-MM-DD`

名称在前，提高模糊匹配命中效率。

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

`worktreePath(triesPath, repoDir, customName string) string`：
- `customName` 非空 → 空格转连字符作为 base
- 否则 → 使用 `filepath.Base(repoDir)` 作为 base
- 通过 `ResolveUniqueName` 确保不重复
- 最终路径为 `triesPath/base-YYYY-MM-DD`

### "." 命令处理

```
try . name     → 使用当前目录作为仓库
try ./path name → 使用指定路径作为仓库
try .          → 报错退出（防止误触）
```

关键逻辑：
1. `try .` 无名称参数 → 报错退出
2. 检查仓库路径下是否有 `.git`（文件或目录）
3. 有 `.git` → 调用 `execWorktree`
4. 无 `.git` → 调用 `execMkdir`

### worktree 执行

Go 直接执行 worktree 创建。详见 `script-generation.md` 中的 `execWorktree`：

- `git rev-parse --show-toplevel` 获取仓库根目录（处理在子目录中执行的情况）
- `--detach` 创建 detached HEAD worktree（不关联分支）
- worktree 创建失败不阻塞，仍然 cd 到目标目录

## 目录名版本化

`ResolveUniqueName` 在目标目录已存在时自动递增后缀：

两种策略：
1. 原名以数字结尾（如 `v2`）→ 递增数字（`v3`、`v4`...）
2. 原名不以数字结尾 → 追加 `-2`、`-3`...

## Git 检测逻辑

通过 `os.Stat(filepath.Join(repoDir, ".git"))` 判断是否是 Git 仓库。`.git` 可以是目录（普通仓库）或文件（worktree）。

ship 命令进一步区分：`.git` 是文件 → worktree（用 `git worktree move`），`.git` 是目录 → 普通仓库（用 `os.Rename`）。
