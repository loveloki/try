package gui

import (
	"testing"

	"fyne.io/fyne/v2"
)

func TestIsMultiSelectModifier(t *testing.T) {
	tests := []struct {
		name string
		mod  fyne.KeyModifier
		want bool
	}{
		{"none", 0, false},
		{"control", fyne.KeyModifierControl, true},
		{"super", fyne.KeyModifierSuper, true},
		{"shift", fyne.KeyModifierShift, false},
		{"control+shift", fyne.KeyModifierControl | fyne.KeyModifierShift, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMultiSelectModifier(tt.mod); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
