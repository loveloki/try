package cli

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/xleine/try/internal/config"
	"github.com/xleine/try/internal/dialog"
	"github.com/xleine/try/internal/i18n"
	"github.com/xleine/try/internal/script"
	"github.com/xleine/try/internal/selector"
	"github.com/xleine/try/internal/shell"
)

var version = "dev"

// runOptions 聚合选择器运行所需的全部参数，避免跨函数传递过多散参
type runOptions struct {
	triesPath     string
	shipPath      string
	colorsEnabled bool
	theme         string
	locale        string
	andExit       bool
	andType       string
	andKeys       string
	andConfirm    string
}

// Run 是 CLI 的主入口，返回退出码
func Run(args []string) int {
	// 提前处理 --help / --version（不依赖配置解析）
	if hasFlag(args, "--help", "-h") {
		printHelp()
		return 2
	}
	if hasFlag(args, "--version", "-v") {
		fmt.Fprintln(os.Stderr, "try "+version)
		return 0
	}

	opts, args := parseGlobalFlags(args)

	if len(args) == 0 {
		return runSelector(opts, "")
	}

	switch args[0] {
	case "install":
		if err := shell.Install(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0
	case "clone":
		return cmdClone(opts.triesPath, args[1:])
	case "worktree":
		return cmdWorktree(opts.triesPath, args[1:])
	case "exec":
		return cmdExec(opts, args[1:])
	default:
		return runSelector(opts, strings.Join(args, "-"))
	}
}

// parseGlobalFlags 从参数中提取全局选项，返回运行配置和剩余参数
func parseGlobalFlags(args []string) (runOptions, []string) {
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

	cliPath, args := extractPath(args)
	cliTheme, args := extractValueFlag(args, "--theme")
	cliLocale, args := extractValueFlag(args, "--locale")

	andExit, args := extractBoolFlag(args, "--and-exit")
	andType, args := extractValueFlag(args, "--and-type")
	andKeys, args := extractValueFlag(args, "--and-keys")
	andConfirm, args := extractValueFlag(args, "--and-confirm")

	cfg := config.LoadConfig()
	triesPath, shipPath := config.ResolvePaths(cliPath, cfg)
	theme := config.ResolveTheme(cliTheme, cfg)
	locale := config.ResolveLocale(cliLocale, cfg)

	return runOptions{
		triesPath:     triesPath,
		shipPath:      shipPath,
		colorsEnabled: colorsEnabled,
		theme:         theme,
		locale:        locale,
		andExit:       andExit,
		andType:       andType,
		andKeys:       andKeys,
		andConfirm:    andConfirm,
	}, args
}

// runSelector 启动交互式选择器
func runSelector(opts runOptions, searchTerm string) int {
	var testKeys []string
	if opts.andKeys != "" {
		testKeys = selector.ParseTestKeys(opts.andKeys)
	}

	cfg := selector.Config{
		SearchTerm:     searchTerm,
		BasePath:       opts.triesPath,
		ShipPath:       opts.shipPath,
		InitialInput:   opts.andType,
		TestRenderOnce: opts.andExit,
		TestKeys:       testKeys,
		TestConfirm:    opts.andConfirm,
		ColorsEnabled:  opts.colorsEnabled,
		Theme:          opts.theme,
		Messages:       i18n.ForLocale(opts.locale),
	}

	model := selector.New(cfg)
	model.SetDialogFactory(&dialogFactoryImpl{})

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

type dialogFactoryImpl struct{}

func (f *dialogFactoryImpl) NewDeleteDialog(items []selector.DeleteItem, basePath, testConfirm string, width int, msgs *i18n.Messages) selector.DialogInstance {
	return dialog.NewDeleteDialog(items, basePath, testConfirm, width, msgs)
}

func (f *dialogFactoryImpl) NewRenameDialog(entry *selector.MatchedEntry, basePath string, width int, msgs *i18n.Messages) selector.DialogInstance {
	return dialog.NewRenameDialog(entry, basePath, width, msgs)
}

func (f *dialogFactoryImpl) NewShipDialog(entry *selector.MatchedEntry, basePath, shipPath string, width int, msgs *i18n.Messages) selector.DialogInstance {
	return dialog.NewShipDialog(entry, basePath, shipPath, width, msgs)
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
  --theme dark|light   配色主题（默认 auto 自动检测）
  --locale en|zh       界面语言（默认 auto 自动检测）
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
