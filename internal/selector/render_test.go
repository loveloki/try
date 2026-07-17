package selector

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
)

// TestSelectedRowBold 验证选中行被正确加粗
func TestSelectedRowBold(t *testing.T) {
	tmpDir := setupTestDirs(t)

	t.Setenv("TRY_WIDTH", "80")
	t.Setenv("TRY_HEIGHT", "24")
	t.Setenv("COLORTERM", "truecolor")

	m := New(Config{
		BasePath:      tmpDir,
		ColorsEnabled: true,
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

	// 验证选中行是否包含 Bold 属性的转义序列 \x1b[1m 或者是其他形式
	if !strings.Contains(rendered, "\x1b[1m") && !strings.Contains(rendered, "\x1b[1;") {
		t.Errorf("rendered output should contain Bold style (\\x1b[1m), got: %q", rendered)
	}
}

// TestThemeAutoDetectsFromEnv 验证主题通过 COLORFGBG 自动检测
func TestThemeAutoDetectsFromEnv(t *testing.T) {
	t.Setenv("TRY_WIDTH", "80")
	t.Setenv("TRY_HEIGHT", "24")
	t.Setenv("COLORTERM", "truecolor")

	renderWithEnv := func(colorfgbg string) string {
		t.Setenv("COLORFGBG", colorfgbg)
		tmpDir := setupTestDirs(t)
		m := New(Config{
			BasePath:      tmpDir,
			ColorsEnabled: true,
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

	dark := renderWithEnv("15;0")
	light := renderWithEnv("0;15")

	if dark == light {
		t.Error("dark and light env should produce different ANSI output")
	}

	// 验证 NewStyles 中的主题色属性
	t.Setenv("COLORFGBG", "")
	stDark := NewStyles(true)
	if stDark.Danger.GetForeground() != lipgloss.Color("#ff3b30") {
		t.Errorf("dark theme danger foreground = %v, want #ff3b30", stDark.Danger.GetForeground())
	}
	if stDark.SurfaceSelected.GetForeground() != lipgloss.Color("#e6edf3") {
		t.Errorf("dark theme selected foreground = %v, want #e6edf3", stDark.SurfaceSelected.GetForeground())
	}

	t.Setenv("COLORFGBG", "0;15")
	stLight := NewStyles(true)
	if stLight.Danger.GetForeground() != lipgloss.Color("#d70000") {
		t.Errorf("light theme danger foreground = %v, want #d70000", stLight.Danger.GetForeground())
	}
	if stLight.SurfaceSelected.GetForeground() != lipgloss.Color("#1f2328") {
		t.Errorf("light theme selected foreground = %v, want #1f2328", stLight.SurfaceSelected.GetForeground())
	}
}
