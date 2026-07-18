package gui

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/loveloki/try/internal/selector"
)

// EntryView 是 GUI 选择器列表使用的进程内视图模型。
type EntryView struct {
	ID         string
	Name       string
	BaseName   string
	Date       string
	Source     string
	Score      float64
	Mtime      time.Time
	Path       string
	Highlights []int
}

// FileEntry 是 GUI 文件视图使用的进程内视图模型。
type FileEntry struct {
	ID     string
	Name   string
	Type   string
	SizeKB float64
	Mtime  time.Time
	IsDir  bool
	Path   string
}

// EntriesResult 聚合选择器条目和来源计数。
type EntriesResult struct {
	Entries []EntryView
	Counts  map[string]int
	Sources []string
}

func entryToView(m selector.MatchedEntry) EntryView {
	e := m.Entry
	base, date := splitNameDate(e.Basename)
	return EntryView{
		ID:         e.Path,
		Name:       e.Basename,
		BaseName:   base,
		Date:       date,
		Source:     e.Source,
		Score:      m.Score,
		Mtime:      e.Mtime,
		Path:       e.Path,
		Highlights: m.HighlightPositions,
	}
}

func splitNameDate(name string) (base, date string) {
	if loc := selector.DateSuffixRe.FindStringIndex(name); loc != nil {
		return name[:loc[0]], name[loc[0]+1:]
	}
	return name, ""
}

func fileTypeOf(name string, isDir bool) string {
	if isDir {
		return "dir"
	}
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".go":
		return "go"
	case ".ts", ".tsx":
		return "ts"
	case ".js", ".jsx":
		return "js"
	case ".md", ".markdown":
		return "md"
	case ".json":
		return "json"
	case ".txt":
		return "txt"
	case ".docx":
		return "docx"
	case ".zip":
		return "zip"
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg":
		return "image"
	default:
		return "unknown"
	}
}
