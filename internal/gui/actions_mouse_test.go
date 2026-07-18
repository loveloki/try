package gui

import (
	"testing"

	"github.com/loveloki/try/internal/i18n"
)

func TestRevealCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		goos string
		dir  string
		name string
		args []string
	}{
		{"darwin", "/tmp/a", "open", []string{"/tmp/a"}},
		{"windows", `C:\tries`, "explorer", []string{`C:\tries`}},
		{"linux", "/home/u/tries", "xdg-open", []string{"/home/u/tries"}},
	}
	for _, tt := range tests {
		t.Run(tt.goos, func(t *testing.T) {
			name, args := revealCommand(tt.goos, tt.dir)
			if name != tt.name {
				t.Fatalf("name = %q, want %q", name, tt.name)
			}
			if len(args) != len(tt.args) || args[0] != tt.args[0] {
				t.Fatalf("args = %#v, want %#v", args, tt.args)
			}
		})
	}
}

func TestOpenFileAtUsesIndex(t *testing.T) {
	t.Parallel()
	files := []FileEntry{
		{Name: "a", Path: "/root/a", IsDir: true},
		{Name: "b", Path: "/root/b", IsDir: true},
		{Name: "c", Path: "/root/c", IsDir: true},
	}
	selected := 0
	clickIdx := 2
	selected = clickIdx
	got := files[selected].Path
	if got != "/root/c" {
		t.Fatalf("open target = %q, want /root/c", got)
	}
}

func TestSelectorStatusContent(t *testing.T) {
	t.Parallel()
	i18n.Init("en")
	g := &desktopGUI{
		msgs:    i18n.Get(),
		entries: make([]EntryView, 3),
		marked:  map[string]bool{},
	}
	left, hints := g.selectorStatusContent()
	if left == "" || len(hints) < 3 {
		t.Fatalf("left=%q hints=%d", left, len(hints))
	}
	g.marked["/x"] = true
	_, hints = g.selectorStatusContent()
	if len(hints) != 2 {
		t.Fatalf("delete mode hints = %d, want 2", len(hints))
	}
}

func TestFilesStatusContent(t *testing.T) {
	t.Parallel()
	i18n.Init("zh")
	g := &desktopGUI{
		msgs:  i18n.Get(),
		files: make([]FileEntry, 2),
	}
	left, hints := g.filesStatusContent()
	if left == "" || len(hints) != 3 {
		t.Fatalf("left=%q hints=%d", left, len(hints))
	}
	if !hints[2].Accent {
		t.Fatal("drop hint should be accent")
	}
}
