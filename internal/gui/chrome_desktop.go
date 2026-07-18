//go:build windows || linux

package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
)

// desktopCustomChrome 为 true 时 wrap 包自绘标题栏；无 desktop.Driver 回退时为 false。
var desktopCustomChrome = true

func usesSystemDecoration() bool {
	return false
}

func (c *WindowChrome) createPlatformWindow(a fyne.App, title string) fyne.Window {
	drv, ok := a.Driver().(desktop.Driver)
	if !ok {
		// 非桌面驱动：保留系统装饰，wrap 时跳过自绘栏，避免双标题栏。
		desktopCustomChrome = false
		return a.NewWindow(title)
	}
	desktopCustomChrome = true
	w := drv.CreateSplashWindow()
	w.SetTitle(title)
	return w
}

func (c *WindowChrome) wrapPlatformContent(body fyne.CanvasObject) fyne.CanvasObject {
	if !desktopCustomChrome {
		return body
	}
	return newCustomTitleBarLayout(c.window, c.title, body)
}
