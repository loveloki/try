package gui

import "testing"

func TestResolveSelectorSelection(t *testing.T) {
	t.Parallel()
	entries := []EntryView{
		{Path: "/a"},
		{Path: "/b"},
		{Path: "/c"},
	}
	tests := []struct {
		name         string
		path         string
		selected     int
		wantIdx      int
		wantPath     string
	}{
		{"keep path", "/b", 0, 1, "/b"},
		{"cleared path resets to index", "", 0, 0, "/a"},
		{"cleared path with stale index clamps", "", 9, 0, "/a"},
		{"missing path falls back", "/missing", 2, 2, "/c"},
		{"empty list", "", 0, 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := entries
			if tt.name == "empty list" {
				list = nil
			}
			gotIdx, gotPath := resolveSelectorSelection(list, tt.path, tt.selected)
			if gotIdx != tt.wantIdx || gotPath != tt.wantPath {
				t.Fatalf("got (%d,%q), want (%d,%q)", gotIdx, gotPath, tt.wantIdx, tt.wantPath)
			}
		})
	}
}

func TestCycleSource(t *testing.T) {
	t.Parallel()
	sources := []string{"", "tries", "ship-a", "ship-b"}

	tests := []struct {
		name    string
		current string
		delta   int
		want    string
	}{
		{"forward from all", "", 1, "tries"},
		{"backward from all", "", -1, "ship-b"},
		{"wrap forward end", "ship-b", 1, ""},
		{"wrap backward start", "", -1, "ship-b"},
		{"shift tab from tries", "tries", -1, ""},
		{"tab from tries", "tries", 1, "ship-a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cycleSource(sources, tt.current, tt.delta)
			if got != tt.want {
				t.Fatalf("cycleSource(%q, %d) = %q, want %q", tt.current, tt.delta, got, tt.want)
			}
		})
	}
}

func TestCycleSourceEmpty(t *testing.T) {
	t.Parallel()
	if got := cycleSource(nil, "tries", 1); got != "tries" {
		t.Fatalf("got %q want tries", got)
	}
}

func TestDecideSelectorOpen(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		markedCount int
		selected    int
		entryCount  int
		want        selectorOpenAction
	}{
		{"enter files", 0, 0, 3, selectorOpenFiles},
		{"delete when marked", 2, 1, 3, selectorOpenDelete},
		{"no selection", 0, -1, 3, selectorOpenNone},
		{"out of range", 0, 5, 3, selectorOpenNone},
		{"empty list", 0, 0, 0, selectorOpenNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decideSelectorOpen(tt.markedCount, tt.selected, tt.entryCount)
			if got != tt.want {
				t.Fatalf("decideSelectorOpen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSearchEntryAcceptsTab(t *testing.T) {
	t.Parallel()
	e := newSearchEntry(nil, nil, nil, nil)
	if !e.AcceptsTab() {
		t.Fatal("searchEntry must AcceptsTab so Tab reaches TypedKey")
	}
}

func TestNavListAcceptsTab(t *testing.T) {
	t.Parallel()
	l := newNavList(func() int { return 0 }, nil, nil, nil)
	if !l.AcceptsTab() {
		t.Fatal("navList must AcceptsTab so Tab reaches TypedKey")
	}
}
