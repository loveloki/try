package selector

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	"github.com/xleine/try/internal/fuzzy"
	"github.com/xleine/try/internal/git"
	"github.com/xleine/try/internal/i18n"
)

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
	Messages       *i18n.Messages
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
	activeDialog      dialog
	dialogFactory     DialogFactory
	styles            *styles
	messages          *i18n.Messages
}

// DialogInstance 对话框实例接口（导出供 CLI 层的工厂实现使用）
type DialogInstance interface {
	tea.Model
	Result() *SelectionResult
	Done() bool
	ViewContent() string
}

// dialog 是内部别名
type dialog = DialogInstance

type renderOnceMsg struct{}
type testKeyMsg struct{}

// New 创建选择器实例
func New(cfg Config) SelectorModel {
	// 确保 basePath 目录存在
	os.MkdirAll(cfg.BasePath, 0o755)

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

	msgs := cfg.Messages
	if msgs == nil {
		msgs = &i18n.EN
	}

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
		messages:          msgs,
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
		refreshCmd := m.refreshList()
		return m, refreshCmd

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
	// 清除上一次的删除状态消息
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

	// 转发给 textinput
	prevValue := m.textInput.Value()
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)

	if m.textInput.Value() != prevValue {
		m.list.Select(0)
		refreshCmd := m.refreshList()
		return m, tea.Batch(cmd, refreshCmd)
	}

	// 未被 textinput 消费的导航键转发给 list
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
		m.deleteStatus = m.messages.EmptyInputHint
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
		m.deleteStatus = m.messages.DeleteCancelled
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

// --- 对话框 ---

func (m SelectorModel) updateDialog(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	dlg, cmd := m.activeDialog.Update(msg)
	m.activeDialog = dlg.(dialog)
	if m.activeDialog.Done() {
		if result := m.activeDialog.Result(); result != nil {
			m.selected = result
			return m, tea.Quit
		}
		m.activeDialog = nil
	}
	return m, cmd
}

func (m SelectorModel) openDeleteDialog() (tea.Model, tea.Cmd) {
	var items []DeleteItem
	for _, entry := range m.cachedResults {
		if m.markedForDeletion[entry.Entry.Path] {
			items = append(items, DeleteItem{Path: entry.Entry.Path, Basename: entry.Entry.Basename})
		}
	}
	// 延迟导入 dialog 包会导致循环依赖，这里直接构建
	// dialog 包通过 SetDialog 方法注入（见下方 SetDialogFactory）
	if m.dialogFactory != nil {
		dlg := m.dialogFactory.NewDeleteDialog(items, m.basePath, m.testConfirm, m.width, m.messages)
		m.activeDialog = dlg
		return m, dlg.Init()
	}
	return m, nil
}

func (m SelectorModel) enterRenameDialog() (tea.Model, tea.Cmd) {
	entry := m.selectedEntry()
	if entry == nil {
		return m, nil
	}
	m.deleteMode = false
	m.markedForDeletion = map[string]bool{}
	if m.dialogFactory != nil {
		dlg := m.dialogFactory.NewRenameDialog(entry, m.basePath, m.width, m.messages)
		m.activeDialog = dlg
		return m, dlg.Init()
	}
	return m, nil
}

func (m SelectorModel) enterShipDialog() (tea.Model, tea.Cmd) {
	entry := m.selectedEntry()
	if entry == nil {
		return m, nil
	}
	m.deleteMode = false
	m.markedForDeletion = map[string]bool{}
	if m.dialogFactory != nil {
		dlg := m.dialogFactory.NewShipDialog(entry, m.basePath, m.shipPath, m.width, m.messages)
		m.activeDialog = dlg
		return m, dlg.Init()
	}
	return m, nil
}

// --- 目录加载与匹配 ---

func (m *SelectorModel) loadAllTries() []Entry {
	if m.allTries != nil {
		return m.allTries
	}

	entries, err := os.ReadDir(m.basePath)
	if err != nil {
		return nil
	}

	now := time.Now()
	var result []Entry
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if !entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		mtime := info.ModTime()
		hoursSinceMod := now.Sub(mtime).Hours()
		baseScore := 3.0 / math.Sqrt(hoursSinceMod+1)
		if DateSuffixRe.MatchString(entry.Name()) {
			baseScore += 2.0
		}

		result = append(result, Entry{
			Basename:  entry.Name(),
			Path:      filepath.Join(m.basePath, entry.Name()),
			Mtime:     mtime,
			BaseScore: baseScore,
		})
	}

	m.allTries = result
	return result
}

func (m *SelectorModel) refreshList() tea.Cmd {
	query := m.textInput.Value()
	if query == m.lastQuery && m.cachedResults != nil {
		return nil
	}

	allTries := m.loadAllTries()
	maxResults := m.height - 6
	if maxResults < 3 {
		maxResults = 3
	}

	// selector.Entry → fuzzy.Entry 转换
	fuzzyEntries := make([]fuzzy.Entry, len(allTries))
	for i, e := range allTries {
		fuzzyEntries[i] = fuzzy.Entry{
			Text:      e.Basename,
			BaseScore: e.BaseScore,
			Data:      e,
		}
	}

	results := fuzzy.Match(fuzzyEntries, query, maxResults)

	matched := make([]MatchedEntry, len(results))
	for i, r := range results {
		matched[i] = MatchedEntry{
			Entry:              r.Entry.Data.(Entry),
			Score:              r.Score,
			HighlightPositions: r.Positions,
		}
	}
	m.cachedResults = matched
	m.lastQuery = query

	items := make([]list.Item, len(matched))
	for i, me := range matched {
		items[i] = me
	}
	return m.list.SetItems(items)
}

// --- 对话框工厂（解耦 dialog 包依赖） ---

// DialogFactory 对话框创建接口，由外部（CLI 层）注入，避免循环依赖
type DialogFactory interface {
	NewDeleteDialog(items []DeleteItem, basePath, testConfirm string, width int, msgs *i18n.Messages) DialogInstance
	NewRenameDialog(entry *MatchedEntry, basePath string, width int, msgs *i18n.Messages) DialogInstance
	NewShipDialog(entry *MatchedEntry, basePath, shipPath string, width int, msgs *i18n.Messages) DialogInstance
}

// SetDialogFactory 注入对话框工厂（避免 selector → dialog 循环依赖）
func (m *SelectorModel) SetDialogFactory(f DialogFactory) {
	m.dialogFactory = f
}
