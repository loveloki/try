package config

import (
	"os"
	"testing"
)

// checkParseOK 封装成功路径的配置解析测试逻辑
func checkParseOK(t *testing.T, content string, want Config) {
	t.Helper()
	got, err := parseConfigData([]byte(content))
	if err != nil {
		t.Fatalf("parseConfigData(%q) returned error: %v", content, err)
	}
	if got.Path != want.Path {
		t.Errorf("Path = %q, want %q", got.Path, want.Path)
	}
	if len(got.Ships) != len(want.Ships) {
		t.Errorf("Ships = %v, want %v", got.Ships, want.Ships)
	} else {
		for i := range got.Ships {
			if got.Ships[i] != want.Ships[i] {
				t.Errorf("Ships[%d] = %q, want %q", i, got.Ships[i], want.Ships[i])
			}
		}
	}
	if got.Locale != want.Locale {
		t.Errorf("Locale = %q, want %q", got.Locale, want.Locale)
	}
}

func TestParseConfigData(t *testing.T) {
	// 解析失败的用例
	t.Run("empty", func(t *testing.T) {
		_, err := parseConfigData([]byte(""))
		if err == nil {
			t.Error("parseConfigData('') should return error")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		_, err := parseConfigData([]byte("not json"))
		if err == nil {
			t.Error("parseConfigData('not json') should return error")
		}
	})

	// 解析成功的用例
	d := Config{Path: "~/src/tries", Ships: []string{"~/src/ship", "~/src/bug"}, Locale: "auto"}
	successCases := []struct {
		name    string
		content string
		want    Config
	}{
		{"empty object", "{}", d},
		{"full config with ships", `{"path":"~/my/tries","ships":["~/my/ship","~/my/bug"],"locale":"zh"}`,
			Config{Path: "~/my/tries", Ships: []string{"~/my/ship", "~/my/bug"}, Locale: "zh"}},
		{"only path", `{"path":"/custom"}`,
			Config{Path: "/custom", Ships: []string{"~/src/ship", "~/src/bug"}, Locale: "auto"}},
		{"only ships", `{"ships":["/custom/a","/custom/b"]}`,
			Config{Path: "~/src/tries", Ships: []string{"/custom/a", "/custom/b"}, Locale: "auto"}},
		{"only locale", `{"locale":"en"}`,
			Config{Path: "~/src/tries", Ships: []string{"~/src/ship", "~/src/bug"}, Locale: "en"}},
		{"unknown key ignored", `{"path":"/a","foo":"bar","ships":["/b"]}`,
			Config{Path: "/a", Ships: []string{"/b"}, Locale: "auto"}},
	}

	for _, tt := range successCases {
		t.Run(tt.name, func(t *testing.T) {
			checkParseOK(t, tt.content, tt.want)
		})
	}
}

// checkResolve 封装路径解析优先级的测试逻辑
func checkResolve(t *testing.T, cliPath string, cfg Config, envs map[string]string, wantTries string, wantShips []string) {
	t.Helper()

	for k, v := range envs {
		t.Setenv(k, v)
	}

	gotTries, gotShips := ResolvePaths(cliPath, cfg)
	if gotTries != wantTries {
		t.Errorf("triesPath = %q, want %q", gotTries, wantTries)
	}
	if len(gotShips) != len(wantShips) {
		t.Errorf("shipPaths = %v, want %v", gotShips, wantShips)
	} else {
		for i := range gotShips {
			if gotShips[i] != wantShips[i] {
				t.Errorf("shipPaths[%d] = %q, want %q", i, gotShips[i], wantShips[i])
			}
		}
	}
}

func TestResolvePaths(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name      string
		cliPath   string
		cfg       Config
		envs      map[string]string
		wantTries string
		wantShips []string
	}{
		{
			name:      "all defaults",
			cfg:       Config{Path: "~/src/tries", Ships: []string{"~/src/ship", "~/src/bug"}},
			wantTries: home + "/src/tries",
			wantShips: []string{home + "/src/ship", home + "/src/bug"},
		},
		{
			name:      "config overrides default",
			cfg:       Config{Path: "/custom/tries", Ships: []string{"/custom/ship"}},
			wantTries: "/custom/tries",
			wantShips: []string{"/custom/ship"},
		},
		{
			name:      "env overrides config",
			cfg:       Config{Path: "/custom/tries", Ships: []string{"/custom/ship"}},
			envs:      map[string]string{"TRY_PATH": "/env/tries", "TRY_PROJECTS": "/env/ship"},
			wantTries: "/env/tries",
			wantShips: []string{"/env/ship"},
		},
		{
			name:      "cli overrides env for tries",
			cliPath:   "/cli/tries",
			cfg:       Config{Path: "/custom/tries", Ships: []string{"/custom/ship"}},
			envs:      map[string]string{"TRY_PATH": "/env/tries"},
			wantTries: "/cli/tries",
			wantShips: []string{"/custom/ship"},
		},
		{
			name:      "tilde expansion in resolved paths",
			cfg:       Config{Path: "~/my/tries", Ships: []string{"~/my/ship", "~/my/bug"}},
			wantTries: home + "/my/tries",
			wantShips: []string{home + "/my/ship", home + "/my/bug"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkResolve(t, tt.cliPath, tt.cfg, tt.envs, tt.wantTries, tt.wantShips)
		})
	}
}

