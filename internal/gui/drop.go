package gui

import (
	"fmt"
	"path/filepath"

	"fyne.io/fyne/v2"
	"github.com/loveloki/try/internal/i18n"
)

func (g *desktopGUI) bindDrop() {
	g.window.SetOnDropped(func(_ fyne.Position, uris []fyne.URI) {
		if g.view != "files" || len(uris) == 0 || g.dropBusy {
			return
		}
		dest := g.filesPath
		g.dropBusy = true
		if g.watcher != nil {
			g.watcher.Pause()
		}
		g.updateDropProgress(0, len(uris), "")
		g.setDropOverlayVisible(true)
		g.setPersistentToast(g.msgs.GUIDropImporting)
		go func() {
			result, err := g.service.copyDroppedFiles(dest, uris, func(done, total int, current string) {
				fyne.Do(func() {
					g.updateDropProgress(done, total, current)
				})
			})
			fyne.Do(func() {
				g.dropBusy = false
				if g.watcher != nil {
					g.watcher.Resume()
				}
				g.setDropOverlayVisible(false)
				if err != nil {
					g.showError(err)
					return
				}
				g.refreshFilesUI()
				g.setToast(formatDropToast(g.msgs, result))
			})
		}()
	})
}

func formatDropToast(msgs *i18n.Messages, result dropResult) string {
	switch {
	case result.Skipped == 0:
		return fmt.Sprintf(msgs.GUIToastCopied, result.Copied)
	case result.Copied == 0:
		return fmt.Sprintf(msgs.GUIToastSkipped, result.Skipped)
	default:
		return fmt.Sprintf(msgs.GUIToastPartial, result.Copied, result.Skipped)
	}
}

func formatDropProgressLabel(msgs *i18n.Messages, done, total int, current string) string {
	label := msgs.GUIDropImporting
	if total > 0 {
		label = fmt.Sprintf(msgs.GUIDropProgress, done, total)
	}
	if current == "" {
		return label
	}
	return fmt.Sprintf("%s · %s", label, filepath.Base(current))
}
