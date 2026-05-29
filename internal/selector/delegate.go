package selector

import (
	"fmt"
	"io"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/list"
	lipgloss "charm.land/lipgloss/v2"
)

// EntryDelegate 自定义条目渲染器
type EntryDelegate struct {
	markedForDeletion map[string]bool
	width             int
	styles            *styles
}

func (d *EntryDelegate) Height() int                              { return 1 }
func (d *EntryDelegate) Spacing() int                             { return 0 }
func (d *EntryDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

// rowCtx 封装单行渲染的上下文状态，避免在子函数间传递大量参数
type rowCtx struct {
	styles   *styles
	bgOnly   lipgloss.Style
	hasRowBg bool
}

// withBg 为样式附加行背景色
func (r *rowCtx) withBg(s lipgloss.Style) lipgloss.Style {
	if !r.hasRowBg {
		return s
	}
	return s.Background(r.bgOnly.GetBackground())
}

func (d *EntryDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	entry, ok := item.(MatchedEntry)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	isMarked := d.markedForDeletion[entry.Entry.Path]

	rc := d.newRowCtx(isSelected, isMarked)
	arrow := d.renderArrow(isSelected, &rc)
	icon := d.renderIcon(isMarked, &rc)
	name := d.renderName(entry, &rc)
	meta := d.renderMeta(entry, &rc)

	line := assembleLine(arrow+icon+name, meta, d.width-1, rc)
	fmt.Fprint(w, line)
}

func (d *EntryDelegate) newRowCtx(isSelected, isMarked bool) rowCtx {
	rc := rowCtx{styles: d.styles, hasRowBg: isMarked || isSelected}
	rc.bgOnly = lipgloss.NewStyle()
	if isMarked {
		rc.bgOnly = rc.bgOnly.Background(d.styles.dangerBg.GetBackground())
	} else if isSelected {
		rc.bgOnly = rc.bgOnly.Background(d.styles.selectedBg.GetBackground())
	}
	return rc
}

func (d *EntryDelegate) renderArrow(isSelected bool, rc *rowCtx) string {
	if isSelected {
		return d.styles.render(rc.withBg(d.styles.highlight), "→ ")
	}
	if rc.hasRowBg {
		return d.styles.render(rc.bgOnly, "  ")
	}
	return "  "
}

func (d *EntryDelegate) renderIcon(isMarked bool, rc *rowCtx) string {
	icon := "📁 "
	if isMarked {
		icon = "🗑️ "
	}
	if rc.hasRowBg {
		return d.styles.render(rc.bgOnly, icon)
	}
	return icon
}

func (d *EntryDelegate) renderMeta(entry MatchedEntry, rc *rowCtx) string {
	timeStr := FormatTimeAgo(time.Since(entry.Entry.Mtime))
	scoreStr := fmt.Sprintf("%.1f", entry.Score)
	return d.styles.render(rc.withBg(d.styles.muted), timeStr+", "+scoreStr)
}

// assembleLine 拼接左侧内容和右侧元数据，用背景色填充中间空白
func assembleLine(left, meta string, maxContent int, rc rowCtx) string {
	leftWidth := lipgloss.Width(left)
	metaWidth := lipgloss.Width(meta)

	if leftWidth+metaWidth+2 <= maxContent {
		padding := maxContent - leftWidth - metaWidth
		if rc.hasRowBg {
			return left + rc.styles.render(rc.bgOnly, strings.Repeat(" ", padding)) + meta
		}
		return left + strings.Repeat(" ", padding) + meta
	}
	if leftWidth < maxContent && rc.hasRowBg {
		remaining := maxContent - leftWidth
		if remaining > 0 {
			return left + rc.styles.render(rc.bgOnly, strings.Repeat(" ", remaining))
		}
	}
	return left
}

// renderName 渲染条目名称，含日期拆分和模糊高亮
func (d *EntryDelegate) renderName(entry MatchedEntry, rc *rowCtx) string {
	name := entry.Entry.Basename
	positions := entry.HighlightPositions

	posSet := make(map[int]bool, len(positions))
	for _, p := range positions {
		posSet[p] = true
	}

	// 拆分名称和日期后缀
	namePart, datePart := name, ""
	if loc := DateSuffixRe.FindStringIndex(name); loc != nil {
		namePart = name[:loc[0]]
		datePart = name[loc[0]:]
	}

	var result strings.Builder
	d.writeHighlighted(&result, namePart, posSet, 0, rc)

	if datePart != "" {
		var dateResult strings.Builder
		d.writeHighlighted(&dateResult, datePart, posSet, len(namePart), rc)
		result.WriteString(d.styles.render(rc.withBg(d.styles.muted), dateResult.String()))
	}

	return result.String()
}

// writeHighlighted 将文本按匹配位置分段写入 builder，匹配段高亮、非匹配段附加背景
func (d *EntryDelegate) writeHighlighted(b *strings.Builder, text string, posSet map[int]bool, offset int, rc *rowCtx) {
	i := 0
	for i < len(text) {
		if posSet[offset+i] {
			j := i
			for j < len(text) && posSet[offset+j] {
				j++
			}
			b.WriteString(d.styles.render(rc.withBg(d.styles.match), text[i:j]))
			i = j
		} else {
			j := i
			for j < len(text) && !posSet[offset+j] {
				j++
			}
			if rc.hasRowBg {
				b.WriteString(d.styles.render(rc.bgOnly, text[i:j]))
			} else {
				b.WriteString(text[i:j])
			}
			i = j
		}
	}
}
