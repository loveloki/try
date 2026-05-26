package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/xleine/try/internal/config"
	"github.com/xleine/try/internal/dialog"
	"github.com/xleine/try/internal/git"
	"github.com/xleine/try/internal/script"
	"github.com/xleine/try/internal/selector"
	"github.com/xleine/try/internal/shell"
)

var version = "dev"

// Run 是 CLI 的主入口，返回退出码
func Run(args []string) int {
	// 1. 提取颜色标志
	colorsEnabled := true
	args = filterFlags(args, func(flag string) bool {
		if flag == "--no-colors" || flag == "--no-expand-tokens" {
			colorsEnabled = false
			return true
		}
		return false
	})
	if os.Getenv("NO_COLOR") != "" {
		colorsEnabled = false
	}

	// 2. --help / -h
	if hasFlag(args, "--help", "-h") {
		printHelp()
		return 2
	}

	// 3. --version / -v
	if hasFlag(args, "--version", "-v") {
		fmt.Fprintln(os.Stderr, "try "+version)
		return 0
	}

	// 4. 提取 --path
	cliPath, args := extractPath(args)

	// 5. 提取测试参数
	andExit, args := extractBoolFlag(args, "--and-exit")
	andType, args := extractValueFlag(args, "--and-type")
	andKeys, args := extractValueFlag(args, "--and-keys")
	andConfirm, args := extractValueFlag(args, "--and-confirm")

	// 6. 解析路径
	cfg := config.LoadConfig()
	triesPath, shipPath := config.ResolvePaths(cliPath, cfg)

	// 7. 命令分派
	if len(args) == 0 {
		return runSelector(triesPath, shipPath, "", colorsEnabled, andExit, andType, andKeys, andConfirm)
	}

	command := args[0]
	remaining := args[1:]

	switch command {
	case "install":
		if err := shell.Install(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0

	case "clone":
		return cmdClone(triesPath, remaining)

	case "worktree":
		return cmdWorktree(triesPath, remaining)

	case "exec":
		return cmdExec(triesPath, shipPath, remaining, colorsEnabled, andExit, andType, andKeys, andConfirm)

	default:
		// 默认视为查询词
		searchTerm := strings.Join(args, "-")
		return runSelector(triesPath, shipPath, searchTerm, colorsEnabled, andExit, andType, andKeys, andConfirm)
	}
}

// cmdExec 处理包装函数内部调用的二级分派
func cmdExec(triesPath, shipPath string, args []string, colorsEnabled, andExit bool, andType, andKeys, andConfirm string) int {
	if len(args) == 0 {
		return runSelector(triesPath, shipPath, "", colorsEnabled, andExit, andType, andKeys, andConfirm)
	}

	switch args[0] {
	case "clone":
		return cmdClone(triesPath, args[1:])
	case "worktree":
		return cmdWorktree(triesPath, args[1:])
	case "cd":
		searchTerm := strings.Join(args[1:], "-")
		return runSelector(triesPath, shipPath, searchTerm, colorsEnabled, andExit, andType, andKeys, andConfirm)
	default:
		// 检查是否是 Git URL
		arg := args[0]
		if git.IsGitURI(arg) {
			var customName string
			if len(args) > 1 {
				customName = args[1]
			}
			return doClone(triesPath, arg, customName)
		}

		// "." 命令处理
		if strings.HasPrefix(arg, ".") {
			return handleDot(triesPath, args)
		}

		// 默认：查询词
		searchTerm := strings.Join(args, "-")
		return runSelector(triesPath, shipPath, searchTerm, colorsEnabled, andExit, andType, andKeys, andConfirm)
	}
}

func cmdClone(triesPath string, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: try clone <url> [name]")
		return 1
	}
	uri := args[0]
	var customName string
	if len(args) > 1 {
		customName = args[1]
	}
	return doClone(triesPath, uri, customName)
}

