package selector_test

import (
	"os"
	"strings"
	"testing"

	"github.com/loveloki/try/internal/dialog"
	"github.com/loveloki/try/internal/selector"
)

// realDialogFactory 与 CLI 层相同，使用真实 dialog 实现。
type realDialogFactory struct{}

func (realDialogFactory) NewDeleteDialog(
	items []selector.DeleteItem, basePath, testConfirm string, width int, colorsEnabled bool, theme string,
) selector.DialogInstance {
	return dialog.NewDeleteDialog(items, basePath, testConfirm, width, colorsEnabled, theme)
}

func (realDialogFactory) NewRenameDialog(entry *selector.MatchedEntry, basePath string, width int) selector.DialogInstance {
	return dialog.NewRenameDialog(entry, basePath, width)
}

func (realDialogFactory) NewShipDialog(entry *selector.MatchedEntry, basePath, shipPath string, width int) selector.DialogInstance {
	return dialog.NewShipDialog(entry, basePath, shipPath, width)
}

func TestSelectorDeleteDialogOverlaysMainUI(t *testing.T) {
	tmpDir := selector.SetupTestDirsForTest(t)

	sm := selector.DriveWithDialogFactory(t, selector.Config{
		BasePath:      tmpDir,
		TestKeys:      []string{"CTRL-D", "ENTER"},
		ColorsEnabled: true,
		Theme:         "dark",
	}, realDialogFactory{})

	view := sm.View().Content
	if !strings.Contains(view, "Try") {
		t.Error("overlay view should still show main selector title")
	}
	if !strings.Contains(view, "Delete") {
		t.Error("overlay view should show delete dialog title")
	}
	if !strings.Contains(view, "╭") {
		t.Error("overlay view should render modal border from real delete dialog")
	}

	_ = os.WriteFile("delete_dialog_view.ansi", []byte(view), 0644)
}
