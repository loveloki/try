package dialog

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
	"github.com/loveloki/try/internal/selector"
)

var (
	renameAllowedRe = regexp.MustCompile(`^[a-zA-Z0-9\-_.\s/]$`)
	whitespaceRe    = regexp.MustCompile(`\s+`)
	errInvalidChar  = errors.New("invalid character")
)

// RenameDialog 重命名对话框
type RenameDialog struct {
	input    textinput.Model
	entry    *selector.MatchedEntry
	basePath string
	done     bool
	result   *selector.SelectionResult
	errMsg   string
	width    int
	styles   Styles
}

// NewRenameDialog 创建重命名对话框，输入框初始值为当前目录名
func NewRenameDialog(entry *selector.MatchedEntry, basePath string, width int, colorsEnabled bool) *RenameDialog {
	ti := textinput.New()
	ti.SetValue(entry.Entry.Basename)
	ti.CharLimit = 256
	ti.Validate = func(s string) error {
		if len(s) == 0 {
			return nil
		}
		last := s[len(s)-1:]
		if !renameAllowedRe.MatchString(last) {
			return errInvalidChar
		}
		return nil
	}

	return &RenameDialog{
		input:    ti,
		entry:    entry,
		basePath: basePath,
		width:    width,
		styles:   NewStyles(selector.NewStyles(colorsEnabled)),
	}
}

func (d *RenameDialog) Init() tea.Cmd {
	return d.input.Focus()
}

func (d *RenameDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch keyMsg.Code {
		case tea.KeyEnter:
			result, errMsg := d.confirmRename()
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

func (d *RenameDialog) View() tea.View { return tea.NewView(d.ViewContent()) }

func (d *RenameDialog) ViewContent() string {
	innerW := modalInnerWidth(d.width)
	if innerW < 0 {
		innerW = 0
	}
	sep := d.styles.Separator.Render(strings.Repeat("─", innerW))

	m := msgs()
	var b strings.Builder
	b.WriteString(padLine(d.styles.Title.Render(iconRenameDlg+"  "+m.RenameTitle), innerW) + "\n")
	b.WriteString(sep + "\n")
	b.WriteString(padLine(d.styles.Muted.Render(d.entry.Entry.Basename), innerW) + "\n")
	b.WriteString(padLine("", innerW) + "\n")
	b.WriteString(padLine(d.styles.InputLabel.Render(m.RenamePrompt)+d.input.View(), innerW) + "\n")
	if d.errMsg != "" {
		b.WriteString(padLine(d.styles.ErrorLine.Render(iconError+" "+d.errMsg), innerW) + "\n")
	}
	b.WriteString(padLine("", innerW) + "\n")
	b.WriteString(sep + "\n")
	footer := d.styles.JoinKeyBadges([]string{"Enter", "Esc"})
	b.WriteString(padLine(d.styles.Footer.Render(footer), innerW))
	return d.styles.renderModalBox(b.String(), d.width)
}

func (d *RenameDialog) OverlaysMainUI() bool { return true }

func (d *RenameDialog) Result() *selector.SelectionResult { return d.result }
func (d *RenameDialog) Done() bool                        { return d.done }

func (d *RenameDialog) confirmRename() (*selector.SelectionResult, string) {
	newName := strings.TrimSpace(d.input.Value())
	newName = whitespaceRe.ReplaceAllString(newName, "-")
	oldName := d.entry.Entry.Basename

	m := msgs()
	if newName == "" {
		return nil, m.RenameEmpty
	}
	if strings.ContainsAny(newName, `/\`) {
		return nil, m.RenameSlash
	}
	if newName == oldName {
		return nil, ""
	}
	if selector.DirExists(filepath.Join(d.basePath, newName)) {
		return nil, m.RenameExists + newName
	}

	return &selector.SelectionResult{
		Type:     selector.SelectRename,
		Old:      oldName,
		New:      newName,
		BasePath: d.basePath,
	}, ""
}
