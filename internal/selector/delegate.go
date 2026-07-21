package selector

import (
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

// EntryDelegate 自定义条目渲染器。
type EntryDelegate struct {
	markedForDeletion map[string]bool
	width             int
	styles            *Styles
}

// Height 返回每行占用的高度：内容行 + 分隔线。
func (d *EntryDelegate) Height() int { return 2 }

// Spacing 返回行间距（已由分隔线承担，此处为 0）。
func (d *EntryDelegate) Spacing() int { return 0 }

// Update 处理气泡消息（本委托无状态更新）。
func (d *EntryDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

// Render 渲染单行条目及其分隔线。
func (d *EntryDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	entry, ok := item.(MatchedEntry)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	isMarked := d.markedForDeletion[entry.Entry.Path]

	content := d.renderContent(entry, isSelected, isMarked)
	fmt.Fprint(w, content)
	fmt.Fprint(w, "\n")
	fmt.Fprint(w, d.renderSeparator())
}

func (d *EntryDelegate) renderContent(entry MatchedEntry, isSelected, isMarked bool) string {
	rowStyle := d.rowStyle(isSelected, isMarked)
	content := d.renderRawContent(entry, isSelected, isMarked)
	if !d.styles.ColorsEnabled() {
		return content
	}
	return d.fillRowBackground(rowStyle, content)
}

func (d *EntryDelegate) rowStyle(isSelected, isMarked bool) lipgloss.Style {
	switch {
	case isMarked:
		return d.styles.DangerSurface
	case isSelected:
		return d.styles.SurfaceSelected
	default:
		return lipgloss.NewStyle()
	}
}

func (d *EntryDelegate) renderRawContent(entry MatchedEntry, isSelected, isMarked bool) string {
	arrow := d.renderArrow(isSelected, isMarked)
	icon := d.renderIcon(isSelected, isMarked)
	name := d.renderName(entry, isSelected, isMarked)
	meta := d.renderMeta(entry, isSelected, isMarked)

	left := arrow + icon + name
	leftW := lipgloss.Width(left)
	metaW := lipgloss.Width(meta)
	maxContent := d.width - 1

	if leftW+metaW+2 > maxContent {
		return left
	}
	padding := maxContent - leftW - metaW
	return left + strings.Repeat(" ", padding) + meta
}

func (d *EntryDelegate) fillRowBackground(rowStyle lipgloss.Style, content string) string {
	if rowStyle.GetBackground() == nil {
		return content
	}
	contentW := lipgloss.Width(content)
	maxW := d.width - 1
	if contentW < maxW {
		content += strings.Repeat(" ", maxW-contentW)
	}
	return rowStyle.Render(content)
}

func (d *EntryDelegate) renderSeparator() string {
	w := d.width - 1
	if w < 0 {
		w = 0
	}
	return d.styles.Render(d.styles.Line, strings.Repeat("─", w))
}

func (d *EntryDelegate) renderArrow(isSelected, isMarked bool) string {
	if !isSelected {
		return d.padSegment("  ", isMarked)
	}
	arrowStyle := d.styles.SelectedArrow
	if isMarked {
		arrowStyle = arrowStyle.Inherit(d.styles.MarkedMeta)
	} else {
		arrowStyle = arrowStyle.Inherit(d.styles.SurfaceSelected)
	}
	return d.styles.Render(arrowStyle, iconSelected+" ")
}

func (d *EntryDelegate) renderIcon(isSelected, isMarked bool) string {
	icon := rowIconForEntry(isMarked)
	if isMarked {
		style := d.styles.MarkedIcon.Inherit(d.styles.MarkedMeta)
		return d.styles.Render(style, icon+" ")
	}
	style := d.styles.FolderIcon
	if isSelected {
		style = style.Inherit(d.styles.SurfaceSelected)
	}
	return d.styles.Render(style, icon+" ")
}

func (d *EntryDelegate) renderName(entry MatchedEntry, isSelected, isMarked bool) string {
	name := entry.Entry.Basename
	positions := entry.HighlightPositions
	posSet := make(map[int]bool, len(positions))
	for _, p := range positions {
		posSet[p] = true
	}

	namePart, datePart := name, ""
	if loc := DateSuffixRe.FindStringIndex(name); loc != nil {
		namePart = name[:loc[0]]
		datePart = name[loc[0]:]
	}

	nameBase := d.styles.Foreground
	dateBase := d.styles.Muted
	switch {
	case isMarked:
		nameBase = d.styles.MarkedName
		dateBase = d.styles.Muted.Inherit(d.styles.MarkedMeta)
	case isSelected:
		nameBase = d.styles.SurfaceSelected
		dateBase = d.styles.Muted.Inherit(d.styles.SurfaceSelected)
	}

	var b strings.Builder
	d.writeHighlighted(&b, namePart, posSet, 0, nameBase, isMarked)
	if datePart != "" {
		d.writeHighlighted(&b, datePart, posSet, len(namePart), dateBase, isMarked)
	}
	return b.String()
}

// writeHighlighted 按匹配位置分段写入，匹配段使用 match 样式。
func (d *EntryDelegate) writeHighlighted(b *strings.Builder, text string, posSet map[int]bool, offset int, baseStyle lipgloss.Style, isMarked bool) {
	i := 0
	for i < len(text) {
		if posSet[offset+i] {
			j := i
			for j < len(text) && posSet[offset+j] {
				_, sz := utf8.DecodeRuneInString(text[j:])
				j += sz
			}
			matchStyle := d.styles.Match.Inherit(baseStyle)
			if isMarked {
				matchStyle = matchStyle.Strikethrough(true)
			}
			b.WriteString(d.styles.Render(matchStyle, text[i:j]))
			i = j
		} else {
			j := i
			for j < len(text) && !posSet[offset+j] {
				_, sz := utf8.DecodeRuneInString(text[j:])
				j += sz
			}
			if baseStyle.GetForeground() != nil || baseStyle.GetBackground() != nil || baseStyle.GetBold() || baseStyle.GetStrikethrough() {
				b.WriteString(d.styles.Render(baseStyle, text[i:j]))
			} else {
				b.WriteString(text[i:j])
			}
			i = j
		}
	}
}

func (d *EntryDelegate) renderMeta(entry MatchedEntry, isSelected, isMarked bool) string {
	timeStr := FormatTimeAgo(time.Since(entry.Entry.Mtime))
	scoreBar := renderScoreBar(entry.Score, 5)

	barText := d.styles.Render(d.styles.ScoreBarFilled, scoreBar.filled) + d.styles.Render(d.styles.ScoreBarEmpty, scoreBar.empty)
	var parts []string
	parts = append(parts, barText)
	parts = append(parts, timeStr)
	if entry.Entry.Source != "" && entry.Entry.Source != "tries" {
		parts = append(parts, d.styles.Render(d.styles.SourcePill, " "+entry.Entry.Source+" "))
	}

	metaText := strings.Join(parts, "  ")
	style := d.styles.Muted
	switch {
	case isMarked:
		style = d.styles.MarkedMeta
	case isSelected:
		style = d.styles.Muted.Inherit(d.styles.SurfaceSelected)
	}
	return d.styles.Render(style, metaText)
}

func (d *EntryDelegate) padSegment(text string, isMarked bool) string {
	if !isMarked {
		return text
	}
	return d.styles.Render(d.styles.MarkedMeta, text)
}

type scoreBar struct {
	filled string
	empty  string
}

// renderScoreBar 将分数转换为 ASCII 块条形图，maxBlocks 控制最大块数。
func renderScoreBar(score float64, maxBlocks int) scoreBar {
	if maxBlocks <= 0 {
		maxBlocks = 5
	}
	ratio := score / 5.0
	if ratio > 1 {
		ratio = 1
	}
	if ratio < 0 {
		ratio = 0
	}
	filledCount := int(ratio*float64(maxBlocks) + 0.5)
	emptyCount := maxBlocks - filledCount
	return scoreBar{
		filled: strings.Repeat("█", filledCount),
		empty:  strings.Repeat("░", emptyCount),
	}
}
