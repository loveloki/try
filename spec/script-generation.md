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

// Quote 用单引号包裹路径，处理路径中的单引号
func Quote(s string) string {
    return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// EmitCd 输出 cd 脚本到 stdout
func EmitCd(path string) {
    EmitCdTo(os.Stdout, path)
}

// EmitCdTo 输出 cd 脚本到指定 writer，便于测试
func EmitCdTo(w io.Writer, path string) {
    fmt.Fprintln(w, ScriptWarning)
    fmt.Fprintln(w, "cd "+Quote(path))
}
```

`Quote()` 仍然需要，因为 `cd` 命令必须走 Shell，路径中的特殊字符需要转义。所有脚本输出函数接受 `io.Writer` 参数，便于测试时捕获输出。

## 操作执行函数

### execCd（选择目录）

```go
func execCd(stdout, stderr io.Writer, path string) error {
    now := time.Now()
    os.Chtimes(path, now, now)
    fmt.Fprintln(stderr, path)
    EmitCdTo(stdout, path)
    return nil
}
```

### execMkdir（创建新目录）

```go
func execMkdir(stdout, stderr io.Writer, path string) error {
    if err := os.MkdirAll(path, 0o755); err != nil {
        return fmt.Errorf("创建目录失败: %w", err)
    }
    return execCd(stdout, stderr, path)
}
```

### execClone（克隆仓库）

```go
func execClone(stdout, stderr io.Writer, path, uri string) error {
    if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
        return fmt.Errorf("创建父目录失败: %w", err)
    }

    fmt.Fprintln(stderr, "Using git clone to create this trial from "+uri+".")
    cmd := exec.Command("git", "clone", uri, path)
    cmd.Stdout = stderr  // git 输出到 stderr（stdout 是脚本通道）
    cmd.Stderr = stderr
    if err := cmd.Run(); err != nil {
        os.RemoveAll(path)
        return fmt.Errorf("git clone 失败: %w", err)
    }
    return execCd(stdout, stderr, path)
}
```

> clone 失败时用 `os.RemoveAll(path)` 清理，这在全脚本方案中很难实现。

### execWorktree（创建 worktree）

```go
func execWorktree(stdout, stderr io.Writer, targetPath, repoDir string) error {
    if err := os.MkdirAll(targetPath, 0o755); err != nil {
        return fmt.Errorf("创建目录失败: %w", err)
    }

    root, err := gitRepoRoot(repoDir)
    if err == nil {
        cmd := exec.Command("git", "-C", root, "worktree", "add", "--detach", targetPath)
        cmd.Stdout = stderr
        cmd.Stderr = stderr
        if err := cmd.Run(); err != nil {
            fmt.Fprintf(stderr, "git worktree add 失败（已创建普通目录）: %v\n", err)
        }
    }

    return execCd(stdout, stderr, targetPath)
}

func gitRepoRoot(dir string) (string, error) {
    cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
    out, err := cmd.Output()
    if err != nil { return "", err }
    return strings.TrimSpace(string(out)), nil
}
```

### execDelete（删除目录）

```go
func execDelete(stdout io.Writer, items []selector.DeleteItem, basePath string) error {
    var failed []string
    for _, item := range items {
        if err := os.RemoveAll(item.Path); err != nil {
            failed = append(failed, item.Basename+": "+err.Error())
        }
    }

    if len(failed) > 0 {
        return fmt.Errorf("部分删除失败:\n%s", strings.Join(failed, "\n"))
    }

    cwd, _ := os.Getwd()
    if selector.DirExists(cwd) {
        EmitCdTo(stdout, cwd)
    } else {
        EmitCdTo(stdout, basePath)
    }
    return nil
}
```

> 安全检查（symlink 解析 + 前缀验证）在 Dialog 层完成，此函数信任输入。

### execRename（重命名目录）

```go
func execRename(stdout, stderr io.Writer, basePath, oldName, newName string) error {
    oldPath := filepath.Join(basePath, oldName)
    newPath := filepath.Join(basePath, newName)
    if err := os.Rename(oldPath, newPath); err != nil {
        return fmt.Errorf("重命名失败: %w", err)
    }
    return execCd(stdout, stderr, newPath)
}
```

### execShip（发布为正式项目）

```go
func execShip(stdout, stderr io.Writer, source, dest, basename string) error {
    gitFile := filepath.Join(source, ".git")

    if selector.IsFile(gitFile) {
        cmd := exec.Command("git", "worktree", "move", source, dest)
        cmd.Stdout = stderr
        cmd.Stderr = stderr
        if err := cmd.Run(); err != nil {
            return fmt.Errorf("git worktree move 失败: %w", err)
        }
    } else {
        if err := os.Rename(source, dest); err != nil {
            return fmt.Errorf("移动目录失败: %w", err)
        }
    }

    fmt.Fprintln(stderr, "Shipped: "+basename+" → "+dest)
    return execCd(stdout, stderr, dest)
}
```


## 选择结果到执行函数的映射

`Execute` 是便捷入口（默认 stdout/stderr），内部委托给 `ExecuteTo`：

```go
func Execute(result *selector.SelectionResult) error {
    if result == nil { return nil }
    return ExecuteTo(os.Stdout, os.Stderr, result)
}

func ExecuteTo(stdout, stderr io.Writer, result *selector.SelectionResult) error {
    switch result.Type {
    case selector.SelectCD:
        return execCd(stdout, stderr, result.Path)
    case selector.SelectMkdir:
        return execMkdir(stdout, stderr, result.Path)
    case selector.SelectDelete:
        return execDelete(stdout, result.Paths, result.BasePath)
    case selector.SelectRename:
        return execRename(stdout, stderr, result.BasePath, result.Old, result.New)
    case selector.SelectShip:
        return execShip(stdout, stderr, result.Source, result.Dest, result.Basename)
    default:
        return fmt.Errorf("未知的操作类型: %d", result.Type)
    }
}
```

另有导出入口供 CLI 直接调用（不经过 SelectionResult）：

```go
func ExecClone(stdout, stderr io.Writer, path, uri string) error
func ExecWorktree(stdout, stderr io.Writer, targetPath, repoDir string) error
```

## 测试注意

Shell 合规测试中 stdout 输出仅为 `cd '/path'`。`spec/tests/test_05_script_format.sh` 的断言需匹配此格式。
