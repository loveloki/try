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
