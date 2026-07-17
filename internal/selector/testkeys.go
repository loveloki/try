package selector

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

// ParseTestKeys 解析 --and-keys 参数为按键 token 列表。
// 支持两种模式：
// - Token 模式（含逗号，或全大写+连字符）：UP,DOWN,ENTER
// - Raw 模式：逐字符解析，自动识别 ANSI 转义序列
func ParseTestKeys(spec string) []string {
	if spec == "" {
		return nil
	}

	// Token 模式：含逗号，或仅由大写字母、连字符、逗号、=组成
	if strings.Contains(spec, ",") || isTokenMode(spec) {
		return parseTokenMode(spec)
	}

	return parseRawMode(spec)
}

func isTokenMode(spec string) bool {
	for _, c := range spec {
		if !((c >= 'A' && c <= 'Z') || c == '-' || c == '=' || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

func parseTokenMode(spec string) []string {
	tokens := strings.Split(spec, ",")
	var keys []string
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}
		// TYPE=text → 展开为逐字符 token
		if strings.HasPrefix(token, "TYPE=") {
			text := token[5:]
			for _, ch := range text {
				keys = append(keys, string(ch))
			}
			continue
		}
		keys = append(keys, token)
	}
	return keys
}

func parseRawMode(spec string) []string {
	var keys []string
	for i := 0; i < len(spec); {
		if spec[i] == '\x1b' && i+2 < len(spec) && spec[i+1] == '[' {
			switch spec[i+2] {
			case 'A':
				keys = append(keys, "UP")
			case 'B':
				keys = append(keys, "DOWN")
			case 'C':
				keys = append(keys, "RIGHT")
			case 'D':
				keys = append(keys, "LEFT")
			}
			i += 3
		} else if spec[i] == '\x1b' {
			keys = append(keys, "ESC")
			i++
		} else if spec[i] == '\r' || spec[i] == '\n' {
			keys = append(keys, "ENTER")
			i++
		} else if spec[i] < 0x20 {
			keys = append(keys, "CTRL-"+string(rune(spec[i]+'A'-1)))
			i++
		} else {
			keys = append(keys, string(spec[i]))
			i++
		}
	}
	return keys
}

// KeyToMsg 将按键 token 转换为 tea.KeyPressMsg
func KeyToMsg(token string) tea.KeyPressMsg {
	switch token {
	case "UP":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "DOWN":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "LEFT":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "RIGHT":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "ENTER":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "SPACE":
		return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	case "DELETE":
		return tea.KeyPressMsg{Code: tea.KeyDelete}
	case "ESC":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "BACKSPACE":
		return tea.KeyPressMsg{Code: tea.KeyBackspace}
	case "SHIFT-TAB":
		return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	default:
		if strings.HasPrefix(token, "CTRL-") {
			ch := strings.ToLower(token[5:])
			return tea.KeyPressMsg{Code: rune(ch[0]), Mod: tea.ModCtrl}
		}
		// 单个可打印字符
		r := []rune(token)
		return tea.KeyPressMsg{Code: r[0], Text: token}
	}
}
