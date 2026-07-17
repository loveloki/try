package gui

import (
	"path/filepath"
	"testing"
)

func TestIsAllowedPath(t *testing.T) {
	tmp := t.TempDir()
	tries := filepath.Join(tmp, "tries")
	ship := filepath.Join(tmp, "ship")
	outside := filepath.Join(tmp, "outside")
	roots := []string{tries, ship}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"根目录本身", tries, true},
		{"子目录", filepath.Join(tries, "a-2026-01-01"), true},
		{"ship 子树", filepath.Join(ship, "proj"), true},
		{"越界", outside, false},
		{"穿越父目录", filepath.Join(tries, "..", "outside"), false},
		{"空路径", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAllowedPath(tt.path, roots)
			if got != tt.want {
				t.Errorf("IsAllowedPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestIsAllowedPathEmptyRoots(t *testing.T) {
	if IsAllowedPath("/tmp/x", nil) {
		t.Error("empty roots should deny")
	}
}

func TestIsAllowedTarget(t *testing.T) {
	tmp := t.TempDir()
	tries := filepath.Join(tmp, "tries")
	ship := filepath.Join(tmp, "ship")
	roots := []string{tries, ship}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"根目录本身拒绝", tries, false},
		{"ship 根目录拒绝", ship, false},
		{"子目录允许", filepath.Join(tries, "a-2026-01-01"), true},
		{"深层子路径允许", filepath.Join(ship, "proj", "note.txt"), true},
		{"越界拒绝", filepath.Join(tmp, "outside"), false},
		{"空路径拒绝", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAllowedTarget(tt.path, roots)
			if got != tt.want {
				t.Errorf("IsAllowedTarget(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
