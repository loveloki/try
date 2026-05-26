package selector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

// setupTestDirs 创建测试目录并设置不同 mtime
func setupTestDirs(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	dirs := []struct {
		name string
		age  time.Duration
	}{
		{"alpha-2025-11-01", 30 * 24 * time.Hour},
		{"beta-2025-11-15", 14 * 24 * time.Hour},
		{"gamma-2025-11-20", 7 * 24 * time.Hour},
		{"no-date-suffix", 1 * time.Hour},
	}

	for _, d := range dirs {
		path := filepath.Join(tmpDir, d.name)
		os.MkdirAll(path, 0o755)
		mtime := time.Now().Add(-d.age)
		os.Chtimes(path, mtime, mtime)
	}

	return tmpDir
}

// driveModel 手动驱动 Model 的 Init + testKeyMsg 循环，无需 TTY。
// 返回最终的 SelectorModel。
func driveModel(t *testing.T, cfg Config) SelectorModel {
	t.Helper()

	t.Setenv("TRY_WIDTH", "80")
	t.Setenv("TRY_HEIGHT", "24")

	m := New(cfg)

	// 模拟 Init: 执行 WindowSizeMsg 初始化尺寸
	var model tea.Model = m
	model, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// 加载目录并刷新列表
	sm := model.(SelectorModel)
	sm.loadAllTries()
	sm.refreshList()

	// 注入 testKeys 逐个处理
	for _, k := range cfg.TestKeys {
		model, _ = sm.Update(KeyToMsg(k))
		sm = model.(SelectorModel)
		if sm.selected != nil {
			break
		}
	}

	// 如果 testKeys 耗尽且未选择，模拟自动 ESC
	if sm.selected == nil && len(cfg.TestKeys) > 0 {
		model, _ = sm.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
		sm = model.(SelectorModel)
	}

	return sm
}

func TestSelectorSelectFirst(t *testing.T) {
	tmpDir := setupTestDirs(t)

	sm := driveModel(t, Config{
		BasePath:      tmpDir,
		TestKeys:      []string{"ENTER"},
		ColorsEnabled: false,
	})

	result := sm.Selected()
	if result == nil {
		t.Fatal("expected a selection result")
	}
	if result.Type != SelectCD {
		t.Errorf("Type = %v, want SelectCD", result.Type)
	}
	// 第一项应该是评分最高的（gamma 有日期后缀加成 + 较新 mtime）
	if !strings.Contains(result.Path, "gamma") {
		t.Errorf("expected gamma (highest score due to date suffix + recent mtime), got %q", result.Path)
	}
}

func TestSelectorEscape(t *testing.T) {
	tmpDir := setupTestDirs(t)

	sm := driveModel(t, Config{
		BasePath:      tmpDir,
		TestKeys:      []string{"ESC"},
		ColorsEnabled: false,
	})

	if sm.Selected() != nil {
		t.Error("ESC should produce nil selection")
	}
}

func TestSelectorNavigation(t *testing.T) {
	tmpDir := setupTestDirs(t)

	sm := driveModel(t, Config{
		BasePath:      tmpDir,
		TestKeys:      []string{"DOWN", "ENTER"},
		ColorsEnabled: false,
	})

	result := sm.Selected()
	if result == nil {
		t.Fatal("expected a selection result")
	}
	if result.Type != SelectCD {
		t.Errorf("Type = %v, want SelectCD", result.Type)
	}
	// DOWN 后不应该还是第一项
	if strings.HasSuffix(result.Path, "no-date-suffix") {
		t.Error("DOWN+ENTER should not select the first item")
	}
}

func TestSelectorCreateNew(t *testing.T) {
	tmpDir := setupTestDirs(t)

	sm := driveModel(t, Config{
		BasePath:      tmpDir,
		InitialInput:  "my-new-project",
		TestKeys:      []string{"CTRL-T"},
		ColorsEnabled: false,
	})

	result := sm.Selected()
	if result == nil {
		t.Fatal("expected a selection result")
	}
	if result.Type != SelectMkdir {
		t.Errorf("Type = %v, want SelectMkdir", result.Type)
	}
	if !strings.Contains(result.Path, "my-new-project") {
		t.Errorf("path should contain 'my-new-project', got %q", result.Path)
	}
}

func TestSelectorSearch(t *testing.T) {
	tmpDir := setupTestDirs(t)

	sm := driveModel(t, Config{
		BasePath:      tmpDir,
		InitialInput:  "alpha",
		TestKeys:      []string{"ENTER"},
		ColorsEnabled: false,
	})

	result := sm.Selected()
	if result == nil {
		t.Fatal("expected a selection result")
	}
	if !strings.Contains(result.Path, "alpha") {
		t.Errorf("search for 'alpha' should select alpha dir, got %q", result.Path)
	}
}

