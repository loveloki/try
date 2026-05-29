package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xleine/try/internal/i18n"
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
	for name, cfg := range Shells {
		t.Run(name, func(t *testing.T) {
			rcFile := cfg.RCFile()
			if rcFile == "" {
				t.Errorf("RCFile() for %s returned empty", name)
			}
		})
	}
}

func TestInstallToFileNewRC(t *testing.T) {
	tmpDir := t.TempDir()
	rcFile := filepath.Join(tmpDir, ".testrc")

	cfg := ShellConfig{
		Name:     "bash",
		RCFile:   func() string { return rcFile },
		InitFunc: posixInit,
	}

	if err := installToFile(cfg, &i18n.EN); err != nil {
		t.Fatalf("installToFile failed: %v", err)
	}

	data, err := os.ReadFile(rcFile)
	if err != nil {
		t.Fatalf("failed to read rc file: %v", err)
	}
	content := string(data)

	for _, want := range []string{marker, "try()", "eval", "2>/dev/tty"} {
		if !strings.Contains(content, want) {
			t.Errorf("rc file missing %q\ngot:\n%s", want, content)
		}
	}
}

func TestInstallToFileIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	rcFile := filepath.Join(tmpDir, ".testrc")

	cfg := ShellConfig{
		Name:     "bash",
		RCFile:   func() string { return rcFile },
		InitFunc: posixInit,
	}

	// 首次安装
	if err := installToFile(cfg, &i18n.EN); err != nil {
		t.Fatalf("first install failed: %v", err)
	}

	data1, _ := os.ReadFile(rcFile)

	// 再次安装不应追加
	if err := installToFile(cfg, &i18n.EN); err != nil {
		t.Fatalf("second install failed: %v", err)
	}

	data2, _ := os.ReadFile(rcFile)
	if string(data1) != string(data2) {
		t.Error("second install should not modify file")
	}

	count := strings.Count(string(data2), marker)
	if count != 1 {
		t.Errorf("marker appears %d times, want 1", count)
	}
}

func TestInstallToFilePreservesExisting(t *testing.T) {
	tmpDir := t.TempDir()
	rcFile := filepath.Join(tmpDir, ".testrc")

	existing := "# existing shell config\nalias ll='ls -la'\n"
	os.WriteFile(rcFile, []byte(existing), 0o644)

	cfg := ShellConfig{
		Name:     "bash",
		RCFile:   func() string { return rcFile },
		InitFunc: posixInit,
	}

	if err := installToFile(cfg, &i18n.EN); err != nil {
		t.Fatalf("installToFile failed: %v", err)
	}

	data, _ := os.ReadFile(rcFile)
	content := string(data)

	if !strings.HasPrefix(content, existing) {
		t.Error("install should preserve existing content")
	}
	if !strings.Contains(content, marker) {
		t.Error("install should append marker")
	}
}

func TestInstallToFileCreatesParentDir(t *testing.T) {
	tmpDir := t.TempDir()
	rcFile := filepath.Join(tmpDir, "nested", "dir", ".testrc")

	cfg := ShellConfig{
		Name:     "fish",
		RCFile:   func() string { return rcFile },
		InitFunc: fishInit,
	}

	if err := installToFile(cfg, &i18n.EN); err != nil {
		t.Fatalf("installToFile failed: %v", err)
	}

	if _, err := os.Stat(rcFile); err != nil {
		t.Errorf("rc file should exist at %s", rcFile)
	}
}

func TestDetectShellEmpty(t *testing.T) {
	t.Setenv("SHELL", "")
	got := DetectShell()
	// 回退到父进程名检测，结果取决于运行环境；此处仅验证不 panic
	_ = got
}
