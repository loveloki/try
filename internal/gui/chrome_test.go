package gui

import (
	"image/color"
	"runtime"
	"testing"
)

func TestWindowChromeConstants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		got  float32
		want float32
	}{
		{"default width", defaultWindowWidth, 900},
		{"default height", defaultWindowHeight, 600},
		{"min width", minWindowWidth, 720},
		{"min height", minWindowHeight, 480},
		{"title bar height", titleBarHeight, 44},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("got %v want %v", tt.got, tt.want)
			}
		})
	}
}

func TestUsesSystemDecorationByPlatform(t *testing.T) {
	t.Parallel()
	want := runtime.GOOS == "darwin"
	if got := usesSystemDecoration(); got != want {
		t.Fatalf("usesSystemDecoration() = %v, want %v for GOOS=%s", got, want, runtime.GOOS)
	}
}

func TestParseHexColor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		hex  string
		want [4]uint8
	}{
		{"dark background", "#0d1117", [4]uint8{0x0d, 0x11, 0x17, 0xff}},
		{"light foreground", "#1f2328", [4]uint8{0x1f, 0x23, 0x28, 0xff}},
		{"invalid", "bad", [4]uint8{0, 0, 0, 0xff}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseHexColor(tt.hex)
			if got.R != tt.want[0] || got.G != tt.want[1] || got.B != tt.want[2] || got.A != tt.want[3] {
				t.Fatalf("parseHexColor(%q) = %+v want %+v", tt.hex, got, tt.want)
			}
		})
	}
}

func TestTokenColorDarkLight(t *testing.T) {
	t.Parallel()
	dark := tokenColor(palette.background, true).(color.NRGBA)
	light := tokenColor(palette.background, false).(color.NRGBA)
	if dark.R != 0x0d || light.R != 0xff {
		t.Fatalf("unexpected token colors: dark=%+v light=%+v", dark, light)
	}
}

func TestWindowRootMinSize(t *testing.T) {
	t.Parallel()
	root := newWindowRoot(nil)
	min := root.MinSize()
	if min.Width < minWindowWidth || min.Height < minWindowHeight {
		t.Fatalf("root min size %v below window minimum", min)
	}
}

func TestNetWMMaximizeAction(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		enable bool
		want   int
	}{
		{"enable maximize", true, netWMStateAdd},
		{"disable maximize", false, netWMStateRemove},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := netWMMaximizeAction(tt.enable); got != tt.want {
				t.Fatalf("netWMMaximizeAction(%v) = %d want %d", tt.enable, got, tt.want)
			}
		})
	}
}

func TestEWMHAtomNames(t *testing.T) {
	t.Parallel()
	if netWMMaximizedHorzAtom != "_NET_WM_STATE_MAXIMIZED_HORZ" {
		t.Fatalf("HORZ atom typo: %q", netWMMaximizedHorzAtom)
	}
	if netWMStateAdd != 1 || netWMStateRemove != 0 {
		t.Fatalf("EWMH action codes must be integers 0/1, got remove=%d add=%d",
			netWMStateRemove, netWMStateAdd)
	}
}
