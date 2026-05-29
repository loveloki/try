package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config 配置结构体，JSON 字段名为小写
type Config struct {
	Path   string `json:"path"`   // tries 根目录
	Ship   string `json:"ship"`   // ship 目标目录
	Theme  string `json:"theme"`  // 主题：dark / light / auto
	Locale string `json:"locale"` // 语言：en / zh / auto
}

var defaultConfig = Config{
	Path:   "~/src/tries",
	Ship:   "~/src/ship",
	Theme:  "auto",
	Locale: "auto",
}

// LoadConfig 从 ~/.config/try/config.json 读取配置，合并默认值。
// 配置文件不存在不报错，静默使用默认值。
func LoadConfig() Config {
	home, err := os.UserHomeDir()
	if err != nil {
		return defaultConfig
	}
	data, err := os.ReadFile(filepath.Join(home, ".config", "try", "config.json"))
	if err != nil {
		return defaultConfig
	}
	return parseConfigData(data)
}

// parseConfigData 解析 JSON 格式的配置内容，未设置的字段保留默认值。
// 空内容视为无配置（不告警），JSON 语法错误时输出 warning。
func parseConfigData(data []byte) Config {
	cfg := defaultConfig
	if len(data) == 0 {
		return cfg
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "try: 配置文件解析失败，使用默认值: %v\n", err)
	}
	return cfg
}

// ResolvePaths 按优先级解析 tries 和 ship 路径。
// tries: --path > TRY_PATH > config > default
// ship:  TRY_PROJECTS > config > default
func ResolvePaths(cliPath string, cfg Config) (triesPath, shipPath string) {
	triesPath = cfg.Path
	if env := os.Getenv("TRY_PATH"); env != "" {
		triesPath = env
	}
	if cliPath != "" {
		triesPath = cliPath
	}

	shipPath = cfg.Ship
	if env := os.Getenv("TRY_PROJECTS"); env != "" {
		shipPath = env
	}

	triesPath = ExpandPath(triesPath)
	shipPath = ExpandPath(shipPath)
	return
}

// ResolveTheme 按优先级解析主题：--theme > TRY_THEME > config > auto
func ResolveTheme(cliTheme string, cfg Config) string {
	theme := cfg.Theme
	if env := os.Getenv("TRY_THEME"); env != "" {
		theme = env
	}
	if cliTheme != "" {
		theme = cliTheme
	}
	switch theme {
	case "light", "dark":
		return theme
	default:
		return detectTheme()
	}
}

// detectTheme 通过 COLORFGBG 环境变量推断终端亮暗模式
func detectTheme() string {
	if val := os.Getenv("COLORFGBG"); val != "" {
		parts := strings.Split(val, ";")
		if len(parts) >= 2 {
			bg := parts[len(parts)-1]
			if bg == "0" || bg == "1" || bg == "2" || bg == "3" ||
				bg == "4" || bg == "5" || bg == "6" {
				return "light"
			}
		}
	}
	return "dark"
}

// ResolveLocale 按优先级解析语言：--locale > TRY_LOCALE > config > auto
func ResolveLocale(cliLocale string, cfg Config) string {
	locale := cfg.Locale
	if env := os.Getenv("TRY_LOCALE"); env != "" {
		locale = env
	}
	if cliLocale != "" {
		locale = cliLocale
	}
	switch locale {
	case "en", "zh":
		return locale
	default:
		return detectLocale()
	}
}

// detectLocale 通过 LC_ALL > LC_MESSAGES > LANG 推断语言
func detectLocale() string {
	for _, key := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if val := os.Getenv(key); val != "" {
			if strings.HasPrefix(val, "zh") {
				return "zh"
			}
			return "en"
		}
	}
	return "en"
}

// ExpandPath 展开 ~ 为用户 home 目录
func ExpandPath(s string) string {
	if strings.HasPrefix(s, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, s[2:])
	}
	return s
}
