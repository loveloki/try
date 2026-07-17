package selector

import (
	"image/color"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	"github.com/loveloki/try/internal/config"
)

// colorToken 保存一个语义颜色在深色/浅色主题下的十六进制值。
type colorToken struct {
	dark  string
	light string
}

// themeTokens 与 Next.js GUI 设计系统对齐的完整调色板。
var themeTokens = struct {
	background      colorToken
	foreground      colorToken
	surface         colorToken
	surfaceHover    colorToken
	surfaceSelected colorToken
	line            colorToken
	header          colorToken
	highlight       colorToken
	match           colorToken
	accent          colorToken
	danger          colorToken
	dangerDim       colorToken
	dangerSurface   colorToken
	onDanger        colorToken
	muted           colorToken
	disabled        colorToken
}{
	background:      colorToken{dark: "#0d1117", light: "#ffffff"},
	foreground:      colorToken{dark: "#e6edf3", light: "#1f2328"},
	surface:         colorToken{dark: "#161b22", light: "#f6f8fa"},
	surfaceHover:    colorToken{dark: "#1f242c", light: "#f3f4f6"},
	surfaceSelected: colorToken{dark: "#21262d", light: "#eef0f2"},
	line:            colorToken{dark: "#30363d", light: "#d0d7de"},
	header:          colorToken{dark: "#5fafff", light: "#005fd7"},
	highlight:       colorToken{dark: "#5fafff", light: "#005fd7"},
	match:           colorToken{dark: "#ffaf5f", light: "#af5f00"},
	accent:          colorToken{dark: "#87d787", light: "#008700"},
	danger:          colorToken{dark: "#ff3b30", light: "#d70000"},
	dangerDim:       colorToken{dark: "#ff3b30", light: "#d70000"},
	dangerSurface:   colorToken{dark: "#3f1e1c", light: "#ffe5e5"},
	onDanger:        colorToken{dark: "#ffffff", light: "#ffffff"},
	muted:           colorToken{dark: "#8b949e", light: "#6e7781"},
	disabled:        colorToken{dark: "#6e7681", light: "#aeb4ba"},
}

// adaptive 返回一个根据终端主题自动切换深/浅色的 lipgloss 颜色。
func adaptive(t colorToken) color.Color {
	isDark := config.DetectTheme() == "dark"
	return lipgloss.LightDark(isDark)(lipgloss.Color(t.light), lipgloss.Color(t.dark))
}

// Styles 集中暴露给 TUI 各组件使用的样式集。
type Styles struct {
	colorsEnabled bool

	Background      lipgloss.Style
	Foreground      lipgloss.Style
	Surface         lipgloss.Style
	SurfaceHover    lipgloss.Style
	SurfaceSelected lipgloss.Style
	Line            lipgloss.Style
	Header          lipgloss.Style
	Highlight       lipgloss.Style
	Match           lipgloss.Style
	Accent          lipgloss.Style
	Danger          lipgloss.Style
	DangerSurface   lipgloss.Style
	Muted           lipgloss.Style
	Disabled        lipgloss.Style

	// 派生样式
	SelectedArrow   lipgloss.Style
	MarkedIcon      lipgloss.Style
	FolderIcon      lipgloss.Style
	ScoreBarFilled  lipgloss.Style
	ScoreBarEmpty   lipgloss.Style
	SourcePill      lipgloss.Style
	KeyBadge        lipgloss.Style
	DeleteModeBadge lipgloss.Style

	// 行状态样式
	MarkedName lipgloss.Style
	MarkedMeta lipgloss.Style
}

// NewStyles 根据终端主题和颜色开关创建样式集。
func NewStyles(colorsEnabled bool) *Styles {
	if !colorsEnabled {
		return plainStyles()
	}

	t := themeTokens
	c := adaptive
	fg := func(token colorToken) lipgloss.Style {
		return lipgloss.NewStyle().Foreground(c(token))
	}
	bg := func(token colorToken) lipgloss.Style {
		return lipgloss.NewStyle().Background(c(token))
	}

	// 选中行文字使用 foreground（深色主题浅字 / 浅色主题深字），保证对比度。
	selectedFg := c(t.foreground)

	s := &Styles{
		colorsEnabled: colorsEnabled,

		Background:      bg(t.background),
		Foreground:      fg(t.foreground),
		Surface:         bg(t.surface),
		SurfaceHover:    bg(t.surfaceHover),
		SurfaceSelected: bg(t.surfaceSelected).Foreground(selectedFg).Bold(true),
		Line:            fg(t.line),
		Header:          fg(t.header).Bold(true),
		Highlight:       fg(t.highlight).Bold(true),
		Match:           fg(t.match).Bold(true),
		Accent:          fg(t.accent).Bold(true),
		Danger:          fg(t.danger).Strikethrough(true),
		DangerSurface:   bg(t.dangerSurface).Foreground(c(t.onDanger)).Strikethrough(true),
		Muted:           fg(t.muted),
		Disabled:        fg(t.disabled),

		SelectedArrow:   fg(t.header).Bold(true),
		MarkedIcon:      fg(t.danger).Bold(true),
		FolderIcon:      fg(t.muted),
		ScoreBarFilled:  fg(t.header),
		ScoreBarEmpty:   fg(t.line),
		SourcePill:      lipgloss.NewStyle().Foreground(c(t.muted)).Background(c(t.surfaceHover)),
		KeyBadge:        lipgloss.NewStyle().Foreground(c(t.muted)).Background(c(t.surface)),
		DeleteModeBadge: lipgloss.NewStyle().Foreground(c(t.danger)).Background(c(t.surface)).Bold(true),

		MarkedName: bg(t.dangerSurface).Foreground(c(t.onDanger)).Strikethrough(true),
		MarkedMeta: bg(t.dangerSurface).Foreground(c(t.onDanger)).Strikethrough(true),
	}

	return s
}

// plainStyles 在无颜色模式下返回所有空白样式。
func plainStyles() *Styles {
	plain := lipgloss.NewStyle()
	return &Styles{
		colorsEnabled:   false,
		Background:      plain,
		Foreground:      plain,
		Surface:         plain,
		SurfaceHover:    plain,
		SurfaceSelected: plain,
		Line:            plain,
		Header:          plain.Bold(true),
		Highlight:       plain.Bold(true),
		Match:           plain.Bold(true),
		Accent:          plain.Bold(true),
		Danger:          plain,
		DangerSurface:   plain,
		Muted:           plain,
		Disabled:        plain,
		SelectedArrow:   plain.Bold(true),
		MarkedIcon:      plain.Bold(true),
		FolderIcon:      plain,
		ScoreBarFilled:  plain,
		ScoreBarEmpty:   plain,
		SourcePill:      plain,
		KeyBadge:        plain,
		DeleteModeBadge: plain.Bold(true),
		MarkedName:      plain,
		MarkedMeta:      plain,
	}
}

// Render 使用指定样式渲染文本。
func (s *Styles) Render(sty lipgloss.Style, text string) string {
	return sty.Render(text)
}

// ColorsEnabled 返回颜色是否启用。
func (s *Styles) ColorsEnabled() bool {
	return s.colorsEnabled
}

// KeyBadgeText 渲染一个快捷键徽章，例如 "Ctrl-T"。
func (s *Styles) KeyBadgeText(text string) string {
	return s.KeyBadge.Render(" " + text + " ")
}

// JoinKeyBadges 用空格连接多个快捷键徽章。
func (s *Styles) JoinKeyBadges(parts []string) string {
	var rendered []string
	for _, p := range parts {
		rendered = append(rendered, s.KeyBadgeText(p))
	}
	return strings.Join(rendered, " ")
}
