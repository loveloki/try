package cli

import (
	"strings"
)

// --- 参数解析工具函数 ---

func hasFlag(args []string, flags ...string) bool {
	for _, arg := range args {
		for _, flag := range flags {
			if arg == flag {
				return true
			}
		}
	}
	return false
}

func filterFlags(args []string, match func(string) bool) []string {
	var result []string
	for _, arg := range args {
		if !match(arg) {
			result = append(result, arg)
		}
	}
	return result
}

// extractPath 提取 --path VALUE 或 --path=VALUE（最后一个生效）
func extractPath(args []string) (string, []string) {
	var path string
	var result []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--path" && i+1 < len(args) {
			path = args[i+1]
			i++
		} else if strings.HasPrefix(args[i], "--path=") {
			path = args[i][7:]
		} else {
			result = append(result, args[i])
		}
	}
	return path, result
}

func extractBoolFlag(args []string, flag string) (bool, []string) {
	found := false
	var result []string
	for _, arg := range args {
		if arg == flag {
			found = true
		} else {
			result = append(result, arg)
		}
	}
	return found, result
}

func extractValueFlag(args []string, flag string) (string, []string) {
	var value string
	var result []string
	for i := 0; i < len(args); i++ {
		if args[i] == flag && i+1 < len(args) {
			value = args[i+1]
			i++
		} else if strings.HasPrefix(args[i], flag+"=") {
			value = args[i][len(flag)+1:]
		} else {
			result = append(result, args[i])
		}
	}
	return value, result
}
