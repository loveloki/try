//go:build windows || linux

package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func newCustomTitleBarLayout(w fyne.Window, title string, body fyne.CanvasObject) fyne.CanvasObject {
	bar := newTitleBar(w, title)
	return container.NewBorder(bar, nil, nil, nil, body)
}

type titleBar struct {
	widget.BaseWidget
	window   fyne.Window
	title    string
	maxBtn   *widget.Button
	controls *fyne.Container
}

func newTitleBar(w fyne.Window, title string) *titleBar {
	t := &titleBar{window: w, title: title}
	t.ExtendBaseWidget(t)

	minBtn := t.controlButton(theme.WindowMinimizeIcon(), func() {
		nativeMinimize(t.window)
	})
	t.maxBtn = t.controlButton(theme.WindowMaximizeIcon(), t.toggleMaximize)
	closeBtn := t.controlButton(theme.WindowCloseIcon(), func() {
		w.Close()
	})
	t.controls = container.NewHBox(minBtn, t.maxBtn, closeBtn)
	return t
}

func (t *titleBar) controlButton(icon fyne.Resource, fn func()) *widget.Button {
	btn := widget.NewButtonWithIcon("", icon, func() {
		if fn != nil {
			fn()
		}
	})
	btn.Importance = widget.LowImportance
	return btn
}

func (t *titleBar) toggleMaximize() {
	nativeToggleMaximize(t.window)
	t.refreshMaximizeIcon()
}

func (t *titleBar) refreshMaximizeIcon() {
	if t.maxBtn == nil {
		return
	}
	if nativeIsMaximized(t.window) {
		t.maxBtn.SetIcon(theme.ViewRestoreIcon())
		return
	}
	t.maxBtn.SetIcon(theme.WindowMaximizeIcon())
}

func (t *titleBar) CreateRenderer() fyne.WidgetRenderer {
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameInputBackground))
	line := canvas.NewRectangle(theme.Color(theme.ColorNameSeparator))
	line.SetMinSize(fyne.NewSize(0, 1))
	drag := newTitleDragArea(t.window, t.title)
	return &titleBarRenderer{
		bar:      t,
		bg:       bg,
		line:     line,
		drag:     drag,
		controls: t.controls,
	}
}

type titleDragArea struct {
	widget.BaseWidget
	window fyne.Window
	label  *widget.Label
}

func newTitleDragArea(w fyne.Window, title string) *titleDragArea {
	d := &titleDragArea{
		window: w,
		label:  widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{}),
	}
	d.ExtendBaseWidget(d)
	return d
}

func (d *titleDragArea) MouseDown(ev *desktop.MouseEvent) {
	if ev.Button != desktop.MouseButtonPrimary {
		return
	}
	nativeBeginSystemDrag(d.window)
}

func (d *titleDragArea) MouseUp(*desktop.MouseEvent) {}

func (d *titleDragArea) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(d.label)
}

type titleBarRenderer struct {
	bar      *titleBar
	bg       *canvas.Rectangle
	line     *canvas.Rectangle
	drag     *titleDragArea
	controls *fyne.Container
	objects  []fyne.CanvasObject
}

func (r *titleBarRenderer) Objects() []fyne.CanvasObject {
	if r.objects == nil {
		r.objects = []fyne.CanvasObject{r.bg, r.line, r.drag, r.controls}
	}
	return r.objects
}

func (r *titleBarRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.bg.Move(fyne.NewPos(0, 0))

	lineY := size.Height - 1
	r.line.Resize(fyne.NewSize(size.Width, 1))
	r.line.Move(fyne.NewPos(0, lineY))

	controlsMin := r.controls.MinSize()
	controlsW := controlsMin.Width
	if controlsW < titleBarButtonWidth*3 {
		controlsW = titleBarButtonWidth * 3
	}
	r.controls.Resize(fyne.NewSize(controlsW, size.Height))
	r.controls.Move(fyne.NewPos(size.Width-controlsW, 0))

	dragW := size.Width - controlsW
	if dragW < 0 {
		dragW = 0
	}
	r.drag.Resize(fyne.NewSize(dragW, size.Height))
	r.drag.Move(fyne.NewPos(0, 0))
}

func (r *titleBarRenderer) MinSize() fyne.Size {
	return fyne.NewSize(minWindowWidth, titleBarHeight)
}

func (r *titleBarRenderer) Refresh() {
	r.bg.FillColor = theme.Color(theme.ColorNameInputBackground)
	r.bg.Refresh()
	r.line.FillColor = theme.Color(theme.ColorNameSeparator)
	r.line.Refresh()
	r.drag.label.SetText(r.bar.title)
	r.bar.refreshMaximizeIcon()
}

func (r *titleBarRenderer) Destroy() {}
