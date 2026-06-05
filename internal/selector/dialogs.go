package selector

import (
	tea "charm.land/bubbletea/v2"
)

// DialogInstance 对话框实例接口（导出供 CLI 层的工厂实现使用）
type DialogInstance interface {
	tea.Model
	Result() *SelectionResult
	Done() bool
	ViewContent() string
	// OverlaysMainUI 为 true 时弹窗叠放在主列表之上，而非独占全屏。
	OverlaysMainUI() bool
}

// dialog 是内部别名
type dialog = DialogInstance

// DialogFactory 对话框创建接口，由外部（CLI 层）注入，避免循环依赖
type DialogFactory interface {
	NewDeleteDialog(items []DeleteItem, basePath, testConfirm string, width int, colorsEnabled bool) DialogInstance
	NewRenameDialog(entry *MatchedEntry, basePath string, width int) DialogInstance
	NewShipDialog(entry *MatchedEntry, basePath string, shipPaths []string, width int) DialogInstance
}

// SetDialogFactory 注入对话框工厂（避免 selector → dialog 循环依赖）
func (m *SelectorModel) SetDialogFactory(f DialogFactory) {
	m.dialogFactory = f
}

// updateDialog 将按键消息转发给当前活跃对话框
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

// openDeleteDialog 收集已标记条目并打开删除确认对话框
func (m SelectorModel) openDeleteDialog() (tea.Model, tea.Cmd) {
	var items []DeleteItem
	for _, entry := range m.cachedResults {
		if m.markedForDeletion[entry.Entry.Path] {
			items = append(items, DeleteItem{Path: entry.Entry.Path, Basename: entry.Entry.Basename})
		}
	}
	if m.dialogFactory != nil {
		dlg := m.dialogFactory.NewDeleteDialog(items, m.basePath, m.testConfirm, m.width, m.colorsEnabled)
		m.activeDialog = dlg
		return m, dlg.Init()
	}
	return m, nil
}

// enterRenameDialog 打开重命名对话框
func (m SelectorModel) enterRenameDialog() (tea.Model, tea.Cmd) {
	entry := m.selectedEntry()
	if entry == nil {
		return m, nil
	}
	m.deleteMode = false
	m.markedForDeletion = map[string]bool{}
	if m.dialogFactory != nil {
		dlg := m.dialogFactory.NewRenameDialog(entry, m.basePath, m.width)
		m.activeDialog = dlg
		return m, dlg.Init()
	}
	return m, nil
}

// enterShipDialog 打开 ship 对话框
func (m SelectorModel) enterShipDialog() (tea.Model, tea.Cmd) {
	entry := m.selectedEntry()
	if entry == nil {
		return m, nil
	}
	m.deleteMode = false
	m.markedForDeletion = map[string]bool{}
	if m.dialogFactory != nil {
		dlg := m.dialogFactory.NewShipDialog(entry, m.basePath, m.shipPaths, m.width)
		m.activeDialog = dlg
		return m, dlg.Init()
	}
	return m, nil
}
