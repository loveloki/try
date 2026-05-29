# 操作执行与脚本生成

## 概述

采用混合方案：文件系统操作（mkdir、rm、mv、git）在 Go 中直接执行，最后只输出 `cd` 脚本到 stdout 供父 Shell eval。这样做比全脚本方案更安全（无脚本注入风险）、错误处理更精细（Go 的 err + defer 回滚）。

## 架构

```
选择结果 (SelectionResult)
    │
    ├── Go 直接执行副作用操作
    │     os.MkdirAll / os.RemoveAll / os.Rename
    │     exec.Command("git", "clone", ...) / exec.Command("git", "worktree", ...)
    │
    └── stdout 输出 cd 脚本（父 Shell eval 执行）
          "cd '/target/path'"
```

唯一走脚本的操作是 `cd`（子进程无法修改父进程的工作目录）。

## cd 脚本输出

```go
const ScriptWarning = "# if you can read this, you didn't launch try from an alias. run try --help."

func Quote(s string) string
func EmitCd(path string)
func EmitCdTo(w io.Writer, path string)
```

- `Quote`：用单引号包裹路径，处理路径中的单引号（`'` → `'"'"'`）
- `EmitCd`：输出 cd 脚本到 `os.Stdout`，委托给 `EmitCdTo`
- `EmitCdTo`：输出 `ScriptWarning` 注释行 + `cd <quoted_path>` 到指定 writer

所有脚本输出函数接受 `io.Writer` 参数，便于测试时捕获输出。

## 操作执行函数

所有 exec 函数接收 `stdout, stderr io.Writer` 参数（stdout 用于脚本输出，stderr 用于用户可见信息）。

### execCd（选择目录）

```go
func execCd(stdout, stderr io.Writer, path string) error
```

更新目录的 atime/mtime 为当前时间（`os.Chtimes`），将路径写入 stderr（供包装函数显示），输出 cd 脚本到 stdout。

### execMkdir（创建新目录）

```go
func execMkdir(stdout, stderr io.Writer, path string) error
```

`os.MkdirAll(path, 0o755)` 创建目录，然后调用 `execCd`。

### execClone（克隆仓库）

```go
func execClone(stdout, stderr io.Writer, path, uri string) error
```

1. 创建父目录（`os.MkdirAll`）
2. 在 stderr 打印提示信息
3. 执行 `git clone uri path`，git 输出重定向到 stderr（stdout 是脚本通道）
4. clone 失败时用 `os.RemoveAll(path)` 清理半成品目录
5. 成功后调用 `execCd`

### execWorktree（创建 worktree）

```go
func execWorktree(stdout, stderr io.Writer, targetPath, repoDir string) error
```

1. 创建目标目录
2. 通过 `gitRepoRoot` 获取仓库根目录（`git rev-parse --show-toplevel`）
3. 执行 `git worktree add --detach targetPath`
4. worktree 创建失败不阻塞，仍调用 `execCd`（目录已创建）

辅助函数 `gitRepoRoot(dir string) (string, error)` 获取 Git 仓库根路径。

### execDelete（删除目录）

```go
func execDelete(stdout io.Writer, items []selector.DeleteItem, basePath string) error
```

逐个删除 items 中的目录（`os.RemoveAll`），收集失败项。全部完成后：
- 当前工作目录仍存在 → cd 回当前目录
- 当前工作目录已被删除 → cd 到 basePath

安全检查（symlink 解析 + 前缀验证）在 Dialog 层完成，此函数信任输入。

### execRename（重命名目录）

```go
func execRename(stdout, stderr io.Writer, basePath, oldName, newName string) error
```

`os.Rename` 重命名后调用 `execCd` 进入新目录。

### execShip（发布为正式项目）

```go
func execShip(stdout, stderr io.Writer, source, dest, basename string) error
```

根据 `.git` 类型选择操作方式：
- `.git` 是文件（worktree）→ `git worktree move source dest`
- `.git` 是目录或不存在 → `os.Rename`

成功后在 stderr 打印 "Shipped: basename → dest"，然后 `execCd` 到目标。

## 选择结果到执行函数的映射

```go
func Execute(result *selector.SelectionResult) error
func ExecuteTo(stdout, stderr io.Writer, result *selector.SelectionResult) error
```

`Execute` 是便捷入口（默认 `os.Stdout`/`os.Stderr`），委托给 `ExecuteTo`。

`ExecuteTo` 根据 `result.Type` 分派：

| SelectionType | 调用 |
|---------------|------|
| `SelectCD` | `execCd` |
| `SelectMkdir` | `execMkdir` |
| `SelectDelete` | `execDelete` |
| `SelectRename` | `execRename` |
| `SelectShip` | `execShip` |

`result` 为 nil 时直接返回 nil（用户取消）。

另有导出入口供 CLI 直接调用（不经过 SelectionResult）：

```go
func ExecClone(stdout, stderr io.Writer, path, uri string) error
func ExecWorktree(stdout, stderr io.Writer, targetPath, repoDir string) error
```
