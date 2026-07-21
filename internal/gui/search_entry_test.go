package gui

import (
	"runtime"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
)

// 搜索框聚焦时 GLFW 驱动把快捷键派发给 Entry 而非 canvas；
// searchEntry 必须把应用已注册的快捷键（Cmd+, / Ctrl+*）经 onShortcut 回调转交应用处理器。
func TestSearchEntryForwardsCustomShortcut(t *testing.T) {
	t.Parallel()
	fired := false
	onShortcut := func(fyne.Shortcut) bool { fired = true; return true }

	e := newSearchEntry(nil, nil, nil, onShortcut)
	// 模拟 GLFW 驱动向聚焦的搜索框派发快捷键。
	e.TypedShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyComma, Modifier: settingsModifier()})
	if !fired {
		t.Error("搜索框聚焦时应用级快捷键未转发到应用处理器")
	}
}

// Entry 自身依赖的 CustomShortcut（词移动/删词/Home/End）不得被转发截走，
// 未命中应用快捷键表时必须回落 Entry 默认处理。
func TestSearchEntryKeepsEntryWordNavigation(t *testing.T) {
	test.NewApp()
	entry := newSearchEntry(nil, nil, nil, func(fyne.Shortcut) bool { return false })
	w := test.NewWindow(entry)
	defer w.Close()

	// 词移动修饰键：darwin 为 Alt，其余平台为 Ctrl（与 widget.Entry 一致）。
	mod := fyne.KeyModifierControl
	if runtime.GOOS == "darwin" {
		mod = fyne.KeyModifierAlt
	}
	entry.SetText("hello world")
	entry.CursorColumn = len("hello world")
	entry.TypedShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyLeft, Modifier: mod})
	if entry.CursorColumn != len("hello ") {
		t.Errorf("词移动后光标列 = %d, want %d", entry.CursorColumn, len("hello "))
	}
}

// 剪贴板类快捷键仍由 Entry 自身处理。
func TestSearchEntryKeepsClipboardShortcuts(t *testing.T) {
	a := test.NewApp()
	entry := newSearchEntry(nil, nil, nil, nil)
	w := test.NewWindow(entry)
	defer w.Close()

	a.Clipboard().SetContent("world")
	entry.TypedShortcut(&fyne.ShortcutPaste{Clipboard: a.Clipboard()})
	if got := entry.Text; got != "world" {
		t.Errorf("粘贴后输入框内容 = %q, want %q", got, "world")
	}
}
