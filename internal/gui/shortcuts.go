package gui

import "fyne.io/fyne/v2"

type ctrlShortcut struct {
	key fyne.KeyName
}

func (s ctrlShortcut) ShortcutName() string { return "Ctrl+" + string(s.key) }
func (s ctrlShortcut) Key() fyne.KeyName    { return s.key }
func (s ctrlShortcut) Mod() fyne.KeyModifier {
	return fyne.KeyModifierControl
}
