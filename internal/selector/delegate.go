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
	styles     *styles
	isSelected bool
	isMarked   bool
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
	return rowCtx{
		styles:     d.styles,
		isSelected: isSelected,
		isMarked:   isMarked,
	}
}

func (d *EntryDelegate) renderArrow(isSelected bool, rc *rowCtx) string {
	if isSelected {
		return d.styles.render(rc.styles.highlight, "→ ")
	}
	return "  "
}

func (d *EntryDelegate) renderIcon(isMarked bool, rc *rowCtx) string {
	icon := "  "
	if isMarked {
		icon = "* "
	}
	return icon
}

func (d *EntryDelegate) renderMeta(entry MatchedEntry, rc *rowCtx) string {
	timeStr := FormatTimeAgo(time.Since(entry.Entry.Mtime))
	scoreStr := fmt.Sprintf("%.1f", entry.Score)

	var metaText string
	if entry.Entry.Source != "" && entry.Entry.Source != "tries" {
		metaText = "[" + entry.Entry.Source + "] " + timeStr + ", " + scoreStr
	} else {
		metaText = timeStr + ", " + scoreStr
	}

	if rc.isMarked {
		return d.styles.render(rc.styles.danger, metaText)
	}
	if rc.isSelected {
		return d.styles.render(rc.styles.selected, metaText)
	}
	return d.styles.render(rc.styles.muted, metaText)
}

// assembleLine 拼接左侧内容和右侧元数据，用普通空格填充中间空白
func assembleLine(left, meta string, maxContent int, rc rowCtx) string {
	leftWidth := lipgloss.Width(left)
	metaWidth := lipgloss.Width(meta)

	if leftWidth+metaWidth+2 <= maxContent {
		padding := maxContent - leftWidth - metaWidth
		return left + strings.Repeat(" ", padding) + meta
	}
	return left
}

// renderName 渲染条目名称，含日期拆分和模糊高亮
func (d *EntryDelegate) renderName(entry MatchedEntry, rc *rowCtx) string {
	name := entry.Entry.Basename
	positions := entry.HighlightPositions

	// 如果被标记删除，直接扁平化渲染为红色删除线，不进行分段高亮，避免嵌套样式导致 ANSI 码解析泄露
	if rc.isMarked {
		return d.styles.render(rc.styles.danger, name)
	}

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

	// 确定基础样式
	var baseStyle lipgloss.Style
	if rc.isSelected {
		baseStyle = rc.styles.selected
	}

	d.writeHighlighted(&result, namePart, posSet, 0, rc, baseStyle)

	if datePart != "" {
		// 日期后缀基础样式：如果是选中项，继承选中项加粗效果
		dateBaseStyle := rc.styles.muted
		if rc.isSelected {
			dateBaseStyle = dateBaseStyle.Inherit(rc.styles.selected)
		}
		d.writeHighlighted(&result, datePart, posSet, len(namePart), rc, dateBaseStyle)
	}

	return result.String()
}

// writeHighlighted 将文本按匹配位置分段写入 builder，匹配段和普通段使用扁平化的继承样式渲染，完全杜绝嵌套 Render
func (d *EntryDelegate) writeHighlighted(b *strings.Builder, text string, posSet map[int]bool, offset int, rc *rowCtx, baseStyle lipgloss.Style) {
	i := 0
	for i < len(text) {
		if posSet[offset+i] {
			j := i
			for j < len(text) && posSet[offset+j] {
				j++
			}
			// 匹配样式继承基础样式（合并加粗或颜色）
			matchStyle := rc.styles.match.Inherit(baseStyle)
			b.WriteString(d.styles.render(matchStyle, text[i:j]))
			i = j
		} else {
			j := i
			for j < len(text) && !posSet[offset+j] {
				j++
			}
			// 普通样式：如果有前景色或特殊效果，使用 baseStyle 渲染；否则直接写入
			if baseStyle.GetForeground() != nil || baseStyle.GetBold() {
				b.WriteString(d.styles.render(baseStyle, text[i:j]))
			} else {
				b.WriteString(text[i:j])
			}
			i = j
		}
	}
}
