package gui

import (
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// controlShortcut 构造 Ctrl 快捷键。
// GLFW 驱动按 desktop.CustomShortcut 的 ShortcutName 派发，注册必须使用同一类型才能匹配。
func controlShortcut(key fyne.KeyName) fyne.Shortcut {
	return &desktop.CustomShortcut{KeyName: key, Modifier: fyne.KeyModifierControl}
}

// settingsShortcut 设置对话框快捷键：macOS Cmd+,，其他平台 Ctrl+,。
func settingsShortcut() fyne.Shortcut {
	return &desktop.CustomShortcut{KeyName: fyne.KeyComma, Modifier: settingsModifier()}
}

// settingsModifier 设置快捷键修饰键：macOS 用 Cmd，其他平台用 Ctrl。
func settingsModifier() fyne.KeyModifier {
	if runtime.GOOS == "darwin" {
		return fyne.KeyModifierSuper
	}
	return fyne.KeyModifierControl
}
