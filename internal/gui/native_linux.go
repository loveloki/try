//go:build linux && cgo

package gui

/*
#cgo linux LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"

import (
	"unsafe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver"
)

func nativeMinimize(w fyne.Window) {
	runX11Native(w, func(dpy *C.Display, win C.Window) {
		C.XIconifyWindow(dpy, win, C.XDefaultScreen(dpy))
		C.XFlush(dpy)
	})
}

func nativeToggleMaximize(w fyne.Window) bool {
	var maximized bool
	runX11Native(w, func(dpy *C.Display, win C.Window) {
		if x11IsMaximized(dpy, win) {
			x11SetMaximized(dpy, win, false)
			maximized = false
			return
		}
		x11SetMaximized(dpy, win, true)
		maximized = true
	})
	return maximized
}

func nativeIsMaximized(w fyne.Window) bool {
	var maximized bool
	runX11Native(w, func(dpy *C.Display, win C.Window) {
		maximized = x11IsMaximized(dpy, win)
	})
	return maximized
}

func nativeBeginSystemDrag(w fyne.Window) {
	runX11Native(w, func(dpy *C.Display, win C.Window) {
		x11BeginMove(dpy, win)
	})
}

func runX11Native(w fyne.Window, fn func(*C.Display, C.Window)) {
	nw, ok := w.(driver.NativeWindow)
	if !ok {
		return
	}
	nw.RunNative(func(ctx any) {
		xctx, ok := ctx.(driver.X11WindowContext)
		if !ok || xctx.WindowHandle == 0 {
			return
		}
		dpy := C.XOpenDisplay(nil)
		if dpy == nil {
			return
		}
		defer C.XCloseDisplay(dpy)
		fn(dpy, C.Window(xctx.WindowHandle))
	})
}

func x11Atom(dpy *C.Display, name string) C.Atom {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return C.XInternAtom(dpy, cname, 0)
}

func x11IsMaximized(dpy *C.Display, win C.Window) bool {
	atomState := x11Atom(dpy, netWMStateAtom)
	atomVert := x11Atom(dpy, netWMMaximizedVertAtom)
	atomHorz := x11Atom(dpy, netWMMaximizedHorzAtom)

	var maxLen C.ulong
	var prop *C.uchar
	var actualType C.Atom
	var actualFormat C.int
	var bytesAfter C.ulong
	status := C.XGetWindowProperty(
		dpy, win, atomState, 0, 1024, 0, C.AnyPropertyType,
		&actualType, &actualFormat, &maxLen, &bytesAfter, &prop,
	)
	if status != C.Success || prop == nil {
		return false
	}
	defer C.XFree(unsafe.Pointer(prop))

	hasVert := false
	hasHorz := false
	count := int(maxLen)
	atoms := (*[1 << 20]C.ulong)(unsafe.Pointer(prop))[:count:count]
	for _, atom := range atoms {
		if atom == C.ulong(atomVert) {
			hasVert = true
		}
		if atom == C.ulong(atomHorz) {
			hasHorz = true
		}
	}
	return hasVert && hasHorz
}

func x11SetMaximized(dpy *C.Display, win C.Window, enable bool) {
	atomState := x11Atom(dpy, netWMStateAtom)
	atomVert := x11Atom(dpy, netWMMaximizedVertAtom)
	atomHorz := x11Atom(dpy, netWMMaximizedHorzAtom)

	var event C.XEvent
	C.memset(unsafe.Pointer(&event), 0, C.size_t(unsafe.Sizeof(event)))
	event.xclient.type = C.ClientMessage
	event.xclient.window = win
	event.xclient.message_type = atomState
	event.xclient.format = 32
	// EWMH：data.l[0] 必须是 0/1/2 整数动作码，不是 Atom。
	event.xclient.data.l[0] = C.long(netWMMaximizeAction(enable))
	event.xclient.data.l[1] = C.long(atomVert)
	event.xclient.data.l[2] = C.long(atomHorz)
	event.xclient.data.l[3] = 1 // source indication: application

	root := C.XDefaultRootWindow(dpy)
	mask := C.SubstructureRedirectMask | C.SubstructureNotifyMask
	C.XSendEvent(dpy, root, 0, mask, &event)
	C.XFlush(dpy)
}

func x11BeginMove(dpy *C.Display, win C.Window) {
	var root C.Window
	var child C.Window
	var rootX, rootY, winX, winY C.int
	var mask C.uint

	C.XQueryPointer(dpy, win, &root, &child, &rootX, &rootY, &winX, &winY, &mask)

	atomMove := x11Atom(dpy, netWMMoveResizeAtom)

	var event C.XEvent
	C.memset(unsafe.Pointer(&event), 0, C.size_t(unsafe.Sizeof(event)))
	event.xclient.type = C.ClientMessage
	event.xclient.window = win
	event.xclient.message_type = atomMove
	event.xclient.format = 32
	event.xclient.data.l[0] = C.long(rootX)
	event.xclient.data.l[1] = C.long(rootY)
	event.xclient.data.l[2] = C.long(netWMMoveResizeMove)
	event.xclient.data.l[3] = 1
	event.xclient.data.l[4] = 0

	C.XUngrabPointer(dpy, C.CurrentTime)
	C.XSendEvent(dpy, root, 0, C.SubstructureRedirectMask|C.SubstructureNotifyMask, &event)
	C.XFlush(dpy)
}
