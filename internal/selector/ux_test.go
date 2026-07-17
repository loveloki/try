package selector

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

// initTestModel 创建一个已加载数据、未发送 ESC 的测试模型。
func initTestModel(t *testing.T, cfg Config) SelectorModel {
	t.Helper()
	t.Setenv("TRY_WIDTH", "80")
	t.Setenv("TRY_HEIGHT", "24")

	m := New(cfg)
	var model tea.Model = m
	model, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	sm := model.(SelectorModel)
	sm.loadAllTries()
	sm.refreshList()
	return sm
}

func TestSelectorSpaceTogglesMark(t *testing.T) {
	tmpDir := setupTestDirs(t)
	sm := initTestModel(t, Config{BasePath: tmpDir, ColorsEnabled: false})

	model, _ := sm.Update(KeyToMsg("SPACE"))
	sm = model.(SelectorModel)

	if len(sm.markedForDeletion) != 1 {
		t.Errorf("Space should mark 1 item, got %d", len(sm.markedForDeletion))
	}
	if !sm.deleteMode {
		t.Error("Space should enable delete mode")
	}
}

func TestSelectorDeleteKeyTogglesMark(t *testing.T) {
	tmpDir := setupTestDirs(t)
	sm := initTestModel(t, Config{BasePath: tmpDir, ColorsEnabled: false})

	model, _ := sm.Update(KeyToMsg("DELETE"))
	sm = model.(SelectorModel)

	if len(sm.markedForDeletion) != 1 {
		t.Errorf("Delete key should mark 1 item, got %d", len(sm.markedForDeletion))
	}
}

func TestSelectorCtrlAMarksAll(t *testing.T) {
	tmpDir := setupTestDirs(t)
	sm := initTestModel(t, Config{BasePath: tmpDir, ColorsEnabled: false})

	model, _ := sm.Update(KeyToMsg("CTRL-A"))
	sm = model.(SelectorModel)

	if len(sm.markedForDeletion) != len(sm.cachedResults) {
		t.Errorf("Ctrl-A should mark all %d items, got %d", len(sm.cachedResults), len(sm.markedForDeletion))
	}
	if !sm.deleteMode {
		t.Error("Ctrl-A should enable delete mode")
	}
}

func TestSelectorShiftTabCyclesTabsReverse(t *testing.T) {
	tmpDir := setupTestDirs(t)
	sm := initTestModel(t, Config{
		BasePath:  tmpDir,
		ShipPaths: []string{t.TempDir()},
	})

	model, _ := sm.Update(KeyToMsg("SHIFT-TAB"))
	sm = model.(SelectorModel)

	lastIdx := len(sm.sourceOptions) - 1
	if sm.sourceFilter != sm.sourceOptions[lastIdx] {
		t.Errorf("Shift-Tab should cycle to last option %q, got %q", sm.sourceOptions[lastIdx], sm.sourceFilter)
	}
}

func TestSelectorEscClearsSearch(t *testing.T) {
	tmpDir := setupTestDirs(t)
	sm := initTestModel(t, Config{
		BasePath:     tmpDir,
		InitialInput: "nonexistent-query",
	})

	model, _ := sm.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	sm = model.(SelectorModel)

	if sm.textInput.Value() != "" {
		t.Errorf("Esc should clear search, got %q", sm.textInput.Value())
	}
}

func TestSelectorEscExitsDeleteMode(t *testing.T) {
	tmpDir := setupTestDirs(t)
	sm := initTestModel(t, Config{BasePath: tmpDir})

	model, _ := sm.Update(KeyToMsg("SPACE"))
	sm = model.(SelectorModel)
	model, _ = sm.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	sm = model.(SelectorModel)

	if sm.deleteMode {
		t.Error("Esc should exit delete mode")
	}
	if len(sm.markedForDeletion) != 0 {
		t.Error("Esc should clear marks")
	}
}

func TestSelectorCtrlDTogglesDeleteMode(t *testing.T) {
	tmpDir := setupTestDirs(t)
	sm := initTestModel(t, Config{BasePath: tmpDir})

	model, _ := sm.Update(KeyToMsg("CTRL-D"))
	sm = model.(SelectorModel)
	if !sm.deleteMode {
		t.Error("Ctrl-D should enable delete mode")
	}

	model, _ = sm.Update(KeyToMsg("CTRL-D"))
	sm = model.(SelectorModel)
	if sm.deleteMode {
		t.Error("Ctrl-D again should exit delete mode")
	}
	if len(sm.markedForDeletion) != 0 {
		t.Error("Ctrl-D again should clear marks")
	}
}

func TestSelectorCreateInputResizesList(t *testing.T) {
	tmpDir := setupTestDirs(t)
	sm := initTestModel(t, Config{BasePath: tmpDir})

	for _, ch := range "hello" {
		model, _ := sm.Update(tea.KeyPressMsg{Code: ch, Text: string(ch)})
		sm = model.(SelectorModel)
	}

	if !strings.Contains(sm.View().Content, msgs().CreateNew) {
		t.Error("View should show create-new preview when input is non-empty")
	}
}

func TestSelectorNavigationWrapsWithCtrlKeys(t *testing.T) {
	tmpDir := setupTestDirs(t)
	sm := initTestModel(t, Config{BasePath: tmpDir, ColorsEnabled: false})

	items := sm.list.Items()
	if len(items) < 2 {
		t.Fatalf("need at least 2 items, got %d", len(items))
	}

	model, _ := sm.Update(KeyToMsg("CTRL-P"))
	sm = model.(SelectorModel)
	if sm.list.Index() != len(items)-1 {
		t.Errorf("Ctrl-P at top should wrap to last index %d, got %d", len(items)-1, sm.list.Index())
	}

	model, _ = sm.Update(KeyToMsg("CTRL-N"))
	sm = model.(SelectorModel)
	if sm.list.Index() != 0 {
		t.Errorf("Ctrl-N at bottom should wrap to first index, got %d", sm.list.Index())
	}
}

