package gui

// EWMH _NET_WM_STATE ClientMessage 的 data.l[0] 动作码（整数，不是 Atom）。
const (
	netWMStateRemove = 0
	netWMStateAdd    = 1
)

const (
	netWMStateAtom          = "_NET_WM_STATE"
	netWMMaximizedVertAtom  = "_NET_WM_STATE_MAXIMIZED_VERT"
	netWMMaximizedHorzAtom  = "_NET_WM_STATE_MAXIMIZED_HORZ"
	netWMMoveResizeAtom     = "_NET_WM_MOVERESIZE"
	netWMMoveResizeMove     = 8
)

// netWMMaximizeAction 返回启用/取消最大化时的 EWMH 动作码。
func netWMMaximizeAction(enable bool) int {
	if enable {
		return netWMStateAdd
	}
	return netWMStateRemove
}
