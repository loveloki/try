package gui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// promptClone 弹出克隆对话框，输入 Git URL 与可选目录名。
func (g *desktopGUI) promptClone() {
	if g.cloning || g.view != "selector" {
		return
	}
	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder(g.msgs.GUICloneURLPlace)
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder(g.msgs.GUICloneName)

	form := widget.NewForm(
		widget.NewFormItem(g.msgs.GUICloneURL, urlEntry),
		widget.NewFormItem(g.msgs.GUICloneName, nameEntry),
	)
	d := dialog.NewCustomConfirm(
		g.msgs.GUITitleClone, g.msgs.GUIConfirm, g.msgs.GUICancel,
		form, func(ok bool) {
			if ok && strings.TrimSpace(urlEntry.Text) != "" {
				g.cloneAsync(strings.TrimSpace(urlEntry.Text), strings.TrimSpace(nameEntry.Text))
			}
		}, g.window,
	)
	d.Resize(fyne.NewSize(480, 200))
	d.Show()
	g.window.Canvas().Focus(urlEntry)
}

// cloneAsync 后台执行克隆，对齐 drop.go 的 goroutine + fyne.Do 模式。
func (g *desktopGUI) cloneAsync(uri, customName string) {
	g.cloning = true
	g.refreshSelectorUI()

	go func() {
		path, err := g.service.cloneEntry(uri, customName)
		fyne.Do(func() {
			g.cloning = false
			if err != nil {
				g.showError(err)
				// 用户可能已进入文件视图，仅在选择器视图刷新选择器状态
				if g.view == "selector" {
					g.refreshSelectorUI()
				}
				return
			}
			g.query = ""
			g.refreshSelectorUI()
			// 对齐 CLI doClone 行为：成功后进入文件视图
			g.enterFiles(path, path)
			g.setToast(g.msgs.GUIToastCloned)
		})
	}()
}
