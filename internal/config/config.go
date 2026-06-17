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
	Path   string   `json:"path"`   // tries 根目录
	Ships  []string `json:"ships"`  // ship 目标目录列表
	Locale string   `json:"locale"` // 语言：en / zh / auto
}

var defaultShips = []string{"~/src/ship", "~/src/bug"}

var defaultConfig = Config{
	Path:   "~/src/tries",
	Ships:  defaultShips,
	Locale: "auto",
}

// LoadConfig 从 ~/.config/try/config.json 读取配置。
// 文件不存在、空文件或 JSON 语法错误时返回 error。
func LoadConfig() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("failed to get home directory: %w", err)
	}
	data, err := os.ReadFile(filepath.Join(home, ".config", "try", "config.json"))
	if err != nil {
		return Config{}, fmt.Errorf("cannot read config ~/.config/try/config.json: %w", err)
	}
	return parseConfigData(data)
}

// parseConfigData 解析 JSON 格式的配置内容，未设置的字段保留默认值。
// 空文件或 JSON 语法错误时返回 error。
func parseConfigData(data []byte) (Config, error) {
	if len(data) == 0 {
		return Config{}, fmt.Errorf("config file is empty")
	}

	// 先解析到一个 Ships 为 nil 的结构，以区分"未设置"和"设置为空"
	var raw struct {
		Path   string   `json:"path"`
		Ships  []string `json:"ships"`
		Locale string   `json:"locale"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return Config{}, fmt.Errorf("failed to parse config: %w", err)
	}

	cfg := defaultConfig
	if raw.Path != "" {
		cfg.Path = raw.Path
	}
	if raw.Locale != "" {
		cfg.Locale = raw.Locale
	}
	if len(raw.Ships) > 0 {
		cfg.Ships = raw.Ships
	}

	return cfg, nil
}

// ResolvePaths 按优先级解析 tries 和 ship 路径。
// tries: --path > TRY_PATH > config > default
// ships: TRY_PROJECTS > config > default
func ResolvePaths(cliPath string, cfg Config) (triesPath string, shipPaths []string) {
	triesPath = cfg.Path
	if env := os.Getenv("TRY_PATH"); env != "" {
		triesPath = env
	}
	if cliPath != "" {
		triesPath = cliPath
	}

	shipPaths = cfg.Ships
	if env := os.Getenv("TRY_PROJECTS"); env != "" {
		shipPaths = []string{env}
	}

	triesPath = ExpandPath(triesPath)
	for i, p := range shipPaths {
		shipPaths[i] = ExpandPath(p)
	}
	return
}

// DetectTheme 通过 COLORFGBG 环境变量推断终端亮暗模式
func DetectTheme() string {
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
		return DetectLocale()
	}
}

// DetectLocale 通过 LC_ALL > LC_MESSAGES > LANG 推断语言
func DetectLocale() string {
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

// InitConfigFile 在 ~/.config/try/config.json 创建默认配置文件（如果不存在）。
// 文件已存在时返回 (false, nil)，新创建时返回 (true, nil)。
func InitConfigFile() (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}
	path := filepath.Join(home, ".config", "try", "config.json")

	// 文件已存在，不做任何事
	if _, err := os.Stat(path); err == nil {
		return false, nil
	}

	// 创建目录
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false, fmt.Errorf("failed to create config directory: %w", err)
	}

	// 写入默认配置
	data, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return false, fmt.Errorf("failed to marshal default config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return false, fmt.Errorf("failed to write config file: %w", err)
	}
	return true, nil
}

// ExpandPath 展开 ~ 为用户 home 目录
func ExpandPath(s string) string {
	if strings.HasPrefix(s, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, s[2:])
	}
	return s
}
