package selector

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// driveModelWithFactory 与 driveModel 类似，但注入 DialogFactory 以测试对话框路径
func driveModelWithFactory(t *testing.T, cfg Config, factory DialogFactory) SelectorModel {
	t.Helper()

	t.Setenv("TRY_WIDTH", "80")
	t.Setenv("TRY_HEIGHT", "24")

	m := New(cfg)
	m.SetDialogFactory(factory)

	var model tea.Model = m
	model, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	sm := model.(SelectorModel)
	sm.loadAllTries()
	sm.refreshList()
	sm.dialogFactory = factory

	for _, k := range cfg.TestKeys {
		model, _ = sm.Update(KeyToMsg(k))
		sm = model.(SelectorModel)
		if sm.selected != nil {
			break
		}
	}

	return sm
}

func TestSelectorCtrlROpensRenameDialog(t *testing.T) {
	tmpDir := setupTestDirs(t)
	factory := &mockDialogFactory{}

	sm := driveModelWithFactory(t, Config{
		BasePath:      tmpDir,
		TestKeys:      []string{"CTRL-R"},
		ColorsEnabled: false,
	}, factory)

	if !factory.renameCalled {
		t.Error("Ctrl-R should open rename dialog")
	}
	if sm.activeDialog == nil {
		t.Error("activeDialog should be set after Ctrl-R")
	}
}

func TestSelectorCtrlGOpensShipDialog(t *testing.T) {
	tmpDir := setupTestDirs(t)
	factory := &mockDialogFactory{}

	sm := driveModelWithFactory(t, Config{
		BasePath:      tmpDir,
		ShipPaths:     []string{t.TempDir()},
		TestKeys:      []string{"CTRL-G"},
		ColorsEnabled: false,
	}, factory)

	if !factory.shipCalled {
		t.Error("Ctrl-G should open ship dialog")
	}
	if sm.activeDialog == nil {
		t.Error("activeDialog should be set after Ctrl-G")
	}
}

func TestSelectorCtrlROnEmptyListDoesNothing(t *testing.T) {
	tmpDir := t.TempDir()
	factory := &mockDialogFactory{}

	driveModelWithFactory(t, Config{
		BasePath:      tmpDir,
		TestKeys:      []string{"CTRL-R"},
		ColorsEnabled: false,
	}, factory)

	if factory.renameCalled {
		t.Error("Ctrl-R on empty list should not open rename dialog")
	}
}

func TestSelectorCtrlGOnEmptyListDoesNothing(t *testing.T) {
	tmpDir := t.TempDir()
	factory := &mockDialogFactory{}

	driveModelWithFactory(t, Config{
		BasePath:      tmpDir,
		ShipPaths:     []string{t.TempDir()},
		TestKeys:      []string{"CTRL-G"},
		ColorsEnabled: false,
	}, factory)

	if factory.shipCalled {
		t.Error("Ctrl-G on empty list should not open ship dialog")
	}
}

func TestSelectorDeleteConfirmOpensDialog(t *testing.T) {
	tmpDir := setupTestDirs(t)
	factory := &mockDialogFactory{}

	sm := driveModelWithFactory(t, Config{
		BasePath:      tmpDir,
		TestKeys:      []string{"CTRL-D", "ENTER"},
		ColorsEnabled: false,
	}, factory)

	if !factory.deleteCalled {
		t.Error("Ctrl-D + Enter should open delete dialog")
	}
	if sm.activeDialog == nil {
		t.Error("activeDialog should be set after delete confirm")
	}
}

func TestSelectorViewWithDialog(t *testing.T) {
	tmpDir := setupTestDirs(t)
	factory := &mockDialogFactory{}

	t.Setenv("TRY_WIDTH", "80")
	t.Setenv("TRY_HEIGHT", "24")

	m := New(Config{
		BasePath:      tmpDir,
		ColorsEnabled: false,
	})
	m.SetDialogFactory(factory)

	var model tea.Model = m
	model, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	sm := model.(SelectorModel)
	sm.loadAllTries()
	sm.refreshList()
	sm.dialogFactory = factory

	// 正常 View
	normalView := sm.View()
	if !strings.Contains(normalView.Content, "Try") {
		t.Error("normal view should contain title")
	}

	// 打开全屏对话框后，View 仅渲染对话框内容
	model, _ = sm.Update(KeyToMsg("CTRL-R"))
	sm = model.(SelectorModel)
	if sm.activeDialog != nil {
		dialogView := sm.View()
		if !strings.Contains(dialogView.Content, "mock dialog") {
			t.Errorf("dialog view should show mock dialog content, got %q", dialogView.Content)
		}
	}
}
