package gui

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type contextMenuActions struct {
	onOpen   func()
	onReveal func()
	onRename func()
	onDelete func()
}

func (g *desktopGUI) showFileContextMenu(e *fyne.PointEvent, idx int) {
	g.setFileSelected(idx)
	actions := g.buildContextMenuActions(idx)
	menu := g.buildContextMenu(actions, g.buildOpenWithSubmenu(idx))
	widget.ShowPopUpMenuAtPosition(menu, g.window.Canvas(), e.AbsolutePosition)
}

func (g *desktopGUI) buildContextMenuActions(idx int) contextMenuActions {
	path := g.files[idx].Path
	return contextMenuActions{
		onOpen:   func() { g.openFileAt(idx) },
		onReveal: func() { g.revealFile(path) },
		onRename: func() { g.promptRenameFile(idx) },
		onDelete: func() { g.confirmDeleteSingleFile(idx) },
	}
}

func (g *desktopGUI) buildContextMenu(a contextMenuActions, openWith *fyne.MenuItem) *fyne.Menu {
	items := []*fyne.MenuItem{
		g.menuItem(g.msgs.GUIContextMenuOpen, a.onOpen),
		openWith,
		fyne.NewMenuItemSeparator(),
		g.menuItem(g.msgs.GUIContextMenuReveal, a.onReveal),
		fyne.NewMenuItemSeparator(),
		g.menuItem(g.msgs.GUIContextMenuRename, a.onRename),
		g.menuItem(g.msgs.GUIContextMenuDelete, a.onDelete),
	}
	return fyne.NewMenu("", items...)
}

func (g *desktopGUI) menuItem(label string, action func()) *fyne.MenuItem {
	return fyne.NewMenuItem(label, action)
}

func (g *desktopGUI) buildOpenWithSubmenu(idx int) *fyne.MenuItem {
	item := fyne.NewMenuItem(g.msgs.GUIContextMenuOpenWith, nil)
	ext := strings.ToLower(filepath.Ext(g.files[idx].Name))
	apps := buildAvailableApps(ext, g.cfg.OpenWith)
	if len(apps) == 0 {
		item.ChildMenu = fyne.NewMenu("",
			fyne.NewMenuItem(g.msgs.GUIContextMenuNoApps, func() {}),
		)
		return item
	}
	items := make([]*fyne.MenuItem, 0, len(apps))
	for _, a := range apps {
		name, path := a.Name, a.Path
		items = append(items, fyne.NewMenuItem(name, func() {
			g.openWithAppByName(path, g.files[idx].Path)
		}))
	}
	item.ChildMenu = fyne.NewMenu("", items...)
	return item
}

func (g *desktopGUI) openWithAppByName(app, path string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := openWithApp(ctx, app, path); err != nil {
		g.showError(fmt.Errorf("%s: %w", g.msgs.GUIErrOpen, err))
		return
	}
	g.setToast(g.msgs.GUIToastOpened)
}

func (g *desktopGUI) revealFile(path string) {
	dir := filepath.Dir(path)
	if err := g.service.revealInFileManager(dir); err != nil {
		g.showError(fmt.Errorf("%s: %w", g.msgs.GUIErrOpenFolder, err))
		return
	}
	g.setToast(g.msgs.GUIToastOpenedFolder)
}

func (g *desktopGUI) promptRenameFile(idx int) {
	file := g.files[idx]
	entry := widget.NewEntry()
	entry.SetText(file.Name)
	g.showInputDialog(g.msgs.GUIRenameFileTitle, entry, func(value string) {
		if _, err := g.service.renameEntry(file.Path, value); err != nil {
			g.showError(err)
			return
		}
		g.refreshFilesUI()
		g.setToast(g.msgs.GUIToastRenamed)
	})
}

func (g *desktopGUI) confirmDeleteSingleFile(idx int) {
	file := g.files[idx]
	title := fmt.Sprintf(g.msgs.DeleteTitle, 1)
	body := widget.NewLabel(filepath.Base(file.Path))
	d := dialog.NewCustomConfirm(title, g.msgs.DeleteOptionYes, g.msgs.DeleteOptionNo, body, func(ok bool) {
		if !ok {
			return
		}
		if err := g.deletePaths([]string{file.Path}); err != nil {
			g.showError(err)
			return
		}
		g.setToast(g.msgs.GUIToastDeleted)
	}, g.window)
	d.Show()
}
