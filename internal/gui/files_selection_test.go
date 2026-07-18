package gui

import "testing"

func TestClampFileSelected(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		selected int
		length   int
		want     int
	}{
		{"empty keeps none", 0, 0, -1},
		{"none stays none", -1, 3, -1},
		{"in range", 1, 3, 1},
		{"past end clamps", 5, 3, 2},
		{"below none", -2, 3, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clampFileSelected(tt.selected, tt.length); got != tt.want {
				t.Fatalf("clampFileSelected(%d,%d)=%d, want %d", tt.selected, tt.length, got, tt.want)
			}
		})
	}
}

func TestStepFileSelected(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		selected int
		delta    int
		length   int
		want     int
	}{
		{"empty", -1, 1, 0, -1},
		{"none down to first", -1, 1, 3, 0},
		{"none up to last", -1, -1, 3, 2},
		{"wrap forward", 2, 1, 3, 0},
		{"wrap back", 0, -1, 3, 2},
		{"step mid", 1, 1, 3, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stepFileSelected(tt.selected, tt.delta, tt.length)
			if got != tt.want {
				t.Fatalf("stepFileSelected(%d,%d,%d)=%d, want %d",
					tt.selected, tt.delta, tt.length, got, tt.want)
			}
		})
	}
}

func TestApplyFilesNavClearsSelection(t *testing.T) {
	g := &desktopGUI{
		view:         "selector",
		fileSelected: 2,
		fileMarked:   map[string]bool{"x": true},
	}
	g.applyFilesNav("/tmp/root", "/tmp/root")
	if g.fileSelected != -1 {
		t.Fatalf("fileSelected=%d, want -1", g.fileSelected)
	}
	if len(g.fileMarked) != 0 {
		t.Fatalf("fileMarked not cleared: %#v", g.fileMarked)
	}
}
