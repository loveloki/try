package dialog

import (
	tea "charm.land/bubbletea/v2"
	"github.com/xleine/try/internal/selector"
)

// Dialog 对话框接口，由 SelectorModel 在 activeDialog 字段中使用
type Dialog interface {
	tea.Model
	Result() *selector.SelectionResult
	Done() bool
	ViewContent() string
}
