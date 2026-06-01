package selector

import lipgloss "charm.land/lipgloss/v2"

// DeleteDialogStyles 删除确认弹窗样式（与主界面 danger 色一致）
type DeleteDialogStyles struct {
	Title         lipgloss.Style
	Item          lipgloss.Style
	Separator     lipgloss.Style
	Footer        lipgloss.Style
	ChoiceActive  lipgloss.Style
	ChoiceYes     lipgloss.Style
	ModalBorder   lipgloss.Style
}

// NewDeleteDialogStyles 构建删除弹窗样式
func NewDeleteDialogStyles(colorsEnabled bool, theme string) DeleteDialogStyles {
	st := newStyles(colorsEnabled, theme)
	if !colorsEnabled {
		plain := lipgloss.NewStyle()
		return DeleteDialogStyles{
			Title: plain, Item: plain, Separator: plain, Footer: plain,
			ChoiceActive: plain, ChoiceYes: plain, ModalBorder: plain,
		}
	}

	dangerNoStrike := st.danger.Strikethrough(false)
	return DeleteDialogStyles{
		Title:        st.highlight.Bold(true),
		Item:         st.danger,
		Separator:    st.muted,
		Footer:       st.muted,
		ChoiceActive: st.highlight.Reverse(true).Bold(true),
		ChoiceYes:    dangerNoStrike.Bold(true),
		ModalBorder:  st.muted,
	}
}
