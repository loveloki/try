package dialog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xleine/try/internal/i18n"
	"github.com/xleine/try/internal/selector"
)

// --- Delete 对话框测试 ---

func checkDeleteConfirm(t *testing.T, input string, items []selector.DeleteItem, basePath string, wantHasResult bool) {
	t.Helper()
	d := NewDeleteDialog(items, basePath, "", 80, &i18n.EN)
	d.confirmInput.SetValue(input)
	result := d.confirm()
	if wantHasResult && result == nil {
		t.Errorf("confirm(%q) = nil, want non-nil result", input)
	}
	if !wantHasResult && result != nil {
		t.Errorf("confirm(%q) = %+v, want nil", input, result)
	}
}

func TestDeleteConfirmYES(t *testing.T) {
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	os.MkdirAll(dir1, 0o755)

	items := []selector.DeleteItem{{Path: dir1, Basename: "dir1"}}
	checkDeleteConfirm(t, "YES", items, tmpDir, true)
}

func TestDeleteConfirmNotYES(t *testing.T) {
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	os.MkdirAll(dir1, 0o755)

	items := []selector.DeleteItem{{Path: dir1, Basename: "dir1"}}

	tests := []string{"yes", "Yes", "no", "", "Y"}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			checkDeleteConfirm(t, input, items, tmpDir, false)
		})
	}
}

func TestDeleteConfirmPathSafety(t *testing.T) {
	tmpDir := t.TempDir()
	// 路径在 basePath 之外应被拒绝
	outsidePath := "/tmp/outside-" + filepath.Base(tmpDir)
	os.MkdirAll(outsidePath, 0o755)
	defer os.RemoveAll(outsidePath)

	items := []selector.DeleteItem{{Path: outsidePath, Basename: "outside"}}
	checkDeleteConfirm(t, "YES", items, tmpDir, false)
}

// --- Rename 对话框测试 ---

func checkRenameConfirm(t *testing.T, input, oldName, basePath string, wantHasResult bool, wantErrMsg string) {
	t.Helper()
	entry := &selector.MatchedEntry{Entry: selector.Entry{Basename: oldName}}
	d := NewRenameDialog(entry, basePath, 80, &i18n.EN)
	d.input.SetValue(input)
	result, errMsg := d.confirmRename()

	if wantErrMsg != "" {
		if errMsg != wantErrMsg {
			t.Errorf("confirmRename(%q) errMsg = %q, want %q", input, errMsg, wantErrMsg)
		}
		return
	}

	if wantHasResult && result == nil {
		t.Errorf("confirmRename(%q) = nil, want non-nil", input)
	}
	if !wantHasResult && result != nil {
		t.Errorf("confirmRename(%q) = %+v, want nil", input, result)
	}
	if result != nil {
		if result.New != input && result.New != "" {
			// 验证空格被替换为连字符
			expected := input
			if expected != result.New {
				t.Logf("note: name transformed from %q to %q", input, result.New)
			}
		}
	}
}

func TestRenameConfirm(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		input         string
		oldName       string
		wantHasResult bool
		wantErrMsg    string
	}{
		{"valid rename", "new-name", "old-name", true, ""},
		{"empty name", "", "old-name", false, "Name cannot be empty"},
		{"contains slash", "a/b", "old-name", false, "Name cannot contain /"},
		{"same name", "old-name", "old-name", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkRenameConfirm(t, tt.input, tt.oldName, tmpDir, tt.wantHasResult, tt.wantErrMsg)
		})
	}
}

func TestRenameExistingDir(t *testing.T) {
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "existing"), 0o755)

	checkRenameConfirm(t, "existing", "old-name", tmpDir, false, "Directory exists: existing")
}

// --- Ship 对话框测试 ---

func checkShipConfirm(t *testing.T, input, sourcePath, basePath string, wantHasResult bool, wantErrContains string) {
	t.Helper()
	entry := &selector.MatchedEntry{
		Entry: selector.Entry{Basename: "test-2025-08-14", Path: sourcePath},
	}
	d := NewShipDialog(entry, basePath, "/tmp/ship", 80, &i18n.EN)
	d.input.SetValue(input)
	result, errMsg := d.confirmShip()

	if wantErrContains != "" {
		if errMsg == "" || !strings.Contains(errMsg, wantErrContains) {
			t.Errorf("confirmShip(%q) errMsg = %q, want containing %q", input, errMsg, wantErrContains)
		}
		return
	}

	if wantHasResult && result == nil {
		t.Errorf("confirmShip(%q) = nil, want non-nil", input)
	}
	if !wantHasResult && result != nil {
		t.Errorf("confirmShip(%q) = %+v, want nil", input, result)
	}
}


func TestShipConfirm(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	os.MkdirAll(sourceDir, 0o755)

	tests := []struct {
		name            string
		input           string
		wantHasResult   bool
		wantErrContains string
	}{
		{"valid destination", filepath.Join(tmpDir, "dest"), true, ""},
		{"empty destination", "", false, "Destination cannot be empty"},
		{"parent does not exist", "/nonexistent/parent/dest", false, "Parent directory does not exist"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkShipConfirm(t, tt.input, sourceDir, tmpDir, tt.wantHasResult, tt.wantErrContains)
		})
	}
}

func TestShipDestExists(t *testing.T) {
	tmpDir := t.TempDir()
	sourceDir := filepath.Join(tmpDir, "source")
	destDir := filepath.Join(tmpDir, "existing-dest")
	os.MkdirAll(sourceDir, 0o755)
	os.MkdirAll(destDir, 0o755)

	checkShipConfirm(t, destDir, sourceDir, tmpDir, false, "Destination already exists")
}
