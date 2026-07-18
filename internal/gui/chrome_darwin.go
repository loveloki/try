//go:build darwin

package gui

import (
	"fyne.io/fyne/v2"
)

func usesSystemDecoration() bool {
	return true
}

func (c *WindowChrome) createPlatformWindow(a fyne.App, title string) fyne.Window {
	return a.NewWindow(title)
}

func (c *WindowChrome) wrapPlatformContent(body fyne.CanvasObject) fyne.CanvasObject {
	return body
}
