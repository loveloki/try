package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const (
	colorNameMatch         fyne.ThemeColorName = "match"
	colorNameAccent        fyne.ThemeColorName = "accent"
	colorNameDangerSurface fyne.ThemeColorName = "dangerSurface"
	colorNameHeader        fyne.ThemeColorName = "header"
	colorNameMuted         fyne.ThemeColorName = "muted"
	colorNameHighlightDim  fyne.ThemeColorName = "highlightDim"
	entryRowHeight         float32             = 40
)

func highlightSegments(text string, positions []int, marked, selected bool) []widget.RichTextSegment {
	posSet := make(map[int]bool, len(positions))
	for _, p := range positions {
		posSet[p] = true
	}
	segs := make([]widget.RichTextSegment, 0, len(text))
	i := 0
	for i < len(text) {
		if posSet[i] {
			j := i
			for j < len(text) && posSet[j] {
				j++
			}
			style := widget.RichTextStyle{
				Inline:    true,
				ColorName: colorNameMatch,
				TextStyle: fyne.TextStyle{Bold: true, Monospace: true},
			}
			if marked {
				style.ColorName = theme.ColorNameError
			}
			segs = append(segs, &widget.TextSegment{Text: text[i:j], Style: style})
			i = j
			continue
		}
		j := i
		for j < len(text) && !posSet[j] {
			j++
		}
		style := widget.RichTextStyle{
			Inline:    true,
			TextStyle: fyne.TextStyle{Monospace: true},
		}
		if marked {
			style.ColorName = theme.ColorNameError
		} else if selected {
			style.TextStyle.Bold = true
		}
		segs = append(segs, &widget.TextSegment{Text: text[i:j], Style: style})
		i = j
	}
	return segs
}

func fileTypeIcon(fileType string, isDir bool) fyne.Resource {
	if isDir {
		return theme.FolderIcon()
	}
	switch fileType {
	case "go", "ts", "js":
		return theme.DocumentCreateIcon()
	case "md", "txt", "docx":
		return theme.DocumentIcon()
	case "json":
		return theme.FileIcon()
	case "zip":
		return theme.ContentAddIcon()
	case "image":
		return theme.FileImageIcon()
	default:
		return theme.FileIcon()
	}
}

func fileTypePrefix(fileType string, isDir bool) string {
	if isDir {
		return "dir"
	}
	return fileType
}

func formatSourceTabLabel(allLabel, source string) string {
	if source == "" {
		return allLabel
	}
	return source
}

func entryNameSegments(e EntryView, marked, selected bool) []widget.RichTextSegment {
	baseSegs := highlightSegments(e.BaseName, offsetHighlights(e.Highlights, 0, len(e.BaseName)), marked, selected)
	if e.Date == "" {
		return baseSegs
	}
	dateOffset := len(e.BaseName) + 1
	dateSegs := highlightSegments(e.Date, offsetHighlights(e.Highlights, dateOffset, dateOffset+len(e.Date)), marked, false)
	dash := widget.RichTextSegment(&widget.TextSegment{
		Text:  "-",
		Style: widget.RichTextStyle{Inline: true, TextStyle: fyne.TextStyle{Monospace: true}},
	})
	out := append(baseSegs, dash)
	return append(out, dateSegs...)
}

func offsetHighlights(positions []int, start, end int) []int {
	out := make([]int, 0)
	for _, p := range positions {
		if p >= start && p < end {
			out = append(out, p-start)
		}
	}
	return out
}

func fileMetaText(f FileEntry) string {
	ago := formatModTime(f.Mtime)
	if f.IsDir {
		return fmt.Sprintf("%s  %s", fileTypePrefix(f.Type, f.IsDir), ago)
	}
	return fmt.Sprintf("%s  %s  %s", fileTypePrefix(f.Type, f.IsDir), formatSizeKB(f.SizeKB), ago)
}

func entryMetaText(e EntryView) string {
	ago := formatModTime(e.Mtime)
	if ago == "" {
		return e.Source
	}
	return fmt.Sprintf("%s  %s", e.Source, ago)
}
