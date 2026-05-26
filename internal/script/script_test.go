package script

import (
	"bytes"
	"strings"
	"testing"
)

func checkQuote(t *testing.T, input, want string) {
	t.Helper()
	got := Quote(input)
	if got != want {
		t.Errorf("Quote(%q) = %q, want %q", input, got, want)
	}
}

func TestQuote(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple path", "/home/user/dir", "'/home/user/dir'"},
		{"with spaces", "/home/user/my dir", "'/home/user/my dir'"},
		{"with single quote", "/home/user/it's", "'/home/user/it'\"'\"'s'"},
		{"empty", "", "''"},
		{"special chars", "/tmp/$HOME & foo", "'/tmp/$HOME & foo'"},
		{"multiple quotes", "a'b'c", "'a'\"'\"'b'\"'\"'c'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkQuote(t, tt.input, tt.want)
		})
	}
}

func checkEmitCd(t *testing.T, path string, wantContains []string) {
	t.Helper()
	var buf bytes.Buffer
	EmitCdTo(&buf, path)
	output := buf.String()
	for _, s := range wantContains {
		if !strings.Contains(output, s) {
			t.Errorf("EmitCdTo(%q) output missing %q\ngot: %s", path, s, output)
		}
	}
}

func TestEmitCd(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantContains []string
	}{
		{
			"simple path",
			"/home/user/project",
			[]string{"cd '/home/user/project'", ScriptWarning},
		},
		{
			"path with spaces",
			"/home/user/my project",
			[]string{"cd '/home/user/my project'"},
		},
		{
			"path with quotes",
			"/home/user/it's",
			[]string{`cd '/home/user/it'"'"'s'`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkEmitCd(t, tt.path, tt.wantContains)
		})
	}
}
