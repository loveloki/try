package selector

import (
	"path/filepath"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	"github.com/xleine/try/internal/git"
)

// handleKey 根据按键消息分派到具体处理函数
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
