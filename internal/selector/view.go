package selector

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// renderMainContent 渲染主界面内容（不含对话框叠加）。
func (m *SelectorModel) renderMainContent() string {
	var b strings.Builder
	b.WriteString(renderHeader(m))
	if len(m.cachedResults) == 0 {
		b.WriteString(renderEmptyState(m))
	} else {
		b.WriteString(m.list.View())
	}
	b.WriteString(renderFooter(m))
	return b.String()
}

// View 渲染选择器完整视图，支持对话框叠加。
func (m SelectorModel) View() tea.View {
	var content string
	switch {
	case m.activeDialog != nil && m.activeDialog.OverlaysMainUI():
		content = overlayModal(m.renderMainContent(), m.activeDialog.ViewContent(), m.width, m.height)
	case m.activeDialog != nil:
		content = m.activeDialog.ViewContent()
	default:
		content = m.renderMainContent()
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
