package gui

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func (g *desktopGUI) focusSearch() {
	if g.search != nil {
		g.window.Canvas().Focus(g.search)
	}
}

func (g *desktopGUI) activeStatus() *statusBar {
	if g.view == "files" {
		return g.filesStatus
	}
	return g.selectorStatus
}

func (g *desktopGUI) setToast(msg string) {
	bar := g.activeStatus()
	if bar == nil {
		return
	}
	g.toastGen++
	gen := g.toastGen
	bar.SetToast(msg)
	go func() {
		time.Sleep(2 * time.Second)
		fyne.Do(func() {
			if g.toastGen != gen {
				return
			}
			if b := g.activeStatus(); b != nil {
				b.ClearToast()
			}
		})
	}()
}

// setPersistentToast 显示 Toast 且不自动清除（用于复制进行中）。
func (g *desktopGUI) setPersistentToast(msg string) {
	bar := g.activeStatus()
	if bar == nil {
		return
	}
	g.toastGen++
	bar.SetToast(msg)
}

func (g *desktopGUI) showError(err error) {
	dialog.ShowError(err, g.window)
}

func (g *desktopGUI) openPath(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return g.service.openFile(ctx, path)
}

func (g *desktopGUI) applyTheme() {
	g.app.Settings().SetTheme(newGUITheme(g.themeName))
}

func wrapIndex(index, length int) int {
	if length <= 0 {
		return 0
	}
	for index < 0 {
		index += length
	}
	return index % length
}

// stepFileSelected 文件列表光标步进：无选中时 ↓→首项、↑→末项；否则首尾循环。
func stepFileSelected(selected, delta, length int) int {
	if length <= 0 {
		return -1
	}
	if selected < 0 {
		if delta > 0 {
			return 0
		}
		return length - 1
	}
	return wrapIndex(selected+delta, length)
}

func normalizeTheme(themeName string) string {
	if themeName == "light" {
		return "light"
	}
	return "dark"
}

func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func previewPaths(paths []string) []string {
	limit := min(5, len(paths))
	preview := make([]string, 0, limit)
	for _, path := range paths[:limit] {
		preview = append(preview, filepath.Base(path))
	}
	if len(paths) > limit {
		preview = append(preview, fmt.Sprintf("+%d", len(paths)-limit))
	}
	return preview
}
