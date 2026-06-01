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

	// 验证选中行是否包含 Bold 属性的转义序列 \x1b[1m 或者是其他形式
	if !strings.Contains(rendered, "\x1b[1m") && !strings.Contains(rendered, "\x1b[1;") {
		t.Errorf("rendered output should contain Bold style (\\x1b[1m), got: %q", rendered)
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

	// 验证 newStyles 中的主题色属性
	stDark := newStyles(true, "dark")
	stLight := newStyles(true, "light")

	if stDark.danger.GetForeground() != lipgloss.Color("196") {
		t.Errorf("dark theme danger foreground = %v, want 196", stDark.danger.GetForeground())
	}
	if stLight.danger.GetForeground() != lipgloss.Color("160") {
		t.Errorf("light theme danger foreground = %v, want 160", stLight.danger.GetForeground())
	}
}
