package dialog

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
	"github.com/user/try/internal/selector"
)

// DeleteDialog 删除确认对话框
type DeleteDialog struct {
	confirmInput textinput.Model
	markedItems  []selector.DeleteItem
	basePath     string
	testConfirm  string
	done         bool
	result       *selector.SelectionResult
	width        int
}

// NewDeleteDialog 创建删除确认对话框
func NewDeleteDialog(items []selector.DeleteItem, basePath, testConfirm string, width int) *DeleteDialog {
	ti := textinput.New()
	ti.Placeholder = "Type YES to confirm"
	ti.CharLimit = 10

	return &DeleteDialog{
		confirmInput: ti,
		markedItems:  items,
		basePath:     basePath,
		testConfirm:  testConfirm,
		width:        width,
	}
}

func (d *DeleteDialog) Init() tea.Cmd {
	cmds := []tea.Cmd{d.confirmInput.Focus()}
	if d.testConfirm != "" {
		d.confirmInput.SetValue(d.testConfirm)
		cmds = append(cmds, func() tea.Msg {
			return tea.KeyPressMsg{Code: tea.KeyEnter}
		})
	}
	return tea.Batch(cmds...)
}

func (d *DeleteDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.Code {
		case tea.KeyEnter:
			d.result = d.confirm()
			d.done = true
			return d, nil
		case tea.KeyEscape:
			d.done = true
			return d, nil
		}
		if keyMsg.Mod == tea.ModCtrl && keyMsg.Code == 'c' {
			d.done = true
			return d, nil
		}
	}

	var cmd tea.Cmd
	d.confirmInput, cmd = d.confirmInput.Update(msg)
	return d, cmd
}

func (d *DeleteDialog) View() tea.View { return tea.NewView(d.ViewContent()) }

func (d *DeleteDialog) ViewContent() string {
	var b strings.Builder
	sep := strings.Repeat("─", d.width)

	b.WriteString(fmt.Sprintf("         🗑️  Delete %d directories?\n", len(d.markedItems)))
	b.WriteString(sep + "\n")
	for _, item := range d.markedItems {
		b.WriteString("🗑️ " + item.Basename + "\n")
	}
	b.WriteString("\n\n")
	b.WriteString("        Type YES to confirm: " + d.confirmInput.View() + "\n\n")
	b.WriteString(sep + "\n")
	b.WriteString("        Enter: Confirm  Esc: Cancel")
	return b.String()
}

func (d *DeleteDialog) Result() *selector.SelectionResult { return d.result }
func (d *DeleteDialog) Done() bool                        { return d.done }

// confirm 执行确认逻辑：验证输入为 YES 且路径在 basePath 内
func (d *DeleteDialog) confirm() *selector.SelectionResult {
	if d.confirmInput.Value() != "YES" {
		return nil
	}

	baseReal, _ := filepath.EvalSymlinks(d.basePath)
	var validated []selector.DeleteItem
	for _, item := range d.markedItems {
		targetReal, _ := filepath.EvalSymlinks(item.Path)
		if !strings.HasPrefix(targetReal, baseReal+"/") {
			return nil
		}
		validated = append(validated, selector.DeleteItem{Path: targetReal, Basename: item.Basename})
	}

	return &selector.SelectionResult{
		Type:     selector.SelectDelete,
		Paths:    validated,
		BasePath: baseReal,
	}
}
