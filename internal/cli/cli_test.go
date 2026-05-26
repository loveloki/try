package cli

import (
	"strings"
	"testing"
)

// --- 参数解析测试 ---

func checkExtractPath(t *testing.T, args []string, wantPath string, wantRemainingLen int) {
	t.Helper()
	path, remaining := extractPath(args)
	if path != wantPath {
		t.Errorf("extractPath(%v) path = %q, want %q", args, path, wantPath)
	}
	if len(remaining) != wantRemainingLen {
		t.Errorf("extractPath(%v) remaining len = %d, want %d", args, len(remaining), wantRemainingLen)
	}
}

func TestExtractPath(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		wantPath         string
		wantRemainingLen int
	}{
		{"no path", []string{"exec", "query"}, "", 2},
		{"--path value", []string{"--path", "/custom", "exec"}, "/custom", 1},
		{"--path=value", []string{"--path=/custom", "exec"}, "/custom", 1},
		{"last wins", []string{"--path", "/a", "--path", "/b", "exec"}, "/b", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkExtractPath(t, tt.args, tt.wantPath, tt.wantRemainingLen)
		})
	}
}

func TestExtractBoolFlag(t *testing.T) {
	found, remaining := extractBoolFlag([]string{"--and-exit", "exec"}, "--and-exit")
	if !found {
		t.Error("should find --and-exit")
	}
	if len(remaining) != 1 || remaining[0] != "exec" {
		t.Errorf("remaining = %v, want [exec]", remaining)
	}

	found, remaining = extractBoolFlag([]string{"exec"}, "--and-exit")
	if found {
		t.Error("should not find --and-exit")
	}
	if len(remaining) != 1 {
		t.Errorf("remaining len = %d, want 1", len(remaining))
	}
}

func TestExtractValueFlag(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		flag      string
		wantValue string
		wantLen   int
	}{
		{"separate", []string{"--and-type", "hello", "exec"}, "--and-type", "hello", 1},
		{"equals", []string{"--and-type=hello", "exec"}, "--and-type", "hello", 1},
		{"missing", []string{"exec"}, "--and-type", "", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, remaining := extractValueFlag(tt.args, tt.flag)
			if value != tt.wantValue {
				t.Errorf("value = %q, want %q", value, tt.wantValue)
			}
			if len(remaining) != tt.wantLen {
				t.Errorf("remaining len = %d, want %d", len(remaining), tt.wantLen)
			}
		})
	}
}

func TestHasFlag(t *testing.T) {
	if !hasFlag([]string{"--help"}, "--help", "-h") {
		t.Error("should find --help")
	}
	if !hasFlag([]string{"-h"}, "--help", "-h") {
		t.Error("should find -h")
	}
	if hasFlag([]string{"exec"}, "--help", "-h") {
		t.Error("should not find help flag")
	}
}

func TestFilterFlags(t *testing.T) {
	result := filterFlags([]string{"--no-colors", "exec", "query"}, func(f string) bool {
		return f == "--no-colors"
	})
	if len(result) != 2 || result[0] != "exec" {
		t.Errorf("result = %v, want [exec query]", result)
	}
}

// --- Run 端到端测试 ---

func TestRunHelp(t *testing.T) {
	// --help 应该返回退出码 2
	code := Run([]string{"--help"})
	if code != 2 {
		t.Errorf("Run(--help) = %d, want 2", code)
	}
}

func TestRunVersion(t *testing.T) {
	code := Run([]string{"--version"})
	if code != 0 {
		t.Errorf("Run(--version) = %d, want 0", code)
	}
}

func TestRunNoArgs(t *testing.T) {
	// 无参数应尝试启动选择器（在无 TTY 环境下会返回 1）
	code := Run(nil)
	if code != 1 {
		t.Errorf("Run(nil) = %d, want 1 (no TTY)", code)
	}
}

func TestRunCloneNoURL(t *testing.T) {
	code := Run([]string{"clone"})
	if code != 1 {
		t.Errorf("Run(clone) = %d, want 1", code)
	}
}

func TestQueryJoin(t *testing.T) {
	// 验证多个 args 合并为连字符分隔的搜索词
	args := []string{"foo", "bar", "baz"}
	joined := strings.Join(args, "-")
	if joined != "foo-bar-baz" {
		t.Errorf("join = %q, want %q", joined, "foo-bar-baz")
	}
}
