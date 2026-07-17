package selector

import (
	"fmt"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
)

// headerLines 返回当前头部占用行数。
// GUI 结构：brand 行（标题+副标题）、分隔线、搜索行、分隔线、来源标签行。
func headerLines() int {
	return 5
}

// footerLines 返回当前底部占用行数，根据是否显示创建输入面板动态变化。
func footerLines(m *SelectorModel) int {
	lines := 2 // 分隔线 + 状态栏
	if strings.TrimSpace(m.textInput.Value()) != "" {
		lines++ // 创建新目录预览行
	}
	return lines
}

// bodyHeight 计算列表主体可用高度。
func bodyHeight(m *SelectorModel) int {
	h := m.height - headerLines() - footerLines(m)
	if h < 1 {
		h = 1
	}
	return h
}

// renderHeader 渲染与 GUI 对齐的头部：brand 行、分隔线、搜索栏、分隔线、来源标签。
func renderHeader(m *SelectorModel) string {
	var b strings.Builder
	w := m.width
	if w <= 0 {
		w = 80
	}

	sep := m.styles.Render(m.styles.Line, strings.Repeat("─", w))

	b.WriteString(renderBrandLine(m, w) + "\n")
	b.WriteString(sep + "\n")
	b.WriteString(renderSearchLine(m, w) + "\n")
	b.WriteString(sep + "\n")
	b.WriteString(renderSourceTabs(m) + "\n")

	return b.String()
}

// renderBrandLine 渲染品牌行：logo + try + subtitle，右侧无操作按钮（TUI 无主题切换）。
func renderBrandLine(m *SelectorModel, w int) string {
	logo := m.styles.Render(m.styles.SurfaceSelected, " try ")
	brand := logo + "  " + m.styles.Render(m.styles.Header, msgs().Title)
	subtitle := m.styles.Render(m.styles.Muted, m.basePath)
	line := brand + "  " + subtitle

	lineW := lipgloss.Width(line)
	if lineW < w {
		line += strings.Repeat(" ", w-lineW)
	}
	return line
}

// renderSearchLine 渲染搜索行：左侧 Search 图标 + 输入框，右侧 / kbd。
func renderSearchLine(m *SelectorModel, w int) string {
	prefix := m.styles.Render(m.styles.Muted, iconSearch+" "+msgs().SearchPrefix)
	input := m.textInput.View()
	search := prefix + input
	searchW := lipgloss.Width(search)

	kbd := m.styles.Render(m.styles.KeyBadge, " / ")
	kbdW := lipgloss.Width(kbd)

	gap := w - searchW - kbdW
	if gap < 1 {
		gap = 1
	}
	return search + strings.Repeat(" ", gap) + kbd
}

// renderSourceTabs 渲染来源过滤标签栏。
func renderSourceTabs(m *SelectorModel) string {
	var parts []string
	for _, opt := range m.sourceOptions {
		label := sourceTabLabel(opt)
		count := m.sourceCounts[opt]
		parts = append(parts, renderSourceTab(m, opt, label, count))
	}
	return strings.Join(parts, " ")
}

func sourceTabLabel(opt string) string {
	switch opt {
	case "":
		return msgs().FilterAll
	default:
		return opt
	}
}

func renderSourceTab(m *SelectorModel, opt, label string, count int) string {
	isActive := opt == m.sourceFilter
	text := fmt.Sprintf(" %s ", label)
	if isActive {
		pill := m.styles.Render(m.styles.Header, fmt.Sprintf(" %d ", count))
		return m.styles.Render(m.styles.SurfaceSelected, text) + pill
	}
	pill := m.styles.Render(m.styles.SourcePill, fmt.Sprintf(" %d ", count))
	return m.styles.Render(m.styles.SurfaceHover, text) + pill
}

// renderFooter 渲染底部状态栏/快捷键提示。
func renderFooter(m *SelectorModel) string {
	var b strings.Builder

	if input := strings.TrimSpace(m.textInput.Value()); input != "" {
		b.WriteString("\n" + m.styles.Render(m.styles.Accent, msgs().CreateNew) + m.styles.Render(m.styles.Highlight, input) + "\n")
	}

	w := m.width
	if w <= 0 {
		w = 80
	}
	sep := m.styles.Render(m.styles.Line, strings.Repeat("─", w))
	b.WriteString(sep + "\n")

	if m.deleteStatus != "" {
		b.WriteString(m.styles.Render(m.styles.Muted, m.deleteStatus))
	} else if m.deleteMode {
		count := len(m.markedForDeletion)
		left := m.styles.Render(m.styles.DeleteModeBadge, msgs().DeleteModeLabel)
		if count > 0 {
			left += " " + m.styles.Render(m.styles.Danger, fmt.Sprintf(msgs().MarkedCount, count))
		}
		right := m.styles.JoinKeyBadges([]string{"Esc", "Enter"})
		b.WriteString(joinFooterSides(left, right, w, m.styles))
	} else {
		left := m.styles.Render(m.styles.Muted, fmt.Sprintf(msgs().ItemCount, len(m.cachedResults)))
		right := m.styles.JoinKeyBadges([]string{"↑↓", "Enter", "⌃T", "⌃D", "Tab"})
		b.WriteString(joinFooterSides(left, right, w, m.styles))
	}

	return b.String()
}

// joinFooterSides 将底部左右两侧内容按宽度对齐。
func joinFooterSides(left, right string, width int, st *Styles) string {
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	gap := width - leftW - rightW
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

// renderEmptyState 渲染空状态提示。
func renderEmptyState(m *SelectorModel) string {
	w := m.width
	if w <= 0 {
		w = 80
	}
	h := bodyHeight(m)
	if h < 3 {
		h = 3
	}

	var message string
	query := strings.TrimSpace(m.textInput.Value())
	switch {
	case m.allTries == nil:
		message = iconLoading + "  " + msgs().LoadingHint
	case query != "":
		message = fmt.Sprintf("%s  %s\n\n%s", iconSearch, fmt.Sprintf(msgs().NoMatchesHint, query), msgs().CreateHint)
	default:
		message = fmt.Sprintf("%s  %s\n\n%s", iconEmptyFolder, msgs().EmptyStateHint, msgs().CreateHint)
	}

	return m.styles.Render(m.styles.Muted, lipgloss.Place(w, h, lipgloss.Center, lipgloss.Center, message))
}