func TestSelectorNavigationWrapsWithArrowKeys(t *testing.T) {
	tmpDir := setupTestDirs(t)
	sm := initTestModel(t, Config{BasePath: tmpDir, ColorsEnabled: false})

	items := sm.list.Items()
	if len(items) < 2 {
		t.Fatalf("need at least 2 items, got %d", len(items))
	}

	model, _ := sm.Update(KeyToMsg("UP"))
	sm = model.(SelectorModel)
	if sm.list.Index() != len(items)-1 {
		t.Errorf("Up at top should wrap to last index %d, got %d", len(items)-1, sm.list.Index())
	}

	model, _ = sm.Update(KeyToMsg("DOWN"))
	sm = model.(SelectorModel)
	if sm.list.Index() != 0 {
		t.Errorf("Down at bottom should wrap to first index, got %d", sm.list.Index())
	}
}

func TestMarkedRowHasDeleteIcon(t *testing.T) {
	tmpDir := setupTestDirs(t)
	sm := initTestModel(t, Config{BasePath: tmpDir, ColorsEnabled: true})

	model, _ := sm.Update(KeyToMsg("SPACE"))
	sm = model.(SelectorModel)

	items := sm.list.Items()
	if len(items) == 0 {
		t.Fatal("no items")
	}
	var buf strings.Builder
	sm.delegate.Render(&buf, sm.list, 0, items[0])
	rendered := buf.String()

	if !strings.Contains(rendered, iconMarked) {
		t.Errorf("marked row should contain %q icon, got:\n%s", iconMarked, rendered)
	}
}

func TestMarkedRowUsesDangerSurface(t *testing.T) {
	tmpDir := setupTestDirs(t)
	sm := initTestModel(t, Config{BasePath: tmpDir, ColorsEnabled: true})

	style := sm.delegate.rowStyle(false, true)
	if style.GetBackground() != sm.styles.DangerSurface.GetBackground() {
		t.Errorf("marked row background = %v, want %v", style.GetBackground(), sm.styles.DangerSurface.GetBackground())
	}
}

func TestSelectedRowHasArrow(t *testing.T) {
	tmpDir := setupTestDirs(t)
	sm := initTestModel(t, Config{BasePath: tmpDir, ColorsEnabled: true})

	items := sm.list.Items()
	if len(items) == 0 {
		t.Fatal("no items")
	}
	var buf strings.Builder
	sm.delegate.Render(&buf, sm.list, 0, items[0])
	rendered := buf.String()

	if !strings.Contains(rendered, iconSelected) {
		t.Errorf("selected row should contain %q arrow, got:\n%s", iconSelected, rendered)
	}
}

func TestScoreBarRendered(t *testing.T) {
	tmpDir := setupTestDirs(t)
	sm := initTestModel(t, Config{BasePath: tmpDir, ColorsEnabled: true})

	items := sm.list.Items()
	if len(items) == 0 {
		t.Fatal("no items")
	}
	var buf strings.Builder
	sm.delegate.Render(&buf, sm.list, 0, items[0])
	rendered := buf.String()

	if !strings.Contains(rendered, "█") && !strings.Contains(rendered, "░") {
		t.Errorf("row should contain score bar blocks, got:\n%s", rendered)
	}
}

func TestSourceTabsRenderCounts(t *testing.T) {
	tmpDir := setupTestDirs(t)
	sm := initTestModel(t, Config{BasePath: tmpDir})

	header := renderHeader(&sm)
	if !strings.Contains(header, msgs().FilterAll) {
		t.Error("header should contain source filter tabs")
	}
}

func TestEmptyStateRendered(t *testing.T) {
	tmpDir := t.TempDir()
	sm := initTestModel(t, Config{BasePath: tmpDir})

	view := sm.View().Content
	if !strings.Contains(view, msgs().EmptyStateHint) {
		t.Errorf("empty state should be rendered, got:\n%s", view)
	}
}

func TestNoMatchesStateRendered(t *testing.T) {
	tmpDir := setupTestDirs(t)
	sm := initTestModel(t, Config{BasePath: tmpDir})

	for _, ch := range "xyz" {
		model, _ := sm.Update(tea.KeyPressMsg{Code: ch, Text: string(ch)})
		sm = model.(SelectorModel)
	}

	view := sm.View().Content
	want := strings.ReplaceAll(msgs().NoMatchesHint, "%s", "xyz")
	if !strings.Contains(view, want) {
		t.Errorf("no-matches state should be rendered, got:\n%s", view)
	}
}

func TestSelectedRowForegroundHasContrast(t *testing.T) {
	t.Setenv("COLORFGBG", "")
	st := NewStyles(true)
	if st.SurfaceSelected.GetForeground() != lipgloss.Color("#e6edf3") {
		t.Errorf("dark selected foreground = %v, want #e6edf3", st.SurfaceSelected.GetForeground())
	}
	if st.SurfaceSelected.GetForeground() == lipgloss.Color("#0d1117") {
		t.Error("selected foreground must not equal dark page background")
	}

	t.Setenv("COLORFGBG", "0;15")
	stLight := NewStyles(true)
	if stLight.SurfaceSelected.GetForeground() != lipgloss.Color("#1f2328") {
		t.Errorf("light selected foreground = %v, want #1f2328", stLight.SurfaceSelected.GetForeground())
	}
}
