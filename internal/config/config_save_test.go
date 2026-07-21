package config

import (
	"os"
	"path/filepath"
	"testing"
)

// saveTestInitial 保存类测试共用的初始配置，含 Path/Ships/Theme 等需保留的字段
const saveTestInitial = `{"path":"/custom/tries","ships":["/a","/b"],"theme":"dark","locale":"en"}`

// seedConfigFile 在临时 HOME 下写入指定内容的 config.json
func seedConfigFile(t *testing.T, content string) {
	t.Helper()
	tmpDir := t.TempDir()
	setTestHome(t, tmpDir)
	configDir := filepath.Join(tmpDir, ".config", "try")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// assertPreserved 断言保存后配置中的 Path/Ships/Theme 保持初始值不变
func assertPreserved(t *testing.T, cfg Config) {
	t.Helper()
	if cfg.Path != "/custom/tries" {
		t.Errorf("Path = %q, want preserved %q", cfg.Path, "/custom/tries")
	}
	if len(cfg.Ships) != 2 || cfg.Ships[0] != "/a" || cfg.Ships[1] != "/b" {
		t.Errorf("Ships = %v, want preserved %v", cfg.Ships, []string{"/a", "/b"})
	}
	if cfg.Theme != "dark" {
		t.Errorf("Theme = %q, want preserved %q", cfg.Theme, "dark")
	}
}

// assertOpenWith 断言两个 openWith 映射内容一致（nil 与空 map 视为相等）
func assertOpenWith(t *testing.T, got, want map[string]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("OpenWith = %v, want %v", got, want)
		return
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("OpenWith[%q] = %q, want %q", k, got[k], v)
		}
	}
}

func TestSaveLocale(t *testing.T) {
	tests := []struct {
		name   string
		locale string
	}{
		{"save zh", "zh"},
		{"save en", "en"},
		{"save auto", "auto"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seedConfigFile(t, saveTestInitial)

			if err := SaveLocale(tt.locale); err != nil {
				t.Fatalf("SaveLocale(%q) returned error: %v", tt.locale, err)
			}

			cfg, err := LoadConfig()
			if err != nil {
				t.Fatalf("LoadConfig() after SaveLocale returned error: %v", err)
			}
			if cfg.Locale != tt.locale {
				t.Errorf("Locale = %q, want %q", cfg.Locale, tt.locale)
			}
			assertPreserved(t, cfg)
		})
	}

	t.Run("error when config missing", func(t *testing.T) {
		setTestHome(t, t.TempDir())
		if err := SaveLocale("zh"); err == nil {
			t.Error("SaveLocale() should return error when config file does not exist")
		}
	})
}

func TestSaveOpenWith(t *testing.T) {
	tests := []struct {
		name     string
		openWith map[string]string
	}{
		{"single mapping", map[string]string{".go": "code"}},
		{"multiple mappings", map[string]string{".go": "code", ".md": "typora", ".png": "preview"}},
		{"empty map", map[string]string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seedConfigFile(t, saveTestInitial)

			if err := SaveOpenWith(tt.openWith); err != nil {
				t.Fatalf("SaveOpenWith(%v) returned error: %v", tt.openWith, err)
			}

			cfg, err := LoadConfig()
			if err != nil {
				t.Fatalf("LoadConfig() after SaveOpenWith returned error: %v", err)
			}
			assertOpenWith(t, cfg.OpenWith, tt.openWith)
			assertPreserved(t, cfg)
		})
	}

	t.Run("error when config missing", func(t *testing.T) {
		setTestHome(t, t.TempDir())
		if err := SaveOpenWith(map[string]string{".go": "code"}); err == nil {
			t.Error("SaveOpenWith() should return error when config file does not exist")
		}
	})
}
