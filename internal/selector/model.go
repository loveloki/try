package selector

import (
	"fmt"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	"github.com/loveloki/try/internal/i18n"
)

// msgs 返回全局语言包的快捷方式
func msgs() *i18n.Messages { return i18n.Get() }

// Config 选择器初始化配置
type Config struct {
	SearchTerm     string
	BasePath       string
	ShipPaths      []string
	InitialInput   string
	TestRenderOnce bool
	TestKeys       []string
	TestConfirm    string
	ColorsEnabled  bool
}

// SelectorModel 交互式选择器的核心状态
type SelectorModel struct {
	textInput         textinput.Model
	list              list.Model
	delegate          *EntryDelegate
	keys              keyMap
	deleteMode        bool
	deleteStatus      string
	markedForDeletion map[string]bool
	allTries          []Entry
	cachedResults     []MatchedEntry
	lastQuery         string
	selected          *SelectionResult
	width, height     int
	basePath          string
	shipPaths         []string
	sourceFilter      string // "" 表示全部，"tries" 或 ship 目录 basename
	sourceOptions     []string
	sourceCounts      map[string]int
	testRenderOnce    bool
	testKeys          []string
	testConfirm       string
	activeDialog    dialog
	dialogFactory   DialogFactory
	styles          *Styles
	colorsEnabled   bool
}

type renderOnceMsg struct{}
type testKeyMsg struct{}

// New 创建选择器实例
func New(cfg Config) SelectorModel {
	ensureDirs(cfg)

	ti := textinput.New()
	ti.CharLimit = 256
	ti.Focus()

	if cfg.InitialInput != "" {
		ti.SetValue(cfg.InitialInput)
	} else if cfg.SearchTerm != "" {
		ti.SetValue(cfg.SearchTerm)
	}

	st := NewStyles(cfg.ColorsEnabled)
	delegate := &EntryDelegate{
		markedForDeletion: map[string]bool{},
		styles:            st,
	}

	l := newList(delegate)
	sourceOpts := buildSourceOptions(cfg.ShipPaths)

	return SelectorModel{
		textInput:         ti,
		list:              l,
		delegate:          delegate,
		keys:              newKeyMap(),
		markedForDeletion: map[string]bool{},
		basePath:          cfg.BasePath,
		shipPaths:         cfg.ShipPaths,
		sourceOptions:     sourceOpts,
		sourceCounts:      map[string]int{},
		testRenderOnce:    cfg.TestRenderOnce,
		testKeys:          cfg.TestKeys,
		testConfirm:     cfg.TestConfirm,
		styles:          st,
		colorsEnabled:   cfg.ColorsEnabled,
	}
}

func ensureDirs(cfg Config) {
	if err := os.MkdirAll(cfg.BasePath, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msgs().ErrMkdir, err)
	}
	for _, sp := range cfg.ShipPaths {
		if err := os.MkdirAll(sp, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", msgs().ErrMkdir, err)
		}
	}
}

func newList(delegate *EntryDelegate) list.Model {
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetShowFilter(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	return l
}

func buildSourceOptions(shipPaths []string) []string {
	opts := []string{"", "tries"}
	for _, sp := range shipPaths {
		opts = append(opts, filepath.Base(sp))
	}
	return opts
}

func (m SelectorModel) Init() tea.Cmd {
	cmds := []tea.Cmd{
		m.textInput.Focus(),
		m.refreshList(),
		func() tea.Msg { return tea.RequestWindowSize() },
	}

	if m.testRenderOnce {
		cmds = append(cmds, func() tea.Msg { return renderOnceMsg{} })
	}

	if len(m.testKeys) > 0 {
		for range m.testKeys {
			cmds = append(cmds, func() tea.Msg { return testKeyMsg{} })
		}
	}

	return tea.Batch(cmds...)
}

func (m SelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case renderOnceMsg:
		return m, tea.Quit

	case testKeyMsg:
		if len(m.testKeys) > 0 {
			k := m.testKeys[0]
			m.testKeys = m.testKeys[1:]
			return m.Update(KeyToMsg(k))
		}
		return m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case tea.KeyPressMsg:
		if m.activeDialog != nil {
			return m.updateDialog(msg)
		}
		return m.handleKey(msg)
	}

	// 非按键消息转发给子组件（粘贴等操作可能改变输入值）
	prevValue := m.textInput.Value()
	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	if m.textInput.Value() != prevValue {
		m.list.Select(0)
		cmds = append(cmds, m.refreshList())
	}

	return m, tea.Batch(cmds...)
}

func (m SelectorModel) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	if w := EnvInt("TRY_WIDTH"); w > 0 {
		m.width = w
	}
	if h := EnvInt("TRY_HEIGHT"); h > 0 {
		m.height = h
	}
	m.list.SetSize(m.width, bodyHeight(&m))
	m.delegate.width = m.width
	m.cachedResults = nil
	return m, m.refreshList()
}

// Selected 返回最终选择结果（Bubbletea 退出后调用）
func (m SelectorModel) Selected() *SelectionResult {
	return m.selected
}

