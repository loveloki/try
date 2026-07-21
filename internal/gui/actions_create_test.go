package gui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/loveloki/try/internal/i18n"
)

func TestCreateEntryThenApplyFilesNav(t *testing.T) {
	i18n.Init("en")
	root := t.TempDir()
	s := newService(root, nil)
	g := &desktopGUI{
		service:    s,
		msgs:       i18n.Get(),
		view:       "selector",
		query:      "demo",
		fileMarked: map[string]bool{},
	}

	path, err := g.service.createEntry("demo")
	if err != nil {
		t.Fatalf("createEntry: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("created dir missing: %v", err)
	}
	if filepath.Dir(path) != root {
		t.Fatalf("path parent = %q, want %q", filepath.Dir(path), root)
	}

	g.query = ""
	g.applyFilesNav(path, path)

	if g.view != "files" {
		t.Fatalf("view = %q, want files", g.view)
	}
	if g.filesRoot != path || g.filesPath != path {
		t.Fatalf("filesRoot/Path = %q/%q, want %q", g.filesRoot, g.filesPath, path)
	}
	if g.selectedPath != path {
		t.Fatalf("selectedPath = %q, want %q", g.selectedPath, path)
	}
	if g.query != "" {
		t.Fatalf("query = %q, want empty", g.query)
	}
	if g.fileSelected != 0 {
		t.Fatalf("fileSelected = %d, want 0", g.fileSelected)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(info.ModTime()) > time.Minute {
		t.Fatalf("mtime not recently touched: %v", info.ModTime())
	}
}

func TestCreateEntryEmptyName(t *testing.T) {
	i18n.Init("en")
	s := newService(t.TempDir(), nil)
	if _, err := s.createEntry("  "); err == nil {
		t.Fatal("expected error for empty name")
	}
}
