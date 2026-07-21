package gui

import "fyne.io/fyne/v2/theme"

// rowVisualKind 列表行背景优先级：marked > selected > hovered > none。
type rowVisualKind int

const (
	rowVisualNone rowVisualKind = iota
	rowVisualHover
	rowVisualSelected
	rowVisualMarked
)

func rowVisualState(marked, selected, hovered bool) rowVisualKind {
	switch {
	case marked:
		return rowVisualMarked
	case selected:
		return rowVisualSelected
	case hovered:
		return rowVisualHover
	default:
		return rowVisualNone
	}
}

func (r *tapRow) setVisualState(marked, selected bool) {
	r.marked = marked
	r.selected = selected
	r.applyBackground()
}

func (r *tapRow) applyBackground() {
	switch rowVisualState(r.marked, r.selected, r.hovered) {
	case rowVisualMarked:
		r.setRowBackground(colorNameDangerSurface, true)
	case rowVisualSelected:
		r.setRowBackground(theme.ColorNameSelection, true)
	case rowVisualHover:
		r.setRowBackground(theme.ColorNameHover, true)
	default:
		r.setRowBackground(theme.ColorNameBackground, true)
	}
}
