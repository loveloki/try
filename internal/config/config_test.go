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
}

func TestParseConfigData(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    Config
	}{
		{"empty", "", Config{Path: "~/src/tries", Ship: "~/src/ship"}},
		{"comment only", "# this is a comment\n# another", Config{Path: "~/src/tries", Ship: "~/src/ship"}},
		{"normal kv", "path = ~/my/tries\nship = ~/my/ship", Config{Path: "~/my/tries", Ship: "~/my/ship"}},
		{"extra spaces", "  path  =  /tmp/tries  \n  ship  =  /tmp/ship  ", Config{Path: "/tmp/tries", Ship: "/tmp/ship"}},
		{"duplicate key last wins", "path = /a\npath = /b", Config{Path: "/b", Ship: "~/src/ship"}},
		{"unknown key ignored", "path = /a\nfoo = bar\nship = /b", Config{Path: "/a", Ship: "/b"}},
		{"no equals sign", "invalid line", Config{Path: "~/src/tries", Ship: "~/src/ship"}},
		{"mixed comments and values", "# comment\npath = /x\n\n# another\nship = /y\n", Config{Path: "/x", Ship: "/y"}},
		{"only path set", "path = /custom", Config{Path: "/custom", Ship: "~/src/ship"}},
		{"only ship set", "ship = /custom", Config{Path: "~/src/tries", Ship: "/custom"}},
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
