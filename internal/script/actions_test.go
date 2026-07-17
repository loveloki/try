package script

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/loveloki/try/internal/selector"
)

func TestExecuteSideEffectNoCd(t *testing.T) {
	tmp := t.TempDir()

	tests := []struct {
		name  string
		setup func() *selector.SelectionResult
		check func(t *testing.T)
	}{
		{
			name: "mkdir",
			setup: func() *selector.SelectionResult {
				return &selector.SelectionResult{
					Type: selector.SelectMkdir,
					Path: filepath.Join(tmp, "new-dir-2026-07-17"),
				}
			},
			check: func(t *testing.T) {
				if !selector.DirExists(filepath.Join(tmp, "new-dir-2026-07-17")) {
					t.Error("directory should exist")
				}
			},
		},
		{
			name: "delete",
			setup: func() *selector.SelectionResult {
				d := filepath.Join(tmp, "to-delete")
				os.MkdirAll(d, 0o755)
				return &selector.SelectionResult{
					Type: selector.SelectDelete,
					Paths: []selector.DeleteItem{
						{Path: d, Basename: "to-delete"},
					},
					BasePath: tmp,
				}
			},
			check: func(t *testing.T) {
				if selector.DirExists(filepath.Join(tmp, "to-delete")) {
					t.Error("directory should be deleted")
				}
			},
		},
		{
			name: "rename",
			setup: func() *selector.SelectionResult {
				os.MkdirAll(filepath.Join(tmp, "old-name"), 0o755)
				return &selector.SelectionResult{
					Type:     selector.SelectRename,
					Old:      "old-name",
					New:      "new-name",
					BasePath: tmp,
				}
			},
			check: func(t *testing.T) {
				if selector.DirExists(filepath.Join(tmp, "old-name")) {
					t.Error("old name should not exist")
				}
				if !selector.DirExists(filepath.Join(tmp, "new-name")) {
					t.Error("new name should exist")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout := captureStdout(t, func() {
				if err := ExecuteSideEffect(tt.setup()); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			})
			checkNoCdOutput(t, stdout)
			tt.check(t)
		})
	}
}

func TestExecuteSideEffectNilAndCD(t *testing.T) {
	if err := ExecuteSideEffect(nil); err != nil {
		t.Errorf("nil: %v", err)
	}
	stdout := captureStdout(t, func() {
		err := ExecuteSideEffect(&selector.SelectionResult{
			Type: selector.SelectCD,
			Path: t.TempDir(),
		})
		if err != nil {
			t.Errorf("SelectCD: %v", err)
		}
	})
	checkNoCdOutput(t, stdout)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	fn()
	_ = w.Close()
	os.Stdout = old
	data, err := io.ReadAll(r)
	_ = r.Close()
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func checkNoCdOutput(t *testing.T, stdout string) {
	t.Helper()
	if stdout != "" {
		t.Errorf("stdout should be empty, got %q", stdout)
	}
	if strings.Contains(stdout, "cd ") {
		t.Errorf("stdout must not contain cd script: %q", stdout)
	}
}
