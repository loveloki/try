package config

import (
	"strings"

	"github.com/jeandeaual/go-locale"
)

// detectOSLocale 在无 LANG/LC_* 时回退到操作系统语言（GUI 从 Dock/.app 启动常见）。
// 环境变量始终优先于本函数；显式 locale: en/zh 不经过此处。
func detectOSLocale() string {
	locales, err := locale.GetLocales()
	if err != nil || len(locales) == 0 {
		if tag, err := locale.GetLocale(); err == nil {
			return localeFromTag(tag)
		}
		return "en"
	}
	for _, tag := range locales {
		if localeFromTag(tag) == "zh" {
			return "zh"
		}
	}
	return "en"
}

func localeFromTag(tag string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(tag)), "zh") {
		return "zh"
	}
	return "en"
}
