package gui

import (
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
