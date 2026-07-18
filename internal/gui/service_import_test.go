package gui

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"github.com/loveloki/try/internal/i18n"
)

func TestCopyDroppedFiles(t *testing.T) {
	i18n.Init("en")
	root := t.TempDir()
	dest := filepath.Join(root, "dest")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	srcFile := filepath.Join(root, "hello.txt")
	if err := os.WriteFile(srcFile, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := newService(root, nil)

	uri, err := storage.ParseURI("file://" + srcFile)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.copyDroppedFiles(dest, []fyne.URI{uri}, nil)
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	if got.Copied != 1 || got.Skipped != 0 {
		t.Fatalf("first copy = %+v, want copied=1 skipped=0", got)
	}
	if _, err := os.Stat(filepath.Join(dest, "hello.txt")); err != nil {
		t.Fatalf("dest missing: %v", err)
	}
	if _, err := os.Stat(srcFile); err != nil {
		t.Fatalf("source removed: %v", err)
	}

	got, err = s.copyDroppedFiles(dest, []fyne.URI{uri}, nil)
	if err != nil {
		t.Fatalf("second copy: %v", err)
	}
	if got.Copied != 0 || got.Skipped != 1 {
		t.Fatalf("rename skip = %+v, want copied=0 skipped=1", got)
	}
}

func TestCopyDroppedDir(t *testing.T) {
	i18n.Init("en")
	root := t.TempDir()
	dest := filepath.Join(root, "dest")
	_ = os.MkdirAll(dest, 0o755)
	srcDir := filepath.Join(root, "fixture")
	_ = os.MkdirAll(filepath.Join(srcDir, "nested"), 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "a.go"), []byte("package a"), 0o644)
	_ = os.WriteFile(filepath.Join(srcDir, "nested", "b.txt"), []byte("b"), 0o644)
	_ = os.WriteFile(filepath.Join(srcDir, ".hidden"), []byte("x"), 0o644)

	s := newService(root, nil)
	uri, err := storage.ParseURI("file://" + srcDir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.copyDroppedFiles(dest, []fyne.URI{uri}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Copied != 1 {
		t.Fatalf("copied=%d, want 1", got.Copied)
	}
	if _, err := os.Stat(filepath.Join(dest, "fixture", "a.go")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dest, "fixture", "nested", "b.txt")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dest, "fixture", ".hidden")); !os.IsNotExist(err) {
		t.Fatal("dotfile should not be copied")
	}
}

func TestCopyDroppedSkipsSymlink(t *testing.T) {
	i18n.Init("en")
	root := t.TempDir()
	dest := filepath.Join(root, "dest")
	_ = os.MkdirAll(dest, 0o755)
	outside := filepath.Join(root, "outside")
	_ = os.MkdirAll(outside, 0o755)
	_ = os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("x"), 0o644)
	link := filepath.Join(root, "linkdir")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	s := newService(root, nil)
	uri, err := storage.ParseURI("file://" + link)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.copyDroppedFiles(dest, []fyne.URI{uri}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Skipped != 1 || got.Copied != 0 {
		t.Fatalf("got %+v, want skipped=1 copied=0", got)
	}
	if _, err := os.Stat(filepath.Join(dest, "linkdir")); !os.IsNotExist(err) {
		t.Fatal("symlink dir should not be copied")
	}
}

func TestCopyDirRollbackOnFailure(t *testing.T) {
	i18n.Init("en")
	root := t.TempDir()
	dest := filepath.Join(root, "dest")
	_ = os.MkdirAll(dest, 0o755)
	srcDir := filepath.Join(root, "fixture")
	sub := filepath.Join(srcDir, "nested")
	_ = os.MkdirAll(sub, 0o755)
	_ = os.WriteFile(filepath.Join(srcDir, "ok.txt"), []byte("ok"), 0o644)
	_ = os.WriteFile(filepath.Join(sub, "x.txt"), []byte("x"), 0o644)
	if err := os.Chmod(sub, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(sub, 0o755) })
	if _, err := os.ReadDir(sub); err == nil {
		t.Skip("platform allows reading mode 000 directories")
	}

	s := newService(root, nil)
	uri, err := storage.ParseURI("file://" + srcDir)
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.copyDroppedFiles(dest, []fyne.URI{uri}, nil)
	if err == nil {
		t.Fatal("expected copy error")
	}
	if _, err := os.Stat(filepath.Join(dest, "fixture")); !os.IsNotExist(err) {
		t.Fatal("incomplete dir should be rolled back")
	}
}

func TestCopyDroppedFilesProgress(t *testing.T) {
	i18n.Init("en")
	root := t.TempDir()
	dest := filepath.Join(root, "dest")
	_ = os.MkdirAll(dest, 0o755)
	var uris []fyne.URI
	for _, name := range []string{"a.txt", "b.txt"} {
		path := filepath.Join(root, name)
		if err := os.WriteFile(path, []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
		uri, err := storage.ParseURI("file://" + path)
		if err != nil {
			t.Fatal(err)
		}
		uris = append(uris, uri)
	}
	s := newService(root, nil)
	var lastDone, lastTotal int
	got, err := s.copyDroppedFiles(dest, uris, func(done, total int, _ string) {
		lastDone, lastTotal = done, total
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Copied != 2 {
		t.Fatalf("copied=%d, want 2", got.Copied)
	}
	if lastDone != 2 || lastTotal != 2 {
		t.Fatalf("progress done/total=%d/%d, want 2/2", lastDone, lastTotal)
	}
}

func TestSortFileEntries(t *testing.T) {
	files := []FileEntry{
		{Name: "Beta.txt", IsDir: false},
		{Name: "dirB", IsDir: true},
		{Name: "alpha.txt", IsDir: false},
	}
	sortFileEntries(files)
	want := []string{"dirB", "alpha.txt", "Beta.txt"}
	for i, name := range want {
		if files[i].Name != name {
			t.Fatalf("index %d = %q, want %q", i, files[i].Name, name)
		}
	}
}
