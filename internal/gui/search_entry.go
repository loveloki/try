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
	// onShortcut 尝试处理应用级快捷键，返回是否已处理；未处理时回落 Entry 默认行为。
	onShortcut func(fyne.Shortcut) bool
}

func newSearchEntry(onTab func(delta int), onMove func(delta int), onOpen func(), onShortcut func(fyne.Shortcut) bool) *searchEntry {
	e := &searchEntry{onTab: onTab, onMove: onMove, onOpen: onOpen, onShortcut: onShortcut}
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

// TypedShortcut 仅把应用已注册的快捷键（Ctrl+* / Cmd+,）经回调转交应用处理器，
// 其余快捷键（剪贴板类、Entry 自有的词移动/删词等 CustomShortcut）仍由 Entry 处理。
// GLFW 驱动在焦点控件实现 Shortcutable 时不再向 canvas 派发，必须在此桥接。
func (e *searchEntry) TypedShortcut(s fyne.Shortcut) {
	if _, ok := s.(*desktop.CustomShortcut); ok && e.onShortcut != nil && e.onShortcut(s) {
		return
	}
	e.Entry.TypedShortcut(s)
}

func currentKeyModifiers() fyne.KeyModifier {
	drv, ok := fyne.CurrentApp().Driver().(desktop.Driver)
	if !ok {
		return 0
	}
	return drv.CurrentKeyModifiers()
}
