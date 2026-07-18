package dialog

import (
	"path/filepath"
	"testing"
)

func TestIsPathUnder(t *testing.T) {
	t.Parallel()
	base := filepath.Join("tmp", "tries")
	tests := []struct {
		name   string
		target string
		want   bool
	}{
		{"子目录", filepath.Join(base, "a-2026-01-01"), true},
		{"深层", filepath.Join(base, "a", "b"), true},
		{"自身", base, false},
		{"越界", filepath.Join("tmp", "outside"), false},
		{"穿越", filepath.Join(base, "..", "outside"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPathUnder(base, filepath.Clean(tt.target))
			if got != tt.want {
				t.Errorf("isPathUnder(%q, %q) = %v, want %v", base, tt.target, got, tt.want)
			}
		})
	}
}
