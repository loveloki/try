package selector

import (
	"fmt"
	"strings"

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
}

// themePalette 定义一组主题色值（256-color ANSI codes）
type themePalette struct {
	header     string
	highlight  string
	muted      string
	match      string
	selectedBg string
	dangerBg   string
	accent     string
}

// GitHub Dark 风格配色
var darkPalette = themePalette{
	header:     "75",  // 浅蓝 (#5fafff)
	highlight:  "75",  // 浅蓝
	muted:      "245", // 灰色
	match:      "215", // 浅橙 (#ffaf5f)
	selectedBg: "237", // 深灰背景
	dangerBg:   "52",  // 暗红背景
	accent:     "114", // 浅绿 (#87d787)
}

// GitHub Light 风格配色
var lightPalette = themePalette{
	header:     "26",  // 深蓝 (#005fd7)
	highlight:  "26",  // 深蓝
	muted:      "242", // 中灰
	match:      "130", // 棕橙 (#af5f00)
	selectedBg: "254", // 浅灰背景
	dangerBg:   "217", // 浅红背景 (#ffafaf)
	accent:     "28",  // 深绿 (#008700)
}

// newStyles 创建样式集。颜色降采样交由 bubbletea v2 内置渲染器处理，
// 此处不做额外 colorprofile 降采样以避免双重转换导致背景色丢失。
func newStyles(colorsEnabled bool, theme string) *styles {
	if !colorsEnabled {
		return &styles{
			header:     lipgloss.NewStyle(),
			highlight:  lipgloss.NewStyle(),
			muted:      lipgloss.NewStyle(),
			match:      lipgloss.NewStyle(),
			selectedBg: lipgloss.NewStyle(),
			dangerBg:   lipgloss.NewStyle(),
			accent:     lipgloss.NewStyle(),
		}
	}

	p := darkPalette
	if theme == "light" {
		p = lightPalette
	}

	return &styles{
		header:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.header)),
		highlight:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.highlight)),
		muted:      lipgloss.NewStyle().Foreground(lipgloss.Color(p.muted)),
		match:      lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.match)),
		selectedBg: lipgloss.NewStyle().Background(lipgloss.Color(p.selectedBg)),
		dangerBg:   lipgloss.NewStyle().Background(lipgloss.Color(p.dangerBg)),
		accent:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.accent)),
	}
}

// render 渲染带样式的文本，颜色降采样由 bubbletea 渲染器统一处理
func (s *styles) render(style lipgloss.Style, text string) string {
	return style.Render(text)
}

// renderHeader 渲染标题 + 分隔线 + 搜索栏 + 分隔线
func renderHeader(m *SelectorModel) string {
	var b strings.Builder
	w := m.width
	if w <= 0 {
		w = 80
	}
	sep := m.styles.render(m.styles.muted, strings.Repeat("─", w))

	b.WriteString(m.styles.render(m.styles.header, m.messages.Title) + "\n")
	b.WriteString(sep + "\n")
	b.WriteString(m.messages.SearchPrefix + m.textInput.View() + "\n")
	b.WriteString(sep + "\n")

	return b.String()
}

// renderFooter 渲染底部状态栏/快捷键提示
func renderFooter(m *SelectorModel) string {
	var b strings.Builder

	// "Create new" 行
	if input := strings.TrimSpace(m.textInput.Value()); input != "" {
		name := strings.ReplaceAll(input, " ", "-")
		b.WriteString("\n" + m.messages.CreateNew + m.styles.render(m.styles.accent, name) + "\n")
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
			fmt.Sprintf(m.messages.DeleteMode, count)))
	} else {
		b.WriteString(m.styles.render(m.styles.muted, m.messages.HintBar))
	}

	return b.String()
}
