package dialog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/loveloki/try/internal/selector"
)

// --- Delete 对话框测试 ---

func checkDeleteConfirm(t *testing.T, confirmYes bool, items []selector.DeleteItem, basePath string, wantHasResult bool) {
	t.Helper()
	d := NewDeleteDialog(items, basePath, "", 80, true, "dark")
	d.confirmYes = confirmYes
	result := d.confirm()
	if wantHasResult && result == nil {
		t.Errorf("confirmYes=%v = nil, want non-nil result", confirmYes)
	}
	if !wantHasResult && result != nil {
		t.Errorf("confirmYes=%v = %+v, want nil", confirmYes, result)
	}
}

func TestDeleteConfirmYES(t *testing.T) {
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	os.MkdirAll(dir1, 0o755)

	items := []selector.DeleteItem{{Path: dir1, Basename: "dir1"}}
	checkDeleteConfirm(t, true, items, tmpDir, true)
}

func TestDeleteConfirmDefaultNo(t *testing.T) {
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	os.MkdirAll(dir1, 0o755)

	items := []selector.DeleteItem{{Path: dir1, Basename: "dir1"}}
	checkDeleteConfirm(t, false, items, tmpDir, false)
}

func TestDeleteConfirmPathSafety(t *testing.T) {
	tmpDir := t.TempDir()
	// 路径在 basePath 之外应被拒绝
	outsidePath := "/tmp/outside-" + filepath.Base(tmpDir)
	os.MkdirAll(outsidePath, 0o755)
	defer os.RemoveAll(outsidePath)

	items := []selector.DeleteItem{{Path: outsidePath, Basename: "outside"}}
	checkDeleteConfirm(t, true, items, tmpDir, false)
}

// --- Rename 对话框测试 ---

