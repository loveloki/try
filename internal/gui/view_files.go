package gui

import (
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *desktopGUI) buildFilesBody() fyne.CanvasObject {
	g.filesTitle = widget.NewLabel(g.filesPathLabel())
	g.filesTitle.TextStyle = fyne.TextStyle{Bold: true}
	g.breadcrumbBox = container.NewHBox()
	g.rebuildBreadcrumbs()

	header := buildHeader(g.filesPathLabel(), g.themeButton())
	toolbar := g.buildFilesToolbar()
	colHeader := g.buildFilesColumnHeader()
	g.fileList = g.buildFileList()
	g.filesStatus = newStatusBar()
	g.updateFilesStatus()

	g.dropOverlay = g.buildDropOverlay()
	g.dropOverlay.Hide()

	listArea := container.NewBorder(withHInset(colHeader), nil, nil, nil, g.fileList)
	body := container.NewBorder(
		container.NewVBox(withHInset(header), withHInset(toolbar), withHInset(g.breadcrumbBox)),
		g.filesStatus,
		nil, nil,
		listArea,
	)
	return container.NewStack(body, g.dropOverlay)
}

func (g *desktopGUI) buildFilesToolbar() fyne.CanvasObject {
	back := widget.NewButtonWithIcon(g.msgs.GUIBack, theme.NavigateBackIcon(), g.handleFilesBack)
	back.Importance = widget.LowImportance
	edit := widget.NewButtonWithIcon(g.msgs.GUIEdit, theme.DocumentCreateIcon(), g.openSelectedFile)
	edit.Importance = widget.LowImportance
	pack := widget.NewButtonWithIcon(g.msgs.GUIDocxPack, theme.FolderIcon(), func() {
		g.setToast(g.msgs.GUINotImplemented)
	})
	pack.Importance = widget.LowImportance
	unpack := widget.NewButtonWithIcon(g.msgs.GUIDocxUnpack, theme.FolderOpenIcon(), func() {
		g.setToast(g.msgs.GUINotImplemented)
	})
	unpack.Importance = widget.LowImportance
	reveal := widget.NewButtonWithIcon(g.msgs.GUIOpenFolder, theme.FolderOpenIcon(), g.revealCurrentFolder)
	reveal.Importance = widget.LowImportance
	del := widget.NewButtonWithIcon(g.msgs.GUIDelete, theme.DeleteIcon(), g.confirmDelete)
	del.Importance = widget.LowImportance

	sep1 := canvas.NewRectangle(theme.Color(theme.ColorNameSeparator))
	sep1.SetMinSize(fyne.NewSize(1, 16))
	sep2 := canvas.NewRectangle(theme.Color(theme.ColorNameSeparator))
	sep2.SetMinSize(fyne.NewSize(1, 16))

	return container.NewHBox(back, sep1, edit, pack, unpack, reveal, sep2, del)
}

func (g *desktopGUI) buildFilesColumnHeader() fyne.CanvasObject {
	name := canvas.NewText(g.msgs.GUIColName, theme.Color(colorNameMuted))
	size := canvas.NewText(g.msgs.GUIColSize, theme.Color(colorNameMuted))
	mtime := canvas.NewText(g.msgs.GUIColMTime, theme.Color(colorNameMuted))
	for _, t := range []*canvas.Text{name, size, mtime} {
		t.TextSize = theme.TextSize() * 0.7
		t.TextStyle = fyne.TextStyle{Monospace: true}
	}
	line := canvas.NewRectangle(theme.Color(theme.ColorNameSeparator))
	line.SetMinSize(fyne.NewSize(1, 1))
	row := container.NewBorder(nil, nil, nil, container.NewHBox(size, mtime), name)
	return container.NewVBox(container.NewPadded(row), line)
}

func (g *desktopGUI) buildDropOverlay() fyne.CanvasObject {
	g.dropOverlayLabel = canvas.NewText(g.msgs.GUIDropImporting, theme.Color(colorNameHeader))
	g.dropOverlayLabel.TextSize = theme.TextSize()
	g.dropOverlayLabel.TextStyle = fyne.TextStyle{Bold: true}
	g.dropOverlayLabel.Alignment = fyne.TextAlignCenter
	g.dropOverlayProgress = widget.NewProgressBar()
	g.dropOverlayProgress.Min = 0
	g.dropOverlayProgress.Max = 1
	box := canvas.NewRectangle(theme.Color(colorNameHighlightDim))
	box.StrokeColor = theme.Color(colorNameHeader)
	box.StrokeWidth = 2
	box.CornerRadius = 8
	inner := container.NewCenter(container.NewVBox(g.dropOverlayLabel, g.dropOverlayProgress))
	return container.NewStack(box, container.NewPadded(inner))
}

func (g *desktopGUI) setDropOverlayVisible(visible bool) {
	if g.dropOverlay == nil {
		return
	}
	if visible {
		g.dropOverlay.Show()
	} else {
		g.dropOverlay.Hide()
	}
	g.dropOverlay.Refresh()
}

