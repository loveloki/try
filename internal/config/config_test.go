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
	if got.Ship != want.Ship {
		t.Errorf("Ship = %q, want %q", got.Ship, want.Ship)
	}
	if got.Theme != want.Theme {
		t.Errorf("Theme = %q, want %q", got.Theme, want.Theme)
	}
}

func TestParseConfigData(t *testing.T) {
	d := Config{Path: "~/src/tries", Ship: "~/src/ship", Theme: "auto"}
	tests := []struct {
		name    string
		content string
		want    Config
	}{
		{"empty", "", d},
		{"comment only", "# this is a comment\n# another", d},
		{"normal kv", "path = ~/my/tries\nship = ~/my/ship", Config{Path: "~/my/tries", Ship: "~/my/ship", Theme: "auto"}},
		{"extra spaces", "  path  =  /tmp/tries  \n  ship  =  /tmp/ship  ", Config{Path: "/tmp/tries", Ship: "/tmp/ship", Theme: "auto"}},
		{"duplicate key last wins", "path = /a\npath = /b", Config{Path: "/b", Ship: "~/src/ship", Theme: "auto"}},
		{"unknown key ignored", "path = /a\nfoo = bar\nship = /b", Config{Path: "/a", Ship: "/b", Theme: "auto"}},
		{"no equals sign", "invalid line", d},
		{"mixed comments and values", "# comment\npath = /x\n\n# another\nship = /y\n", Config{Path: "/x", Ship: "/y", Theme: "auto"}},
		{"only path set", "path = /custom", Config{Path: "/custom", Ship: "~/src/ship", Theme: "auto"}},
		{"only ship set", "ship = /custom", Config{Path: "~/src/tries", Ship: "/custom", Theme: "auto"}},
		{"theme dark", "theme = dark", Config{Path: "~/src/tries", Ship: "~/src/ship", Theme: "dark"}},
		{"theme light", "theme = light", Config{Path: "~/src/tries", Ship: "~/src/ship", Theme: "light"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkParse(t, tt.content, tt.want)
		})
	}
}

// checkResolve 封装路径解析优先级的测试逻辑
func checkResolve(t *testing.T, cliPath string, cfg Config, envs map[string]string, wantTries, wantShip string) {
	t.Helper()

	// 设置环境变量并在测试结束后恢复
	for k, v := range envs {
		t.Setenv(k, v)
	}

	gotTries, gotShip := ResolvePaths(cliPath, cfg)
	if gotTries != wantTries {
		t.Errorf("triesPath = %q, want %q", gotTries, wantTries)
	}
	if gotShip != wantShip {
		t.Errorf("shipPath = %q, want %q", gotShip, wantShip)
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
		wantShip  string
	}{
		{
			name:      "all defaults",
			cfg:       Config{Path: "~/src/tries", Ship: "~/src/ship"},
			wantTries: home + "/src/tries",
			wantShip:  home + "/src/ship",
		},
		{
			name:      "config overrides default",
			cfg:       Config{Path: "/custom/tries", Ship: "/custom/ship"},
			wantTries: "/custom/tries",
			wantShip:  "/custom/ship",
		},
		{
			name:      "env overrides config",
			cfg:       Config{Path: "/custom/tries", Ship: "/custom/ship"},
			envs:      map[string]string{"TRY_PATH": "/env/tries", "TRY_PROJECTS": "/env/ship"},
			wantTries: "/env/tries",
			wantShip:  "/env/ship",
		},
		{
			name:      "cli overrides env for tries",
			cliPath:   "/cli/tries",
			cfg:       Config{Path: "/custom/tries", Ship: "/custom/ship"},
			envs:      map[string]string{"TRY_PATH": "/env/tries"},
			wantTries: "/cli/tries",
			wantShip:  "/custom/ship",
		},
		{
			name:      "tilde expansion in resolved paths",
			cfg:       Config{Path: "~/my/tries", Ship: "~/my/ship"},
			wantTries: home + "/my/tries",
			wantShip:  home + "/my/ship",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkResolve(t, tt.cliPath, tt.cfg, tt.envs, tt.wantTries, tt.wantShip)
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
