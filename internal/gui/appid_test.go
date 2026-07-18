package gui

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestAppIDMatchesFyneAppToml(t *testing.T) {
	t.Parallel()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	tomlPath := filepath.Join(filepath.Dir(file), "..", "..", "cmd", "try-gui", "FyneApp.toml")
	data, err := os.ReadFile(tomlPath)
	if err != nil {
		t.Fatalf("read FyneApp.toml: %v", err)
	}
	wantLine := `ID = "` + AppID + `"`
	if !strings.Contains(string(data), wantLine) {
		t.Fatalf("FyneApp.toml missing %q", wantLine)
	}
}
