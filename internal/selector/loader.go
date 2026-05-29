package selector

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/xleine/try/internal/fuzzy"
)

// loadAllTries 从 basePath 读取所有子目录，计算时间权重基础分
func (m *SelectorModel) loadAllTries() []Entry {
	if m.allTries != nil {
		return m.allTries
	}

	entries, err := os.ReadDir(m.basePath)
	if err != nil {
		return nil
	}

	now := time.Now()
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
			Path:      filepath.Join(m.basePath, entry.Name()),
			Mtime:     mtime,
			BaseScore: baseScore,
		})
	}

	m.allTries = result
	return result
}

// refreshList 根据当前搜索词重新匹配并更新列表
func (m *SelectorModel) refreshList() tea.Cmd {
	query := m.textInput.Value()
	if query == m.lastQuery && m.cachedResults != nil {
		return nil
	}

	allTries := m.loadAllTries()
	maxResults := m.height - 6
	if maxResults < 3 {
		maxResults = 3
	}

	// selector.Entry → fuzzy.Entry 转换
	fuzzyEntries := make([]fuzzy.Entry, len(allTries))
	for i, e := range allTries {
		fuzzyEntries[i] = fuzzy.Entry{
			Text:      e.Basename,
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
	m.cachedResults = matched
	m.lastQuery = query

	items := make([]list.Item, len(matched))
	for i, me := range matched {
		items[i] = me
	}
	return m.list.SetItems(items)
}
