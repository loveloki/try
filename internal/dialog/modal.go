package dialog

const (
	modalMinWidth = 40
	modalMaxWidth = 64
)

// modalBoxWidth 根据终端宽度计算弹窗外框宽度。
func modalBoxWidth(termWidth int) int {
	w := termWidth - 8
	if w > modalMaxWidth {
		w = modalMaxWidth
	}
	if w < modalMinWidth {
		w = modalMinWidth
	}
	return w
}

// modalInnerWidth 弹窗内容区可用宽度（扣除圆角边框与水平 padding）。
func modalInnerWidth(termWidth int) int {
	// RoundedBorder 左右各 1 列 + Padding(0,1) 左右各 1 列
	const chrome = 4
	w := modalBoxWidth(termWidth) - chrome
	if w < 0 {
		return 0
	}
	return w
}
