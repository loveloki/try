package gui

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTouchDirUpdatesMtime(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "try-demo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(dir, old, old); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}

	s := newService(root, nil)
	s.touchDir(dir)

	after, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !after.ModTime().After(before.ModTime()) {
		t.Fatalf("mtime not updated: before=%v after=%v", before.ModTime(), after.ModTime())
	}
}
