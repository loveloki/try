//go:build linux && !cgo

package gui

import "fyne.io/fyne/v2"

func nativeMinimize(_ fyne.Window) {}

func nativeToggleMaximize(_ fyne.Window) bool { return false }

func nativeIsMaximized(_ fyne.Window) bool { return false }

func nativeBeginSystemDrag(_ fyne.Window) {}
