package i18n

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestForLocale(t *testing.T) {
	tests := []struct {
		locale string
		want   *Messages
	}{
		{"en", &EN},
		{"zh", &ZH},
		{"unknown", &EN},
		{"", &EN},
		{"ja", &EN},
	}
	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			got := ForLocale(tt.locale)
			if got != tt.want {
				t.Errorf("ForLocale(%q) returned unexpected Messages pointer", tt.locale)
			}
		})
	}
}

// TestAllFieldsNonEmpty 确保 EN 和 ZH 中每个字段都有值，不为空字符串
func TestAllFieldsNonEmpty(t *testing.T) {
	for _, tc := range []struct {
		name string
		msgs Messages
	}{
		{"EN", EN},
		{"ZH", ZH},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := reflect.ValueOf(tc.msgs)
			typ := v.Type()
			for i := 0; i < v.NumField(); i++ {
				field := typ.Field(i)
				val := v.Field(i).String()
				if val == "" {
					t.Errorf("%s.%s is empty", tc.name, field.Name)
				}
			}
		})
	}
}

// TestFormatPlaceholders 确保含 %d 占位符的字段能正常格式化
func TestFormatPlaceholders(t *testing.T) {
	for _, tc := range []struct {
		name string
		msgs Messages
	}{
		{"EN", EN},
		{"ZH", ZH},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, field := range []struct {
				name, tmpl string
			}{
				{"DeleteTitle", tc.msgs.DeleteTitle},
				{"DeleteMode", tc.msgs.DeleteMode},
			} {
				if !strings.Contains(field.tmpl, "%d") {
					t.Errorf("%s.%s should contain %%d placeholder", tc.name, field.name)
					continue
				}
				result := fmt.Sprintf(field.tmpl, 3)
				if !strings.Contains(result, "3") {
					t.Errorf("%s.%s: Sprintf did not produce expected number: %q", tc.name, field.name, result)
				}
			}
		})
	}
}
