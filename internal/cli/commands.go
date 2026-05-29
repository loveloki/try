package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xleine/try/internal/git"
	"github.com/xleine/try/internal/i18n"
	"github.com/xleine/try/internal/script"
	"github.com/xleine/try/internal/selector"
)

// cmdExec 处理包装函数内部调用的二级分派
func cmdExec(opts runOptions, args []string) int {
	if len(args) == 0 {
		return runSelector(opts, "")
	}

	switch args[0] {
	case "clone":
		return cmdClone(opts, args[1:])
	case "worktree":
		return cmdWorktree(opts, args[1:])
	case "cd":
		return runSelector(opts, strings.Join(args[1:], "-"))
	default:
		return cmdExecDefault(opts, args)
	}
}

func cmdExecDefault(opts runOptions, args []string) int {
	arg := args[0]
	if git.IsGitURI(arg) {
		var customName string
		if len(args) > 1 {
			customName = args[1]
		}
		return doClone(opts, arg, customName)
	}

	if strings.HasPrefix(arg, ".") {
		return handleDot(opts, args)
	}

	return runSelector(opts, strings.Join(args, "-"))
}

func cmdClone(opts runOptions, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, i18n.Get().UsageClone)
		return 1
	}
	uri := args[0]
	var customName string
	if len(args) > 1 {
		customName = args[1]
	}
	return doClone(opts, uri, customName)
}

func doClone(opts runOptions, uri, customName string) int {
	dirName := git.GenerateCloneDirName(uri, customName)
	if dirName == "" {
		fmt.Fprintln(os.Stderr, i18n.Get().ErrParseGitURI+uri)
		return 1
	}
	targetPath := filepath.Join(opts.triesPath, dirName)
	if err := script.ExecClone(os.Stdout, os.Stderr, targetPath, uri); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func cmdWorktree(opts runOptions, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, i18n.Get().UsageWorktree)
		return 1
	}
	repoDir := args[0]
	var customName string
	if len(args) > 1 {
		customName = args[1]
	}

	targetPath := worktreePath(opts.triesPath, repoDir, customName)
	if err := script.ExecWorktree(os.Stdout, os.Stderr, targetPath, repoDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func handleDot(opts runOptions, args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, i18n.Get().UsageDot)
		return 1
	}
	repoDir := "."
	name := args[1]

	if args[0] != "." {
		repoDir = args[0]
		if len(args) > 2 {
			name = args[2]
		}
	}

	absRepo, err := filepath.Abs(repoDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, i18n.Get().ErrParsePath+err.Error())
		return 1
	}
	targetPath := worktreePath(opts.triesPath, absRepo, name)

	gitPath := filepath.Join(absRepo, ".git")
	if selector.FileExists(gitPath) {
		if err := script.ExecWorktree(os.Stdout, os.Stderr, targetPath, absRepo); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0
	}
	result := &selector.SelectionResult{Type: selector.SelectMkdir, Path: targetPath}
	if err := script.Execute(result); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func worktreePath(triesPath, repoDir, customName string) string {
	var base string
	if customName != "" {
		base = strings.ReplaceAll(customName, " ", "-")
	} else {
		base = filepath.Base(repoDir)
	}
	dateSuffix := time.Now().Format("2006-01-02")
	base = git.ResolveUniqueName(triesPath, base, dateSuffix)
	return filepath.Join(triesPath, base+"-"+dateSuffix)
}
