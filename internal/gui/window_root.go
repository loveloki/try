package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// windowRoot 为根内容施加最小窗口尺寸约束。
type windowRoot struct {
	widget.BaseWidget
	child fyne.CanvasObject
}

func newWindowRoot(child fyne.CanvasObject) *windowRoot {
	r := &windowRoot{child: child}
	r.ExtendBaseWidget(r)
	return r
}

func (r *windowRoot) CreateRenderer() fyne.WidgetRenderer {
	return &windowRootRenderer{root: r}
}

type windowRootRenderer struct {
	root *windowRoot
}

func (r *windowRootRenderer) Layout(size fyne.Size) {
	if r.root.child == nil {
		return
	}
	r.root.child.Resize(size)
	r.root.child.Move(fyne.NewPos(0, 0))
}

func (r *windowRootRenderer) MinSize() fyne.Size {
	if r.root.child == nil {
		return fyne.NewSize(minWindowWidth, minWindowHeight)
	}
	childMin := r.root.child.MinSize()
	return fyne.NewSize(
		maxFloat32(minWindowWidth, childMin.Width),
		maxFloat32(minWindowHeight, childMin.Height),
	)
}

func (r *windowRootRenderer) Refresh() {}

func (r *windowRootRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.root.child}
}

func (r *windowRootRenderer) Destroy() {}

func maxFloat32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
