package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// navList 在列表获焦时仍处理 Enter/Tab/Esc 等，避免被 List.TypedKey 吞掉。
type navList struct {
	widget.List
	onTypedKey func(*fyne.KeyEvent)
}

func newNavList(
	length func() int,
	create func() fyne.CanvasObject,
	update func(widget.ListItemID, fyne.CanvasObject),
	onKey func(*fyne.KeyEvent),
) *navList {
	l := &navList{onTypedKey: onKey}
	l.Length = length
	l.CreateItem = create
	l.UpdateItem = update
	l.ExtendBaseWidget(l)
	return l
}

func (l *navList) TypedKey(ev *fyne.KeyEvent) {
	if l.onTypedKey != nil {
		l.onTypedKey(ev)
		return
	}
	l.List.TypedKey(ev)
}

// AcceptsTab 列表获焦时也用 Tab 切换来源，不做焦点遍历。
func (l *navList) AcceptsTab() bool { return true }
