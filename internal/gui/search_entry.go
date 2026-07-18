package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// searchEntry 在聚焦时拦截导航键，保留普通字符输入。
// 必须 AcceptsTab()==true，否则 GLFW 驱动在 TypedKey 之前用 FocusNext 吃掉 Tab。
type searchEntry struct {
	widget.Entry
	onTab  func(delta int)
	onMove func(delta int)
	onOpen func()
}

func newSearchEntry(onTab func(delta int), onMove func(delta int), onOpen func()) *searchEntry {
	e := &searchEntry{onTab: onTab, onMove: onMove, onOpen: onOpen}
	e.ExtendBaseWidget(e)
	return e
}

// AcceptsTab 让 Tab/Shift+Tab 进入 TypedKey，而不是焦点遍历。
func (e *searchEntry) AcceptsTab() bool { return true }

func (e *searchEntry) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyTab:
		delta := 1
		if currentKeyModifiers()&fyne.KeyModifierShift != 0 {
			delta = -1
		}
		if e.onTab != nil {
			e.onTab(delta)
		}
		return
	case fyne.KeyUp:
		if e.onMove != nil {
			e.onMove(-1)
		}
		return
	case fyne.KeyDown:
		if e.onMove != nil {
			e.onMove(1)
		}
		return
	case fyne.KeyReturn, fyne.KeyEnter:
		if e.onOpen != nil {
			e.onOpen()
		}
		return
	}
	e.Entry.TypedKey(key)
}

func currentKeyModifiers() fyne.KeyModifier {
	drv, ok := fyne.CurrentApp().Driver().(desktop.Driver)
	if !ok {
		return 0
	}
	return drv.CurrentKeyModifiers()
}
