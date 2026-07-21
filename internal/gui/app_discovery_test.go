package gui

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestOpenWithCommand(t *testing.T) {
	t.Parallel()
	tests := []struct {
		goos string
		app  string
		path string
		name string
		args []string
	}{
		{"darwin", "Visual Studio Code.app", "/tmp/f.go", "open", []string{"-a", "Visual Studio Code.app", "/tmp/f.go"}},
		{"windows", "notepad", `C:\f.txt`, "cmd", []string{"/c", "start", "", "notepad", `C:\f.txt`}},
		{"linux", "vim", "/tmp/f.go", "vim", []string{"/tmp/f.go"}},
	}
	for _, tt := range tests {
		t.Run(tt.goos, func(t *testing.T) {
			name, args := openWithCommand(tt.goos, tt.app, tt.path)
			if name != tt.name {
				t.Fatalf("name = %q, want %q", name, tt.name)
			}
			if len(args) != len(tt.args) {
				t.Fatalf("args len = %d, want %d", len(args), len(tt.args))
			}
			for i := range args {
				if args[i] != tt.args[i] {
					t.Fatalf("args[%d] = %q, want %q", i, args[i], tt.args[i])
				}
			}
		})
	}
}

func TestBuiltinAppsNotEmpty(t *testing.T) {
	t.Parallel()
	apps := builtinApps()
	if len(apps) == 0 {
		t.Fatal("builtinApps() returned empty")
	}
	for _, a := range apps {
		if a.Name == "" || a.Path == "" {
			t.Fatalf("empty name or path in app: %+v", a)
		}
	}
}

func TestFilterAppsByConfigNil(t *testing.T) {
	t.Parallel()
	apps := []availableApp{
		{Name: "VS Code", Path: "code", Available: true},
		{Name: "Vim", Path: "vim", Available: false},
	}
	result := filterAppsByConfig(apps, ".go", nil)
	if len(result) != 2 {
		t.Fatalf("expected 2 apps, got %d", len(result))
	}
}

func TestFilterAppsByConfigMatch(t *testing.T) {
	t.Parallel()
	apps := []availableApp{
		{Name: "VS Code", Path: "code", Available: true},
		{Name: "Vim", Path: "vim", Available: true},
	}
	cfg := map[string]string{".go": "code"}
	result := filterAppsByConfig(apps, ".go", cfg)
	if len(result) != 1 {
		t.Fatalf("expected 1 app, got %d", len(result))
	}
	if result[0].Name != "VS Code" {
		t.Fatalf("expected VS Code, got %q", result[0].Name)
	}
	if !result[0].Available {
		t.Fatal("expected VS Code to be marked available via config")
	}
}

func TestFilterAppsByConfigNoExtMatch(t *testing.T) {
	t.Parallel()
	apps := []availableApp{
		{Name: "VS Code", Path: "code", Available: true},
	}
	cfg := map[string]string{".py": "code"}
	result := filterAppsByConfig(apps, ".go", cfg)
	if len(result) != 1 {
		t.Fatalf("expected all apps when ext not in config, got %d", len(result))
	}
}

