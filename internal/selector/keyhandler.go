package selector

import (
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/loveloki/try/internal/git"
)

// handleKey 根据按键消息分派到具体处理函数
func (m SelectorModel) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.deleteStatus = ""

	switch {
	case key.Matches(msg, m.keys.Enter):
		return m.handleEnter()
	case key.Matches(msg, m.keys.CtrlP):
		m.moveCursor(-1)
		return m, nil
	case key.Matches(msg, m.keys.CtrlN):
		m.moveCursor(1)
		return m, nil
	case msg.Code == tea.KeyUp:
		m.moveCursor(-1)
		return m, nil
	case msg.Code == tea.KeyDown:
		m.moveCursor(1)
		return m, nil
	case key.Matches(msg, m.keys.Space), key.Matches(msg, m.keys.Delete):
		return m.toggleDeleteMark()
	case key.Matches(msg, m.keys.CtrlD):
		return m.toggleDelete()
	case key.Matches(msg, m.keys.CtrlA):
		return m.markAll()
	case key.Matches(msg, m.keys.CtrlT):
		return m.handleCreateNew()
	case key.Matches(msg, m.keys.CtrlR):
		return m.enterRenameDialog()
	case key.Matches(msg, m.keys.CtrlG):
		return m.enterShipDialog()
	case key.Matches(msg, m.keys.Tab):
		return m.cycleSourceFilter(1)
	case key.Matches(msg, m.keys.ShiftTab):
		return m.cycleSourceFilter(-1)
	case key.Matches(msg, m.keys.Slash), key.Matches(msg, m.keys.CtrlF):
		return m.focusSearch()
	case key.Matches(msg, m.keys.Quit):
		return m.handleQuit()
	}

	return m.handleTextInput(msg)
}

func (m *SelectorModel) moveCursor(delta int) {
	items := m.list.Items()
	if len(items) == 0 {
		return
	}
	next := (m.list.Index() + delta) % len(items)
	if next < 0 {
		next += len(items)
	}
	m.list.Select(next)
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
	if m.textInput.Value() != "" {
		m.textInput.SetValue("")
		m.list.Select(0)
		return m, m.refreshList()
	}
	if m.deleteMode {
		m.deleteMode = false
		m.markedForDeletion = map[string]bool{}
		m.deleteStatus = msgs().DeleteCancelled
		m.delegate.markedForDeletion = m.markedForDeletion
		return m, nil
	}
	return m, tea.Quit
}

// focusSearch 聚焦搜索框；由于搜索框始终聚焦，此处仅清空已有查询以重新输入。
func (m SelectorModel) focusSearch() (tea.Model, tea.Cmd) {
	if m.textInput.Value() != "" {
		m.textInput.SetValue("")
		m.list.Select(0)
		return m, m.refreshList()
	}
	return m, nil
}

// toggleDelete 切换删除模式；当前有标记时退出模式并清空标记。
func (m SelectorModel) toggleDelete() (tea.Model, tea.Cmd) {
	if m.deleteMode {
		m.deleteMode = false
		m.markedForDeletion = map[string]bool{}
	} else {
		return m.toggleDeleteMark()
	}
	m.delegate.markedForDeletion = m.markedForDeletion
	return m, nil
}

// toggleDeleteMark 切换当前选中条目的删除标记。
func (m SelectorModel) toggleDeleteMark() (tea.Model, tea.Cmd) {
	entry := m.selectedEntry()
	if entry == nil {
		return m, nil
	}
	path := entry.Entry.Path
	m.markedForDeletion = copyMarkedMap(m.markedForDeletion)
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

// markAll 标记当前过滤结果中的所有条目。
func (m SelectorModel) markAll() (tea.Model, tea.Cmd) {
	if len(m.cachedResults) == 0 {
		return m, nil
	}
	m.markedForDeletion = copyMarkedMap(m.markedForDeletion)
	for _, entry := range m.cachedResults {
		m.markedForDeletion[entry.Entry.Path] = true
	}
	m.deleteMode = true
	m.delegate.markedForDeletion = m.markedForDeletion
	return m, nil
}

// copyMarkedMap 深拷贝标记集合（保持不可变性）。
func copyMarkedMap(src map[string]bool) map[string]bool {
	out := make(map[string]bool, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func (m SelectorModel) selectedEntry() *MatchedEntry {
	item := m.list.SelectedItem()
	if item == nil {
		return nil
	}
	entry := item.(MatchedEntry)
	return &entry
}

func (m SelectorModel) cycleSourceFilter(delta int) (tea.Model, tea.Cmd) {
	if len(m.sourceOptions) <= 1 {
		return m, nil
	}
	currentIdx := 0
	for i, opt := range m.sourceOptions {
		if opt == m.sourceFilter {
			currentIdx = i
			break
		}
	}
	n := len(m.sourceOptions)
	m.sourceFilter = m.sourceOptions[((currentIdx+delta)%n+n)%n]
	m.lastQuery = ""
	m.cachedResults = nil
	m.list.Select(0)
	return m, m.refreshList()
}
