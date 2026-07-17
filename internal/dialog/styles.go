package dialog

import (
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/loveloki/try/internal/selector"
)

// Styles 对话框专用样式集，基于 selector 的全局调色板构建。
type Styles struct {
	selector *selector.Styles

	ModalBorder     lipgloss.Style
	DimLayer        lipgloss.Style
	Title           lipgloss.Style
	Muted           lipgloss.Style
	Item            lipgloss.Style
	ItemMarked      lipgloss.Style
	Separator       lipgloss.Style
	Footer          lipgloss.Style
	Prompt          lipgloss.Style
	ErrorLine       lipgloss.Style
	ChoiceActive    lipgloss.Style
	ChoiceInactive  lipgloss.Style
	ChoiceDanger    lipgloss.Style
	InputLabel      lipgloss.Style
	RadioSelected   lipgloss.Style
	RadioUnselected lipgloss.Style
}

// NewStyles 从 selector 样式集派生对话框样式。
func NewStyles(st *selector.Styles) Styles {
	borderFg := st.Line.GetForeground()
	return Styles{
		selector:        st,
		ModalBorder:     st.Line.Border(lipgloss.RoundedBorder()).BorderForeground(borderFg),
		DimLayer:        st.Muted.Background(borderFg),
		Title:           st.Header.Bold(true),
		Muted:           st.Muted,
		Item:            st.Muted,
		ItemMarked:      st.Danger,
		Separator:       st.Line,
		Footer:          st.Muted,
		Prompt:          st.Accent,
		ErrorLine:       st.Danger.Strikethrough(false).Bold(true),
		ChoiceActive:    st.Highlight.Reverse(true).Bold(true),
		ChoiceInactive:  st.Muted,
		ChoiceDanger:    st.Danger.Strikethrough(false).Bold(true).Reverse(true),
		InputLabel:      st.Accent.Bold(true),
		RadioSelected:   st.Highlight.Bold(true),
		RadioUnselected: st.Muted,
	}
}

// KeyBadgeText 渲染快捷键徽章。
func (s Styles) KeyBadgeText(text string) string {
	return s.selector.KeyBadgeText(text)
}

// JoinKeyBadges 用空格连接多个快捷键徽章。
func (s Styles) JoinKeyBadges(parts []string) string {
	return s.selector.JoinKeyBadges(parts)
}

// renderModalBox 使用当前边框样式渲染带圆角边框的模态卡片。
func (s Styles) renderModalBox(content string, width int) string {
	boxW := modalBoxWidth(width)
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Align(lipgloss.Left).
		Width(boxW)
	box = box.Inherit(s.ModalBorder)
	return box.Render(content)
}

// padLine 将文本在指定宽度内左对齐并补空格。
func padLine(text string, width int) string {
	w := lipgloss.Width(text)
	if w >= width {
		return text
	}
	return text + strings.Repeat(" ", width-w)
}
