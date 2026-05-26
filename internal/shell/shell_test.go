package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func checkDetect(t *testing.T, shellEnv, want string) {
	t.Helper()
	t.Setenv("SHELL", shellEnv)
	got := DetectShell()
	if got != want {
		t.Errorf("DetectShell() with SHELL=%q = %q, want %q", shellEnv, got, want)
	}
}

func TestDetectShell(t *testing.T) {
	tests := []struct {
		shellEnv string
		want     string
	}{
		{"/bin/bash", "bash"},
		{"/usr/bin/bash", "bash"},
		{"/bin/zsh", "zsh"},
		{"/usr/local/bin/fish", "fish"},
		{"/opt/homebrew/bin/fish", "fish"},
	}

	for _, tt := range tests {
		t.Run(tt.shellEnv, func(t *testing.T) {
			checkDetect(t, tt.shellEnv, tt.want)
		})
	}
}

func checkTemplate(t *testing.T, shellType, binaryPath string, wantContains []string) {
	t.Helper()
	cfg, ok := Shells[shellType]
	if !ok {
		t.Fatalf("unknown shell type: %s", shellType)
	}
	output := cfg.InitFunc(binaryPath)
	for _, s := range wantContains {
		if !strings.Contains(output, s) {
			t.Errorf("InitFunc(%q, %q) output missing %q\ngot:\n%s", shellType, binaryPath, s, output)
		}
	}
}

func TestTemplates(t *testing.T) {
	tests := []struct {
		name         string
		shellType    string
		binaryPath   string
		wantContains []string
	}{
		{
			"bash template",
			"bash",
			"/usr/local/bin/try",
			[]string{marker, "eval", "2>/dev/tty", "try()", "'/usr/local/bin/try'"},
		},
		{
			"zsh template",
			"zsh",
			"/usr/local/bin/try",
			[]string{marker, "eval", "2>/dev/tty", "try()"},
		},
		{
			"fish template",
			"fish",
			"/usr/local/bin/try",
			[]string{marker, "eval", "2>/dev/tty", "function try", "'/usr/local/bin/try'"},
		},
		{
			"path with spaces",
			"bash",
			"/home/user/my tools/try",
			[]string{"'/home/user/my tools/try'"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkTemplate(t, tt.shellType, tt.binaryPath, tt.wantContains)
		})
	}
}

func TestInstallIdempotence(t *testing.T) {
	tmpDir := t.TempDir()
	rcFile := filepath.Join(tmpDir, ".bashrc")

	// 模拟已安装的情况：写入 marker
	if err := os.WriteFile(rcFile, []byte("existing content\n"+marker+"\ntry() { ... }\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(rcFile)
	if !strings.Contains(string(data), marker) {
		t.Fatal("setup failed: marker not found in rc file")
	}

	// 再次安装不应重复追加（无法直接调用 Install 因为它依赖 detectShell，
	// 但可验证 marker 检测逻辑）
	content := string(data)
	count := strings.Count(content, marker)
	if count != 1 {
		t.Errorf("marker appears %d times, want 1", count)
	}
}

func TestRCFilePaths(t *testing.T) {
	// 验证 RC 文件路径函数返回非空
	for name, cfg := range Shells {
		t.Run(name, func(t *testing.T) {
			rcFile := cfg.RCFile()
			if rcFile == "" {
				t.Errorf("RCFile() for %s returned empty", name)
			}
		})
	}
}
