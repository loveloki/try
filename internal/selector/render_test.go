package selector

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

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
