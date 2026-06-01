package dialog

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/loveloki/try/internal/config"
	"github.com/loveloki/try/internal/selector"
)

var (
	shipAllowedRe     = regexp.MustCompile(`^[a-zA-Z0-9\-_.\s/~]$`)
	errShipInvalidChr = errors.New("invalid character")
)

// ShipDialog 发布为正式项目的对话框
type ShipDialog struct {
	input       textinput.Model
	entry       *selector.MatchedEntry
	basePath    string
	shipPaths   []string
	selectedIdx int
	done        bool
	result      *selector.SelectionResult
	errMsg      string
	width       int
}

// NewShipDialog 创建 ship 对话框，支持多个目标目录选择
func NewShipDialog(entry *selector.MatchedEntry, basePath string, shipPaths []string, width int) *ShipDialog {
	projectName := selector.DateSuffixRe.ReplaceAllString(entry.Entry.Basename, "")

	selectedShipPath := ""
	if len(shipPaths) > 0 {
		selectedShipPath = shipPaths[0]
	}
	defaultDest := filepath.Join(selectedShipPath, projectName)

	ti := textinput.New()
	ti.SetValue(defaultDest)
	ti.CharLimit = 512
	ti.Validate = func(s string) error {
		if len(s) == 0 {
			return nil
		}
		last := s[len(s)-1:]
		if !shipAllowedRe.MatchString(last) {
			return errShipInvalidChr
		}
		return nil
	}

	return &ShipDialog{
		input:     ti,
		entry:     entry,
		basePath:  basePath,
		shipPaths: shipPaths,
		width:     width,
	}
}

func (d *ShipDialog) Init() tea.Cmd {
	return d.input.Focus()
}

func (d *ShipDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.Code {
		case tea.KeyEnter:
			result, errMsg := d.confirmShip()
			if errMsg != "" {
				d.errMsg = errMsg
				return d, nil
			}
			d.result = result
			d.done = true
			return d, nil
		case tea.KeyEscape:
			d.done = true
			return d, nil
		case tea.KeyTab:
			if len(d.shipPaths) > 1 {
				d.switchShipPath(1)
				return d, nil
			}
		}
		if keyMsg.Mod == tea.ModShift && keyMsg.Code == tea.KeyTab {
			if len(d.shipPaths) > 1 {
				d.switchShipPath(-1)
				return d, nil
			}
		}
		if keyMsg.Mod == tea.ModCtrl && keyMsg.Code == 'c' {
			d.done = true
			return d, nil
		}
	}

	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)
	return d, cmd
}

// switchShipPath 切换选中的 ship 目录并更新输入框
func (d *ShipDialog) switchShipPath(delta int) {
	oldPath := d.shipPaths[d.selectedIdx]
	d.selectedIdx = (d.selectedIdx + delta + len(d.shipPaths)) % len(d.shipPaths)
	newPath := d.shipPaths[d.selectedIdx]

	current := d.input.Value()
	if strings.HasPrefix(current, oldPath+"/") {
		suffix := current[len(oldPath):]
		d.input.SetValue(newPath + suffix)
	} else {
		projectName := selector.DateSuffixRe.ReplaceAllString(d.entry.Entry.Basename, "")
		d.input.SetValue(filepath.Join(newPath, projectName))
	}
	d.errMsg = ""
}

func (d *ShipDialog) View() tea.View { return tea.NewView(d.ViewContent()) }

func (d *ShipDialog) ViewContent() string {
	dialogStyle := lipgloss.NewStyle().MarginLeft(4)
	var b strings.Builder
	w := d.width - 4
	if w < 0 {
		w = 0
	}
	sep := strings.Repeat("─", w)

	m := msgs()
	b.WriteString(m.ShipTitle + "\n")
	b.WriteString(sep + "\n")
	b.WriteString(d.entry.Entry.Basename + "\n\n")

	// 显示目标目录选项（每个独占一行）
	for i, sp := range d.shipPaths {
		if i == d.selectedIdx {
			b.WriteString("  ▸ " + filepath.Base(sp) + "  ← " + sp + "\n")
		} else {
			b.WriteString("    " + filepath.Base(sp) + "\n")
		}
	}
	b.WriteString("\n")

	b.WriteString(m.ShipMoveLabel + d.input.View() + "\n")
	if d.errMsg != "" {
		b.WriteString(d.errMsg + "\n")
	}
	b.WriteString("\n" + m.ShipHint + "\n\n")
	b.WriteString(sep + "\n")
	b.WriteString(m.ShipFooter)
	return dialogStyle.Render(b.String())
}

func (d *ShipDialog) OverlaysMainUI() bool { return false }

func (d *ShipDialog) Result() *selector.SelectionResult { return d.result }
func (d *ShipDialog) Done() bool                        { return d.done }

func (d *ShipDialog) confirmShip() (*selector.SelectionResult, string) {
	dest := config.ExpandPath(strings.TrimSpace(d.input.Value()))

	m := msgs()
	if dest == "" {
		return nil, m.ShipEmptyErr
	}
	if selector.FileExists(dest) {
		return nil, m.ShipExistsErr + dest
	}
	parent := filepath.Dir(dest)
	if !selector.DirExists(parent) {
		return nil, m.ShipNoParentErr + parent
	}

	return &selector.SelectionResult{
		Type:     selector.SelectShip,
		Source:   d.entry.Entry.Path,
		Dest:     dest,
		Basename: d.entry.Entry.Basename,
		BasePath: d.basePath,
	}, ""
}
