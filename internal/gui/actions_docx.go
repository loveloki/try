package gui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"github.com/loveloki/try/internal/docx"
)

func (g *desktopGUI) packDocx() {
	entry, ok := g.singleFilesTarget()
	if !ok {
		return
	}
	if !entry.IsDir {
		g.setToast(g.msgs.GUIErrDocxNeedDir)
		return
	}
	if err := g.service.requireAllowed(entry.Path); err != nil {
		g.showError(err)
		return
	}
	g.runDocxJob(func() (string, error) {
		out, err := docx.Pack(entry.Path)
		if err != nil {
			return "", err
		}
		if err := g.service.requireMutable(out); err != nil {
			_ = os.Remove(out)
			return "", err
		}
		return out, nil
	}, func(out string) {
		g.setToast(fmt.Sprintf(g.msgs.GUIToastDocxPacked, filepath.Base(out)))
	})
}

func (g *desktopGUI) unpackDocx() {
	entry, ok := g.singleFilesTarget()
	if !ok {
		return
	}
	if entry.IsDir || !strings.EqualFold(filepath.Ext(entry.Name), ".docx") {
		g.setToast(g.msgs.GUIErrDocxNeedFile)
		return
	}
	if err := g.service.requireAllowed(entry.Path); err != nil {
		g.showError(err)
		return
	}
	g.runDocxJob(func() (string, error) {
		out, err := docx.Unpack(entry.Path)
		if err != nil {
			return "", err
		}
		if err := g.service.requireMutable(out); err != nil {
			_ = os.RemoveAll(out)
			return "", err
		}
		return out, nil
	}, func(out string) {
		g.setToast(fmt.Sprintf(g.msgs.GUIToastDocxUnpacked, filepath.Base(out)))
	})
}

func (g *desktopGUI) runDocxJob(job func() (string, error), onOK func(string)) {
	if g.view != "files" || g.dropBusy {
		return
	}
	g.dropBusy = true
	if g.watcher != nil {
		g.watcher.Pause()
	}
	go func() {
		out, err := job()
		fyne.Do(func() {
			g.dropBusy = false
			if g.watcher != nil {
				g.watcher.Resume()
			}
			if err != nil {
				g.showError(err)
				return
			}
			g.refreshFilesUI()
			onOK(out)
		})
	}()
}

// singleFilesTarget 取唯一勾选项，否则取光标项。
func (g *desktopGUI) singleFilesTarget() (FileEntry, bool) {
	if g.view != "files" {
		g.setToast(g.msgs.GUIErrNoSelection)
		return FileEntry{}, false
	}
	marked := mapKeys(g.fileMarked)
	if len(marked) > 1 {
		g.setToast(g.msgs.GUIErrDocxMulti)
		return FileEntry{}, false
	}
	if len(marked) == 1 {
		for _, f := range g.files {
			if f.Path == marked[0] {
				return f, true
			}
		}
		g.setToast(g.msgs.GUIErrNoSelection)
		return FileEntry{}, false
	}
	if g.fileSelected < 0 || g.fileSelected >= len(g.files) {
		g.setToast(g.msgs.GUIErrNoSelection)
		return FileEntry{}, false
	}
	return g.files[g.fileSelected], true
}
