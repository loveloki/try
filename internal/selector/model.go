package selector

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	"github.com/xleine/try/internal/git"
	"github.com/xleine/try/internal/i18n"
)

// msgs 返回全局语言包的快捷方式
func msgs() *i18n.Messages { return i18n.Get() }

const (
	headerLines = 4
	footerLines = 3
)

// Config 选择器初始化配置
type Config struct {
	SearchTerm     string
	BasePath       string
	ShipPath       string
	InitialInput   string
	TestRenderOnce bool
	TestKeys       []string
	TestConfirm    string
	ColorsEnabled  bool
	Theme          string // "dark" 或 "light"
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
	shipPath          string
	testRenderOnce    bool
	testKeys          []string
	testConfirm       string
	activeDialog  dialog
	dialogFactory DialogFactory
	styles        *styles
}

type renderOnceMsg struct{}
type testKeyMsg struct{}

// New 创建选择器实例
func New(cfg Config) SelectorModel {
	if err := os.MkdirAll(cfg.BasePath, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msgs().ErrMkdir, err)
	}

	ti := textinput.New()
	ti.CharLimit = 256
	ti.Focus()

	// InitialInput 优先于 SearchTerm
	if cfg.InitialInput != "" {
		ti.SetValue(cfg.InitialInput)
	} else if cfg.SearchTerm != "" {
		ti.SetValue(cfg.SearchTerm)
	}

	st := newStyles(cfg.ColorsEnabled, cfg.Theme)

	delegate := &EntryDelegate{
		markedForDeletion: map[string]bool{},
		styles:            st,
	}
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.SetShowFilter(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()

	return SelectorModel{
		textInput:         ti,
		list:              l,
		delegate:          delegate,
		keys:              newKeyMap(),
		markedForDeletion: map[string]bool{},
		basePath:          cfg.BasePath,
		shipPath:          cfg.ShipPath,
		testRenderOnce:    cfg.TestRenderOnce,
		testKeys:          cfg.TestKeys,
		testConfirm:       cfg.TestConfirm,
		styles:            st,
	}
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

	// 非按键消息转发给子组件
	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)
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
	bodyHeight := m.height - headerLines - footerLines
	if bodyHeight < 1 {
		bodyHeight = 1
	}
	m.list.SetSize(m.width, bodyHeight)
	m.delegate.width = m.width
	m.cachedResults = nil
	return m, m.refreshList()
}

func (m SelectorModel) View() tea.View {
	var b strings.Builder

	if m.activeDialog != nil {
		b.WriteString(m.activeDialog.ViewContent())
	} else {
		b.WriteString(renderHeader(&m))
		b.WriteString(m.list.View())
		b.WriteString(renderFooter(&m))
	}

	v := tea.NewView(b.String())
	v.AltScreen = true
	return v
}

// Selected 返回最终选择结果（Bubbletea 退出后调用）
func (m SelectorModel) Selected() *SelectionResult {
	return m.selected
}

// --- 按键处理 ---

func (m SelectorModel) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.deleteStatus = ""

	switch {
	case key.Matches(msg, m.keys.Enter):
		return m.handleEnter()
	case key.Matches(msg, m.keys.CtrlP):
		m.list.CursorUp()
		return m, nil
	case key.Matches(msg, m.keys.CtrlN):
		m.list.CursorDown()
		return m, nil
	case key.Matches(msg, m.keys.CtrlD):
		return m.toggleDelete()
	case key.Matches(msg, m.keys.CtrlT):
		return m.handleCreateNew()
	case key.Matches(msg, m.keys.CtrlR):
		return m.enterRenameDialog()
	case key.Matches(msg, m.keys.CtrlG):
		return m.enterShipDialog()
	case key.Matches(msg, m.keys.Quit):
		return m.handleQuit()
	}

	return m.handleTextInput(msg)
}

// handleTextInput 将按键转发给 textInput，变化时刷新列表
func (m SelectorModel) handleTextInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	prevValue := m.textInput.Value()
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)

	if m.textInput.Value() != prevValue {
		m.list.Select(0)
		return m, tea.Batch(cmd, m.refreshList())
	}

	var listCmd tea.Cmd
	m.list, listCmd = m.list.Update(msg)
	return m, tea.Batch(cmd, listCmd)
}

func (m SelectorModel) handleEnter() (tea.Model, tea.Cmd) {
	if m.deleteMode && len(m.markedForDeletion) > 0 {
		return m.openDeleteDialog()
	}
	if entry := m.selectedEntry(); entry != nil {
		m.selected = &SelectionResult{Type: SelectCD, Path: entry.Entry.Path}
		return m, tea.Quit
	}
	if m.textInput.Value() != "" {
		return m.handleCreateNew()
	}
	return m, nil
}

func (m SelectorModel) handleCreateNew() (tea.Model, tea.Cmd) {
	input := strings.TrimSpace(m.textInput.Value())
	if input == "" {
		m.deleteStatus = msgs().EmptyInputHint
		return m, nil
	}

	name := strings.ReplaceAll(input, " ", "-")
	dateSuffix := time.Now().Format("2006-01-02")
	name = git.ResolveUniqueName(m.basePath, name, dateSuffix)
	dirName := name + "-" + dateSuffix

	m.selected = &SelectionResult{
		Type: SelectMkdir,
		Path: filepath.Join(m.basePath, dirName),
	}
	return m, tea.Quit
}

func (m SelectorModel) handleQuit() (tea.Model, tea.Cmd) {
	if m.deleteMode {
		m.deleteMode = false
		m.markedForDeletion = map[string]bool{}
		m.deleteStatus = msgs().DeleteCancelled
		return m, nil
	}
	return m, tea.Quit
}

func (m SelectorModel) toggleDelete() (tea.Model, tea.Cmd) {
	entry := m.selectedEntry()
	if entry == nil {
		return m, nil
	}
	path := entry.Entry.Path
	if m.markedForDeletion[path] {
		delete(m.markedForDeletion, path)
	} else {
		m.markedForDeletion[path] = true
		m.deleteMode = true
	}
	if len(m.markedForDeletion) == 0 {
		m.deleteMode = false
	}

	m.delegate.markedForDeletion = m.markedForDeletion
	return m, nil
}

func (m SelectorModel) selectedEntry() *MatchedEntry {
	item := m.list.SelectedItem()
	if item == nil {
		return nil
	}
	entry := item.(MatchedEntry)
	return &entry
}
