package selector

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/loveloki/try/internal/fuzzy"
)

// loadAllTries 从 basePath 和 shipPaths 读取所有子目录，并计算各来源数量。
func (m *SelectorModel) loadAllTries() []Entry {
	if m.allTries != nil {
		return m.allTries
	}

	now := time.Now()
	var result []Entry

	result = append(result, scanDir(m.basePath, "tries", now)...)
	for _, sp := range m.shipPaths {
		source := filepath.Base(sp)
		result = append(result, scanDir(sp, source, now)...)
	}

	m.allTries = result
	if m.allTries == nil {
		m.allTries = []Entry{}
	}
	m.sourceCounts = computeSourceCounts(m.allTries, m.sourceOptions)
	return m.allTries
}

func computeSourceCounts(entries []Entry, options []string) map[string]int {
	counts := make(map[string]int, len(options))
	for _, opt := range options {
		counts[opt] = 0
	}
	for _, e := range entries {
		counts[e.Source]++
	}
	counts[""] = len(entries)
	return counts
}

func scanDir(dir, source string, now time.Time) []Entry {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var result []Entry
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if !entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		mtime := info.ModTime()
		hoursSinceMod := now.Sub(mtime).Hours()
		baseScore := 3.0 / math.Sqrt(hoursSinceMod+1)
		if DateSuffixRe.MatchString(entry.Name()) {
			baseScore += 2.0
		}

		result = append(result, Entry{
			Basename:  entry.Name(),
			Path:      filepath.Join(dir, entry.Name()),
			Mtime:     mtime,
			BaseScore: baseScore,
			Source:    source,
		})
	}
	return result
}

// refreshList 根据当前搜索词和来源过滤重新匹配并更新列表
func (m *SelectorModel) refreshList() tea.Cmd {
	query := m.textInput.Value()
	filterKey := query + "\x00" + m.sourceFilter
	if filterKey == m.lastQuery && m.cachedResults != nil {
		return nil
	}

	filtered := m.filteredEntries()
	maxResults := bodyHeight(m)
	if maxResults < 3 {
		maxResults = 3
	}

	matched := m.matchEntries(filtered, query, maxResults)
	m.cachedResults = matched
	m.lastQuery = filterKey

	items := make([]list.Item, len(matched))
	for i, me := range matched {
		items[i] = me
	}
	return m.list.SetItems(items)
}

func (m *SelectorModel) filteredEntries() []Entry {
	allTries := m.loadAllTries()
	if m.sourceFilter == "" {
		return allTries
	}
	var filtered []Entry
	for _, e := range allTries {
		if e.Source == m.sourceFilter {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func (m *SelectorModel) matchEntries(entries []Entry, query string, maxResults int) []MatchedEntry {
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
