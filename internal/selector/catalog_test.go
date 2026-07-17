package selector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAllEntries(t *testing.T) {
	tmp := t.TempDir()
	tries := filepath.Join(tmp, "tries")
	ship := filepath.Join(tmp, "ship")
	bug := filepath.Join(tmp, "bug")
	mustMkdir(t, tries, "alpha-2026-01-01")
	mustMkdir(t, tries, ".hidden")
	mustMkdir(t, ship, "beta-2026-02-01")
	mustMkdir(t, bug, "gamma-2026-03-01")
	// 普通文件应被忽略
	if err := os.WriteFile(filepath.Join(tries, "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	entries := LoadAllEntries(tries, []string{ship, bug})
	checkEntryCount(t, entries, 3)
	byName := map[string]Entry{}
	for _, e := range entries {
		byName[e.Basename] = e
	}
	checkSource(t, byName["alpha-2026-01-01"], "tries")
	checkSource(t, byName["beta-2026-02-01"], "ship")
	checkSource(t, byName["gamma-2026-03-01"], "bug")
}

func TestMatchEntries(t *testing.T) {
	entries := []Entry{
		{Basename: "axum-middleware-2025-07-15", BaseScore: 5, Source: "tries"},
		{Basename: "sqlx-pool-2025-07-10", BaseScore: 3, Source: "tries"},
		{Basename: "bubbletea-v2-2025-05-30", BaseScore: 4, Source: "bug"},
	}

	tests := []struct {
		name   string
		query  string
		source string
		want   []string
	}{
		{"空查询全部", "", "", []string{"axum-middleware-2025-07-15", "bubbletea-v2-2025-05-30", "sqlx-pool-2025-07-10"}},
		{"来源过滤", "", "bug", []string{"bubbletea-v2-2025-05-30"}},
		{"模糊匹配", "axm", "", []string{"axum-middleware-2025-07-15"}},
		{"过滤无匹配", "zzz", "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchEntries(entries, tt.query, tt.source, 0)
			checkMatchedNames(t, got, tt.want)
		})
	}
}

func TestSourceCounts(t *testing.T) {
	entries := []Entry{
		{Source: "tries"},
		{Source: "tries"},
		{Source: "ship"},
	}
	opts := []string{"", "tries", "ship", "bug"}
	counts := SourceCounts(entries, opts)
	checkCount(t, counts, "", 3)
	checkCount(t, counts, "tries", 2)
	checkCount(t, counts, "ship", 1)
	checkCount(t, counts, "bug", 0)
}

func mustMkdir(t *testing.T, parent, name string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(parent, name), 0o755); err != nil {
		t.Fatal(err)
	}
}

func checkEntryCount(t *testing.T, entries []Entry, want int) {
	t.Helper()
	if len(entries) != want {
		t.Fatalf("len(entries) = %d, want %d", len(entries), want)
	}
}

func checkSource(t *testing.T, e Entry, want string) {
	t.Helper()
	if e.Source != want {
		t.Errorf("Source = %q, want %q", e.Source, want)
	}
}

func checkMatchedNames(t *testing.T, got []MatchedEntry, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i, name := range want {
		if got[i].Entry.Basename != name {
			t.Errorf("[%d] Basename = %q, want %q", i, got[i].Entry.Basename, name)
		}
	}
}

func checkCount(t *testing.T, counts map[string]int, key string, want int) {
	t.Helper()
	if counts[key] != want {
		t.Errorf("counts[%q] = %d, want %d", key, counts[key], want)
	}
}
