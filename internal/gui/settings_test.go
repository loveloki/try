package gui

import (
	"slices"
	"testing"

	"github.com/loveloki/try/internal/i18n"
)

func TestNormalizeExtension(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  string
		want string
		ok   bool
	}{
		{"uppercase", ".GO", ".go", true},
		{"missing dot auto-filled", "go", ".go", true},
		{"missing dot uppercase", "GO", ".go", true},
		{"missing dot digits", "7z", ".7z", true},
		{"only dot", ".", "", false},
		{"double dot", "..", "", false},
		{"dot dot suffix", "..x", "", false},
		{"with space", ".a b", "", false},
		{"with slash", ".a/b", "", false},
		{"digits", ".7z", ".7z", true},
		{"wildcard", "*", "*", true},
		{"wildcard with spaces", " * ", "*", true},
		{"wildcard prefix", "*go", "", false},
		{"trim spaces", " .Md ", ".md", true},
		{"empty", "", "", false},
		{"spaces only", "   ", "", false},
		{"already normalized", ".png", ".png", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := normalizeExtension(tt.raw)
			if ok != tt.ok || got != tt.want {
				t.Errorf("normalizeExtension(%q) = (%q, %v), want (%q, %v)",
					tt.raw, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestThemeValueMapping(t *testing.T) {
	t.Parallel()
	// index ↔ value 往返一致
	for i, want := range themeValues {
		if got := optionValueAt(themeValues, i); got != want {
			t.Errorf("optionValueAt(themeValues, %d) = %q, want %q", i, got, want)
		}
		if got := valueIndex(themeValues, want); got != i {
			t.Errorf("valueIndex(themeValues, %q) = %d, want %d", want, got, i)
		}
	}
	// 未知值与越界索引回退到末项（auto）
	if got := valueIndex(themeValues, "solarized"); got != len(themeValues)-1 {
		t.Errorf("valueIndex(unknown) = %d, want %d", got, len(themeValues)-1)
	}
	if got := optionValueAt(themeValues, 99); got != "auto" {
		t.Errorf("optionValueAt(oob) = %q, want %q", got, "auto")
	}
}

func TestLocaleValueMapping(t *testing.T) {
	t.Parallel()
	for i, want := range localeValues {
		if got := optionValueAt(localeValues, i); got != want {
			t.Errorf("optionValueAt(localeValues, %d) = %q, want %q", i, got, want)
		}
		if got := valueIndex(localeValues, want); got != i {
			t.Errorf("valueIndex(localeValues, %q) = %d, want %d", want, got, i)
		}
	}
	if got := valueIndex(localeValues, "fr"); got != len(localeValues)-1 {
		t.Errorf("valueIndex(unknown) = %d, want %d", got, len(localeValues)-1)
	}
}

func TestLabelIndex(t *testing.T) {
	t.Parallel()
	labels := []string{"a", "b", "c"}
	if got := labelIndex(labels, "b"); got != 1 {
		t.Errorf("labelIndex = %d, want 1", got)
	}
	if got := labelIndex(labels, "z"); got != -1 {
		t.Errorf("labelIndex(missing) = %d, want -1", got)
	}
}

func TestSettingsSelectLabels(t *testing.T) {
	i18n.Init("zh")
	t.Cleanup(func() { i18n.Init("") })
	msgs := i18n.Get()
	g := &desktopGUI{msgs: msgs}

	themeWant := []string{msgs.GUISettingsThemeDark, msgs.GUISettingsThemeLight, msgs.GUISettingsThemeAuto}
	if got := themeLabels(g.msgs); !slices.Equal(got, themeWant) {
		t.Errorf("themeLabels = %v, want %v", got, themeWant)
	}
	localeWant := []string{"English", "中文", msgs.GUISettingsLangAuto}
	if got := localeLabels(g.msgs); !slices.Equal(got, localeWant) {
		t.Errorf("localeLabels = %v, want %v", got, localeWant)
	}
}

func TestSettingsViewRoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		from string
		want string
	}{
		{"from selector", "selector", "selector"},
		{"from files", "files", "files"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &desktopGUI{view: tt.from}
			g.applyOpenSettings()
			if g.view != "settings" {
				t.Fatalf("after open: view = %q, want settings", g.view)
			}
			if g.settingsReturn != tt.from {
				t.Fatalf("settingsReturn = %q, want %q", g.settingsReturn, tt.from)
			}
			if back := g.applyCloseSettings(); back != tt.want {
				t.Fatalf("close returned %q, want %q", back, tt.want)
			}
			if g.view != tt.want {
				t.Fatalf("after close: view = %q, want %q", g.view, tt.want)
			}
		})
	}
}

func TestSettingsViewRoundTripDefault(t *testing.T) {
	t.Parallel()
	// 未记录返回视图时回退到 selector
	g := &desktopGUI{view: "settings"}
	if back := g.applyCloseSettings(); back != "selector" {
		t.Fatalf("close returned %q, want selector", back)
	}
}

func TestSortedOpenWithExts(t *testing.T) {
	t.Parallel()
	got := sortedOpenWithExts(map[string]string{".md": "a", ".go": "b", ".png": "c"})
	want := []string{".go", ".md", ".png"}
	if !slices.Equal(got, want) {
		t.Errorf("sortedOpenWithExts = %v, want %v", got, want)
	}
	if got := sortedOpenWithExts(nil); len(got) != 0 {
		t.Errorf("sortedOpenWithExts(nil) = %v, want empty", got)
	}
}
