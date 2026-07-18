package gui

func (g *desktopGUI) toggleFileMarkAt(idx int) {
	if idx < 0 || idx >= len(g.files) {
		return
	}
	path := g.files[idx].Path
	if g.fileMarked[path] {
		delete(g.fileMarked, path)
	} else {
		g.fileMarked[path] = true
	}
	g.fileSelected = idx
	if g.fileList != nil {
		g.fileList.Select(idx)
		g.fileList.Refresh()
	}
	g.updateFilesStatus()
}

func (g *desktopGUI) setFileMarkedAt(idx int, marked bool) {
	if idx < 0 || idx >= len(g.files) {
		return
	}
	path := g.files[idx].Path
	if marked {
		g.fileMarked[path] = true
	} else {
		delete(g.fileMarked, path)
	}
	g.fileSelected = idx
	if g.fileList != nil {
		g.fileList.Select(idx)
		g.fileList.Refresh()
	}
	g.updateFilesStatus()
}
