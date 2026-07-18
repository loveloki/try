package gui

import (
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func (g *desktopGUI) moveSelection(delta int) {
	if g.view == "files" {
		g.fileSelected = wrapIndex(g.fileSelected+delta, len(g.files))
		if g.list != nil && len(g.files) > 0 {
			g.list.Select(g.fileSelected)
			g.list.ScrollTo(g.fileSelected)
			g.list.Refresh()
		}
		return
	}
	g.selected = wrapIndex(g.selected+delta, len(g.entries))
	if g.list != nil && len(g.entries) > 0 {
		g.list.Select(g.selected)
		g.list.ScrollTo(g.selected)
		g.list.Refresh()
	}
}

func (g *desktopGUI) openSelected() {
	if g.view == "files" {
		g.openSelectedFile()
		return
	}
	switch decideSelectorOpen(len(g.marked), g.selected, len(g.entries)) {
	case selectorOpenDelete:
		g.confirmDelete()
	case selectorOpenFiles:
		entry := g.entries[g.selected]
		g.enterFiles(entry.Path, entry.Path)
	default:
		g.setToast(g.msgs.GUIErrNoSelection)
	}
}

func (g *desktopGUI) submitSearch() {
	if g.view != "selector" {
		return
	}
	if len(g.entries) == 0 && strings.TrimSpace(g.query) != "" {
		g.createFromName(g.query)
		return
	}
	g.openSelected()
}

func (g *desktopGUI) openSelectedFile() {
	if g.fileSelected < 0 || g.fileSelected >= len(g.files) {
		g.setToast(g.msgs.GUIErrNoSelection)
		return
	}
	file := g.files[g.fileSelected]
	if file.IsDir {
		g.enterFiles(g.filesRoot, file.Path)
		return
	}
	if err := g.openPath(file.Path); err != nil {
		g.showError(fmt.Errorf("%s: %w", g.msgs.GUIErrOpen, err))
		return
	}
	g.setToast(g.msgs.GUIToastOpened)
}

func (g *desktopGUI) toggleMark() {
	if g.view == "files" {
		g.toggleFileMark()
		return
	}
	if g.selected < 0 || g.selected >= len(g.entries) {
		return
	}
	path := g.entries[g.selected].Path
	if g.marked[path] {
		delete(g.marked, path)
	} else {
		g.marked[path] = true
	}
	g.refreshSelectorUI()
}

func (g *desktopGUI) toggleFileMark() {
	g.toggleFileMarkAt(g.fileSelected)
}

func (g *desktopGUI) confirmDelete() {
	paths := g.selectedDeletePaths()
	if len(paths) == 0 {
		return
	}
	title := fmt.Sprintf(g.msgs.DeleteTitle, len(paths))
	body := widget.NewLabel(strings.Join(previewPaths(paths), "\n"))
	d := dialog.NewCustomConfirm(title, g.msgs.DeleteOptionYes, g.msgs.DeleteOptionNo, body, func(ok bool) {
		if !ok {
			return
		}
		if err := g.deletePaths(paths); err != nil {
			g.showError(err)
			return
		}
		g.setToast(g.msgs.GUIToastDeleted)
	}, g.window)
	d.Show()
}

func (g *desktopGUI) deletePaths(paths []string) error {
	if g.view == "files" {
		if err := g.service.deleteFiles(paths); err != nil {
			return err
		}
		g.fileMarked = map[string]bool{}
		g.refreshFilesUI()
		return nil
	}
	if err := g.service.deleteEntries(paths); err != nil {
		return err
	}
	g.marked = map[string]bool{}
	g.refreshSelectorUI()
	return nil
}

func (g *desktopGUI) selectedDeletePaths() []string {
	if g.view == "files" {
		if len(g.fileMarked) > 0 {
			return mapKeys(g.fileMarked)
		}
		if g.fileSelected >= 0 && g.fileSelected < len(g.files) {
			return []string{g.files[g.fileSelected].Path}
		}
		return nil
	}
	if len(g.marked) > 0 {
		return mapKeys(g.marked)
	}
	if g.selected >= 0 && g.selected < len(g.entries) {
		return []string{g.entries[g.selected].Path}
	}
	return nil
}

func (g *desktopGUI) promptCreate() {
	if g.view != "selector" {
		return
	}
	entry := widget.NewEntry()
	entry.SetPlaceHolder(g.msgs.GUISearchPlace)
	g.showInputDialog(g.msgs.CreateNew, entry, func(value string) {
		g.createFromName(value)
	})
}

func (g *desktopGUI) createFromName(name string) {
	path, err := g.service.createEntry(name)
	if err != nil {
		g.showError(err)
		return
	}
	g.query = ""
	// 对齐 TUI execMkdir→execCd：创建成功后进入项目文件视图
	g.enterFiles(path, path)
	g.setToast(g.msgs.GUIToastCreated)
}

func (g *desktopGUI) promptRename() {
	if g.view != "selector" || g.selected < 0 || g.selected >= len(g.entries) {
		return
	}
	selected := g.entries[g.selected]
	entry := widget.NewEntry()
	entry.SetText(selected.Name)
	g.showInputDialog(g.msgs.RenameTitle, entry, func(value string) {
		if _, err := g.service.renameEntry(selected.Path, value); err != nil {
			g.showError(err)
			return
		}
		g.refreshSelectorUI()
		g.setToast(g.msgs.GUIToastRenamed)
	})
}

func (g *desktopGUI) promptShip() {
	if g.view != "selector" || g.selected < 0 || g.selected >= len(g.entries) {
		return
	}
	if len(g.service.shipPaths) == 0 {
		g.setToast(g.msgs.ShipEmptyErr)
		return
	}
	selected := g.entries[g.selected]
	options := make([]string, len(g.service.shipPaths))
	for i, path := range g.service.shipPaths {
		options[i] = filepath.Base(path)
	}
	destIndex := 0
	selectWidget := widget.NewSelect(options, func(value string) {
		for i, opt := range options {
			if opt == value {
				destIndex = i
				return
			}
		}
	})
	selectWidget.SetSelected(options[0])
	content := container.NewVBox(widget.NewLabel(g.msgs.ShipHint), selectWidget)
	d := dialog.NewCustomConfirm(g.msgs.ShipTitle, g.msgs.GUIConfirm, g.msgs.GUICancel, content, func(ok bool) {
		if !ok {
			return
		}
		if _, err := g.service.shipEntry(selected.Path, destIndex); err != nil {
			g.showError(err)
			return
		}
		g.refreshSelectorUI()
		g.setToast(g.msgs.GUIToastShipped)
	}, g.window)
	d.Show()
}

func (g *desktopGUI) showInputDialog(title string, entry *widget.Entry, apply func(string)) {
	content := container.NewVBox(entry)
	d := dialog.NewCustomConfirm(title, g.msgs.GUIConfirm, g.msgs.GUICancel, content, func(ok bool) {
		if ok {
			apply(entry.Text)
		}
	}, g.window)
	d.Show()
	g.window.Canvas().Focus(entry)
}

func (g *desktopGUI) handleEsc() {
	if g.view == "files" {
		if len(g.fileMarked) > 0 {
			g.fileMarked = map[string]bool{}
			g.refreshFilesUI()
			return
		}
		g.handleFilesBack()
		return
	}
	if len(g.marked) > 0 {
		g.marked = map[string]bool{}
		g.refreshSelectorUI()
		return
	}
	if g.query != "" {
		g.query = ""
		g.switchToSelector(true)
		return
	}
	g.window.Hide()
}

func (g *desktopGUI) handleFilesBack() {
	if g.view != "files" {
		return
	}
	if g.filesPath == g.filesRoot {
		g.switchToSelector(false)
		return
	}
	g.enterFiles(g.filesRoot, filepath.Dir(g.filesPath))
}

