package cli

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/loveloki/try/internal/config"
	"github.com/loveloki/try/internal/dialog"
	"github.com/loveloki/try/internal/i18n"
	"github.com/loveloki/try/internal/script"
	"github.com/loveloki/try/internal/selector"
	"github.com/loveloki/try/internal/shell"
)

func isTestMode(opts runOptions) bool {
	return opts.andExit || opts.andKeys != ""
}

var version = "dev"

// runOptions 聚合运行所需的全部参数，避免跨函数传递过多散参
type runOptions struct {
	triesPath     string
	shipPaths     []string
	colorsEnabled bool
	locale        string
	andExit       bool
	andType       string
	andKeys       string
	andConfirm    string
}

// Run 是 CLI 的主入口，返回退出码
func Run(args []string) int {
	opts, args := parseGlobalFlags(args)

	if hasFlag(args, "--help", "-h") {
		fmt.Fprintln(os.Stderr, i18n.Get().HelpText)
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
		if err := shell.Install(); err != nil {
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
	cliLocale, args := extractValueFlag(args, "--locale")

	andExit, args := extractBoolFlag(args, "--and-exit")
	andType, args := extractValueFlag(args, "--and-type")
	andKeys, args := extractValueFlag(args, "--and-keys")
	andConfirm, args := extractValueFlag(args, "--and-confirm")

	cfg := config.LoadConfig()
	triesPath, shipPaths := config.ResolvePaths(cliPath, cfg)
	locale := config.ResolveLocale(cliLocale, cfg)
	i18n.Init(locale)

	return runOptions{
		triesPath:     triesPath,
		shipPaths:     shipPaths,
		colorsEnabled: colorsEnabled,
		locale:        locale,
		andExit:       andExit,
		andType:       andType,
		andKeys:       andKeys,
		andConfirm:    andConfirm,
	}, args
}

// runSelector 启动交互式选择器
func runSelector(opts runOptions, searchTerm string) int {
	// 非测试模式下，非终端 stdin 会导致 bubbletea cancelreader 卡在 epoll
	if !isTestMode(opts) {
		if fi, err := os.Stdin.Stat(); err != nil || fi.Mode()&os.ModeCharDevice == 0 {
			fmt.Fprintln(os.Stderr, i18n.Get().ErrNotTerminal)
			return 1
		}
	}

	var testKeys []string
	if opts.andKeys != "" {
		testKeys = selector.ParseTestKeys(opts.andKeys)
	}

	cfg := selector.Config{
		SearchTerm:     searchTerm,
		BasePath:       opts.triesPath,
		ShipPaths:      opts.shipPaths,
		InitialInput:   opts.andType,
		TestRenderOnce: opts.andExit,
		TestKeys:       testKeys,
		TestConfirm:    opts.andConfirm,
		ColorsEnabled:  opts.colorsEnabled,
	}

	model := selector.New(cfg)
	model.SetDialogFactory(&dialogFactoryImpl{})

	progOpts := []tea.ProgramOption{tea.WithOutput(os.Stderr)}
	// 测试模式下用空 reader 替代 stdin，避免 cancelreader 阻塞在 epoll
	if isTestMode(opts) {
		progOpts = append(progOpts, tea.WithInput(strings.NewReader("")))
	}

	p := tea.NewProgram(model, progOpts...)
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

func (f *dialogFactoryImpl) NewDeleteDialog(
	items []selector.DeleteItem,
	basePath, testConfirm string,
	width int,
	colorsEnabled bool,
) selector.DialogInstance {
	return dialog.NewDeleteDialog(items, basePath, testConfirm, width, colorsEnabled)
}

func (f *dialogFactoryImpl) NewRenameDialog(entry *selector.MatchedEntry, basePath string, width int) selector.DialogInstance {
	return dialog.NewRenameDialog(entry, basePath, width)
}

func (f *dialogFactoryImpl) NewShipDialog(entry *selector.MatchedEntry, basePath string, shipPaths []string, width int) selector.DialogInstance {
	return dialog.NewShipDialog(entry, basePath, shipPaths, width)
}

