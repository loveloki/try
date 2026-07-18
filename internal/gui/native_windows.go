//go:build windows

package gui

import (
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
)

const (
	swMinimize   = 6
	swMaximize   = 3
	swRestore    = 9
	wmSysCommand = 0x0112
	scMove       = 0xF010
	htCaption    = 2
)

var (
	user32               = syscall.NewLazyDLL("user32.dll")
	procShowWindow       = user32.NewProc("ShowWindow")
	procIsZoomed         = user32.NewProc("IsZoomed")
	procReleaseCapture   = user32.NewProc("ReleaseCapture")
	procSendMessageW     = user32.NewProc("SendMessageW")
)

func nativeMinimize(w fyne.Window) {
	runWindowsNative(w, func(hwnd uintptr) {
		_, _, _ = procShowWindow.Call(hwnd, swMinimize)
	})
}

func nativeToggleMaximize(w fyne.Window) bool {
	var maximized bool
	runWindowsNative(w, func(hwnd uintptr) {
		zoomed, _, _ := procIsZoomed.Call(hwnd)
		if zoomed != 0 {
			_, _, _ = procShowWindow.Call(hwnd, swRestore)
			maximized = false
			return
		}
		_, _, _ = procShowWindow.Call(hwnd, swMaximize)
		maximized = true
	})
	return maximized
}

func nativeIsMaximized(w fyne.Window) bool {
	var maximized bool
	runWindowsNative(w, func(hwnd uintptr) {
		zoomed, _, _ := procIsZoomed.Call(hwnd)
		maximized = zoomed != 0
	})
	return maximized
}

func nativeBeginSystemDrag(w fyne.Window) {
	runWindowsNative(w, func(hwnd uintptr) {
		_, _, _ = procReleaseCapture.Call()
		_, _, _ = procSendMessageW.Call(
			hwnd,
			uintptr(wmSysCommand),
			uintptr(scMove|htCaption),
			0,
		)
	})
}

func runWindowsNative(w fyne.Window, fn func(hwnd uintptr)) {
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return
	}
	nw.RunNative(func(ctx any) {
		wctx, ok := ctx.(driver.WindowsWindowContext)
		if !ok || wctx.HWND == 0 {
			return
		}
		fn(wctx.HWND)
	})
}