func TestSelectorDeleteMode(t *testing.T) {
	tmpDir := setupTestDirs(t)

	// Ctrl-D 标记然后 ESC 取消删除模式
	sm := driveModel(t, Config{
		BasePath:      tmpDir,
		TestKeys:      []string{"CTRL-D", "ESC"},
		ColorsEnabled: false,
	})

	if sm.deleteMode {
		t.Error("ESC should exit delete mode")
	}
	if sm.Selected() != nil {
		t.Error("expected nil selection after ESC in delete mode")
	}
}

func TestSelectorEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	sm := driveModel(t, Config{
		BasePath:      tmpDir,
		TestKeys:      []string{"ESC"},
		ColorsEnabled: false,
	})

	if sm.Selected() != nil {
		t.Error("empty dir + ESC should produce nil selection")
	}
}

func TestSelectorCreateInEmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	sm := driveModel(t, Config{
		BasePath:     tmpDir,
		InitialInput: "new-project",
		TestKeys:     []string{"ENTER"},
		ColorsEnabled: false,
	})

	result := sm.Selected()
	if result == nil {
		t.Fatal("expected a selection result")
	}
	if result.Type != SelectMkdir {
		t.Errorf("Type = %v, want SelectMkdir", result.Type)
	}
}

// TestSelectorTypeAndCreate 测试真实打字流程：
// 不用 InitialInput，而是通过 TestKeys 模拟逐字输入，验证 textInput 正确接收焦点和按键。
func TestSelectorTypeAndCreate(t *testing.T) {
	tmpDir := t.TempDir()

	sm := driveModel(t, Config{
		BasePath:      tmpDir,
		TestKeys:      []string{"h", "i", "CTRL-T"},
		ColorsEnabled: false,
	})

	result := sm.Selected()
	if result == nil {
		t.Fatal("expected a selection result: textInput may not be focused")
	}
	if result.Type != SelectMkdir {
		t.Errorf("Type = %v, want SelectMkdir", result.Type)
	}
	if !strings.Contains(result.Path, "hi") {
		t.Errorf("path should contain typed text 'hi', got %q", result.Path)
	}
}

// TestSelectorTypeAndSearch 测试打字后列表过滤：
// 通过 TestKeys 输入搜索词，验证列表被过滤后选择正确的条目。
func TestSelectorTypeAndSearch(t *testing.T) {
	tmpDir := setupTestDirs(t)

	sm := driveModel(t, Config{
		BasePath:      tmpDir,
		TestKeys:      []string{"a", "l", "p", "ENTER"},
		ColorsEnabled: false,
	})

	result := sm.Selected()
	if result == nil {
		t.Fatal("expected a selection result after typing search query")
	}
	if !strings.Contains(result.Path, "alpha") {
		t.Errorf("typing 'alp' + ENTER should select alpha dir, got %q", result.Path)
	}
}

// TestSelectorCtrlPCtrlN 测试 Ctrl-P/N 导航（而非 UP/DOWN）。
func TestSelectorCtrlPCtrlN(t *testing.T) {
	tmpDir := setupTestDirs(t)

	// 先按 CTRL-N（下移），再 ENTER，应选择第二项
	sm := driveModel(t, Config{
		BasePath:      tmpDir,
		TestKeys:      []string{"CTRL-N", "ENTER"},
		ColorsEnabled: false,
	})

	result := sm.Selected()
	if result == nil {
		t.Fatal("expected a selection result")
	}
	// 不应该选中第一项（gamma）
	if strings.Contains(result.Path, "gamma") {
		t.Error("CTRL-N+ENTER should not select the first item (gamma)")
	}
}

// TestSelectorCtrlTEmptyInput 验证输入框为空时按 Ctrl-T 不创建目录。
func TestSelectorCtrlTEmptyInput(t *testing.T) {
	tmpDir := setupTestDirs(t)

	// Ctrl-T 在输入为空时不应产生 selection
	m := New(Config{
		BasePath:      tmpDir,
		ColorsEnabled: false,
	})

	t.Setenv("TRY_WIDTH", "80")
	t.Setenv("TRY_HEIGHT", "24")

	var model tea.Model = m
	model, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	sm := model.(SelectorModel)
	sm.loadAllTries()
	sm.refreshList()

	// 不输入任何文字，直接按 Ctrl-T
	model, _ = sm.Update(KeyToMsg("CTRL-T"))
	sm = model.(SelectorModel)

	if sm.Selected() != nil {
		t.Error("Ctrl-T with empty input should not produce a selection")
	}
	if sm.deleteStatus == "" {
		t.Error("Ctrl-T with empty input should show a hint message")
	}
}

