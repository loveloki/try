package fuzzy

import (
	"testing"
)

// checkMatch 验证匹配结果的名称和顺序（测试功能而非内部评分）
func checkMatch(t *testing.T, entries []Entry, query string, limit int, wantNames []string) {
	t.Helper()
	results := Match(entries, query, limit)
	gotNames := make([]string, len(results))
	for i, r := range results {
		gotNames[i] = r.Entry.Text
	}
	if len(gotNames) != len(wantNames) {
		t.Fatalf("Match(%q) returned %d results %v, want %d results %v",
			query, len(gotNames), gotNames, len(wantNames), wantNames)
	}
	for i := range wantNames {
		if gotNames[i] != wantNames[i] {
			t.Errorf("Match(%q)[%d] = %q, want %q (full: %v)", query, i, gotNames[i], wantNames[i], gotNames)
			return
		}
	}
}

// checkMatchCount 验证匹配数量
func checkMatchCount(t *testing.T, entries []Entry, query string, limit int, wantCount int) {
	t.Helper()
	results := Match(entries, query, limit)
	if len(results) != wantCount {
		names := make([]string, len(results))
		for i, r := range results {
			names[i] = r.Entry.Text
		}
		t.Errorf("Match(%q) returned %d results %v, want %d", query, len(results), names, wantCount)
	}
}

// 测试用条目集
func testEntries() []Entry {
	return []Entry{
		{Text: "redis-server-2025-08-14", BaseScore: 4.12},
		{Text: "quick-test", BaseScore: 5.0},
		{Text: "another-project-2025-08-10", BaseScore: 1.5},
		{Text: "golang-api", BaseScore: 3.0},
		{Text: "react-app-2025-08-17", BaseScore: 2.0},
	}
}

func TestEmptyQuery(t *testing.T) {
	entries := testEntries()
	// 空 query 返回全部条目，按 BaseScore 降序
	checkMatch(t, entries, "", 0, []string{
		"quick-test",
		"redis-server-2025-08-14",
		"golang-api",
		"react-app-2025-08-17",
		"another-project-2025-08-10",
	})
}

func TestEmptyQueryWithLimit(t *testing.T) {
	entries := testEntries()
	checkMatch(t, entries, "", 2, []string{
		"quick-test",
		"redis-server-2025-08-14",
	})
}

func TestNoMatch(t *testing.T) {
	entries := testEntries()
	checkMatchCount(t, entries, "xyz123", 0, 0)
}

func TestExactPrefixMatch(t *testing.T) {
	entries := testEntries()
	// "redis" 应该匹配 redis-server，且排在前面
	results := Match(entries, "redis", 0)
	if len(results) == 0 {
		t.Fatal("expected at least one match for 'redis'")
	}
	if results[0].Entry.Text != "redis-server-2025-08-14" {
		t.Errorf("top result for 'redis' = %q, want 'redis-server-2025-08-14'", results[0].Entry.Text)
	}
}

func TestCaseInsensitive(t *testing.T) {
	entries := []Entry{
		{Text: "MyProject", BaseScore: 1.0},
		{Text: "another", BaseScore: 1.0},
	}
	checkMatchCount(t, entries, "myproject", 0, 1)
	checkMatchCount(t, entries, "MYPROJECT", 0, 1)
}

func TestBaseScoreAffectsRanking(t *testing.T) {
	// 两个同样匹配 "test" 的条目，BaseScore 高的应排前面
	entries := []Entry{
		{Text: "test-old", BaseScore: 1.0},
		{Text: "test-new", BaseScore: 5.0},
	}
	results := Match(entries, "test", 0)
	if len(results) < 2 {
		t.Fatal("expected 2 results")
	}
	if results[0].Entry.Text != "test-new" {
		t.Errorf("higher BaseScore entry should rank first, got %q", results[0].Entry.Text)
	}
}

func TestLimitTruncation(t *testing.T) {
	entries := testEntries()
	// "a" 可能匹配多个，但 limit=2 只返回 2 个
	results := Match(entries, "a", 2)
	if len(results) > 2 {
		t.Errorf("expected at most 2 results, got %d", len(results))
	}
}

func TestPositionsTracking(t *testing.T) {
	entries := []Entry{
		{Text: "abc", BaseScore: 1.0},
	}
	results := Match(entries, "ac", 0)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
	pos := results[0].Positions
	if len(pos) != 2 || pos[0] != 0 || pos[1] != 2 {
		t.Errorf("positions = %v, want [0, 2]", pos)
	}
}

func TestSingleEntry(t *testing.T) {
	entries := []Entry{
		{Text: "hello", BaseScore: 1.0},
	}
	checkMatch(t, entries, "hlo", 0, []string{"hello"})
	checkMatchCount(t, entries, "xyz", 0, 0)
}

func TestEmptyEntries(t *testing.T) {
	checkMatchCount(t, nil, "test", 0, 0)
	checkMatchCount(t, nil, "", 0, 0)
	checkMatchCount(t, []Entry{}, "test", 0, 0)
}

func TestDensityFavorsCompactMatches(t *testing.T) {
	// "ab" 匹配 "ab-xyz"（紧凑）应优于 "a---b"（稀疏），同 BaseScore
	entries := []Entry{
		{Text: "a---b-long-name", BaseScore: 1.0},
		{Text: "ab-xyz", BaseScore: 1.0},
	}
	results := Match(entries, "ab", 0)
	if len(results) < 2 {
		t.Fatal("expected 2 results")
	}
	if results[0].Entry.Text != "ab-xyz" {
		t.Errorf("compact match should rank higher, got %q first", results[0].Entry.Text)
	}
}
