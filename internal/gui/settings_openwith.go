package gui

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/loveloki/try/internal/config"
)

// normalizeExtension 规范化扩展名输入：trim + 小写，须为 "." 后跟字母或数字（与 filepath.Ext 的产出格式一致）。
func normalizeExtension(raw string) (string, bool) {
	ext := strings.ToLower(strings.TrimSpace(raw))
	if len(ext) <= 1 || ext[0] != '.' {
		return "", false
	}
	for _, r := range ext[1:] {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return "", false
		}
	}
	return ext, true
}

// sortedOpenWithExts 按字典序返回扩展名 key，保证映射列表顺序稳定。
func sortedOpenWithExts(openWith map[string]string) []string {
	exts := make([]string, 0, len(openWith))
	for ext := range openWith {
		exts = append(exts, ext)
	}
	sort.Strings(exts)
	return exts
}

// appNameOptions 返回内置应用的名称列表，作为 SelectEntry 的候选。
func appNameOptions() []string {
	apps := builtinApps()
	names := make([]string, 0, len(apps))
	for _, a := range apps {
		names = append(names, a.Name)
	}
	return names
}

func (g *desktopGUI) buildOpenWithCard() fyne.CanvasObject {
	listBox := container.NewVBox()
	g.rebuildOpenWithList(listBox)
	body := container.NewVBox(listBox, g.buildOpenWithAddRow(listBox))
	return widget.NewCard(g.msgs.GUISettingsOpenWith, "", body)
}

// rebuildOpenWithList 按当前 g.cfg.OpenWith 重建映射列表区域。
func (g *desktopGUI) rebuildOpenWithList(listBox *fyne.Container) {
	listBox.Objects = nil
	if len(g.cfg.OpenWith) == 0 {
		listBox.Add(widget.NewLabel(g.msgs.GUISettingsNoMappings))
	}
	for _, ext := range sortedOpenWithExts(g.cfg.OpenWith) {
		listBox.Add(g.buildOpenWithRow(listBox, ext, g.cfg.OpenWith[ext]))
	}
	listBox.Refresh()
}

func (g *desktopGUI) buildOpenWithRow(listBox *fyne.Container, ext, appName string) fyne.CanvasObject {
	del := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		g.removeOpenWith(listBox, ext)
	})
	del.Importance = widget.LowImportance
	return container.NewBorder(nil, nil, widget.NewLabel(ext), del, widget.NewLabel(appName))
}

func (g *desktopGUI) buildOpenWithAddRow(listBox *fyne.Container) fyne.CanvasObject {
	extEntry := widget.NewEntry()
	extEntry.SetPlaceHolder(g.msgs.GUISettingsExtPlaceholder)
	// 候选为内置应用名，也允许手动输入其他应用名。
	appEntry := widget.NewSelectEntry(appNameOptions())
	appEntry.SetPlaceHolder(g.msgs.GUISettingsApplication)
	add := widget.NewButton(g.msgs.GUISettingsAdd, func() {
		g.addOpenWith(listBox, extEntry, appEntry)
	})
	return container.NewBorder(nil, nil, extEntry, add, appEntry)
}

func (g *desktopGUI) addOpenWith(listBox *fyne.Container, extEntry *widget.Entry, appEntry *widget.SelectEntry) {
	ext, ok := normalizeExtension(extEntry.Text)
	if !ok {
		g.showError(errors.New(g.msgs.GUISettingsInvalidExt))
		return
	}
	appName := strings.TrimSpace(appEntry.Text)
	if appName == "" {
		return
	}
	// 保存成功后清空输入，失败时保留用户输入。
	if g.saveOpenWith(listBox, func(m map[string]string) { m[ext] = appName }) {
		extEntry.SetText("")
		appEntry.SetText("")
	}
}

func (g *desktopGUI) removeOpenWith(listBox *fyne.Container, ext string) {
	g.saveOpenWith(listBox, func(m map[string]string) { delete(m, ext) })
}

// saveOpenWith 在映射副本上变更，持久化成功后才替换内存状态并重建列表，返回是否成功。
func (g *desktopGUI) saveOpenWith(listBox *fyne.Container, mutate func(map[string]string)) bool {
	next := make(map[string]string, len(g.cfg.OpenWith)+1)
	for k, v := range g.cfg.OpenWith {
		next[k] = v
	}
	mutate(next)
	if err := config.SaveOpenWith(next); err != nil {
		g.showError(fmt.Errorf("%s: %w", g.msgs.GUISettingsErrSave, err))
		return false
	}
	g.cfg.OpenWith = next
	g.rebuildOpenWithList(listBox)
	return true
}
