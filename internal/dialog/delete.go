package dialog

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/loveloki/try/internal/i18n"
	"github.com/loveloki/try/internal/selector"
)

func msgs() *i18n.Messages { return i18n.Get() }

// DeleteDialog 删除确认对话框
type DeleteDialog struct {
	confirmYes  bool // false = NO（默认），true = YES
	markedItems []selector.DeleteItem
	basePath    string
	testConfirm string
	done        bool
	result      *selector.SelectionResult
	width       int
	styles      selector.DeleteDialogStyles
}

// NewDeleteDialog 创建删除确认对话框
func NewDeleteDialog(
	items []selector.DeleteItem,
	basePath, testConfirm string,
	width int,
	colorsEnabled bool,
	theme string,
) *DeleteDialog {
	d := &DeleteDialog{
		markedItems: items,
		basePath:    basePath,
		testConfirm: testConfirm,
		width:       width,
		styles:      selector.NewDeleteDialogStyles(colorsEnabled, theme),
	}
	if testConfirm == "YES" {
		d.confirmYes = true
	}
	return d
}

func (d *DeleteDialog) Init() tea.Cmd {
	if d.testConfirm == "YES" {
		return func() tea.Msg {
			return tea.KeyPressMsg{Code: tea.KeyEnter}
		}
	}
	return nil
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
		case tea.KeyLeft, tea.KeyRight, tea.KeyTab:
			d.toggleConfirmChoice(keyMsg)
			return d, nil
		}
		if keyMsg.Mod == tea.ModCtrl && keyMsg.Code == 'c' {
			d.done = true
			return d, nil
		}
	}
	return d, nil
}

func (d *DeleteDialog) toggleConfirmChoice(key tea.KeyPressMsg) {
	switch key.Code {
	case tea.KeyLeft:
		d.confirmYes = false
	case tea.KeyRight:
		d.confirmYes = true
	case tea.KeyTab:
		d.confirmYes = !d.confirmYes
	}
}

func (d *DeleteDialog) View() tea.View { return tea.NewView(d.ViewContent()) }

func (d *DeleteDialog) ViewContent() string {
	var b strings.Builder
	innerW := modalInnerWidth(d.width)
	if innerW < 0 {
		innerW = 0
	}
	sep := d.styles.Separator.Render(strings.Repeat("─", innerW))

	padAndTrunc := func(s string, style lipgloss.Style) string {
		if innerW <= 0 {
			return s
		}
		rendered := style.Render(s)
		w := ansi.StringWidth(rendered)
		if w >= innerW {
			return ansi.Truncate(rendered, innerW, "…")
		}
		return rendered + strings.Repeat(" ", innerW-w)
	}

	m := msgs()
	b.WriteString(padAndTrunc(fmt.Sprintf(m.DeleteTitle, len(d.markedItems)), d.styles.Title) + "\n")
	b.WriteString(sep + "\n")
	for _, item := range d.markedItems {
		b.WriteString(padAndTrunc(item.Basename, d.styles.Item) + "\n")
	}
	b.WriteString(padAndTrunc("", lipgloss.NewStyle()) + "\n")
	b.WriteString(padAndTrunc("", lipgloss.NewStyle()) + "\n")
	b.WriteString(padAndTrunc(d.renderChoice(), lipgloss.NewStyle()) + "\n")
	b.WriteString(padAndTrunc("", lipgloss.NewStyle()) + "\n")
	b.WriteString(sep + "\n")
	b.WriteString(padAndTrunc(m.DeleteFooter, d.styles.Footer))
	return renderModalBoxWithBorder(d.styles.ModalBorder, b.String(), d.width)
}

func (d *DeleteDialog) renderChoice() string {
	m := msgs()
	noLabel := m.DeleteOptionNo
	yesLabel := m.DeleteOptionYes
	if d.confirmYes {
		noLabel = " " + noLabel + " "
		yesLabel = d.styles.ChoiceActive.Render(" " + yesLabel + " ")
	} else {
		noLabel = d.styles.ChoiceActive.Render(" " + noLabel + " ")
		yesLabel = d.styles.ChoiceYes.Render(" " + yesLabel + " ")
	}
	return m.DeleteConfirmPrompt + noLabel + "  " + yesLabel
}

func (d *DeleteDialog) OverlaysMainUI() bool { return true }

func (d *DeleteDialog) Result() *selector.SelectionResult { return d.result }
func (d *DeleteDialog) Done() bool                        { return d.done }

// confirm 在选中 YES 时校验路径并返回删除结果
func (d *DeleteDialog) confirm() *selector.SelectionResult {
	if !d.confirmYes {
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
