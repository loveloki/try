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
// quote 用单引号包裹路径，处理路径中的单引号
func quote(s string) string {
    return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// emitCd 输出 cd 脚本到 stdout，由父 Shell eval 执行
func emitCd(path string) {
    fmt.Println(scriptWarning)
    fmt.Println("cd " + quote(path))
}

const scriptWarning = "# if you can read this, you didn't launch try from an alias. run try --help."
```

`quote()` 仍然需要，因为 `cd` 命令必须走 Shell，路径中的特殊字符需要转义。

## 操作执行函数

### execCd（选择目录）

```go
func execCd(path string) error {
    // 更新 mtime（影响下次排序）
    now := time.Now()
    os.Chtimes(path, now, now)
    // 输出提示到 stderr
    fmt.Fprintln(os.Stderr, path)
    // 输出 cd 脚本到 stdout
    emitCd(path)
    return nil
}
```

### execMkdir（创建新目录）

```go
func execMkdir(path string) error {
    if err := os.MkdirAll(path, 0755); err != nil {
        return fmt.Errorf("创建目录失败: %w", err)
    }
    return execCd(path)
}
```

### execClone（克隆仓库）

```go
func execClone(path, uri string) error {
    if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
        return fmt.Errorf("创建父目录失败: %w", err)
    }

    fmt.Fprintln(os.Stderr, "Using git clone to create this trial from "+uri+".")
    cmd := exec.Command("git", "clone", uri, path)
    cmd.Stdout = os.Stderr  // git 输出到 stderr（stdout 是脚本通道）
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        // clone 失败：清理已创建的空目录
        os.RemoveAll(path)
        return fmt.Errorf("git clone 失败: %w", err)
    }
    return execCd(path)
}
```

> clone 失败时用 `os.RemoveAll(path)` 清理，这在全脚本方案中很难实现。

### execWorktree（创建 worktree）

```go
func execWorktree(targetPath, repoDir string) error {
    if err := os.MkdirAll(targetPath, 0755); err != nil {
        return fmt.Errorf("创建目录失败: %w", err)
    }

    // 检查是否在 Git 仓库中
    root, err := gitRepoRoot(repoDir)
    if err == nil {
        // 创建 detached HEAD worktree（失败不阻塞）
        cmd := exec.Command("git", "-C", root, "worktree", "add", "--detach", targetPath)
        cmd.Stdout = os.Stderr
        cmd.Stderr = os.Stderr
        cmd.Run() // 忽略错误：worktree 失败仍然 cd 到目录
    }

    return execCd(targetPath)
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
func execDelete(items []DeleteItem, basePath string) error {
    var failed []string
    for _, item := range items {
        if err := os.RemoveAll(item.Path); err != nil {
            failed = append(failed, item.Basename+": "+err.Error())
        }
    }

    if len(failed) > 0 {
        return fmt.Errorf("部分删除失败:\n%s", strings.Join(failed, "\n"))
    }

    // 输出 cd 脚本：回到原 cwd，若已被删除则回退到 basePath
    cwd, _ := os.Getwd()
    if dirExists(cwd) {
        emitCd(cwd)
    } else {
        emitCd(basePath)
    }
    return nil
}
```

> 安全检查（symlink 解析 + 前缀验证）在 Dialog 层完成，此函数信任输入。

### execRename（重命名目录）

```go
func execRename(basePath, oldName, newName string) error {
    oldPath := filepath.Join(basePath, oldName)
    newPath := filepath.Join(basePath, newName)
    if err := os.Rename(oldPath, newPath); err != nil {
        return fmt.Errorf("重命名失败: %w", err)
    }
    return execCd(newPath)
}
```

### execShip（发布为正式项目）

```go
func execShip(source, dest, basename, basePath string) error {
    gitFile := filepath.Join(source, ".git")

    // 区分 Git worktree 和普通仓库
    if isFile(gitFile) {
        // .git 是文件 → worktree，用 git worktree move
        cmd := exec.Command("git", "worktree", "move", source, dest)
        cmd.Stdout = os.Stderr
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
            return fmt.Errorf("git worktree move 失败: %w", err)
        }
    } else {
        // .git 是目录或不存在 → 普通 mv
        if err := os.Rename(source, dest); err != nil {
            return fmt.Errorf("移动目录失败: %w", err)
        }
    }

    fmt.Fprintln(os.Stderr, "Shipped: "+basename+" → "+dest)
    return execCd(dest)
}
```


## 选择结果到执行函数的映射

```go
func Execute(result *SelectionResult) error {
    if result == nil { return nil }

    switch result.Type {
    case SelectCD:
        return execCd(result.Path)
    case SelectMkdir:
        return execMkdir(result.Path)
    case SelectDelete:
        return execDelete(result.Paths, result.BasePath)
    case SelectRename:
        return execRename(result.BasePath, result.Old, result.New)
    case SelectShip:
        return execShip(result.Source, result.Dest, result.Basename, result.BasePath)
    default:
        return fmt.Errorf("未知的操作类型: %d", result.Type)
    }
}
```

## 测试注意

Shell 合规测试中 stdout 输出仅为 `cd '/path'`。`spec/tests/test_05_script_format.sh` 的断言需匹配此格式。