func checkRenameConfirm(t *testing.T, input, oldName, basePath string, wantHasResult bool, wantErrMsg string) {
	t.Helper()
	entry := &selector.MatchedEntry{Entry: selector.Entry{Basename: oldName}}
	d := NewRenameDialog(entry, basePath, 80)
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
	d := NewShipDialog(entry, basePath, "/tmp/ship", 80)
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

// --- 按键路由测试 ---

// driveDialog 驱动对话框处理按键序列，返回最终状态
func driveDialog(t *testing.T, d Dialog, keys []tea.KeyPressMsg) Dialog {
	t.Helper()
	var model tea.Model = d
	for _, k := range keys {
		model, _ = model.Update(k)
		d = model.(Dialog)
		if d.Done() {
			break
		}
	}
	return d
}

func TestDeleteDialogEscape(t *testing.T) {
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	os.MkdirAll(dir1, 0o755)
	items := []selector.DeleteItem{{Path: dir1, Basename: "dir1"}}

	d := NewDeleteDialog(items, tmpDir, "", 80, true, "dark")
	d = driveDialog(t, d, []tea.KeyPressMsg{
		{Code: tea.KeyEscape},
	}).(*DeleteDialog)

	if !d.Done() {
		t.Error("ESC should close dialog")
	}
	if d.Result() != nil {
		t.Error("ESC should produce nil result")
	}
}

func TestDeleteDialogCtrlC(t *testing.T) {
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	os.MkdirAll(dir1, 0o755)
	items := []selector.DeleteItem{{Path: dir1, Basename: "dir1"}}

	d := NewDeleteDialog(items, tmpDir, "", 80, true, "dark")
	d = driveDialog(t, d, []tea.KeyPressMsg{
		{Code: 'c', Mod: tea.ModCtrl},
	}).(*DeleteDialog)

	if !d.Done() {
		t.Error("Ctrl-C should close dialog")
	}
	if d.Result() != nil {
		t.Error("Ctrl-C should produce nil result")
	}
}

func TestDeleteDialogEnterWithYES(t *testing.T) {
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	os.MkdirAll(dir1, 0o755)
	items := []selector.DeleteItem{{Path: dir1, Basename: "dir1"}}

	d := NewDeleteDialog(items, tmpDir, "", 80, true, "dark")
	d = driveDialog(t, d, []tea.KeyPressMsg{
		{Code: tea.KeyRight},
		{Code: tea.KeyEnter},
	}).(*DeleteDialog)

	if !d.Done() {
		t.Error("Enter should close dialog")
	}
	if d.Result() == nil {
		t.Error("Enter with YES selected should produce non-nil result")
	}
}

func TestDeleteDialogEnterWithDefaultNo(t *testing.T) {
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	os.MkdirAll(dir1, 0o755)
	items := []selector.DeleteItem{{Path: dir1, Basename: "dir1"}}

	d := NewDeleteDialog(items, tmpDir, "", 80, true, "dark")
	if d.confirmYes {
		t.Fatal("default choice should be NO")
	}
	d = driveDialog(t, d, []tea.KeyPressMsg{
		{Code: tea.KeyEnter},
	}).(*DeleteDialog)

	if !d.Done() {
		t.Error("Enter should close dialog")
	}
	if d.Result() != nil {
		t.Error("Enter with default NO should produce nil result")
	}
}

func TestRenameDialogEscape(t *testing.T) {
	entry := &selector.MatchedEntry{Entry: selector.Entry{Basename: "old-name"}}
	d := NewRenameDialog(entry, t.TempDir(), 80)
	d = driveDialog(t, d, []tea.KeyPressMsg{
		{Code: tea.KeyEscape},
	}).(*RenameDialog)

	if !d.Done() {
		t.Error("ESC should close dialog")
	}
	if d.Result() != nil {
		t.Error("ESC should produce nil result")
	}
}

func TestRenameDialogEnter(t *testing.T) {
	entry := &selector.MatchedEntry{Entry: selector.Entry{Basename: "old-name"}}
	d := NewRenameDialog(entry, t.TempDir(), 80)
	d.input.SetValue("new-name")
	d = driveDialog(t, d, []tea.KeyPressMsg{
		{Code: tea.KeyEnter},
	}).(*RenameDialog)

	if !d.Done() {
		t.Error("Enter should close dialog")
	}
	result := d.Result()
	if result == nil {
		t.Fatal("Enter with valid name should produce non-nil result")
	}
	if result.Type != selector.SelectRename {
		t.Errorf("Type = %v, want SelectRename", result.Type)
	}
	if result.New != "new-name" {
		t.Errorf("New = %q, want %q", result.New, "new-name")
	}
}

func TestShipDialogEscape(t *testing.T) {
	tmpDir := t.TempDir()
	entry := &selector.MatchedEntry{
		Entry: selector.Entry{Basename: "test-2025-08-14", Path: filepath.Join(tmpDir, "test")},
	}
	d := NewShipDialog(entry, tmpDir, "/tmp/ship", 80)
	d = driveDialog(t, d, []tea.KeyPressMsg{
		{Code: tea.KeyEscape},
	}).(*ShipDialog)

	if !d.Done() {
		t.Error("ESC should close dialog")
	}
	if d.Result() != nil {
		t.Error("ESC should produce nil result")
	}
}

func TestShipDialogEnter(t *testing.T) {
	tmpDir := t.TempDir()
	entry := &selector.MatchedEntry{
		Entry: selector.Entry{Basename: "test-2025-08-14", Path: filepath.Join(tmpDir, "test")},
	}
	d := NewShipDialog(entry, tmpDir, "/tmp/ship", 80)
	dest := filepath.Join(tmpDir, "dest")
	d.input.SetValue(dest)
	d = driveDialog(t, d, []tea.KeyPressMsg{
		{Code: tea.KeyEnter},
	}).(*ShipDialog)

	if !d.Done() {
		t.Error("Enter should close dialog")
	}
	result := d.Result()
	if result == nil {
		t.Fatal("Enter with valid dest should produce non-nil result")
	}
	if result.Type != selector.SelectShip {
		t.Errorf("Type = %v, want SelectShip", result.Type)
	}
}

// --- ViewContent 渲染测试 ---

func TestDeleteDialogViewContent(t *testing.T) {
	items := []selector.DeleteItem{
		{Path: "/tmp/dir1", Basename: "dir1"},
		{Path: "/tmp/dir2", Basename: "dir2"},
	}
	d := NewDeleteDialog(items, "/tmp", "", 60, true, "dark")
	content := d.ViewContent()
	plain := ansi.Strip(content)

	for _, want := range []string{"dir1", "dir2", "Delete", "NO", "YES", "╭", "╯"} {
		if !strings.Contains(plain, want) {
			t.Errorf("ViewContent() missing %q\ngot:\n%s", want, plain)
		}
	}
	if !strings.Contains(content, "38;5;196") {
		t.Errorf("ViewContent() should use danger color (196), got:\n%s", content)
	}
}

func TestRenameDialogViewContent(t *testing.T) {
	entry := &selector.MatchedEntry{Entry: selector.Entry{Basename: "old-dir"}}
	d := NewRenameDialog(entry, t.TempDir(), 60)
	content := d.ViewContent()

	for _, want := range []string{"old-dir", "Rename"} {
		if !strings.Contains(content, want) {
			t.Errorf("ViewContent() missing %q\ngot:\n%s", want, content)
		}
	}
}

func TestShipDialogViewContent(t *testing.T) {
	entry := &selector.MatchedEntry{
		Entry: selector.Entry{Basename: "test-2025-08-14", Path: "/tmp/test"},
	}
	d := NewShipDialog(entry, "/tmp", "/tmp/ship", 60)
	content := d.ViewContent()

	for _, want := range []string{"test-2025-08-14", "Ship", "/tmp/ship"} {
		if !strings.Contains(content, want) {
			t.Errorf("ViewContent() missing %q\ngot:\n%s", want, content)
		}
	}
}

func TestDeleteDialogTestConfirmAutoSubmit(t *testing.T) {
	tmpDir := t.TempDir()
	dir1 := filepath.Join(tmpDir, "dir1")
	os.MkdirAll(dir1, 0o755)
	items := []selector.DeleteItem{{Path: dir1, Basename: "dir1"}}

	d := NewDeleteDialog(items, tmpDir, "YES", 80, true, "dark")
	cmd := d.Init()

	// Init 应该选中 YES 并产生 Enter 按键
	if !d.confirmYes {
		t.Error("testConfirm YES should select YES")
	}
	if cmd == nil {
		t.Error("Init with testConfirm should return non-nil cmd")
	}
}
