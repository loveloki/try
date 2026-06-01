package selector

import (
	"strings"
	"testing"
)

func TestOverlayModalPreservesBackground(t *testing.T) {
	t.Helper()
	bg := "HEADER\nLIST\nFOOTER"
	modal := "MODAL"
	got := overlayModal(bg, modal, 20, 5)
	if !strings.Contains(got, "HEADER") {
		t.Error("overlay should keep header line")
	}
	if !strings.Contains(got, "MODAL") {
		t.Error("overlay should include modal text")
	}
}