func (g *desktopGUI) updateDropProgress(done, total int, current string) {
	label := formatDropProgressLabel(g.msgs, done, total, current)
	if g.dropOverlayLabel != nil {
		g.dropOverlayLabel.Text = label
		canvas.Refresh(g.dropOverlayLabel)
	}
	if g.dropOverlayProgress != nil && total > 0 {
		g.dropOverlayProgress.SetValue(float64(done) / float64(total))
	}
	g.setPersistentToast(label)
}

func (g *desktopGUI) filesPathLabel() string {
	base := filepath.Base(g.filesRoot)
	if g.filesRoot == "" {
		base = g.msgs.GUIFilesTitle
	}
	rel, err := filepath.Rel(g.filesRoot, g.filesPath)
	if err != nil || rel == "." || rel == "" {
		return base
	}
	return filepath.Join(base, rel)
}

func (g *desktopGUI) updateFilesStatus() {
	if g.filesStatus == nil {
		return
	}
	left, hints := g.filesStatusContent()
	g.filesStatus.SetContent(left, hints)
}

func (g *desktopGUI) filesStatusContent() (string, []shortcutHint) {
	left := fmt.Sprintf(g.msgs.GUIFilesItemCount, len(g.files))
	if len(g.files) == 0 {
		left = g.msgs.GUIFilesEmpty
	}
	if len(g.fileMarked) > 0 {
		left += "  ·  " + fmt.Sprintf(g.msgs.MarkedCount, len(g.fileMarked))
	}
	return left, []shortcutHint{
		{Key: "↑↓", Label: g.msgs.GUIShortcutNav},
		{Key: "Esc", Label: g.msgs.GUIShortcutEscBack},
		{Key: "", Label: g.msgs.GUIShortcutDrop, Accent: true},
	}
}

func (g *desktopGUI) buildFileList() *navList {
	list := newNavList(
		func() int { return len(g.files) },
		func() fyne.CanvasObject { return newFileRowWidget() },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			g.paintFileRow(id, obj.(*fileRowWidget))
		},
		g.handleNavKey,
	)
	list.HideSeparators = true
	list.OnSelected = func(id widget.ListItemID) { g.fileSelected = int(id) }
	if len(g.files) > 0 {
		list.Select(g.fileSelected)
	}
	return list
}

func (g *desktopGUI) paintFileRow(id widget.ListItemID, row *fileRowWidget) {
	if id < 0 || int(id) >= len(g.files) {
		return
	}
	f := g.files[id]
	idx := int(id)
	marked := g.fileMarked[f.Path]
	selected := idx == g.fileSelected

	row.onSelect = func() { g.setFileSelected(idx) }
	row.onCtrlSelect = func() { g.toggleFileMarkAt(idx) }
	row.onOpen = func() { g.openFileAt(idx) }
	row.setVisualState(marked, selected)

	row.check.OnChanged = nil
	row.check.SetChecked(marked)
	row.check.OnChanged = func(on bool) { g.setFileMarkedAt(idx, on) }

	row.icon.SetResource(fileTypeIcon(f.Type, f.IsDir))
	row.name.Text = fileDisplayName(f)
	if marked {
		row.name.Color = theme.Color(theme.ColorNameError)
	} else {
		row.name.Color = theme.Color(theme.ColorNameForeground)
	}
	row.meta.Text = fileMetaText(f)
	canvas.Refresh(row.name)
	canvas.Refresh(row.meta)
}

func fileDisplayName(f FileEntry) string {
	if f.IsDir {
		return f.Name + string(filepath.Separator)
	}
	return f.Name
}

func (g *desktopGUI) rebuildBreadcrumbs() {
	if g.breadcrumbBox == nil {
		return
	}
	g.breadcrumbBox.Objects = nil
	for _, item := range g.breadcrumbItems() {
		g.breadcrumbBox.Add(item)
	}
	g.breadcrumbBox.Refresh()
}

func (g *desktopGUI) breadcrumbItems() []fyne.CanvasObject {
	rootBase := filepath.Base(g.filesRoot)
	if g.filesRoot == "" {
		return []fyne.CanvasObject{widget.NewLabel(g.msgs.GUIFilesTitle)}
	}
	rel, err := filepath.Rel(g.filesRoot, g.filesPath)
	if err != nil || rel == "." || rel == "" {
		return []fyne.CanvasObject{widget.NewButton(rootBase, func() {
			g.enterFiles(g.filesRoot, g.filesRoot)
		})}
	}
	parts := strings.Split(rel, string(filepath.Separator))
	items := []fyne.CanvasObject{widget.NewButton(rootBase, func() {
		g.enterFiles(g.filesRoot, g.filesRoot)
	})}
	path := g.filesRoot
	for _, part := range parts {
		if part == "" {
			continue
		}
		part := part
		path = filepath.Join(path, part)
		target := path
		items = append(items, widget.NewLabel("/"))
		items = append(items, widget.NewButton(part, func() {
			g.enterFiles(g.filesRoot, target)
		}))
	}
	return items
}
