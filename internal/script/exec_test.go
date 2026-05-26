package script

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/user/try/internal/selector"
)

// checkExecute 封装完整的 Execute 测试逻辑
func checkExecute(t *testing.T, result *selector.SelectionResult, wantStdoutContains []string, wantErr bool) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	err := ExecuteTo(&stdout, &stderr, result)
	if wantErr && err == nil {
		t.Error("expected error, got nil")
	}
	if !wantErr && err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	output := stdout.String()
	for _, s := range wantStdoutContains {
		if !strings.Contains(output, s) {
			t.Errorf("stdout missing %q\ngot: %s", s, output)
		}
	}
}

func TestExecuteNil(t *testing.T) {
	err := Execute(nil)
	if err != nil {
		t.Errorf("Execute(nil) = %v, want nil", err)
	}
}

func TestExecMkdir(t *testing.T) {
	tmpDir := t.TempDir()
	newDir := filepath.Join(tmpDir, "new-project-2025-08-17")

	checkExecute(t, &selector.SelectionResult{
		Type: selector.SelectMkdir,
		Path: newDir,
	}, []string{"cd '" + newDir + "'"}, false)

	if !selector.DirExists(newDir) {
		t.Error("directory should have been created")
	}
}

func TestExecCd(t *testing.T) {
	tmpDir := t.TempDir()

	checkExecute(t, &selector.SelectionResult{
		Type: selector.SelectCD,
		Path: tmpDir,
	}, []string{"cd '" + tmpDir + "'"}, false)
}

func TestExecDelete(t *testing.T) {
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir2")
	os.MkdirAll(dir1, 0o755)
	os.MkdirAll(dir2, 0o755)

	checkExecute(t, &selector.SelectionResult{
		Type: selector.SelectDelete,
		Paths: []selector.DeleteItem{
			{Path: dir1, Basename: "dir1"},
			{Path: dir2, Basename: "dir2"},
		},
		BasePath: tmpDir,
	}, []string{"cd '"}, false)

	if selector.DirExists(dir1) {
		t.Error("dir1 should have been deleted")
	}
	if selector.DirExists(dir2) {
		t.Error("dir2 should have been deleted")
	}
}

func TestExecRename(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir := filepath.Join(tmpDir, "old-name")
	os.MkdirAll(oldDir, 0o755)

	checkExecute(t, &selector.SelectionResult{
		Type:     selector.SelectRename,
		Old:      "old-name",
		New:      "new-name",
		BasePath: tmpDir,
	}, []string{"cd '" + filepath.Join(tmpDir, "new-name") + "'"}, false)

	if selector.DirExists(oldDir) {
		t.Error("old directory should not exist after rename")
	}
	if !selector.DirExists(filepath.Join(tmpDir, "new-name")) {
		t.Error("new directory should exist after rename")
	}
}

func TestExecShipNonGit(t *testing.T) {
	tmpDir := t.TempDir()
	source := filepath.Join(tmpDir, "source")
	dest := filepath.Join(tmpDir, "dest")
	os.MkdirAll(source, 0o755)

	checkExecute(t, &selector.SelectionResult{
		Type:     selector.SelectShip,
		Source:   source,
		Dest:     dest,
		Basename: "source",
	}, []string{"cd '" + dest + "'"}, false)

	if selector.DirExists(source) {
		t.Error("source should not exist after ship")
	}
	if !selector.DirExists(dest) {
		t.Error("dest should exist after ship")
	}
}
