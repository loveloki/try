//go:build linux && cgo

package gui

/*
#cgo linux LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <stdlib.h>
#include <string.h>

static void send_client_message(Display *dpy, Window win, Atom msg_type,
	long d0, long d1, long d2, long d3, long d4) {
	XClientMessageEvent event;
	memset(&event, 0, sizeof(event));
	event.type = ClientMessage;
	event.window = win;
	event.message_type = msg_type;
	event.format = 32;
	event.data.l[0] = d0;
	event.data.l[1] = d1;
	event.data.l[2] = d2;
	event.data.l[3] = d3;
	event.data.l[4] = d4;
	XSendEvent(dpy, DefaultRootWindow(dpy), False,
		SubstructureRedirectMask | SubstructureNotifyMask, (XEvent *)&event);
	XFlush(dpy);
}
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
	C.send_client_message(
		dpy, win, atomState,
		C.long(netWMMaximizeAction(enable)),
		C.long(atomVert),
		C.long(atomHorz),
		1,
		0,
	)
}

func x11BeginMove(dpy *C.Display, win C.Window) {
	var root C.Window
	var child C.Window
	var rootX, rootY, winX, winY C.int
	var mask C.uint

	C.XQueryPointer(dpy, win, &root, &child, &rootX, &rootY, &winX, &winY, &mask)
	atomMove := x11Atom(dpy, netWMMoveResizeAtom)
	C.XUngrabPointer(dpy, C.CurrentTime)
	C.send_client_message(
		dpy, win, atomMove,
		C.long(rootX),
		C.long(rootY),
		C.long(netWMMoveResizeMove),
		1,
		0,
	)
}
