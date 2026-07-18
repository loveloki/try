package gui

func (g *desktopGUI) setupWindow() {
	g.selectorBody = g.buildSelectorBody()
	g.filesBody = g.buildFilesBody()
	g.switchToSelector(true)
	g.bindKeys()
	g.bindDrop()
}

func (g *desktopGUI) switchToSelector(focusSearch bool) {
	g.view = "selector"
	g.list = g.entryList
	g.setWindowContent(g.selectorBody)
	g.refreshSelectorUI()
	if focusSearch {
		g.focusSearch()
	} else if g.list != nil {
		g.window.Canvas().Focus(g.list)
	}
}

func (g *desktopGUI) switchToFiles() {
	g.view = "files"
	g.list = g.fileList
	g.setWindowContent(g.filesBody)
	g.refreshFilesUI()
	if g.list != nil {
		g.window.Canvas().Focus(g.list)
	}
}

func (g *desktopGUI) refreshSelectorUI() {
	g.refreshEntries()
	if g.search != nil && g.search.Text != g.query {
		g.search.SetText(g.query)
	}
	g.rebuildSourceTabs()
	if g.list != nil {
		g.list.Refresh()
		if len(g.entries) > 0 {
			g.list.Select(g.selected)
			g.list.ScrollTo(g.selected)
		} else {
			g.list.UnselectAll()
		}
	}
	g.updateSelectorStatus()
}

func (g *desktopGUI) refreshFilesUI() {
	g.refreshFiles()
	if g.filesTitle != nil {
		g.filesTitle.SetText(g.filesPathLabel())
	}
	g.rebuildBreadcrumbs()
	if g.list != nil {
		g.list.Refresh()
		if len(g.files) > 0 {
			g.list.Select(g.fileSelected)
			g.list.ScrollTo(g.fileSelected)
		} else {
			g.list.UnselectAll()
		}
	}
	g.updateFilesStatus()
}

func (g *desktopGUI) toggleTheme() {
	if g.themeName == "light" {
		g.themeName = "dark"
	} else {
		g.themeName = "light"
	}
	g.applyThemeAndRedraw()
}

func (g *desktopGUI) applyThemeAndRedraw() {
	g.applyTheme()
	// 主题切换后重建 body，使 header/status 绑定新 token。
	g.selectorBody = g.buildSelectorBody()
	g.filesBody = g.buildFilesBody()
	if g.view == "files" {
		g.switchToFiles()
		return
	}
	g.switchToSelector(false)
}

func (g *desktopGUI) cycleSource(delta int) {
	if g.view != "selector" || len(g.sources) == 0 {
		return
	}
	next := cycleSource(g.sources, g.source, delta)
	g.applySource(next)
}

func (g *desktopGUI) applySource(source string) {
	if source == g.source {
		return
	}
	g.source = source
	g.clearSelectorSelection()
	g.refreshSelectorUI()
}

func (g *desktopGUI) clearSelectorSelection() {
	g.selected = 0
	g.selectedPath = ""
}

func (g *desktopGUI) refreshEntries() {
	result := g.service.listEntries(g.query, g.source)
	g.entries = result.Entries
	g.counts = result.Counts
	g.sources = result.Sources
	g.selected, g.selectedPath = resolveSelectorSelection(g.entries, g.selectedPath, g.selected)
}

func (g *desktopGUI) refreshFiles() {
	files, err := g.service.listFiles(g.filesPath)
	if err != nil {
		g.showError(err)
		return
	}
	g.files = files
	if g.fileSelected >= len(g.files) {
		g.fileSelected = max(0, len(g.files)-1)
	}
}

func (g *desktopGUI) enterFiles(root, path string) {
	g.applyFilesNav(root, path)
	g.switchToFiles()
}

// applyFilesNav 设置文件视图导航状态（不含 UI 切换，便于单测）。
func (g *desktopGUI) applyFilesNav(root, path string) {
	if path == root && root != "" && g.service != nil {
		g.service.touchDir(root)
		g.selectedPath = root
	}
	g.filesRoot = root
	g.filesPath = path
	g.fileSelected = 0
	g.fileMarked = map[string]bool{}
	g.view = "files"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
