package dialog

import (
	lipgloss "charm.land/lipgloss/v2"
)

const (
	modalMinWidth = 40
	modalMaxWidth = 64
)

// modalBoxWidth 根据终端宽度计算弹窗外框宽度。
func modalBoxWidth(termWidth int) int {
	w := termWidth - 8
	if w > modalMaxWidth {
		w = modalMaxWidth
	}
	if w < modalMinWidth {
		w = modalMinWidth
	}
	return w
}

// renderModalBoxWithBorder 使用额外边框样式渲染弹窗（如删除确认的红色边框）。
func renderModalBoxWithBorder(border lipgloss.Style, content string, termWidth int) string {
	boxW := modalBoxWidth(termWidth)
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Align(lipgloss.Left).
		Width(boxW)
	box = box.Inherit(border)
	return box.Render(content)
}

// modalInnerWidth 弹窗内容区可用宽度（扣除圆角边框与水平 padding）。
func modalInnerWidth(termWidth int) int {
	// RoundedBorder 左右各 1 列 + Padding(0,1) 左右各 1 列
	const chrome = 4
	w := modalBoxWidth(termWidth) - chrome
	if w < 0 {
		return 0
	}
	return w
}
