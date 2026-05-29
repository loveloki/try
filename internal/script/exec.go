package script

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/xleine/try/internal/i18n"
	"github.com/xleine/try/internal/selector"
)

func msgs() *i18n.Messages { return i18n.Get() }

// Execute 根据选择结果类型分派到对应的执行函数
func Execute(result *selector.SelectionResult) error {
	if result == nil {
		return nil
	}
	return ExecuteTo(os.Stdout, os.Stderr, result)
}

// ExecuteTo 可指定输出目标的执行函数，便于测试
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
		return fmt.Errorf(msgs().ErrUnknownOp, result.Type)
	}
}

func execCd(stdout, stderr io.Writer, path string) error {
	now := time.Now()
	os.Chtimes(path, now, now)
	fmt.Fprintln(stderr, path)
	EmitCdTo(stdout, path)
	return nil
}

func execMkdir(stdout, stderr io.Writer, path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("%s: %w", msgs().ErrMkdir, err)
	}
	return execCd(stdout, stderr, path)
}

func execClone(stdout, stderr io.Writer, path, uri string) error {
	m := msgs()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("%s: %w", m.ErrMkdirParent, err)
	}

	fmt.Fprintf(stderr, m.MsgCloneFrom+"\n", uri)
	cmd := exec.Command("git", "clone", uri, path)
	cmd.Stdout = stderr
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		os.RemoveAll(path)
		return fmt.Errorf("%s: %w", m.ErrClone, err)
	}
	return execCd(stdout, stderr, path)
}

func execWorktree(stdout, stderr io.Writer, targetPath, repoDir string) error {
	if err := os.MkdirAll(targetPath, 0o755); err != nil {
		return fmt.Errorf("%s: %w", msgs().ErrMkdir, err)
	}

	root, err := gitRepoRoot(repoDir)
	if err == nil {
		cmd := exec.Command("git", "-C", root, "worktree", "add", "--detach", targetPath)
		cmd.Stdout = stderr
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(stderr, msgs().ErrWorktreeAdd+"\n", err)
		}
	}

	return execCd(stdout, stderr, targetPath)
}

func gitRepoRoot(dir string) (string, error) {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func execDelete(stdout io.Writer, items []selector.DeleteItem, basePath string) error {
	var failed []string
	for _, item := range items {
		if err := os.RemoveAll(item.Path); err != nil {
			failed = append(failed, item.Basename+": "+err.Error())
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf(msgs().ErrDeletePartial, strings.Join(failed, "\n"))
	}

	cwd, _ := os.Getwd()
	if selector.DirExists(cwd) {
		EmitCdTo(stdout, cwd)
	} else {
		EmitCdTo(stdout, basePath)
	}
	return nil
}

func execRename(stdout, stderr io.Writer, basePath, oldName, newName string) error {
	oldPath := filepath.Join(basePath, oldName)
	newPath := filepath.Join(basePath, newName)
	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("%s: %w", msgs().ErrRename, err)
	}
	return execCd(stdout, stderr, newPath)
}

func execShip(stdout, stderr io.Writer, source, dest, basename string) error {
	m := msgs()
	gitFile := filepath.Join(source, ".git")

	if selector.IsFile(gitFile) {
		cmd := exec.Command("git", "worktree", "move", source, dest)
		cmd.Stdout = stderr
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %w", m.ErrWorktreeMove, err)
		}
	} else {
		if err := os.Rename(source, dest); err != nil {
			return fmt.Errorf("%s: %w", m.ErrMove, err)
		}
	}

	fmt.Fprintf(stderr, m.MsgShipped+"\n", basename, dest)
	return execCd(stdout, stderr, dest)
}

// ExecClone 导出的 clone 入口（供 CLI 调用）
func ExecClone(stdout, stderr io.Writer, path, uri string) error {
	return execClone(stdout, stderr, path, uri)
}

// ExecWorktree 导出的 worktree 入口（供 CLI 调用）
func ExecWorktree(stdout, stderr io.Writer, targetPath, repoDir string) error {
	return execWorktree(stdout, stderr, targetPath, repoDir)
}
