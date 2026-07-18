package gui

import (
	"fyne.io/fyne/v2"
)

const (
	defaultWindowWidth  float32 = 900
	defaultWindowHeight float32 = 600
	minWindowWidth      float32 = 720
	minWindowHeight     float32 = 480
	titleBarHeight      float32 = 44
	titleBarButtonWidth float32 = 46
)

// WindowChrome 封装平台窗口创建与内容区包装策略。
type WindowChrome struct {
	window fyne.Window
	title  string
}

// NewWindowChrome 按平台创建主窗口并应用通用尺寸约束。
func NewWindowChrome(a fyne.App, title string) *WindowChrome {
	c := &WindowChrome{title: title}
	c.window = c.createPlatformWindow(a, title)
	c.configureWindow()
	return c
}

// Window 返回底层 Fyne 窗口。
func (c *WindowChrome) Window() fyne.Window {
	return c.window
}

// Title 返回窗口标题文案。
func (c *WindowChrome) Title() string {
	return c.title
}

// UsesSystemDecoration 表示是否由系统绘制标题栏。
func (c *WindowChrome) UsesSystemDecoration() bool {
	return usesSystemDecoration()
}

// WrapContent 将内容区包在平台 chrome 内（macOS 原样返回）。
func (c *WindowChrome) WrapContent(body fyne.CanvasObject) fyne.CanvasObject {
	wrapped := c.wrapPlatformContent(body)
	return newWindowRoot(wrapped)
}

func (c *WindowChrome) configureWindow() {
	c.window.Resize(fyne.NewSize(defaultWindowWidth, defaultWindowHeight))
	c.window.CenterOnScreen()
	c.window.SetPadded(false)
}