func doClone(triesPath, uri, customName string) int {
	dirName := git.GenerateCloneDirName(uri, customName)
	if dirName == "" {
		fmt.Fprintln(os.Stderr, "无法解析 Git URI: "+uri)
		return 1
	}
	targetPath := filepath.Join(triesPath, dirName)
	if err := script.ExecClone(os.Stdout, os.Stderr, targetPath, uri); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func cmdWorktree(triesPath string, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: try worktree <dir> [name]")
		return 1
	}
	repoDir := args[0]
	var customName string
	if len(args) > 1 {
		customName = args[1]
	}

	targetPath := worktreePath(triesPath, repoDir, customName)
	if err := script.ExecWorktree(os.Stdout, os.Stderr, targetPath, repoDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func handleDot(triesPath string, args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: try . <name>")
		return 1
	}
	repoDir := "."
	name := args[1]

	if args[0] != "." {
		// try ./path name
		repoDir = args[0]
		if len(args) > 2 {
			name = args[2]
		}
	}

	absRepo, _ := filepath.Abs(repoDir)
	targetPath := worktreePath(triesPath, absRepo, name)

	// 检查是否是 Git 仓库
	gitPath := filepath.Join(absRepo, ".git")
	if selector.FileExists(gitPath) {
		return doExecWorktree(triesPath, targetPath, absRepo)
	}
	// 非 Git 仓库：创建普通目录
	return doExecMkdir(targetPath)
}

func doExecWorktree(triesPath, targetPath, repoDir string) int {
	if err := script.ExecWorktree(os.Stdout, os.Stderr, targetPath, repoDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func doExecMkdir(path string) int {
	result := &selector.SelectionResult{Type: selector.SelectMkdir, Path: path}
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

// runSelector 启动交互式选择器
func runSelector(triesPath, shipPath, searchTerm string, colorsEnabled, andExit bool, andType, andKeys, andConfirm string) int {
	var testKeys []string
	if andKeys != "" {
		testKeys = selector.ParseTestKeys(andKeys)
	}

	cfg := selector.Config{
		SearchTerm:     searchTerm,
		BasePath:       triesPath,
		ShipPath:       shipPath,
		InitialInput:   andType,
		TestRenderOnce: andExit,
		TestKeys:       testKeys,
		TestConfirm:    andConfirm,
		ColorsEnabled:  colorsEnabled,
	}

	model := selector.New(cfg)
	// 注入对话框工厂
	factory := &dialogFactoryImpl{}
	model.SetDialogFactory(factory)

	p := tea.NewProgram(model, tea.WithOutput(os.Stderr))
	result, err := p.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	sm := result.(selector.SelectorModel)
	selected := sm.Selected()
	if selected == nil {
		return 1
	}

	if err := script.Execute(selected); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

// --- 对话框工厂实现（连接 selector 和 dialog 包） ---
// dialogFactoryImpl 实现 selector.DialogFactory，
// 将 dialog 包的具体类型适配为 selector 包内部 dialog 接口。

type dialogFactoryImpl struct{}

func (f *dialogFactoryImpl) NewDeleteDialog(items []selector.DeleteItem, basePath, testConfirm string, width int) selector.DialogInstance {
	return dialog.NewDeleteDialog(items, basePath, testConfirm, width)
}

func (f *dialogFactoryImpl) NewRenameDialog(entry *selector.MatchedEntry, basePath string, width int) selector.DialogInstance {
	return dialog.NewRenameDialog(entry, basePath, width)
}

func (f *dialogFactoryImpl) NewShipDialog(entry *selector.MatchedEntry, basePath, shipPath string, width int) selector.DialogInstance {
	return dialog.NewShipDialog(entry, basePath, shipPath, width)
}

// --- 帮助文本 ---

func printHelp() {
	help := `try - 临时实验目录管理工具

用法:
  try [query]                搜索并选择目录
  try clone <url> [name]     克隆 Git 仓库
  try worktree <dir> [name]  创建 Git worktree
  try install                安装 Shell 集成
  try . <name>               创建 worktree 或目录

选项:
  --help, -h           显示帮助
  --version, -v        显示版本
  --path PATH          指定 tries 根目录
  --no-colors          禁用颜色

快捷键:
  Enter    选择/确认
  Ctrl-T   创建新目录
  Ctrl-D   标记/取消删除
  Ctrl-R   重命名
  Ctrl-G   Ship（发布为正式项目）
  Ctrl-P/N 上下移动
  Esc      退出`
	fmt.Fprintln(os.Stderr, help)
}

// --- 参数解析工具函数 ---

func hasFlag(args []string, flags ...string) bool {
	for _, arg := range args {
		for _, flag := range flags {
			if arg == flag {
				return true
			}
		}
	}
	return false
}

func filterFlags(args []string, match func(string) bool) []string {
	var result []string
	for _, arg := range args {
		if !match(arg) {
			result = append(result, arg)
		}
	}
	return result
}

// extractPath 提取 --path VALUE 或 --path=VALUE（最后一个生效）
func extractPath(args []string) (string, []string) {
	var path string
	var result []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--path" && i+1 < len(args) {
			path = args[i+1]
			i++ // 跳过 value
		} else if strings.HasPrefix(args[i], "--path=") {
			path = args[i][7:]
		} else {
			result = append(result, args[i])
		}
	}
	return path, result
}

func extractBoolFlag(args []string, flag string) (bool, []string) {
	found := false
	var result []string
	for _, arg := range args {
		if arg == flag {
			found = true
		} else {
			result = append(result, arg)
		}
	}
	return found, result
}

func extractValueFlag(args []string, flag string) (string, []string) {
	var value string
	var result []string
	for i := 0; i < len(args); i++ {
		if args[i] == flag && i+1 < len(args) {
			value = args[i+1]
			i++
		} else if strings.HasPrefix(args[i], flag+"=") {
			value = args[i][len(flag)+1:]
		} else {
			result = append(result, args[i])
		}
	}
	return value, result
}
