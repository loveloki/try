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

// runOptions 聚合运行所需的全部参数，避免跨函数传递过多散参
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
	messages      *i18n.Messages
}

// Run 是 CLI 的主入口，返回退出码
func Run(args []string) int {
	// 所有命令统一走完整的全局标志解析（包括加载配置文件）
	opts, args := parseGlobalFlags(args)

	if hasFlag(args, "--help", "-h") {
		fmt.Fprintln(os.Stderr, opts.messages.HelpText)
		return 2
	}
	if hasFlag(args, "--version", "-v") {
		fmt.Fprintln(os.Stderr, "try "+version)
		return 0
	}

	if len(args) == 0 {
		return runSelector(opts, "")
	}

	switch args[0] {
	case "install":
		if err := shell.Install(opts.messages); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0
	case "clone":
		return cmdClone(opts, args[1:])
	case "worktree":
		return cmdWorktree(opts, args[1:])
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
		messages:      i18n.ForLocale(locale),
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
		Messages:       opts.messages,
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

	if err := script.Execute(selected, opts.messages); err != nil {
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

