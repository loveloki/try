package gui

import (
	"testing"

	"github.com/loveloki/try/internal/i18n"
)

func TestFormatDropToast(t *testing.T) {
	i18n.Init("en")
	msgs := i18n.Get()
	tests := []struct {
		name   string
		result dropResult
		want   string
	}{
		{"all copied", dropResult{Copied: 2}, "Copied 2 item(s)"},
		{"all skipped", dropResult{Skipped: 1}, "Skipped 1 existing item(s)"},
		{"partial", dropResult{Copied: 2, Skipped: 1}, "Copied 2, skipped 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDropToast(msgs, tt.result); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatDropProgressLabel(t *testing.T) {
	i18n.Init("en")
	msgs := i18n.Get()
	tests := []struct {
		name           string
		done, total    int
		current, want  string
	}{
		{"importing", 0, 0, "", "Copying files…"},
		{"progress", 1, 3, "", "Copying 1 / 3"},
		{"with name", 2, 4, "foo.txt", "Copying 2 / 4 · foo.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDropProgressLabel(msgs, tt.done, tt.total, tt.current); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
