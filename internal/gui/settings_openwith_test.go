package gui

import (
	"slices"
	"testing"
)

func TestNormalizeAppName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  string
		want string
		ok   bool
	}{
		{"plain name", "zed", "zed", true},
		{"with spaces", "  Zed  ", "Zed", true},
		{"bundle name", "Zed.app", "Zed.app", true},
		{"absolute path", "/Applications/Zed.app", "/Applications/Zed.app", true},
		{"empty", "", "", false},
		{"only spaces", "   ", "", false},
		{"dot", ".", "", false},
		{"dotdot", "..", "", false},
		{"traversal", "../evil", "", false},
		{"traversal inside", "/tmp/../evil", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := normalizeAppName(tt.raw)
			if ok != tt.ok || got != tt.want {
				t.Errorf("normalizeAppName(%q) = (%q, %v), want (%q, %v)",
					tt.raw, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestFilterAppOptions(t *testing.T) {
	t.Parallel()
	options := []string{"Zed", "Visual Studio Code", "Vim", "7zz", "zed-cli"}
	tests := []struct {
		name  string
		query string
		want  []string
	}{
		{"empty query", "", options},
		{"spaces only", "  ", options},
		{"case insensitive", "zed", []string{"Zed", "zed-cli"}},
		{"substring", "vim", []string{"Vim"}},
		{"no match", "emacs", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterAppOptions(options, tt.query)
			if !slices.Equal(got, tt.want) {
				t.Errorf("filterAppOptions(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}

func TestLooksLikePath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  string
		want bool
	}{
		{"plain name", "zed", false},
		{"unix path", "/usr/local/bin/zed", true},
		{"windows path", `C:\Apps\zed.exe`, true},
		{"relative", "bin/zed", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := looksLikePath(tt.raw); got != tt.want {
				t.Errorf("looksLikePath(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}
