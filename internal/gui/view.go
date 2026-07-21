package gui

import "github.com/loveloki/try/internal/config"

func (g *desktopGUI) setupWindow() {
	g.selectorBody = g.buildSelectorBody()
	g.filesBody = g.buildFilesBody()
	g.settingsBody = g.buildSettingsBody()
	g.switchToSelector(true)
	g.bindKeys()
	g.bindDrop()
}

func (g *desktopGUI) switchToSelector(focusSearch bool) {
	g.view = "selector"
	g.list = g.entryList
	g.setWindowContent(g.selectorBody)
	g.refreshSelectorUI()
	g.syncFilesWatch()
	if focusSearch {
		g.focusSearch()
	} else if g.list != nil {
		g.window.Canvas().Focus(g.list)
	}
}

// switchToSettings 进入全屏设置页，记住返回视图。
func (g *desktopGUI) switchToSettings() {
	g.applyOpenSettings()
	g.list = nil
	g.setWindowContent(g.settingsBody)
	g.syncFilesWatch()
}

// closeSettings 关闭设置页，返回进入前的视图。
func (g *desktopGUI) closeSettings() {
	if g.view != "settings" {
		return
	}
	back := g.applyCloseSettings()
	if back == "files" {
		g.switchToFiles()
		return
	}
	g.switchToSelector(false)
}

// applyOpenSettings 设置设置页视图状态（不含 UI 切换，便于单测）。
func (g *desktopGUI) applyOpenSettings() {
	if g.view != "settings" {
		g.settingsReturn = g.view
	}
	g.view = "settings"
}

// applyCloseSettings 恢复进入设置页前的视图并返回其名称。
func (g *desktopGUI) applyCloseSettings() string {
	back := g.settingsReturn
	if back == "" {
		back = "selector"
	}
	g.view = back
	return back
}

func (g *desktopGUI) switchToFiles() {
	g.view = "files"
	g.list = g.fileList
	g.setWindowContent(g.filesBody)
	g.refreshFilesUI()
	g.syncFilesWatch()
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
		if g.fileSelected >= 0 && g.fileSelected < len(g.files) {
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
	_ = config.SaveTheme(g.themeName)
	g.applyThemeAndRedraw()
}

func (g *desktopGUI) applyThemeAndRedraw() {
	g.applyTheme()
	// 主题切换后重建 body，使 header/status 绑定新 token。
	g.selectorBody = g.buildSelectorBody()
	g.filesBody = g.buildFilesBody()
	g.settingsBody = g.buildSettingsBody()
	switch g.view {
	case "files":
		g.switchToFiles()
	case "settings":
		// 停留在设置页，文案与主题以新值重建。
		g.view = g.settingsReturn
		g.switchToSettings()
	default:
		g.switchToSelector(false)
	}
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
	prevPath := ""
	if g.fileSelected >= 0 && g.fileSelected < len(g.files) {
		prevPath = g.files[g.fileSelected].Path
	}
	files, err := g.service.listFiles(g.filesPath)
	if err != nil {
		g.showError(err)
		return
	}
	g.files = files
	if prevPath != "" {
		g.fileSelected = indexOfFilePath(files, prevPath)
		return
	}
	g.fileSelected = clampFileSelected(g.fileSelected, len(g.files))
}

func indexOfFilePath(files []FileEntry, path string) int {
	for i, f := range files {
		if f.Path == path {
			return i
		}
	}
	return -1
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

// clampFileSelected 将光标夹到合法范围；空列表或原无选中时为 -1。
func clampFileSelected(selected, length int) int {
	if length <= 0 || selected < 0 {
		return -1
	}
	if selected >= length {
		return length - 1
	}
	return selected
}
