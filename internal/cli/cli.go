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
	// --help 和 --version 不需要配置文件
	if hasFlag(args, "--help", "-h") {
		fmt.Fprintln(os.Stderr, i18n.Get().HelpText)
		return 2
	}
	if hasFlag(args, "--version", "-v") {
		fmt.Fprintln(os.Stderr, "try "+version)
		return 0
	}

	// install 命令不需要已有配置文件，它会创建默认配置
	if len(args) > 0 && args[0] == "install" {
		if _, err := config.InitConfigFile(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		if err := shell.Install(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return 0
	}

	opts, remaining, err := parseGlobalFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	if len(remaining) == 0 {
		return runSelector(opts, "")
	}

	switch remaining[0] {
	case "clone":
		return cmdClone(opts, remaining[1:])
	case "worktree":
		return cmdWorktree(opts, remaining[1:])
	case "exec":
		return cmdExec(opts, remaining[1:])
	default:
		return runSelector(opts, strings.Join(remaining, "-"))
	}
}

// extractTestFlags 在测试文件中被替换为真实实现，生产版本始终返回空值。
var extractTestFlags = func(args []string) (andExit bool, andType, andKeys, andConfirm string, remaining []string) {
	return false, "", "", "", args
}

// parseGlobalFlags 从参数中提取全局选项，返回运行配置和剩余参数。
// 配置文件不存在时返回 error。
func parseGlobalFlags(args []string) (runOptions, []string, error) {
	colorsEnabled := true

	andExit, andType, andKeys, andConfirm, args := extractTestFlags(args)

	cfg, err := config.LoadConfig()
	if err != nil {
		return runOptions{}, args, fmt.Errorf("try: %w\n\tRun 'try install' to create a default config", err)
	}
	triesPath, shipPaths := config.ResolvePaths("", cfg)
	locale := config.ResolveLocale("", cfg)
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
	}, args, nil
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

func (f *dialogFactoryImpl) NewRenameDialog(entry *selector.MatchedEntry, basePath string, width int, colorsEnabled bool) selector.DialogInstance {
	return dialog.NewRenameDialog(entry, basePath, width, colorsEnabled)
}

func (f *dialogFactoryImpl) NewShipDialog(entry *selector.MatchedEntry, basePath string, shipPaths []string, width int, colorsEnabled bool) selector.DialogInstance {
	return dialog.NewShipDialog(entry, basePath, shipPaths, width, colorsEnabled)
}

