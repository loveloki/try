package gui

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestDirWatcherNotifiesOnCreate(t *testing.T) {
	dir := t.TempDir()
	var hits atomic.Int32
	w := newDirWatcher(func() { hits.Add(1) }, 80*time.Millisecond)
	t.Cleanup(w.Close)

	if err := w.SetPath(dir); err != nil {
		t.Fatal(err)
	}
	// 给 watcher 一点启动时间
	time.Sleep(50 * time.Millisecond)

	target := filepath.Join(dir, "created.txt")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if hits.Load() > 0 {
			return
		}
		time.Sleep(40 * time.Millisecond)
	}
	t.Fatalf("expected watch callback, hits=%d", hits.Load())
}

func TestDirWatcherPauseSkipsEvents(t *testing.T) {
	dir := t.TempDir()
	var hits atomic.Int32
	w := newDirWatcher(func() { hits.Add(1) }, 50*time.Millisecond)
	t.Cleanup(w.Close)
	if err := w.SetPath(dir); err != nil {
		t.Fatal(err)
	}
	time.Sleep(50 * time.Millisecond)

	w.Pause()
	if err := os.WriteFile(filepath.Join(dir, "paused.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	if hits.Load() != 0 {
		t.Fatalf("paused watcher fired: hits=%d", hits.Load())
	}

	w.Resume()
	if err := os.WriteFile(filepath.Join(dir, "resumed.txt"), []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if hits.Load() > 0 {
			return
		}
		time.Sleep(40 * time.Millisecond)
	}
	t.Fatalf("expected callback after resume, hits=%d", hits.Load())
}

func TestIndexOfFilePath(t *testing.T) {
	t.Parallel()
	files := []FileEntry{{Path: "/a"}, {Path: "/b"}}
	if got := indexOfFilePath(files, "/b"); got != 1 {
		t.Fatalf("got %d", got)
	}
	if got := indexOfFilePath(files, "/missing"); got != -1 {
		t.Fatalf("got %d", got)
	}
}
