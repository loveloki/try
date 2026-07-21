package gui

import (
	"runtime"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// GLFW 驱动按字面量名称派发快捷键（"CustomDesktop:<修饰键>+<键名>"），
// 注册名的格式必须与之一致，否则处理器永远不会触发。
func TestControlShortcutName(t *testing.T) {
	t.Parallel()
	if got := controlShortcut(fyne.KeyT).ShortcutName(); got != "CustomDesktop:Control+T" {
		t.Errorf("controlShortcut(T).ShortcutName() = %q, want %q", got, "CustomDesktop:Control+T")
	}
}

func TestSettingsShortcutName(t *testing.T) {
	t.Parallel()
	want := "CustomDesktop:Super+,"
	if runtime.GOOS == "darwin" {
		want = "CustomDesktop:Command+,"
	}
	if got := settingsShortcut().ShortcutName(); got != want {
		t.Errorf("settingsShortcut().ShortcutName() = %q, want %q", got, want)
	}
}

// 模拟 GLFW 驱动派发，验证注册的处理器真正被触发。
func TestShortcutDispatch(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		register fyne.Shortcut
		dispatch fyne.Shortcut
	}{
		{"ctrl+t", controlShortcut(fyne.KeyT),
			&desktop.CustomShortcut{KeyName: fyne.KeyT, Modifier: fyne.KeyModifierControl}},
		{"settings", settingsShortcut(),
			&desktop.CustomShortcut{KeyName: fyne.KeyComma, Modifier: settingsModifier()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fired := false
			var h fyne.ShortcutHandler
			h.AddShortcut(tt.register, func(fyne.Shortcut) { fired = true })
			h.TypedShortcut(tt.dispatch)
			if !fired {
				t.Errorf("派发 %q 后处理器未触发（注册名 %q）",
					tt.dispatch.ShortcutName(), tt.register.ShortcutName())
			}
		})
	}
}

func TestSettingsShortcutDiffersFromControl(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "darwin" {
		t.Skip("仅 macOS 上设置快捷键修饰键与 Ctrl 不同")
	}
	if got, ctrl := settingsShortcut().ShortcutName(), controlShortcut(fyne.KeyComma).ShortcutName(); got == ctrl {
		t.Errorf("darwin 下设置快捷键 %q 不应与 Ctrl+, 相同", got)
	}
}
