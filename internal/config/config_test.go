package config

import (
	"os"
	"testing"
)

// checkParse 封装配置解析的测试逻辑，API 变更时只改这一处
func checkParse(t *testing.T, content string, want Config) {
	t.Helper()
	got := parseConfigData([]byte(content))
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
	if got.Theme != want.Theme {
		t.Errorf("Theme = %q, want %q", got.Theme, want.Theme)
	}
	if got.Locale != want.Locale {
		t.Errorf("Locale = %q, want %q", got.Locale, want.Locale)
	}
}

func TestParseConfigData(t *testing.T) {
	d := Config{Path: "~/src/tries", Ships: []string{"~/src/ship", "~/src/bug"}, Theme: "auto", Locale: "auto"}
	tests := []struct {
		name    string
		content string
		want    Config
	}{
		{"empty", "", d},
		{"invalid json", "not json", d},
		{"empty object", "{}", d},
		{"full config with ships", `{"path":"~/my/tries","ships":["~/my/ship","~/my/bug"],"theme":"dark","locale":"zh"}`,
			Config{Path: "~/my/tries", Ships: []string{"~/my/ship", "~/my/bug"}, Theme: "dark", Locale: "zh"}},
		{"legacy ship field", `{"path":"~/my/tries","ship":"~/my/ship"}`,
			Config{Path: "~/my/tries", Ships: []string{"~/my/ship"}, Theme: "auto", Locale: "auto"}},
		{"only path", `{"path":"/custom"}`,
			Config{Path: "/custom", Ships: []string{"~/src/ship", "~/src/bug"}, Theme: "auto", Locale: "auto"}},
		{"only ships", `{"ships":["/custom/a","/custom/b"]}`,
			Config{Path: "~/src/tries", Ships: []string{"/custom/a", "/custom/b"}, Theme: "auto", Locale: "auto"}},
		{"only theme", `{"theme":"light"}`,
			Config{Path: "~/src/tries", Ships: []string{"~/src/ship", "~/src/bug"}, Theme: "light", Locale: "auto"}},
		{"only locale", `{"locale":"en"}`,
			Config{Path: "~/src/tries", Ships: []string{"~/src/ship", "~/src/bug"}, Theme: "auto", Locale: "en"}},
		{"unknown key ignored", `{"path":"/a","foo":"bar","ships":["/b"]}`,
			Config{Path: "/a", Ships: []string{"/b"}, Theme: "auto", Locale: "auto"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkParse(t, tt.content, tt.want)
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

func TestResolveTheme(t *testing.T) {
	tests := []struct {
		name     string
		cliTheme string
		cfg      Config
		envs     map[string]string
		want     string
	}{
		{
			name: "default auto resolves to dark",
			cfg:  Config{Theme: "auto"},
			want: "dark",
		},
		{
			name: "config dark",
			cfg:  Config{Theme: "dark"},
			want: "dark",
		},
		{
			name: "config light",
			cfg:  Config{Theme: "light"},
			want: "light",
		},
		{
			name: "env overrides config",
			cfg:  Config{Theme: "dark"},
			envs: map[string]string{"TRY_THEME": "light"},
			want: "light",
		},
		{
			name:     "cli overrides env",
			cliTheme: "dark",
			cfg:      Config{Theme: "light"},
			envs:     map[string]string{"TRY_THEME": "light"},
			want:     "dark",
		},
		{
			name: "COLORFGBG light background detected",
			cfg:  Config{Theme: "auto"},
			envs: map[string]string{"COLORFGBG": "15;0"},
			want: "light",
		},
		{
			name: "COLORFGBG dark background detected",
			cfg:  Config{Theme: "auto"},
			envs: map[string]string{"COLORFGBG": "0;15"},
			want: "dark",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envs {
				t.Setenv(k, v)
			}
			got := ResolveTheme(tt.cliTheme, tt.cfg)
			if got != tt.want {
				t.Errorf("ResolveTheme(%q, %+v) = %q, want %q", tt.cliTheme, tt.cfg, got, tt.want)
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
