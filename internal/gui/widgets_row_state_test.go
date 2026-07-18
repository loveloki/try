package gui

import "testing"

func TestRowVisualState(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                       string
		marked, selected, hovered  bool
		want                       rowVisualKind
	}{
		{"none", false, false, false, rowVisualNone},
		{"hover", false, false, true, rowVisualHover},
		{"selected", false, true, false, rowVisualSelected},
		{"selected beats hover", false, true, true, rowVisualSelected},
		{"marked beats all", true, true, true, rowVisualMarked},
		{"marked beats hover", true, false, true, rowVisualMarked},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rowVisualState(tt.marked, tt.selected, tt.hovered)
			if got != tt.want {
				t.Fatalf("rowVisualState() = %v, want %v", got, tt.want)
			}
		})
	}
}
