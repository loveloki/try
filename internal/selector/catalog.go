package selector

import (
	"path/filepath"
	"time"

	"github.com/loveloki/try/internal/fuzzy"
)

// LoadAllEntries 扫描 basePath 与 shipPaths 下的子目录，返回全部 Entry。
func LoadAllEntries(basePath string, shipPaths []string) []Entry {
	now := time.Now()
	var result []Entry
	result = append(result, scanDir(basePath, "tries", now)...)
	for _, sp := range shipPaths {
		result = append(result, scanDir(sp, filepath.Base(sp), now)...)
	}
	if result == nil {
		return []Entry{}
	}
	return result
}

// MatchEntries 按 sourceFilter 过滤后，对 query 做模糊匹配。
// sourceFilter 为空表示不过滤；limit <= 0 表示不截断。
func MatchEntries(entries []Entry, query, sourceFilter string, limit int) []MatchedEntry {
	filtered := filterBySource(entries, sourceFilter)
	return matchQuery(filtered, query, limit)
}

// SourceCounts 统计各来源条目数量；options 中的键均会出现，"" 表示全部。
func SourceCounts(entries []Entry, options []string) map[string]int {
	return computeSourceCounts(entries, options)
}

// SourceOptions 构建来源过滤选项列表：""（全部）、"tries"、各 ship 目录 basename。
func SourceOptions(shipPaths []string) []string {
	opts := []string{"", "tries"}
	for _, sp := range shipPaths {
		opts = append(opts, filepath.Base(sp))
	}
	return opts
}

func filterBySource(entries []Entry, sourceFilter string) []Entry {
	if sourceFilter == "" {
		return entries
	}
	var filtered []Entry
	for _, e := range entries {
		if e.Source == sourceFilter {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func matchQuery(entries []Entry, query string, maxResults int) []MatchedEntry {
	fuzzyEntries := make([]fuzzy.Entry, len(entries))
	for i, e := range entries {
		name := e.Basename
		if loc := DateSuffixRe.FindStringIndex(name); loc != nil {
			name = name[:loc[0]]
		}
		fuzzyEntries[i] = fuzzy.Entry{
			Text:      name,
			BaseScore: e.BaseScore,
			Data:      e,
		}
	}

	results := fuzzy.Match(fuzzyEntries, query, maxResults)
	matched := make([]MatchedEntry, len(results))
	for i, r := range results {
		matched[i] = MatchedEntry{
			Entry:              r.Entry.Data.(Entry),
			Score:              r.Score,
			HighlightPositions: r.Positions,
		}
	}
	return matched
}
