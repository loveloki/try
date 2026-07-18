package gui

import (
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const filesWatchDebounce = 250 * time.Millisecond

// dirWatcher 监听单个目录变更，防抖后回调（回调可能在非 UI 线程）。
type dirWatcher struct {
	mu       sync.Mutex
	watcher  *fsnotify.Watcher
	path     string
	paused   bool
	onChange func()
	debounce time.Duration
	stopCh   chan struct{}
	doneCh   chan struct{}
	timer    *time.Timer
}

func newDirWatcher(onChange func(), debounce time.Duration) *dirWatcher {
	if debounce <= 0 {
		debounce = filesWatchDebounce
	}
	return &dirWatcher{onChange: onChange, debounce: debounce}
}

// SetPath 切换监听目录；path 为空则停止监听。
func (d *dirWatcher) SetPath(path string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if path == d.path && d.watcher != nil {
		return nil
	}
	d.stopLocked()
	d.path = path
	if path == "" {
		return nil
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		d.path = ""
		return err
	}
	if err := w.Add(path); err != nil {
		_ = w.Close()
		d.path = ""
		return err
	}
	d.watcher = w
	d.stopCh = make(chan struct{})
	d.doneCh = make(chan struct{})
	go d.loop(w, d.stopCh, d.doneCh)
	return nil
}

func (d *dirWatcher) Pause() {
	d.mu.Lock()
	d.paused = true
	d.mu.Unlock()
}

func (d *dirWatcher) Resume() {
	d.mu.Lock()
	d.paused = false
	d.mu.Unlock()
}

func (d *dirWatcher) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stopLocked()
	d.path = ""
}

func (d *dirWatcher) stopLocked() {
	if d.stopCh != nil {
		close(d.stopCh)
		d.stopCh = nil
	}
	if d.doneCh != nil {
		<-d.doneCh
		d.doneCh = nil
	}
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
	if d.watcher != nil {
		_ = d.watcher.Close()
		d.watcher = nil
	}
}

func (d *dirWatcher) loop(w *fsnotify.Watcher, stop <-chan struct{}, done chan struct{}) {
	defer close(done)
	for {
		select {
		case <-stop:
			return
		case err, ok := <-w.Errors:
			if !ok {
				return
			}
			_ = err
		case ev, ok := <-w.Events:
			if !ok {
				return
			}
			if ev.Op == fsnotify.Chmod {
				continue
			}
			d.mu.Lock()
			paused := d.paused
			d.mu.Unlock()
			if paused {
				continue
			}
			d.schedule()
		}
	}
}

func (d *dirWatcher) schedule() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.debounce, func() {
		d.mu.Lock()
		cb := d.onChange
		paused := d.paused
		d.mu.Unlock()
		if paused || cb == nil {
			return
		}
		cb()
	})
}
