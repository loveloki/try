package gui

import (
	"testing"

	"github.com/loveloki/try/internal/i18n"
)

func TestCloneEntryUnparsableURI(t *testing.T) {
	i18n.Init("en")

	tests := []struct {
		name       string
		uri        string
		customName string
	}{
		{"非法URI且无自定义名", "not-a-valid-uri", ""},
		{"空URI且无自定义名", "", ""},
		{"缺少仓库段的路径", "https://github.com/only-user", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newService(t.TempDir(), nil)

			path, err := s.cloneEntry(tt.uri, tt.customName)
			if err == nil {
				t.Fatalf("cloneEntry(%q, %q) error = nil, want non-nil", tt.uri, tt.customName)
			}
			if path != "" {
				t.Fatalf("cloneEntry(%q, %q) path = %q, want empty", tt.uri, tt.customName, path)
			}
			// 错误文本为 ErrParseGitURI 前缀拼接原始 URI，精确匹配避免空 URI 用例断言恒真
			if want := i18n.Get().ErrParseGitURI + tt.uri; err.Error() != want {
				t.Fatalf("error = %q, want %q", err.Error(), want)
			}
		})
	}
}
