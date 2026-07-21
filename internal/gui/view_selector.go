package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *desktopGUI) buildSelectorBody() fyne.CanvasObject {
	g.search = newSearchEntry(g.cycleSource, g.moveSelection, g.submitSearch, g.dispatchAppShortcut)
	g.search.SetPlaceHolder(g.msgs.GUISearchPlace)
	g.search.SetText(g.query)
	g.search.OnChanged = func(q string) {
		g.query = q
		g.refreshSelectorUI()
	}
	g.search.OnSubmitted = func(string) { g.submitSearch() }

	g.sourceTabsBox = container.NewHBox()
	g.rebuildSourceTabs()

	header := buildHeader(g.msgs.Title, g.headerButtons())
	top := container.NewVBox(withHInset(header), withHInset(g.search), withHInset(g.sourceTabsBox))
	g.entryList = g.buildEntryList()
	g.list = g.entryList
	g.selectorStatus = newStatusBar()
	g.updateSelectorStatus()
	return container.NewBorder(top, g.selectorStatus, nil, nil, g.entryList)
}

func (g *desktopGUI) rebuildSourceTabs() {
	if g.sourceTabsBox == nil {
		return
	}
	g.sourceTabsBox.Objects = nil
	for _, src := range g.sources {
		source := src
		label := formatSourceTabLabel(g.msgs.FilterAll, source)
		count := g.counts[source]
		active := source == g.source
		g.sourceTabsBox.Add(newSourceTab(label, count, active, func() {
			g.applySource(source)
		}))
	}
	g.sourceTabsBox.Refresh()
}

func (g *desktopGUI) buildEntryList() *navList {
	list := newNavList(
		func() int { return len(g.entries) },
		func() fyne.CanvasObject { return newEntryRowWidget() },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			g.paintEntryRow(id, obj.(*entryRowWidget))
		},
		g.handleNavKey,
	)
	list.HideSeparators = true
	list.OnSelected = func(id widget.ListItemID) { g.selected = int(id) }
	if len(g.entries) > 0 {
		list.Select(g.selected)
	}
	return list
}

func (g *desktopGUI) paintEntryRow(id widget.ListItemID, row *entryRowWidget) {
	if id < 0 || int(id) >= len(g.entries) {
		return
	}
	e := g.entries[id]
	idx := int(id)
	marked := g.marked[e.Path]
	selected := idx == g.selected

	row.onSelect = func() { g.setSelectorSelected(idx) }
	row.onOpen = func() { g.openEntryAt(idx) }
	row.setVisualState(marked, selected)

	if marked {
		row.arrow.Text = "✕"
		row.arrow.Color = theme.Color(theme.ColorNameError)
	} else if selected {
		row.arrow.Text = "›"
		row.arrow.Color = theme.Color(colorNameHeader)
	} else {
		row.arrow.Text = " "
	}

	row.name.Segments = entryNameSegments(e, marked, selected)
	row.meta.Text = entryMetaText(e)
	if marked {
		row.meta.Color = theme.Color(theme.ColorNameError)
	} else {
		row.meta.Color = theme.Color(colorNameMuted)
	}
	canvas.Refresh(row.arrow)
	canvas.Refresh(row.meta)
	row.name.Refresh()
}

func (g *desktopGUI) updateSelectorStatus() {
	if g.selectorStatus == nil {
		return
	}
	left, hints := g.selectorStatusContent()
	g.selectorStatus.SetContent(left, hints)
}

func (g *desktopGUI) selectorStatusContent() (string, []shortcutHint) {
	if len(g.marked) > 0 {
		left := fmt.Sprintf(g.msgs.DeleteMode, len(g.marked))
		return left, []shortcutHint{
			{Key: "Esc", Label: g.msgs.GUICancel},
			{Key: "Enter", Label: g.msgs.GUIConfirm},
		}
	}
	left := fmt.Sprintf(g.msgs.ItemCount, len(g.entries))
	return left, []shortcutHint{
		{Key: "↑↓", Label: g.msgs.GUIShortcutNav},
		{Key: "Enter", Label: g.msgs.GUIShortcutOpen},
		{Key: "⌃T", Label: g.msgs.GUIShortcutNew},
		{Key: "⌃D", Label: g.msgs.GUIShortcutDelete},
		{Key: ",", Label: g.msgs.GUISettingsTitle},
	}
}

func (g *desktopGUI) themeButton() fyne.CanvasObject {
	btn := widget.NewButtonWithIcon("", theme.ColorPaletteIcon(), g.toggleTheme)
	btn.Importance = widget.LowImportance
	return btn
}

func (g *desktopGUI) settingsButton() fyne.CanvasObject {
	btn := widget.NewButtonWithIcon("", theme.SettingsIcon(), g.promptSettings)
	btn.Importance = widget.LowImportance
	return btn
}

// headerButtons 标题行右侧按钮组：主题切换 + 设置。
func (g *desktopGUI) headerButtons() fyne.CanvasObject {
	return container.NewHBox(g.themeButton(), g.settingsButton())
}
