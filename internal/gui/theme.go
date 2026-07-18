package gui

import (
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type guiTheme struct {
	variant fyne.ThemeVariant
	base    fyne.Theme
	dark    bool
}

type hexToken struct {
	dark  string
	light string
}

// palette 与 internal/selector/styles.go themeTokens 对齐。
var palette = struct {
	background      hexToken
	foreground      hexToken
	surface         hexToken
	surfaceHover    hexToken
	surfaceSelected hexToken
	line            hexToken
	header          hexToken
	highlight       hexToken
	highlightDim    hexToken
	match           hexToken
	accent          hexToken
	danger          hexToken
	dangerSurface   hexToken
	muted           hexToken
	disabled        hexToken
	onDanger        hexToken
}{
	background:      hexToken{dark: "#0d1117", light: "#ffffff"},
	foreground:      hexToken{dark: "#e6edf3", light: "#1f2328"},
	surface:         hexToken{dark: "#161b22", light: "#f6f8fa"},
	surfaceHover:    hexToken{dark: "#1f242c", light: "#f3f4f6"},
	surfaceSelected: hexToken{dark: "#21262d", light: "#eef0f2"},
	line:            hexToken{dark: "#30363d", light: "#d0d7de"},
	header:          hexToken{dark: "#5fafff", light: "#005fd7"},
	highlight:       hexToken{dark: "#5fafff", light: "#005fd7"},
	highlightDim:    hexToken{dark: "#1a2332", light: "#e8f0fe"},
	match:           hexToken{dark: "#ffaf5f", light: "#af5f00"},
	accent:          hexToken{dark: "#87d787", light: "#008700"},
	danger:          hexToken{dark: "#ff3b30", light: "#d70000"},
	dangerSurface:   hexToken{dark: "#3f1e1c", light: "#ffe5e5"},
	muted:           hexToken{dark: "#8b949e", light: "#6e7781"},
	disabled:        hexToken{dark: "#6e7681", light: "#aeb4ba"},
	onDanger:        hexToken{dark: "#ffffff", light: "#ffffff"},
}

func newGUITheme(themeName string) fyne.Theme {
	variant := theme.VariantDark
	dark := true
	if themeName == "light" {
		variant = theme.VariantLight
		dark = false
	}
	return guiTheme{
		variant: variant,
		base:    theme.DefaultTheme(),
		dark:    dark,
	}
}

func (t guiTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	if tok, ok := themeColorToken(name); ok {
		return tokenColor(tok, t.dark)
	}
	return t.base.Color(name, t.variant)
}

func themeColorToken(name fyne.ThemeColorName) (hexToken, bool) {
	switch name {
	case theme.ColorNameBackground:
		return palette.background, true
	case theme.ColorNameForeground, theme.ColorNameForegroundOnPrimary:
		return palette.foreground, true
	case theme.ColorNameButton, theme.ColorNameInputBackground,
		theme.ColorNameHeaderBackground, theme.ColorNameMenuBackground,
		theme.ColorNameOverlayBackground:
		return palette.surface, true
	case theme.ColorNameDisabledButton, theme.ColorNameHover:
		return palette.surfaceHover, true
	case theme.ColorNameDisabled:
		return palette.disabled, true
	case theme.ColorNameError:
		return palette.danger, true
	case theme.ColorNameForegroundOnError:
		return palette.onDanger, true
	case theme.ColorNameFocus, theme.ColorNamePrimary:
		return palette.highlight, true
	case theme.ColorNameInputBorder, theme.ColorNameSeparator, theme.ColorNameShadow:
		return palette.line, true
	case theme.ColorNameSelection:
		return palette.surfaceSelected, true
	case colorNameHeader:
		return palette.header, true
	case colorNameMatch:
		return palette.match, true
	case colorNameAccent:
		return palette.accent, true
	case colorNameDangerSurface:
		return palette.dangerSurface, true
	case colorNameMuted:
		return palette.muted, true
	case colorNameHighlightDim:
		return palette.highlightDim, true
	default:
		return hexToken{}, false
	}
}

func tokenColor(tok hexToken, dark bool) color.Color {
	if dark {
		return parseHexColor(tok.dark)
	}
	return parseHexColor(tok.light)
}

func parseHexColor(hex string) color.NRGBA {
	if len(hex) != 7 || hex[0] != '#' {
		return color.NRGBA{A: 0xff}
	}
	r, errR := strconv.ParseUint(hex[1:3], 16, 8)
	g, errG := strconv.ParseUint(hex[3:5], 16, 8)
	b, errB := strconv.ParseUint(hex[5:7], 16, 8)
	if errR != nil || errG != nil || errB != nil {
		return color.NRGBA{A: 0xff}
	}
	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0xff}
}

func (t guiTheme) Font(style fyne.TextStyle) fyne.Resource {
	return t.base.Font(style)
}

func (t guiTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return t.base.Icon(name)
}

func (t guiTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameSelectionRadius {
		return 0 // VS Code 式满宽直角选中条
	}
	return t.base.Size(name)
}
