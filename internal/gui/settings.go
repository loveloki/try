package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/loveloki/try/internal/config"
	"github.com/loveloki/try/internal/i18n"
)

// 主题/语言选项的存储值，顺序固定，与显示文案按索引一一对应。
var themeValues = []string{"dark", "light", "auto"}
var localeValues = []string{"en", "zh", "auto"}

const (
	// settingsMaxWidth 设置页内容最大宽度，宽窗口下居中留白。
	settingsMaxWidth float32 = 680
	// settingsSectionGap 设置页分组之间的纵向间距。
	settingsSectionGap float32 = 20
	// settingsSelectWidth 外观下拉框固定宽度。
	settingsSelectWidth float32 = 220
	// settingsRowHeight 设置项行高。
	settingsRowHeight float32 = 36
)

// themeLabels 返回主题选项的本地化文案（顺序与 themeValues 一致）。
func themeLabels(msgs *i18n.Messages) []string {
	return []string{msgs.GUISettingsThemeDark, msgs.GUISettingsThemeLight, msgs.GUISettingsThemeAuto}
}

// localeLabels 返回语言选项文案；语言名使用自身书写（English/中文），无需本地化。
func localeLabels(msgs *i18n.Messages) []string {
	return []string{"English", "中文", msgs.GUISettingsLangAuto}
}

// valueIndex 返回 value 在 values 中的索引；未知值回退到末项（auto）。
func valueIndex(values []string, value string) int {
	for i, v := range values {
		if v == value {
			return i
		}
	}
	return len(values) - 1
}

// optionValueAt 按索引取存储值；越界回退到末项。
func optionValueAt(values []string, index int) string {
	if index < 0 || index >= len(values) {
		return values[len(values)-1]
	}
	return values[index]
}

// labelIndex 按显示文案反查索引，未找到返回 -1。
func labelIndex(labels []string, label string) int {
	for i, l := range labels {
		if l == label {
			return i
		}
	}
	return -1
}

// promptSettings 打开全屏设置页；已在设置页时关闭（快捷键切换语义）。
// 所有修改即时生效并持久化，无保存/取消按钮。
func (g *desktopGUI) promptSettings() {
	if g.view == "settings" {
		g.closeSettings()
		return
	}
	g.switchToSettings()
}

// buildSettingsBody 构建设置页：头部（返回按钮）+ 分组内容 + 底栏快捷键提示。
func (g *desktopGUI) buildSettingsBody() fyne.CanvasObject {
	back := widget.NewButtonWithIcon(g.msgs.GUISettingsBack, theme.NavigateBackIcon(), g.closeSettings)
	back.Importance = widget.LowImportance
	header := buildHeader(g.msgs.GUISettingsTitle, back)

	sections := container.NewVBox(
		g.buildAppearanceSection(),
		g.buildOpenWithSection(),
	)
	sections.Layout = layout.NewCustomPaddedVBoxLayout(settingsSectionGap)
	body := container.NewPadded(container.New(&maxWidthLayout{width: settingsMaxWidth}, sections))
	scroll := container.NewVScroll(body)

	status := newStatusBar()
	status.SetContent("", []shortcutHint{{Key: "Esc", Label: g.msgs.GUISettingsBack}})
	return container.NewBorder(header, status, nil, nil, scroll)
}

// settingsSection 设置页分组：小标题 + 圆角面板包裹的内容。
func settingsSection(title string, content fyne.CanvasObject) fyne.CanvasObject {
	label := canvas.NewText(title, theme.Color(colorNameMuted))
	label.TextSize = theme.TextSize() - 1
	bg := canvas.NewRectangle(theme.Color(theme.ColorNameButton))
	bg.CornerRadius = theme.Size(theme.SizeNameInputRadius) * 2
	panel := container.NewStack(bg, container.NewPadded(content))
	return container.NewVBox(label, panel)
}

