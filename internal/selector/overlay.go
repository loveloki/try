package selector

import (
	lipgloss "charm.land/lipgloss/v2"
)

// overlayModal 将 modal 居中叠放在 background 之上（Compositor 按坐标合成，不铺满空白前景）。
func overlayModal(background, modal string, width, height int) string {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}

	bg := lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, background)

	mw := lipgloss.Width(modal)
	mh := lipgloss.Height(modal)
	x := (width - mw) / 2
	y := (height - mh) / 2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	comp := lipgloss.NewCompositor(
		lipgloss.NewLayer(bg),
		lipgloss.NewLayer(modal).X(x).Y(y),
	)

	canvas := lipgloss.NewCanvas(width, height)
	canvas.Compose(comp)
	return canvas.Render()
}
