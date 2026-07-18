package gui

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// sourceTabVisual 来源 Tab 的背景/前景 token（inactive hover 用 foreground，避免 muted-on-hover）。
type sourceTabVisual struct {
	bg   fyne.ThemeColorName
	fg   fyne.ThemeColorName
	bold bool
}

func sourceTabVisualState(active, hovered bool) sourceTabVisual {
	switch {
	case active:
		return sourceTabVisual{theme.ColorNameSelection, colorNameHeader, true}
	case hovered:
		return sourceTabVisual{theme.ColorNameHover, theme.ColorNameForeground, false}
	default:
		return sourceTabVisual{theme.ColorNameButton, colorNameMuted, false}
	}
}

type sourceTab struct {
	widget.BaseWidget
	label   string
	count   int
	active  bool
	hovered bool
	onTap   func()
	bg      *canvas.Rectangle
	text    *canvas.Text
	badge   *canvas.Text
}

func newSourceTab(label string, count int, active bool, onTap func()) *sourceTab {
	t := &sourceTab{
		label:  label,
		count:  count,
		active: active,
		onTap:  onTap,
		bg:     canvas.NewRectangle(theme.Color(theme.ColorNameButton)),
		text:   canvas.NewText(label, theme.Color(colorNameMuted)),
		badge:  canvas.NewText(strconv.Itoa(count), theme.Color(colorNameMuted)),
	}
	t.text.TextSize = theme.TextSize() * 0.85
	t.badge.TextSize = theme.TextSize() * 0.75
	t.ExtendBaseWidget(t)
	t.applyVisual()
	return t
}

func (t *sourceTab) applyVisual() {
	v := sourceTabVisualState(t.active, t.hovered)
	t.bg.FillColor = theme.Color(v.bg)
	t.text.Color = theme.Color(v.fg)
	t.text.TextStyle = fyne.TextStyle{Bold: v.bold}
	t.badge.Color = theme.Color(v.fg)
	canvas.Refresh(t.bg)
	canvas.Refresh(t.text)
	canvas.Refresh(t.badge)
}

func (t *sourceTab) CreateRenderer() fyne.WidgetRenderer {
	content := container.NewPadded(container.NewHBox(t.text, t.badge))
	return widget.NewSimpleRenderer(container.NewStack(t.bg, content))
}

func (t *sourceTab) Tapped(_ *fyne.PointEvent) {
	if t.onTap != nil {
		t.onTap()
	}
}

func (t *sourceTab) MouseIn(_ *desktop.MouseEvent) {
	t.hovered = true
	t.applyVisual()
}

func (t *sourceTab) MouseMoved(_ *desktop.MouseEvent) {}

func (t *sourceTab) MouseOut() {
	t.hovered = false
	t.applyVisual()
}

func (t *sourceTab) Cursor() desktop.Cursor {
	return desktop.PointerCursor
}
