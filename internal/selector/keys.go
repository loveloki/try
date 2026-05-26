package selector

import (
	"charm.land/bubbles/v2/key"
)

type keyMap struct {
	Enter key.Binding
	CtrlP key.Binding
	CtrlN key.Binding
	CtrlD key.Binding
	CtrlT key.Binding
	CtrlR key.Binding
	CtrlG key.Binding
	Quit  key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Enter: key.NewBinding(key.WithKeys("enter")),
		CtrlP: key.NewBinding(key.WithKeys("ctrl+p")),
		CtrlN: key.NewBinding(key.WithKeys("ctrl+n")),
		CtrlD: key.NewBinding(key.WithKeys("ctrl+d")),
		CtrlT: key.NewBinding(key.WithKeys("ctrl+t")),
		CtrlR: key.NewBinding(key.WithKeys("ctrl+r")),
		CtrlG: key.NewBinding(key.WithKeys("ctrl+g")),
		Quit:  key.NewBinding(key.WithKeys("esc", "ctrl+c")),
	}
}