func TestFilterAppsByConfigCustomAbsPath(t *testing.T) {
	t.Parallel()
	// 创建一个真实存在的可执行文件作为自定义应用
	dir := t.TempDir()
	appPath := filepath.Join(dir, "myeditor")
	if err := os.WriteFile(appPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	apps := []availableApp{
		{Name: "VS Code", Path: "code", Available: true},
	}
	cfg := map[string]string{".go": appPath}
	result := filterAppsByConfig(apps, ".go", cfg)
	if len(result) != 1 {
		t.Fatalf("expected 1 custom app, got %d", len(result))
	}
	if result[0].Path != appPath || !result[0].Available {
		t.Fatalf("unexpected custom app: %+v", result[0])
	}
}

func TestFilterAppsByConfigCustomNotFound(t *testing.T) {
	t.Parallel()
	apps := []availableApp{
		{Name: "VS Code", Path: "code", Available: true},
	}
	missing := filepath.Join(t.TempDir(), "no_such_app_xyz")
	cfg := map[string]string{".go": missing}
	result := filterAppsByConfig(apps, ".go", cfg)
	if len(result) != 0 {
		t.Fatalf("expected 0 apps for unresolvable custom app, got %d", len(result))
	}
}

func TestFindCustomAppEmpty(t *testing.T) {
	t.Parallel()
	if got := findCustomApp("  "); got != nil {
		t.Fatalf("expected nil for blank name, got %v", got)
	}
}

func TestBundleCandidates(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want []string
	}{
		{"Zed", []string{"Zed", "Zed.app", "zed"}},
		{"Zed.app", []string{"Zed.app"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bundleCandidates(tt.name)
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("bundleCandidates(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestExecCandidates(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want []string
	}{
		{"Zed", []string{"Zed", "zed"}},
		{"vim", []string{"vim"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := execCandidates(tt.name)
			if strings.Join(got, ",") != strings.Join(tt.want, ",") {
				t.Fatalf("execCandidates(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestPathExecutables(t *testing.T) {
	// 使用 t.Setenv 修改 PATH，不能 Parallel
	dir := t.TempDir()
	bin := filepath.Join(dir, "mytool")
	if err := os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	// 不可执行文件与目录不应入选
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	names := pathExecutables(50)
	if runtime.GOOS == "windows" {
		if len(names) != 0 {
			t.Fatalf("expected 0 executables on windows (no .exe), got %v", names)
		}
		return
	}
	if len(names) != 1 || names[0] != "mytool" {
		t.Fatalf("expected [mytool], got %v", names)
	}
}

func TestPathExecutablesLimit(t *testing.T) {
	// 使用 t.Setenv 修改 PATH，不能 Parallel
	if runtime.GOOS == "windows" {
		t.Skip("windows 可执行名依赖扩展名，另行覆盖")
	}
	dir := t.TempDir()
	for _, name := range []string{"a", "b", "c"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir)
	names := pathExecutables(2)
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Fatalf("expected [a b], got %v", names)
	}
}

func TestFilterAppsByConfigWildcard(t *testing.T) {
	t.Parallel()
	apps := []availableApp{
		{Name: "VS Code", Path: "code", Available: true},
		{Name: "Vim", Path: "vim", Available: false},
	}
	tests := []struct {
		name string
		ext  string
		cfg  map[string]string
		want string // 期望唯一的应用名；空串表示返回全部内置
	}{
		{"wildcard applies to any ext", ".go", map[string]string{"*": "code"}, "VS Code"},
		{"wildcard applies to empty ext", "", map[string]string{"*": "code"}, "VS Code"},
		{"exact match wins over wildcard", ".go", map[string]string{".go": "vim", "*": "code"}, "Vim"},
		{"no match falls back to all", ".md", map[string]string{".go": "vim"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterAppsByConfig(apps, tt.ext, tt.cfg)
			if tt.want == "" {
				if len(result) != len(apps) {
					t.Fatalf("expected all apps, got %v", result)
				}
				return
			}
			if len(result) != 1 || result[0].Name != tt.want {
				t.Fatalf("expected only %q, got %v", tt.want, result)
			}
		})
	}
}

func TestFilterAppsByConfigWildcardCustom(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	appPath := filepath.Join(dir, "myeditor")
	if err := os.WriteFile(appPath, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	apps := []availableApp{{Name: "VS Code", Path: "code", Available: true}}
	cfg := map[string]string{"*": appPath}
	result := filterAppsByConfig(apps, ".md", cfg)
	if len(result) != 1 || result[0].Path != appPath {
		t.Fatalf("expected wildcard custom app, got %v", result)
	}
}

func TestApplicationNamesStripped(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "darwin" {
		t.Skip("applicationNames 仅扫描 macOS 应用目录")
	}
	names := applicationNames()
	seen := make(map[string]bool, len(names))
	for _, n := range names {
		if strings.HasSuffix(n, ".app") {
			t.Fatalf("name %q should have .app stripped", n)
		}
		if n == "" {
			t.Fatal("empty app name")
		}
		if seen[n] {
			t.Fatalf("duplicate app name %q", n)
		}
		seen[n] = true
	}
	t.Logf("found %d applications", len(names))
}

func TestIsAppAvailableVim(t *testing.T) {
	t.Parallel()
	// vim is typically available on CI/development machines
	available := isAppAvailable("vim")
	t.Logf("vim available: %v", available)
}

func TestIsAppAvailableNonexistent(t *testing.T) {
	t.Parallel()
	available := isAppAvailable("nonexistent_app_xyz_12345")
	if available {
		t.Fatal("expected nonexistent app to be unavailable")
	}
}
