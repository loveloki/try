package gui

import (
	"testing"
	"time"

	"github.com/loveloki/try/internal/i18n"
)

func TestFormatSizeKB(t *testing.T) {
	tests := []struct {
		name string
		kb   float64
		want string
	}{
		{"bytes", 0.5, "512B"},
		{"kb", 12.3, "12.3KB"},
		{"mb", 2048, "2.0MB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatSizeKB(tt.kb); got != tt.want {
				t.Errorf("formatSizeKB(%v) = %q, want %q", tt.kb, got, tt.want)
			}
		})
	}
}

func TestFormatModTime(t *testing.T) {
	i18n.Init("en")
	got := formatModTime(time.Now().Add(-30 * time.Second))
	if got != "just now" {
		t.Errorf("formatModTime recent = %q, want just now", got)
	}
	if formatModTime(time.Time{}) != "" {
		t.Error("zero time should return empty")
	}
}
