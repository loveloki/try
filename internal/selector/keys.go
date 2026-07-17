package selector

import (
	"charm.land/bubbles/v2/key"
)

type keyMap struct {
	Enter    key.Binding
	CtrlP    key.Binding
	CtrlN    key.Binding
	CtrlD    key.Binding
	CtrlT    key.Binding
	CtrlR    key.Binding
	CtrlG    key.Binding
	CtrlA    key.Binding
	CtrlF    key.Binding
	Tab      key.Binding
	ShiftTab key.Binding
	Space    key.Binding
	Delete   key.Binding
	Slash    key.Binding
	Quit     key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Enter:    key.NewBinding(key.WithKeys("enter")),
		CtrlP:    key.NewBinding(key.WithKeys("ctrl+p")),
		CtrlN:    key.NewBinding(key.WithKeys("ctrl+n")),
		CtrlD:    key.NewBinding(key.WithKeys("ctrl+d")),
		CtrlT:    key.NewBinding(key.WithKeys("ctrl+t")),
		CtrlR:    key.NewBinding(key.WithKeys("ctrl+r")),
		CtrlG:    key.NewBinding(key.WithKeys("ctrl+g")),
		CtrlA:    key.NewBinding(key.WithKeys("ctrl+a")),
		CtrlF:    key.NewBinding(key.WithKeys("ctrl+f")),
		Tab:      key.NewBinding(key.WithKeys("tab")),
		ShiftTab: key.NewBinding(key.WithKeys("shift+tab")),
		Space:    key.NewBinding(key.WithKeys("space")),
		Delete:   key.NewBinding(key.WithKeys("delete")),
		Slash:    key.NewBinding(key.WithKeys("/")),
		Quit:     key.NewBinding(key.WithKeys("esc", "ctrl+c")),
	}
}
