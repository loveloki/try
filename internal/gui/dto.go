package gui

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/loveloki/try/internal/i18n"
	"github.com/loveloki/try/internal/selector"
)

// EntryDTO 对齐 try-gui TryEntry。
type EntryDTO struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	BaseName     string  `json:"baseName"`
	Date         string  `json:"date"`
	Source       string  `json:"source"`
	Score        float64 `json:"score"`
	LastModified string  `json:"lastModified"`
	Path         string  `json:"path"`
	Highlights   []int   `json:"highlights,omitempty"`
	FileCount    int     `json:"fileCount"`
	SizeKB       float64 `json:"sizeKB"`
}

// FileDTO 对齐 try-gui FileEntry。
type FileDTO struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	SizeKB   float64 `json:"sizeKB"`
	Modified string  `json:"modified"`
	IsDir    bool    `json:"isDir"`
	Path     string  `json:"path"`
}

// PathsDTO 配置中的可操作根路径。
type PathsDTO struct {
	Tries string   `json:"tries"`
	Ships []string `json:"ships"`
}

// BootstrapDTO 启动所需的前端配置。
type BootstrapDTO struct {
	Locale   string            `json:"locale"`
	Theme    string            `json:"theme"`
	Messages map[string]string `json:"messages"`
	Paths    PathsDTO          `json:"paths"`
}

// EntriesResponse 列表与来源计数。
type EntriesResponse struct {
	Entries []EntryDTO     `json:"entries"`
	Counts  map[string]int `json:"counts"`
}

type createReq struct {
	Name string `json:"name"`
}

type pathsReq struct {
	Paths []string `json:"paths"`
}

type renameReq struct {
	Path    string `json:"path"`
	NewName string `json:"newName"`
}

type shipReq struct {
	Path      string `json:"path"`
	DestIndex int    `json:"destIndex"`
}

type pathReq struct {
	Path string `json:"path"`
}

type errorResp struct {
	Error string `json:"error"`
}

func entryToDTO(m selector.MatchedEntry) EntryDTO {
	e := m.Entry
	base, date := splitNameDate(e.Basename)
	return EntryDTO{
		ID:           e.Path,
		Name:         e.Basename,
		BaseName:     base,
		Date:         date,
		Source:       e.Source,
		Score:        m.Score,
		LastModified: e.Mtime.UTC().Format(time.RFC3339),
		Path:         e.Path,
		Highlights:   m.HighlightPositions,
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
	case ".md":
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

func bootstrapMessages(m *i18n.Messages) map[string]string {
	return map[string]string{
		"title":           m.Title,
		"searchPrefix":    m.SearchPrefix,
		"createNew":       m.CreateNew,
		"createHint":      m.CreateHint,
		"emptyStateHint":  m.EmptyStateHint,
		"noMatchesHint":   m.NoMatchesHint,
		"hintBar":         m.HintBar,
		"filterAll":       m.FilterAll,
		"deleteTitle":     m.DeleteTitle,
		"deleteFooter":    m.DeleteFooter,
		"renameTitle":     m.RenameTitle,
		"renamePrompt":    m.RenamePrompt,
		"renameEmpty":     m.RenameEmpty,
		"renameSlash":     m.RenameSlash,
		"renameExists":    m.RenameExists,
		"shipTitle":       m.ShipTitle,
		"shipMoveLabel":   m.ShipMoveLabel,
		"shipHint":        m.ShipHint,
		"itemCount":       m.ItemCount,
		"markedCount":     m.MarkedCount,
		"deleteModeLabel": m.DeleteModeLabel,
		"emptyInputHint":  m.EmptyInputHint,
	}
}
