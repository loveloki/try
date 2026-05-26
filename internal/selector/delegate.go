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
	isMarked := d.markedForDeletion[entry.Entry.Path]
	maxContent := d.width - 1

	// 确定行背景样式（选中或标记删除时整行需要背景）
	hasRowBg := isMarked || isSelected
	bgOnly := lipgloss.NewStyle()
	if isMarked {
		bgOnly = bgOnly.Background(d.styles.dangerBg.GetBackground())
	} else if isSelected {
		bgOnly = bgOnly.Background(d.styles.selectedBg.GetBackground())
	}

	// 辅助：为样式附加行背景色
	withBg := func(s lipgloss.Style) lipgloss.Style {
		if !hasRowBg {
			return s
		}
		return s.Background(bgOnly.GetBackground())
	}

	// 选中箭头（2 字符）
	arrow := "  "
	if isSelected {
		arrow = d.styles.render(withBg(d.styles.highlight), "→ ")
	} else if hasRowBg {
		arrow = d.styles.render(bgOnly, "  ")
	}

	// 图标（含尾部空格，约 3 字符显示宽度）
	icon := "📁 "
	if isMarked {
		icon = "🗑️ "
	}
	if hasRowBg {
		icon = d.styles.render(bgOnly, icon)
	}

	// 名称（含模糊高亮，传入背景样式）
	name := d.renderNameWithBg(entry, bgOnly, hasRowBg)

	// 右侧元数据
	timeStr := FormatTimeAgo(time.Since(entry.Entry.Mtime))
	scoreStr := fmt.Sprintf("%.1f", entry.Score)
	meta := d.styles.render(withBg(d.styles.muted), timeStr+", "+scoreStr)

	// 组装行
	left := arrow + icon + name
	leftWidth := lipgloss.Width(left)
	metaWidth := lipgloss.Width(meta)

	line := left
	if leftWidth+metaWidth+2 <= maxContent {
		padding := maxContent - leftWidth - metaWidth
		if hasRowBg {
			line = left + d.styles.render(bgOnly, strings.Repeat(" ", padding)) + meta
		} else {
			line = left + strings.Repeat(" ", padding) + meta
		}
	} else if leftWidth < maxContent && hasRowBg {
		remaining := maxContent - leftWidth
		if remaining > 0 {
			line = left + d.styles.render(bgOnly, strings.Repeat(" ", remaining))
		}
	}

	fmt.Fprint(w, line)
}

// renderNameWithBg 渲染条目名称，含日期拆分、模糊高亮和可选的行背景色
func (d *EntryDelegate) renderNameWithBg(entry MatchedEntry, bgOnly lipgloss.Style, hasRowBg bool) string {
	name := entry.Entry.Basename
	positions := entry.HighlightPositions

	withBg := func(s lipgloss.Style) lipgloss.Style {
		if !hasRowBg {
			return s
		}
		return s.Background(bgOnly.GetBackground())
	}

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
			j := i
			for j < len(namePart) && posSet[j] {
				j++
			}
			result.WriteString(d.styles.render(withBg(d.styles.match), namePart[i:j]))
			i = j
		} else {
			// 连续非高亮字符合并渲染
			j := i
			for j < len(namePart) && !posSet[j] {
				j++
			}
			if hasRowBg {
				result.WriteString(d.styles.render(bgOnly, namePart[i:j]))
			} else {
				result.WriteString(namePart[i:j])
			}
			i = j
		}
	}

	// 日期部分
	if datePart != "" {
		offset := len(namePart)
		var dateResult strings.Builder
		for j := 0; j < len(datePart); j++ {
			if posSet[offset+j] {
				k := j
				for k < len(datePart) && posSet[offset+k] {
					k++
				}
				dateResult.WriteString(d.styles.render(withBg(d.styles.match), datePart[j:k]))
				j = k - 1
			} else {
				dateResult.WriteByte(datePart[j])
			}
		}
		result.WriteString(d.styles.render(withBg(d.styles.muted), dateResult.String()))
	}

	return result.String()
}