func TestDetectTheme(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		want string
	}{
		{
			name: "default dark",
			want: "dark",
		},
		{
			name: "COLORFGBG light background",
			envs: map[string]string{"COLORFGBG": "0;15"},
			want: "light",
		},
		{
			name: "COLORFGBG dark background",
			envs: map[string]string{"COLORFGBG": "15;0"},
			want: "dark",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("COLORFGBG", "")
			for k, v := range tt.envs {
				t.Setenv(k, v)
			}
			got := DetectTheme()
			if got != tt.want {
				t.Errorf("DetectTheme() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveLocale(t *testing.T) {
	tests := []struct {
		name      string
		cliLocale string
		cfg       Config
		envs      map[string]string
		want      string
	}{
		{
			name: "default auto resolves to en",
			cfg:  Config{Locale: "auto"},
			want: "en",
		},
		{
			name: "config zh",
			cfg:  Config{Locale: "zh"},
			want: "zh",
		},
		{
			name: "env overrides config",
			cfg:  Config{Locale: "en"},
			envs: map[string]string{"TRY_LOCALE": "zh"},
			want: "zh",
		},
		{
			name:      "cli overrides env",
			cliLocale: "en",
			cfg:       Config{Locale: "zh"},
			envs:      map[string]string{"TRY_LOCALE": "zh"},
			want:      "en",
		},
		{
			name: "LANG zh detected",
			cfg:  Config{Locale: "auto"},
			envs: map[string]string{"LANG": "zh_CN.UTF-8"},
			want: "zh",
		},
		{
			name: "LANG en detected",
			cfg:  Config{Locale: "auto"},
			envs: map[string]string{"LANG": "en_US.UTF-8"},
			want: "en",
		},
		{
			name: "LC_MESSAGES zh overrides LANG en",
			cfg:  Config{Locale: "auto"},
			envs: map[string]string{"LANG": "en_US.UTF-8", "LC_MESSAGES": "zh_CN.UTF-8"},
			want: "zh",
		},
		{
			name: "LC_ALL overrides LC_MESSAGES",
			cfg:  Config{Locale: "auto"},
			envs: map[string]string{"LC_MESSAGES": "zh_CN.UTF-8", "LC_ALL": "en_US.UTF-8"},
			want: "en",
		},
	}

	// detectLocale 依赖的环境变量列表，每个子测试开始前清空以隔离副作用
	localeEnvKeys := []string{"TRY_LOCALE", "LC_ALL", "LC_MESSAGES", "LANG"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, k := range localeEnvKeys {
				t.Setenv(k, "")
			}
			for k, v := range tt.envs {
				t.Setenv(k, v)
			}
			got := ResolveLocale(tt.cliLocale, tt.cfg)
			if got != tt.want {
				t.Errorf("ResolveLocale(%q, %+v) = %q, want %q", tt.cliLocale, tt.cfg, got, tt.want)
			}
		})
	}
}

func TestInitConfigFile(t *testing.T) {
	t.Run("creates default config when file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)

		created, err := InitConfigFile()
		if err != nil {
			t.Fatalf("InitConfigFile() returned error: %v", err)
		}
		if !created {
			t.Error("InitConfigFile() should return true when creating a new config")
		}

		content, err := os.ReadFile(tmpDir + "/.config/try/config.json")
		if err != nil {
			t.Fatalf("config file was not created: %v", err)
		}

		cfg, err := parseConfigData(content)
		if err != nil {
			t.Fatalf("config file content is invalid: %v", err)
		}
		if cfg.Path != "~/src/tries" {
			t.Errorf("Path = %q, want %q", cfg.Path, "~/src/tries")
		}
		if len(cfg.Ships) != 2 || cfg.Ships[0] != "~/src/ship" || cfg.Ships[1] != "~/src/bug" {
			t.Errorf("Ships = %v, want %v", cfg.Ships, []string{"~/src/ship", "~/src/bug"})
		}
		if cfg.Locale != "auto" {
			t.Errorf("Locale = %q, want %q", cfg.Locale, "auto")
		}
	})

	t.Run("does not overwrite existing config", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)

		configDir := tmpDir + "/.config/try"
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			t.Fatal(err)
		}
		customContent := []byte(`{"path":"/custom/path"}` + "\n")
		if err := os.WriteFile(configDir+"/config.json", customContent, 0o644); err != nil {
			t.Fatal(err)
		}

		created, err := InitConfigFile()
		if err != nil {
			t.Fatalf("InitConfigFile() returned error: %v", err)
		}
		if created {
			t.Error("InitConfigFile() should return false when config already exists")
		}

		content, err := os.ReadFile(configDir + "/config.json")
		if err != nil {
			t.Fatal(err)
		}
		if string(content) != string(customContent) {
			t.Errorf("existing config was overwritten\ngot:  %q\nwant: %q", string(content), string(customContent))
		}
	})

	t.Run("error when home dir is invalid", func(t *testing.T) {
		// HOME is set to empty to cause os.UserHomeDir() to still work,
		// so we instead test by making the path unwritable via chmod.
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)

		// Make the directory read-only so MkdirAll fails
		if err := os.Chmod(tmpDir, 0o444); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { os.Chmod(tmpDir, 0o755) })

		if _, err := InitConfigFile(); err == nil {
			t.Error("InitConfigFile() should return error when config dir cannot be created")
		}
	})
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input string
		want  string
	}{
		{"~/foo", home + "/foo"},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
		{"~notahome", "~notahome"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ExpandPath(tt.input)
			if got != tt.want {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