// settingsRow 分组内一行：左侧标签 + 右侧控件，行间分隔线由调用方插入。
func settingsRow(label string, w fyne.CanvasObject) fyne.CanvasObject {
	return container.NewBorder(nil, nil, widget.NewLabel(label), nil,
		container.New(&fixedHeightLayout{height: settingsRowHeight}, w))
}

func (g *desktopGUI) buildAppearanceSection() fyne.CanvasObject {
	themeSel := fixedWidth(g.newThemeSelect(), settingsSelectWidth)
	localeSel := fixedWidth(g.newLocaleSelect(), settingsSelectWidth)
	rows := container.NewVBox(
		settingsRow(g.msgs.GUISettingsTheme, themeSel),
		widget.NewSeparator(),
		settingsRow(g.msgs.GUISettingsLanguage, localeSel),
	)
	return settingsSection(g.msgs.GUISettingsAppearance, rows)
}

// fixedWidth 约束控件宽度，高度取最小高度。
func fixedWidth(o fyne.CanvasObject, w float32) fyne.CanvasObject {
	return container.New(&fixedWidthLayout{width: w}, o)
}

func (g *desktopGUI) newThemeSelect() *widget.Select {
	labels := themeLabels(g.msgs)
	sel := widget.NewSelect(labels, nil)
	sel.SetSelected(labels[valueIndex(themeValues, g.cfg.Theme)])
	// 初始 SetSelected 之后再绑定回调，避免打开设置页即触发保存。
	sel.OnChanged = func(label string) {
		g.setTheme(optionValueAt(themeValues, labelIndex(labels, label)))
	}
	return sel
}

func (g *desktopGUI) newLocaleSelect() *widget.Select {
	labels := localeLabels(g.msgs)
	sel := widget.NewSelect(labels, nil)
	sel.SetSelected(labels[valueIndex(localeValues, g.cfg.Locale)])
	sel.OnChanged = func(label string) {
		g.setLocale(optionValueAt(localeValues, labelIndex(labels, label)))
	}
	return sel
}

// setTheme 持久化主题并立即重建界面。
func (g *desktopGUI) setTheme(value string) {
	if err := config.SaveTheme(value); err != nil {
		g.showError(fmt.Errorf("%s: %w", g.msgs.GUISettingsErrSave, err))
		return
	}
	g.cfg.Theme = value
	g.themeName = resolveTheme(g.cfg)
	g.applyThemeAndRedraw()
}

// setLocale 持久化语言、重载 i18n 并立即重建界面；返回是否成功。
func (g *desktopGUI) setLocale(value string) bool {
	if err := config.SaveLocale(value); err != nil {
		g.showError(fmt.Errorf("%s: %w", g.msgs.GUISettingsErrSave, err))
		return false
	}
	g.cfg.Locale = value
	i18n.Init(config.ResolveLocale("", g.cfg))
	g.msgs = i18n.Get()
	g.applyThemeAndRedraw()
	return true
}

// maxWidthLayout 将唯一子对象宽度限制在 width 内并水平居中。
type maxWidthLayout struct{ width float32 }

func (l *maxWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	w := size.Width
	if w > l.width {
		w = l.width
	}
	objects[0].Resize(fyne.NewSize(w, size.Height))
	objects[0].Move(fyne.NewPos((size.Width-w)/2, 0))
}

func (l *maxWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	min := objects[0].MinSize()
	if min.Width > l.width {
		min.Width = l.width
	}
	return min
}

// fixedWidthLayout 固定宽度布局，高度取子对象最小高度。
type fixedWidthLayout struct{ width float32 }

func (l *fixedWidthLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}
	objects[0].Resize(size)
}

func (l *fixedWidthLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}
	min := objects[0].MinSize()
	return fyne.NewSize(l.width, min.Height)
}

// fixedHeightLayout 固定行高布局，控件拉伸填充整行。
type fixedHeightLayout struct{ height float32 }

func (l *fixedHeightLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, o := range objects {
		o.Resize(size)
	}
}

func (l *fixedHeightLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w := float32(0)
	for _, o := range objects {
		w += o.MinSize().Width
	}
	return fyne.NewSize(w, l.height)
}
