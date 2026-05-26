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

func (d *EntryDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	entry, ok := item.(MatchedEntry)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	maxContent := d.width - 1

	// 选中箭头（2 字符）
	arrow := "  "
	if isSelected {
		arrow = d.styles.render(d.styles.highlight, "→ ")
	}

	// 图标（含尾部空格，约 3 字符显示宽度）
	icon := "📁 "
	if d.markedForDeletion[entry.Entry.Path] {
		icon = "🗑️ "
	}

	// 名称（含模糊高亮）
	name := d.renderName(entry)

	// 右侧元数据
	timeStr := FormatTimeAgo(time.Since(entry.Entry.Mtime))
	scoreStr := fmt.Sprintf("%.1f", entry.Score)
	meta := d.styles.render(d.styles.muted, timeStr+", "+scoreStr)

	// 组装行
	left := arrow + icon + name
	leftWidth := lipgloss.Width(left)
	metaWidth := lipgloss.Width(meta)

	line := left
	if leftWidth+metaWidth+2 <= maxContent {
		padding := maxContent - leftWidth - metaWidth
		line = left + strings.Repeat(" ", padding) + meta
	} else if leftWidth < maxContent {
		line = left
	}

	// 选中行或删除标记行的背景色
	if d.markedForDeletion[entry.Entry.Path] {
		line = d.styles.render(d.styles.dangerBg, line)
	} else if isSelected {
		line = d.styles.render(d.styles.selectedBg, line)
	}

	fmt.Fprint(w, line)
}

// renderName 渲染条目名称，含日期拆分和模糊高亮
func (d *EntryDelegate) renderName(entry MatchedEntry) string {
	name := entry.Entry.Basename
	positions := entry.HighlightPositions

	// 拆分日期后缀
	datePart := ""
	namePart := name
	if loc := DateSuffixRe.FindStringIndex(name); loc != nil {
		namePart = name[:loc[0]]
		datePart = name[loc[0]:]
	}

	// 对 namePart 应用高亮
	posSet := make(map[int]bool, len(positions))
	for _, p := range positions {
		posSet[p] = true
	}

	var result strings.Builder
	i := 0
	for i < len(namePart) {
		if posSet[i] {
			// 找出连续高亮区间
			j := i
			for j < len(namePart) && posSet[j] {
				j++
			}
			result.WriteString(d.styles.render(d.styles.match, namePart[i:j]))
			i = j
		} else {
			result.WriteByte(namePart[i])
			i++
		}
	}

	// 日期部分：高亮位置超过 namePart 长度的字符也高亮
	if datePart != "" {
		offset := len(namePart)
		var dateResult strings.Builder
		for j := 0; j < len(datePart); j++ {
			if posSet[offset+j] {
				k := j
				for k < len(datePart) && posSet[offset+k] {
					k++
				}
				dateResult.WriteString(d.styles.render(d.styles.match, datePart[j:k]))
				j = k - 1
			} else {
				dateResult.WriteByte(datePart[j])
			}
		}
		result.WriteString(d.styles.render(d.styles.muted, dateResult.String()))
	}

	return result.String()
}
