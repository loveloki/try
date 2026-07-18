package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

func TestSourceTabVisualState(t *testing.T) {
	tests := []struct {
		name            string
		active, hovered bool
		wantBG, wantFG  fyne.ThemeColorName
		wantBold        bool
	}{
		{"idle", false, false, theme.ColorNameButton, colorNameMuted, false},
		{"hover", false, true, theme.ColorNameHover, theme.ColorNameForeground, false},
		{"active", true, false, theme.ColorNameSelection, colorNameHeader, true},
		{"active hover", true, true, theme.ColorNameSelection, colorNameHeader, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sourceTabVisualState(tt.active, tt.hovered)
			if got.bg != tt.wantBG || got.fg != tt.wantFG || got.bold != tt.wantBold {
				t.Fatalf("got %+v, want bg=%v fg=%v bold=%v", got, tt.wantBG, tt.wantFG, tt.wantBold)
			}
		})
	}
}
