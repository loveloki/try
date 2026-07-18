package gui

import (
	"fmt"
)

func (g *desktopGUI) setSelectorSelected(idx int) {
	if idx < 0 || idx >= len(g.entries) {
		return
	}
	g.selected = idx
	g.selectedPath = g.entries[idx].Path
	if g.list != nil {
		g.list.Select(g.selected)
		g.list.Refresh()
	}
}

func (g *desktopGUI) setFileSelected(idx int) {
	if idx < 0 || idx >= len(g.files) {
		return
	}
	g.fileSelected = idx
	if g.fileList != nil {
		g.fileList.Select(g.fileSelected)
		g.fileList.Refresh()
	} else if g.list != nil {
		g.list.Select(g.fileSelected)
		g.list.Refresh()
	}
}

func (g *desktopGUI) openEntryAt(idx int) {
	g.setSelectorSelected(idx)
	g.openSelected()
}

func (g *desktopGUI) openFileAt(idx int) {
	g.setFileSelected(idx)
	g.openSelectedFile()
}

func (g *desktopGUI) revealCurrentFolder() {
	if g.view != "files" || g.filesPath == "" {
		return
	}
	if err := g.service.revealInFileManager(g.filesPath); err != nil {
		g.showError(fmt.Errorf("%s: %w", g.msgs.GUIErrOpenFolder, err))
		return
	}
	g.setToast(g.msgs.GUIToastOpenedFolder)
}
