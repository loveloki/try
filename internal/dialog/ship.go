package dialog

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
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
	styles      Styles
}

// NewShipDialog 创建 ship 对话框，支持多个目标目录选择
func NewShipDialog(entry *selector.MatchedEntry, basePath string, shipPaths []string, width int, colorsEnabled bool) *ShipDialog {
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
		styles:    NewStyles(selector.NewStyles(colorsEnabled)),
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
		default:
			if keyMsg.Mod == tea.ModShift && keyMsg.Code == tea.KeyTab {
				if len(d.shipPaths) > 1 {
					d.switchShipPath(-1)
					return d, nil
				}
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
	sep := string(filepath.Separator)
	if strings.HasPrefix(current, oldPath+sep) {
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
	innerW := modalInnerWidth(d.width)
	if innerW < 0 {
		innerW = 0
	}
	sep := d.styles.Separator.Render(strings.Repeat("─", innerW))

	m := msgs()
	var b strings.Builder
	b.WriteString(padLine(d.styles.Title.Render(iconShipDlg+"  "+m.ShipTitle), innerW) + "\n")
	b.WriteString(sep + "\n")
	b.WriteString(padLine(d.styles.Muted.Render(d.entry.Entry.Basename), innerW) + "\n")
	b.WriteString(padLine("", innerW) + "\n")

	for i, sp := range d.shipPaths {
		label := filepath.Base(sp)
		var line string
		if i == d.selectedIdx {
			line = d.styles.RadioSelected.Render(iconRadioOn+" "+label) + d.styles.Muted.Render("  "+sp)
		} else {
			line = d.styles.RadioUnselected.Render(iconRadioOff+" "+label)
		}
		b.WriteString(padLine(line, innerW) + "\n")
	}

	b.WriteString(padLine("", innerW) + "\n")
	b.WriteString(padLine(d.styles.InputLabel.Render(m.ShipMoveLabel)+d.input.View(), innerW) + "\n")
	if d.errMsg != "" {
		b.WriteString(padLine(d.styles.ErrorLine.Render(iconError+" "+d.errMsg), innerW) + "\n")
	}
	b.WriteString(padLine("", innerW) + "\n")
	b.WriteString(sep + "\n")
	footer := d.styles.JoinKeyBadges([]string{"Tab", "Enter", "Esc"})
	b.WriteString(padLine(d.styles.Footer.Render(footer), innerW))
	return d.styles.renderModalBox(b.String(), d.width)
}

func (d *ShipDialog) OverlaysMainUI() bool { return true }

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