// TestSelectorBackspace 测试退格键清除输入。
func TestSelectorBackspace(t *testing.T) {
	tmpDir := setupTestDirs(t)

	sm := driveModel(t, Config{
		BasePath:      tmpDir,
		TestKeys:      []string{"x", "y", "z", "BACKSPACE", "BACKSPACE", "BACKSPACE", "ENTER"},
		ColorsEnabled: false,
	})

	result := sm.Selected()
	if result == nil {
		t.Fatal("expected a selection result after backspace clears input")
	}
	// 退格全部清除后相当于无搜索词，应选中评分最高的 gamma
	if !strings.Contains(result.Path, "gamma") {
		t.Errorf("after clearing input, should select gamma (highest score), got %q", result.Path)
	}
}

// TestSelectedRowFullBackground 验证选中行的背景色覆盖整行，
// 而非仅箭头部分。通过检查渲染输出中背景色转义序列出现次数判断。
func TestSelectedRowFullBackground(t *testing.T) {
	tmpDir := setupTestDirs(t)

	t.Setenv("TRY_WIDTH", "80")
	t.Setenv("TRY_HEIGHT", "24")
	t.Setenv("COLORTERM", "truecolor")

	m := New(Config{
		BasePath:      tmpDir,
		ColorsEnabled: true,
		Theme:         "dark",
	})

	var model tea.Model = m
	model, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	sm := model.(SelectorModel)
	sm.loadAllTries()
	sm.refreshList()

	// 渲染选中行（index 0 = 选中项）
	items := sm.list.Items()
	if len(items) == 0 {
		t.Fatal("no items in list")
	}

	var buf strings.Builder
	sm.delegate.Render(&buf, sm.list, 0, items[0])
	rendered := buf.String()

	// 背景色 237（dark 主题 selectedBg）的 ANSI 序列应该出现多次，
	// 因为每个样式段都需要独立设置背景。如果只出现一次，说明仅对整行做了
	// 外层包裹（被内部 reset 打断）。
	bgSeq := "48;5;237"
	count := strings.Count(rendered, bgSeq)
	if count < 2 {
		t.Errorf("selectedBg ANSI sequence %q appears %d time(s), want >= 2 (full row bg).\nRendered: %q",
			bgSeq, count, rendered)
	}
}

// TestThemeAffectsRendering 验证不同主题产生不同的颜色输出
func TestThemeAffectsRendering(t *testing.T) {
	tmpDir := setupTestDirs(t)

	t.Setenv("TRY_WIDTH", "80")
	t.Setenv("TRY_HEIGHT", "24")
	t.Setenv("COLORTERM", "truecolor")

	renderWithTheme := func(theme string) string {
		m := New(Config{
			BasePath:      tmpDir,
			ColorsEnabled: true,
			Theme:         theme,
		})
		var model tea.Model = m
		model, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		sm := model.(SelectorModel)
		sm.loadAllTries()
		sm.refreshList()

		items := sm.list.Items()
		if len(items) == 0 {
			t.Fatal("no items")
		}
		var buf strings.Builder
		sm.delegate.Render(&buf, sm.list, 0, items[0])
		return buf.String()
	}

	dark := renderWithTheme("dark")
	light := renderWithTheme("light")

	if dark == light {
		t.Error("dark and light themes should produce different ANSI output")
	}

	// dark 主题使用 237 背景，light 主题使用 254 背景
	if !strings.Contains(dark, "48;5;237") {
		t.Error("dark theme should use bg color 237")
	}
	if !strings.Contains(light, "48;5;254") {
		t.Error("light theme should use bg color 254")
	}
}

func TestParseTestKeys(t *testing.T) {
	tests := []struct {
		name string
		spec string
		want []string
	}{
		{"token mode", "UP,DOWN,ENTER", []string{"UP", "DOWN", "ENTER"}},
		{"with type", "CTRL-D,TYPE=hello,ENTER", []string{"CTRL-D", "h", "e", "l", "l", "o", "ENTER"}},
		{"raw escape", "\x1b[A\x1b[B\r", []string{"UP", "DOWN", "ENTER"}},
		{"empty", "", nil},
		{"single", "ENTER", []string{"ENTER"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTestKeys(tt.spec)
			if len(got) != len(tt.want) {
				t.Fatalf("ParseTestKeys(%q) = %v, want %v", tt.spec, got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("ParseTestKeys(%q)[%d] = %q, want %q", tt.spec, i, got[i], tt.want[i])
				}
			}
		})
	}
}
