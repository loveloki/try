package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const statusBarHeight float32 = 28

// shortcutHint 状态栏右侧一项：键帽 + 说明。
type shortcutHint struct {
	Key   string
	Label string
	Accent bool
}

type statusBar struct {
	widget.BaseWidget
	leftLabel *canvas.Text
	rightBox  *fyne.Container
	toastText string
	leftText  string
	hints     []shortcutHint
}

func newStatusBar() *statusBar {
	s := &statusBar{
		leftLabel: canvas.NewText("", theme.Color(colorNameMuted)),
		rightBox:  container.NewHBox(),
	}
	s.leftLabel.TextSize = theme.TextSize() * 0.75
	s.leftLabel.TextStyle = fyne.TextStyle{Monospace: true}
	s.ExtendBaseWidget(s)
	return s
}

func (s *statusBar) CreateRenderer() fyne.WidgetRenderer {
	line := canvas.NewRectangle(theme.Color(theme.ColorNameSeparator))
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	content := container.NewBorder(nil, nil, s.leftLabel, s.rightBox)
	padded := container.New(layout.NewCustomPaddedLayout(4, 4, contentInsetH, contentInsetH), content)
	return &statusBarRenderer{
		bar:     s,
		line:    line,
		bg:      bg,
		content: padded,
		objects: []fyne.CanvasObject{bg, line, padded},
	}
}

func (s *statusBar) SetContent(left string, hints []shortcutHint) {
	s.leftText = left
	s.hints = hints
	s.rebuild()
}

func (s *statusBar) SetToast(msg string) {
	s.toastText = msg
	s.rebuild()
}

func (s *statusBar) ClearToast() {
	s.toastText = ""
	s.rebuild()
}

func (s *statusBar) rebuild() {
	left := s.leftText
	if s.toastText != "" {
		left = s.toastText
	}
	s.leftLabel.Text = left
	s.leftLabel.Color = theme.Color(colorNameMuted)
	canvas.Refresh(s.leftLabel)

	s.rightBox.Objects = nil
	if s.toastText == "" {
		for _, h := range s.hints {
			s.rightBox.Add(newKeyCapHint(h))
		}
	}
	s.rightBox.Refresh()
	s.Refresh()
}

func newKeyCapHint(h shortcutHint) fyne.CanvasObject {
	labelColor := theme.Color(colorNameMuted)
	if h.Accent {
		labelColor = theme.Color(colorNameHeader)
	}
	label := canvas.NewText(h.Label, labelColor)
	label.TextSize = theme.TextSize() * 0.7
	label.TextStyle = fyne.TextStyle{Monospace: true}
	if h.Key == "" {
		return label
	}
	key := canvas.NewText(h.Key, labelColor)
	key.TextSize = theme.TextSize() * 0.7
	key.TextStyle = fyne.TextStyle{Monospace: true}
	border := canvas.NewRectangle(theme.Color(theme.ColorNameBackground))
	border.StrokeColor = theme.Color(theme.ColorNameSeparator)
	border.StrokeWidth = 1
	border.CornerRadius = 2
	keyCap := container.NewStack(border, container.NewPadded(key))
	return container.NewHBox(keyCap, label)
}

type statusBarRenderer struct {
	bar     *statusBar
	line    *canvas.Rectangle
	bg      *canvas.Rectangle
	content fyne.CanvasObject
	objects []fyne.CanvasObject
}

func (r *statusBarRenderer) Layout(size fyne.Size) {
	r.bg.Resize(size)
	r.line.Resize(fyne.NewSize(size.Width, 1))
	r.line.Move(fyne.NewPos(0, 0))
	r.content.Resize(fyne.NewSize(size.Width, size.Height))
	r.content.Move(fyne.NewPos(0, 0))
}

func (r *statusBarRenderer) MinSize() fyne.Size {
	return fyne.NewSize(120, statusBarHeight)
}

func (r *statusBarRenderer) Refresh() {
	r.bg.FillColor = theme.Color(theme.ColorNameBackground)
	r.line.FillColor = theme.Color(theme.ColorNameSeparator)
	canvas.Refresh(r.bg)
	canvas.Refresh(r.line)
	canvas.Refresh(r.bar.leftLabel)
	canvas.Refresh(r.bar.rightBox)
}

func (r *statusBarRenderer) Objects() []fyne.CanvasObject { return r.objects }
func (r *statusBarRenderer) Destroy()                     {}
