package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/loveloki/try/internal/config"
	"github.com/loveloki/try/internal/i18n"
)

// 主题/语言选项的存储值，顺序固定，与显示文案按索引一一对应。
var themeValues = []string{"dark", "light", "auto"}
var localeValues = []string{"en", "zh", "auto"}

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

// promptSettings 弹出设置模态对话框：修改即时生效并持久化，无保存/取消按钮。
func (g *desktopGUI) promptSettings() {
	// 防止重复打开叠加多个模态对话框。
	if g.settingsDlg != nil {
		return
	}
	var dlg dialog.Dialog
	// 语言切换后对话框文案过时，需关闭并以新语言重开。
	reopen := func() {
		dlg.Hide()
		g.promptSettings()
	}
	content := container.NewVBox(
		g.buildAppearanceCard(reopen),
		g.buildOpenWithCard(),
	)
	dlg = dialog.NewCustom(g.msgs.GUISettingsTitle, g.msgs.GUISettingsClose, content, g.window)
	dlg.SetOnClosed(func() { g.settingsDlg = nil })
	size := g.window.Canvas().Size()
	dlg.Resize(fyne.NewSize(size.Width*0.6, size.Height*0.7))
	g.settingsDlg = dlg
	dlg.Show()
}

func (g *desktopGUI) buildAppearanceCard(onLocaleChange func()) fyne.CanvasObject {
	rows := container.NewVBox(
		labeledRow(g.msgs.GUISettingsTheme, g.newThemeSelect()),
		labeledRow(g.msgs.GUISettingsLanguage, g.newLocaleSelect(onLocaleChange)),
	)
	return widget.NewCard(g.msgs.GUISettingsAppearance, "", rows)
}

// labeledRow 左侧标签 + 右侧填充控件的一行。
func labeledRow(label string, w fyne.CanvasObject) fyne.CanvasObject {
	return container.NewBorder(nil, nil, widget.NewLabel(label), nil, w)
}

func (g *desktopGUI) newThemeSelect() *widget.Select {
	labels := themeLabels(g.msgs)
	sel := widget.NewSelect(labels, nil)
	sel.SetSelected(labels[valueIndex(themeValues, g.cfg.Theme)])
	// 初始 SetSelected 之后再绑定回调，避免打开对话框即触发保存。
	sel.OnChanged = func(label string) {
		g.setTheme(optionValueAt(themeValues, labelIndex(labels, label)))
	}
	return sel
}

func (g *desktopGUI) newLocaleSelect(onLocaleChange func()) *widget.Select {
	labels := localeLabels(g.msgs)
	sel := widget.NewSelect(labels, nil)
	sel.SetSelected(labels[valueIndex(localeValues, g.cfg.Locale)])
	sel.OnChanged = func(label string) {
		if g.setLocale(optionValueAt(localeValues, labelIndex(labels, label))) {
			onLocaleChange()
		}
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
