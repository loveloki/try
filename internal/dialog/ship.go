package dialog

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
	"github.com/xleine/try/internal/config"
	"github.com/xleine/try/internal/selector"
)

var (
	shipAllowedRe     = regexp.MustCompile(`^[a-zA-Z0-9\-_.\s/~]$`)
	errShipInvalidChr = errors.New("invalid character")
)

// ShipDialog 发布为正式项目的对话框
type ShipDialog struct {
	input    textinput.Model
	entry    *selector.MatchedEntry
	basePath string
	shipPath string
	done     bool
	result   *selector.SelectionResult
	errMsg   string
	width    int
}

// NewShipDialog 创建 ship 对话框，输入框初始值为推导的目标路径
func NewShipDialog(entry *selector.MatchedEntry, basePath, shipPath string, width int) *ShipDialog {
	projectName := selector.DateSuffixRe.ReplaceAllString(entry.Entry.Basename, "")
	defaultDest := filepath.Join(shipPath, projectName)

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
		input:    ti,
		entry:    entry,
		basePath: basePath,
		shipPath: shipPath,
		width:    width,
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

func (d *ShipDialog) View() tea.View { return tea.NewView(d.ViewContent()) }

func (d *ShipDialog) ViewContent() string {
	var b strings.Builder
	sep := strings.Repeat("─", d.width)

	m := msgs()
	b.WriteString("         " + m.ShipTitle + "\n")
	b.WriteString(sep + "\n")
	b.WriteString("📁 " + d.entry.Entry.Basename + "\n\n")
	b.WriteString("   " + m.ShipDestLabel + d.shipPath + "\n")
	b.WriteString("   " + m.ShipMoveLabel + d.input.View() + "\n")
	if d.errMsg != "" {
		b.WriteString("   " + d.errMsg + "\n")
	}
	b.WriteString("\n   " + m.ShipHint + "\n\n")
	b.WriteString(sep + "\n")
	b.WriteString("        " + m.ShipFooter)
	return b.String()
}

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
