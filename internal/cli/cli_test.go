package cli

import (
	"os"
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
	// 用临时空目录隔离，避免依赖用户实际的 tries 目录
	t.Setenv("TRY_PATH", t.TempDir())
	t.Setenv("TRY_PROJECTS", t.TempDir())

	// 将 stdin 替换为关闭的 pipe（非 TTY），防止 bubbletea 进入交互模式卡住
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	w.Close()
	oldStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = oldStdin; r.Close() })

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
	args := []string{"foo", "bar", "baz"}
	joined := strings.Join(args, "-")
	if joined != "foo-bar-baz" {
		t.Errorf("join = %q, want %q", joined, "foo-bar-baz")
	}
}

// --- exec 分派测试 ---

func TestCmdExecCloneNoURL(t *testing.T) {
	opts := runOptions{triesPath: t.TempDir()}
	code := cmdExec(opts, []string{"clone"})
	if code != 1 {
		t.Errorf("cmdExec clone (no url) = %d, want 1", code)
	}
}

func TestCmdExecWorktreeNoDir(t *testing.T) {
	opts := runOptions{triesPath: t.TempDir()}
	code := cmdExec(opts, []string{"worktree"})
	if code != 1 {
		t.Errorf("cmdExec worktree (no dir) = %d, want 1", code)
	}
}

func TestCmdWorktreeNoArgs(t *testing.T) {
	opts := runOptions{triesPath: t.TempDir()}
	code := cmdWorktree(opts, nil)
	if code != 1 {
		t.Errorf("cmdWorktree(nil) = %d, want 1", code)
	}
}

// --- dot 处理测试 ---

func TestHandleDotNoName(t *testing.T) {
	opts := runOptions{triesPath: t.TempDir()}
	code := handleDot(opts, []string{"."})
	if code != 1 {
		t.Errorf("handleDot(\".\") = %d, want 1 (missing name)", code)
	}
}

func TestHandleDotMkdir(t *testing.T) {
	tmpDir := t.TempDir()
	opts := runOptions{triesPath: tmpDir}
	// 没有 .git 文件，应走 mkdir 分支
	code := handleDot(opts, []string{".", "my-test"})
	if code != 0 {
		t.Errorf("handleDot mkdir = %d, want 0", code)
	}
}

// --- worktreePath 测试 ---

func TestWorktreePath(t *testing.T) {
	tmpDir := t.TempDir()
	tests := []struct {
		name       string
		repoDir    string
		customName string
		wantBase   string // 结果应包含的基础名
	}{
		{"auto name from dir", "/some/repo", "", "repo"},
		{"custom name", "/some/repo", "my-branch", "my-branch"},
		{"custom name with spaces", "/some/repo", "my branch", "my-branch"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := worktreePath(tmpDir, tt.repoDir, tt.customName)
			if !strings.Contains(got, tt.wantBase) {
				t.Errorf("worktreePath() = %q, want containing %q", got, tt.wantBase)
			}
			if !strings.HasPrefix(got, tmpDir) {
				t.Errorf("worktreePath() = %q, should be under %q", got, tmpDir)
			}
		})
	}
}

// --- parseGlobalFlags 测试 ---

func TestParseGlobalFlagsNoColor(t *testing.T) {
	opts, remaining := parseGlobalFlags([]string{"--no-colors", "exec", "query"})
	if opts.colorsEnabled {
		t.Error("--no-colors should disable colors")
	}
	if len(remaining) != 2 || remaining[0] != "exec" {
		t.Errorf("remaining = %v, want [exec query]", remaining)
	}
}

func TestParseGlobalFlagsNO_COLOR(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	opts, _ := parseGlobalFlags([]string{"exec"})
	if opts.colorsEnabled {
		t.Error("NO_COLOR env should disable colors")
	}
}

func TestParseGlobalFlagsThemeLocale(t *testing.T) {
	opts, remaining := parseGlobalFlags([]string{"--theme", "light", "--locale", "zh", "exec"})
	if opts.theme != "light" {
		t.Errorf("theme = %q, want %q", opts.theme, "light")
	}
	if opts.locale != "zh" {
		t.Errorf("locale = %q, want %q", opts.locale, "zh")
	}
	if len(remaining) != 1 || remaining[0] != "exec" {
		t.Errorf("remaining = %v, want [exec]", remaining)
	}
}

func TestParseGlobalFlagsAndExit(t *testing.T) {
	opts, _ := parseGlobalFlags([]string{"--and-exit", "exec"})
	if !opts.andExit {
		t.Error("--and-exit should be true")
	}
}

func TestParseGlobalFlagsAndType(t *testing.T) {
	opts, _ := parseGlobalFlags([]string{"--and-type", "hello", "exec"})
	if opts.andType != "hello" {
		t.Errorf("andType = %q, want %q", opts.andType, "hello")
	}
}
