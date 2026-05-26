package config

import (
	"os"
	"path/filepath"
	"strings"
)

// Config 配置结构体，字段对应配置文件中的 key
type Config struct {
	Path  string // tries 根目录
	Ship  string // ship 目标目录
	Theme string // 主题：dark / light / auto
}

var defaultConfig = Config{
	Path:  "~/src/tries",
	Ship:  "~/src/ship",
	Theme: "auto",
}

// LoadConfig 从 ~/.try 读取配置，合并默认值。
// 配置文件不存在不报错，静默使用默认值。
func LoadConfig() Config {
	home, err := os.UserHomeDir()
	if err != nil {
		return defaultConfig
	}
	data, err := os.ReadFile(filepath.Join(home, ".try"))
	if err != nil {
		return defaultConfig
	}
	return parseConfigData(data)
}

// parseConfigData 解析 key=value 格式的配置内容，合并默认值
func parseConfigData(data []byte) Config {
	cfg := defaultConfig
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key := strings.TrimSpace(k)
		value := strings.TrimSpace(v)
		switch key {
		case "path":
			cfg.Path = value
		case "ship":
			cfg.Ship = value
		case "theme":
			cfg.Theme = value
		}
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

// ExpandPath 展开 ~ 为用户 home 目录
func ExpandPath(s string) string {
	if strings.HasPrefix(s, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, s[2:])
	}
	return s
}
