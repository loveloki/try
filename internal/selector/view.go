package selector

import (
	"fmt"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/loveloki/try/internal/config"
)

// styles 集中管理所有 TUI 样式
type styles struct {
	header    lipgloss.Style
	highlight lipgloss.Style
	muted     lipgloss.Style
	match     lipgloss.Style
	selected  lipgloss.Style
	danger    lipgloss.Style
	accent    lipgloss.Style
}

// themePalette 定义一组主题色值（256-color ANSI codes）
type themePalette struct {
	header    string
	highlight string
	muted     string
	match     string
	accent    string
	danger    string
}

// GitHub Dark 风格配色
var darkPalette = themePalette{
	header:    "75",  // 浅蓝 (#5fafff)
	highlight: "75",  // 浅蓝
	muted:     "245", // 灰色
	match:     "215", // 浅橙 (#ffaf5f)
	accent:    "114", // 浅绿 (#87d787)
	danger:    "196", // 鲜红 (#ff0000)
}

// GitHub Light 风格配色
var lightPalette = themePalette{
	header:    "26",  // 深蓝 (#005fd7)
	highlight: "26",  // 深蓝
	muted:     "242", // 中灰
	match:     "130", // 棕橙 (#af5f00)
	accent:    "28",  // 深绿 (#008700)
	danger:    "160", // 深红 (#d70000)
}

// newStyles 创建样式集，主题通过终端环境自动检测。
// 颜色降采样交由 bubbletea v2 内置渲染器处理。
func newStyles(colorsEnabled bool) *styles {
	if !colorsEnabled {
		return &styles{
			header:    lipgloss.NewStyle(),
			highlight: lipgloss.NewStyle(),
			muted:     lipgloss.NewStyle(),
			match:     lipgloss.NewStyle(),
			selected:  lipgloss.NewStyle(),
			danger:    lipgloss.NewStyle(),
			accent:    lipgloss.NewStyle(),
		}
	}

	p := darkPalette
	if config.DetectTheme() == "light" {
		p = lightPalette
	}

	return &styles{
		header:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.header)),
		highlight: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.highlight)),
		muted:     lipgloss.NewStyle().Foreground(lipgloss.Color(p.muted)),
		match:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.match)),
		selected:  lipgloss.NewStyle().Bold(true),
		danger:    lipgloss.NewStyle().Foreground(lipgloss.Color(p.danger)).Strikethrough(true),
		accent:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(p.accent)),
	}
}

// render 渲染带样式的文本，颜色降采样由 bubbletea 渲染器统一处理
func (s *styles) render(style lipgloss.Style, text string) string {
	return style.Render(text)
}

// renderHeader 渲染标题 + 分隔线 + 搜索栏 + 来源过滤 + 分隔线
func renderHeader(m *SelectorModel) string {
	var b strings.Builder
	w := m.width
	if w <= 0 {
		w = 80
	}
	sep := m.styles.render(m.styles.muted, strings.Repeat("─", w))

	b.WriteString(m.styles.render(m.styles.header, msgs().Title) + "\n")
	b.WriteString(sep + "\n")
	b.WriteString(msgs().SearchPrefix + m.textInput.View())

	// 在搜索栏右侧显示来源过滤标签
	if len(m.sourceOptions) > 1 {
		filterLabel := renderSourceTabs(m)
		searchWidth := lipgloss.Width(msgs().SearchPrefix + m.textInput.View())
		padding := w - searchWidth - lipgloss.Width(filterLabel)
		if padding > 0 {
			b.WriteString(strings.Repeat(" ", padding))
		} else {
			b.WriteString("  ")
		}
		b.WriteString(filterLabel)
	}
	b.WriteString("\n")
	b.WriteString(sep + "\n")

	return b.String()
}

func renderSourceTabs(m *SelectorModel) string {
	var parts []string
	for _, opt := range m.sourceOptions {
		label := opt
		if label == "" {
			label = msgs().FilterAll
		}
		if opt == m.sourceFilter {
			parts = append(parts, m.styles.render(m.styles.accent, "["+label+"]"))
		} else {
			parts = append(parts, m.styles.render(m.styles.muted, label))
		}
	}
	return strings.Join(parts, " ")
}

// renderFooter 渲染底部状态栏/快捷键提示
func renderFooter(m *SelectorModel) string {
	var b strings.Builder

	// "Create new" 行
	if input := strings.TrimSpace(m.textInput.Value()); input != "" {
		name := strings.ReplaceAll(input, " ", "-")
		b.WriteString("\n" + msgs().CreateNew + m.styles.render(m.styles.accent, name) + "\n")
	}

	w := m.width
	if w <= 0 {
		w = 80
	}
	sep := m.styles.render(m.styles.muted, strings.Repeat("─", w))
	b.WriteString(sep + "\n")

	if m.deleteStatus != "" {
		b.WriteString(m.styles.render(m.styles.muted, m.deleteStatus))
	} else if m.deleteMode {
		count := len(m.markedForDeletion)
		b.WriteString(m.styles.render(m.styles.muted,
			fmt.Sprintf(msgs().DeleteMode, count)))
	} else {
		b.WriteString(m.styles.render(m.styles.muted, msgs().HintBar))
	}

	return b.String()
}
