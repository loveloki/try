package selector

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func checkTimeAgo(t *testing.T, d time.Duration, want string) {
	t.Helper()
	got := FormatTimeAgo(d)
	if got != want {
		t.Errorf("FormatTimeAgo(%v) = %q, want %q", d, got, want)
	}
}

func TestFormatTimeAgo(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"just now", 0, "just now"},
		{"30 seconds", 30 * time.Second, "just now"},
		{"59 seconds", 59 * time.Second, "just now"},
		{"1 minute", 60 * time.Second, "1m ago"},
		{"30 minutes", 30 * time.Minute, "30m ago"},
		{"59 minutes", 59 * time.Minute, "59m ago"},
		{"1 hour", 60 * time.Minute, "1h ago"},
		{"23 hours", 23 * time.Hour, "23h ago"},
		{"1 day", 24 * time.Hour, "1d ago"},
		{"6 days", 6 * 24 * time.Hour, "6d ago"},
		{"1 week", 7 * 24 * time.Hour, "7d ago"},
		{"29 days", 29 * 24 * time.Hour, "29d ago"},
		{"1 month", 30 * 24 * time.Hour, "1mo ago"},
		{"11 months", 11 * 30 * 24 * time.Hour, "11mo ago"},
		{"1 year", 365 * 24 * time.Hour, "1y ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkTimeAgo(t, tt.d, tt.want)
		})
	}
}

func checkDateSuffix(t *testing.T, name string, wantMatch bool) {
	t.Helper()
	got := DateSuffixRe.MatchString(name)
	if got != wantMatch {
		t.Errorf("DateSuffixRe.MatchString(%q) = %v, want %v", name, got, wantMatch)
	}
}

func TestDateSuffixRe(t *testing.T) {
	tests := []struct {
		name      string
		wantMatch bool
	}{
		{"project-2025-08-14", true},
		{"a-2025-01-01", true},
		{"multi-part-name-2025-12-31", true},
		{"no-date-suffix", false},
		{"2025-08-14", false},
		{"project-2025-08", false},
		{"project-25-08-14", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkDateSuffix(t, tt.name, tt.wantMatch)
		})
	}
}

func TestMatchedEntryListItem(t *testing.T) {
	me := MatchedEntry{
		Entry: Entry{Basename: "test-dir"},
	}
	if me.FilterValue() != "" {
		t.Errorf("FilterValue() = %q, want empty", me.FilterValue())
	}
	if me.Title() != "test-dir" {
		t.Errorf("Title() = %q, want %q", me.Title(), "test-dir")
	}
	if me.Description() != "" {
		t.Errorf("Description() = %q, want empty", me.Description())
	}
}

func TestDirExists(t *testing.T) {
	tmpDir := t.TempDir()
	if !DirExists(tmpDir) {
		t.Error("DirExists should return true for temp dir")
	}
	if DirExists(filepath.Join(tmpDir, "nonexistent")) {
		t.Error("DirExists should return false for nonexistent")
	}
	// 创建文件，DirExists 应返回 false
	f := filepath.Join(tmpDir, "file.txt")
	os.WriteFile(f, []byte("x"), 0o644)
	if DirExists(f) {
		t.Error("DirExists should return false for regular file")
	}
}

func TestEnvInt(t *testing.T) {
	t.Setenv("TEST_INT", "42")
	if EnvInt("TEST_INT") != 42 {
		t.Error("EnvInt should parse integer from env")
	}
	t.Setenv("TEST_INT", "invalid")
	if EnvInt("TEST_INT") != 0 {
		t.Error("EnvInt should return 0 for invalid value")
	}
	if EnvInt("NONEXISTENT_VAR") != 0 {
		t.Error("EnvInt should return 0 for missing var")
	}
}
