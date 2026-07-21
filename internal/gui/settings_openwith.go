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
	xwidget "fyne.io/x/fyne/widget"
	"github.com/loveloki/try/internal/config"
)

// settingsExtWidth 添加行扩展名输入框固定宽度。
const settingsExtWidth float32 = 96

// normalizeExtension 规范化扩展名输入：trim + 小写，缺省补 "." 前缀，
// 最终须为 "." 后跟字母或数字（与 filepath.Ext 的产出格式一致）；
// 特殊值 "*" 表示通用映射（所有文件）。
func normalizeExtension(raw string) (string, bool) {
	ext := strings.ToLower(strings.TrimSpace(raw))
	if ext == "*" {
		return ext, true
	}
	if ext == "" {
		return "", false
	}
	if ext[0] != '.' {
		ext = "." + ext
	}
	if len(ext) <= 1 {
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

// normalizeAppName 规范化应用名输入：trim 后拒绝空值与路径遍历片段。
// 允许三种形式：应用名（如 zed）、macOS bundle 名（如 Zed.app）、绝对路径。
func normalizeAppName(raw string) (string, bool) {
	name := strings.TrimSpace(raw)
	if name == "" || name == "." || name == ".." || strings.Contains(name, "..") {
		return "", false
	}
	return name, true
}

// looksLikePath 判断输入是否为路径形式（含路径分隔符）。
func looksLikePath(s string) bool {
	return strings.ContainsRune(s, '/') || strings.ContainsRune(s, '\\')
}

// appNameOptions 返回应用选择器候选：内置应用 + 已安装应用（去重）。
// macOS 候选来自 /Applications 等目录的 .app 扫描，其他平台来自 PATH 可执行名。
func appNameOptions() []string {
	apps := builtinApps()
	names := make([]string, 0, len(apps))
	seen := make(map[string]bool, len(apps))
	for _, a := range apps {
		names = append(names, a.Name)
		seen[strings.ToLower(a.Name)] = true
	}
	for _, e := range installedAppNames() {
		if looksLikePath(e) || seen[strings.ToLower(e)] {
			continue
		}
		seen[strings.ToLower(e)] = true
		names = append(names, e)
	}
	return names
}

func (g *desktopGUI) buildOpenWithSection() fyne.CanvasObject {
	listBox := container.NewVBox()
	g.rebuildOpenWithList(listBox)
	content := container.NewVBox(listBox, widget.NewSeparator(), g.buildOpenWithAddRow(listBox))
	return settingsSection(g.msgs.GUISettingsOpenWith, content)
}

// rebuildOpenWithList 按当前 g.cfg.OpenWith 重建映射列表区域。
func (g *desktopGUI) rebuildOpenWithList(listBox *fyne.Container) {
	listBox.Objects = nil
	if len(g.cfg.OpenWith) == 0 {
		empty := widget.NewLabel(g.msgs.GUISettingsNoMappings)
		empty.Alignment = fyne.TextAlignCenter
		listBox.Add(empty)
	}
	exts := sortedOpenWithExts(g.cfg.OpenWith)
	for i, ext := range exts {
		if i > 0 {
			listBox.Add(widget.NewSeparator())
		}
		listBox.Add(g.buildOpenWithRow(listBox, ext, g.cfg.OpenWith[ext]))
	}
	listBox.Refresh()
}

func (g *desktopGUI) buildOpenWithRow(listBox *fyne.Container, ext, appName string) fyne.CanvasObject {
	del := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		g.removeOpenWith(listBox, ext)
	})
	del.Importance = widget.LowImportance
	row := container.NewBorder(nil, nil, fixedWidth(widget.NewLabel(ext), settingsExtWidth), del,
		widget.NewLabel(appName))
	return container.New(&fixedHeightLayout{height: settingsRowHeight}, row)
}

// filterAppOptions 按大小写不敏感的子串过滤候选；空查询返回全部。
func filterAppOptions(options []string, query string) []string {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return options
	}
	out := make([]string, 0, len(options))
	for _, o := range options {
		if strings.Contains(strings.ToLower(o), q) {
			out = append(out, o)
		}
	}
	return out
}

func (g *desktopGUI) buildOpenWithAddRow(listBox *fyne.Container) fyne.CanvasObject {
	extEntry := widget.NewEntry()
	extEntry.SetPlaceHolder(g.msgs.GUISettingsExtPlaceholder)
	// fyne-x CompletionEntry：弹层由使用方在 OnChanged 中驱动——
	// 输入即过滤候选并弹出可滚动列表，↑↓ 导航、Enter 选中。
	options := appNameOptions()
	appEntry := xwidget.NewCompletionEntry(options)
	appEntry.SetPlaceHolder(g.msgs.GUISettingsApplication)
	appEntry.OnChanged = func(s string) {
		filtered := filterAppOptions(options, s)
		appEntry.SetOptions(filtered)
		if strings.TrimSpace(s) == "" || len(filtered) == 0 {
			appEntry.HideCompletion()
			return
		}
		appEntry.ShowCompletion()
	}
	add := widget.NewButton(g.msgs.GUISettingsAdd, func() {
		g.addOpenWith(listBox, extEntry, appEntry)
	})
	add.Importance = widget.HighImportance
	row := container.NewBorder(nil, nil, fixedWidth(extEntry, settingsExtWidth), add, appEntry)
	return container.New(&fixedHeightLayout{height: settingsRowHeight}, row)
}

func (g *desktopGUI) addOpenWith(listBox *fyne.Container, extEntry *widget.Entry, appEntry *xwidget.CompletionEntry) {
	ext, ok := normalizeExtension(extEntry.Text)
	if !ok {
		g.showError(errors.New(g.msgs.GUISettingsInvalidExt))
		return
	}
	appName, ok := normalizeAppName(appEntry.Text)
	if !ok {
		g.showError(errors.New(g.msgs.GUISettingsInvalidApp))
		return
	}
	// 保存成功后清空输入并 Toast 反馈，失败时保留用户输入。
	if g.saveOpenWith(listBox, func(m map[string]string) { m[ext] = appName }) {
		extEntry.SetText("")
		appEntry.SetText("")
		g.setToast(fmt.Sprintf(g.msgs.GUISettingsMappingAdded, ext, appName))
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
