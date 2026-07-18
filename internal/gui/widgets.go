package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type tapRow struct {
	widget.BaseWidget
	inner        fyne.CanvasObject
	bg           *canvas.Rectangle
	onSelect     func()
	onCtrlSelect func()
	onOpen       func()
	minHeight    float32
	marked       bool
	selected     bool
	hovered      bool
	fromMouse    bool
}

func newTapRow(inner fyne.CanvasObject, minHeight float32) *tapRow {
	r := &tapRow{
		inner:     inner,
		bg:        canvas.NewRectangle(theme.Color(theme.ColorNameBackground)),
		minHeight: minHeight,
	}
	r.bg.Hide()
	r.ExtendBaseWidget(r)
	return r
}

func (r *tapRow) setRowBackground(name fyne.ThemeColorName, visible bool) {
	if visible {
		r.bg.FillColor = theme.Color(name)
		r.bg.Show()
	} else {
		r.bg.Hide()
	}
	canvas.Refresh(r.bg)
}

func (r *tapRow) CreateRenderer() fyne.WidgetRenderer {
	return &tapRowRenderer{row: r, objects: []fyne.CanvasObject{r.bg, r.inner}}
}

func (r *tapRow) MinSize() fyne.Size {
	min := r.inner.MinSize()
	if min.Height < r.minHeight {
		min.Height = r.minHeight
	}
	return min
}

func (r *tapRow) Tapped(_ *fyne.PointEvent) {
	if r.fromMouse {
		r.fromMouse = false
		return
	}
	if r.onSelect != nil {
		r.onSelect()
	}
}

func (r *tapRow) DoubleTapped(_ *fyne.PointEvent) {
	if r.onOpen != nil {
		r.onOpen()
	}
}

func (r *tapRow) MouseDown(e *desktop.MouseEvent) {
	if e.Button != desktop.MouseButtonPrimary {
		return
	}
	r.fromMouse = true
	if isMultiSelectModifier(e.Modifier) && r.onCtrlSelect != nil {
		r.onCtrlSelect()
		return
	}
	if r.onSelect != nil {
		r.onSelect()
	}
}

func isMultiSelectModifier(m fyne.KeyModifier) bool {
	return m&fyne.KeyModifierControl != 0 || m&fyne.KeyModifierSuper != 0
}

func (r *tapRow) MouseUp(_ *desktop.MouseEvent) {}

func (r *tapRow) MouseIn(_ *desktop.MouseEvent) {
	r.hovered = true
	r.applyBackground()
}

func (r *tapRow) MouseMoved(_ *desktop.MouseEvent) {}

func (r *tapRow) MouseOut() {
	r.hovered = false
	r.applyBackground()
}

type tapRowRenderer struct {
	row     *tapRow
	objects []fyne.CanvasObject
}

func (r *tapRowRenderer) Layout(size fyne.Size) {
	r.row.bg.Resize(size)
	r.row.bg.Move(fyne.NewPos(0, 0))
	r.row.inner.Resize(size)
	r.row.inner.Move(fyne.NewPos(0, 0))
}

func (r *tapRowRenderer) MinSize() fyne.Size           { return r.row.MinSize() }
func (r *tapRowRenderer) Refresh()                     { canvas.Refresh(r.row.inner) }
func (r *tapRowRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *tapRowRenderer) Destroy()                     {}

type entryRowWidget struct {
	*tapRow
	arrow *canvas.Text
	icon  *widget.Icon
	name  *widget.RichText
	meta  *canvas.Text
}

func newEntryRowWidget() *entryRowWidget {
	w := &entryRowWidget{
		arrow: canvas.NewText("", theme.Color(colorNameHeader)),
		icon:  widget.NewIcon(theme.FolderIcon()),
		name:  widget.NewRichText(),
		meta:  canvas.NewText("", theme.Color(colorNameMuted)),
	}
	w.arrow.TextSize = theme.TextSize()
	w.meta.TextSize = theme.TextSize() * 0.85
	inner := container.NewBorder(nil, nil,
		container.NewHBox(w.arrow, w.icon),
		w.meta,
		w.name,
	)
	w.tapRow = newTapRow(withHInset(inner), entryRowHeight)
	return w
}

type fileRowWidget struct {
	*tapRow
	check *widget.Check
	icon  *widget.Icon
	name  *canvas.Text
	meta  *canvas.Text
}

func newFileRowWidget() *fileRowWidget {
	w := &fileRowWidget{
		check: widget.NewCheck("", nil),
		icon:  widget.NewIcon(theme.FileIcon()),
		name:  canvas.NewText("", theme.Color(theme.ColorNameForeground)),
		meta:  canvas.NewText("", theme.Color(colorNameMuted)),
	}
	w.name.TextSize = theme.TextSize()
	w.meta.TextSize = theme.TextSize() * 0.85
	inner := container.NewBorder(nil, nil,
		container.NewHBox(w.check, w.icon),
		w.meta,
		w.name,
	)
	w.tapRow = newTapRow(withHInset(inner), entryRowHeight)
	return w
}

func buildHeader(title string, themeBtn fyne.CanvasObject) fyne.CanvasObject {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameButton))
	titleLabel := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	row := container.NewBorder(nil, nil, nil, themeBtn, titleLabel)
	return container.NewStack(bg, container.NewPadded(row))
}
