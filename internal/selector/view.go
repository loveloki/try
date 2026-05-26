package selector

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/colorprofile"
	lipgloss "charm.land/lipgloss/v2"
)

// styles 集中管理所有 TUI 样式
type styles struct {
	header     lipgloss.Style
	highlight  lipgloss.Style
	muted      lipgloss.Style
	match      lipgloss.Style
	selectedBg lipgloss.Style
	dangerBg   lipgloss.Style
	accent     lipgloss.Style
	profile    colorprofile.Profile
}

func newStyles(colorsEnabled bool) *styles {
	var profile colorprofile.Profile
	if colorsEnabled {
		w := colorprofile.NewWriter(os.Stderr, os.Environ())
		profile = w.Profile
	} else {
		profile = colorprofile.Ascii
	}

	return &styles{
		header:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("114")),
		highlight:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33")),
		muted:      lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		match:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("226")),
		selectedBg: lipgloss.NewStyle().Background(lipgloss.Color("238")),
		dangerBg:   lipgloss.NewStyle().Background(lipgloss.Color("52")),
		accent:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")),
		profile:    profile,
	}
}

// render 使用 colorprofile 降采样后渲染样式
func (s *styles) render(style lipgloss.Style, text string) string {
	rendered := style.Render(text)
	if s.profile >= colorprofile.TrueColor {
		return rendered
	}
	// 通过 colorprofile.Writer 降采样颜色
	var buf bytes.Buffer
	w := &colorprofile.Writer{
		Forward: &buf,
		Profile: s.profile,
	}
	fmt.Fprint(w, rendered)
	return buf.String()
}

// renderHeader 渲染标题 + 分隔线 + 搜索栏 + 分隔线
func renderHeader(m *SelectorModel) string {
	var b strings.Builder
	w := m.width
	if w <= 0 {
		w = 80
	}
	sep := m.styles.render(m.styles.muted, strings.Repeat("─", w))

	b.WriteString(m.styles.render(m.styles.header, "🏠 Try Directory Selection") + "\n")
	b.WriteString(sep + "\n")
	b.WriteString("Search: " + m.textInput.View() + "\n")
	b.WriteString(sep + "\n")

	return b.String()
}

// renderFooter 渲染底部状态栏/快捷键提示
func renderFooter(m *SelectorModel) string {
	var b strings.Builder

	// "Create new" 行
	if input := strings.TrimSpace(m.textInput.Value()); input != "" {
		name := strings.ReplaceAll(input, " ", "-")
		b.WriteString("\n📂 Create new: " + m.styles.render(m.styles.accent, name) + "\n")
	}

	w := m.width
	if w <= 0 {
		w = 80
	}
	sep := m.styles.render(m.styles.muted, strings.Repeat("─", w))
	b.WriteString(sep + "\n")

	if m.deleteStatus != "" {
		b.WriteString(m.styles.render(m.styles.accent, m.deleteStatus))
	} else if m.deleteMode {
		count := len(m.markedForDeletion)
		b.WriteString(m.styles.render(m.styles.dangerBg,
			fmt.Sprintf(" DELETE MODE  %d marked  |  Ctrl-D: Toggle  Enter: Confirm  Esc: Cancel", count)))
	} else {
		hint := "Ctrl-T: New  Ctrl-D: Delete  Ctrl-R: Rename  Ctrl-G: Ship  Esc: Quit"
		b.WriteString(m.styles.render(m.styles.muted, hint))
	}

	return b.String()
}
