package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
)

const contentInsetH float32 = 16

// withHInset 为内容区 band / 列表行施加 16px 水平内边距（spec §2.2）。
func withHInset(obj fyne.CanvasObject) fyne.CanvasObject {
	return container.New(layout.NewCustomPaddedLayout(0, 0, contentInsetH, contentInsetH), obj)
}
