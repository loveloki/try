package gui

import (
	"fmt"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/loveloki/try/internal/config"
	"github.com/loveloki/try/internal/i18n"
)

// Options 启动 GUI 后端的选项。
type Options struct {
	Path string // 可选，覆盖 tries 根目录（对应 -path）
}

type desktopGUI struct {
	app     fyne.App
	window  fyne.Window
	chrome  *WindowChrome
	service *service
	msgs    *i18n.Messages

	themeName string
	view      string
	query     string
	source    string

	entries      []EntryView
	counts       map[string]int
	sources      []string
	selected     int
	selectedPath string
	marked       map[string]bool

	files        []FileEntry
	filesPath    string
	filesRoot    string
	fileSelected int
	fileMarked   map[string]bool

	selectorBody  fyne.CanvasObject
	filesBody     fyne.CanvasObject
	sourceTabsBox *fyne.Container
	filesTitle    *widget.Label
	breadcrumbBox *fyne.Container
	dropOverlay   fyne.CanvasObject
	dropOverlayLabel    *canvas.Text
	dropOverlayProgress *widget.ProgressBar
	dropBusy            bool
	toastGen            uint64

	search         *searchEntry
	list           *navList // 当前视图活动列表（selector 或 files）
	selectorStatus *statusBar
	filesStatus    *statusBar

	entryList *navList
	fileList  *navList

	watcher *dirWatcher
}

// Run 加载配置、创建原生桌面窗口，并阻塞至应用退出。
func Run(opts Options) error {
	cfg, err := loadOrInitConfig()
	if err != nil {
		return err
	}
	triesPath, shipPaths := config.ResolvePaths(opts.Path, cfg)
	locale := config.ResolveLocale("", cfg)
	i18n.Init(locale)

	ensureGUIDirs(triesPath, shipPaths)

	gui := newDesktopGUI(triesPath, shipPaths, config.DetectTheme())
	gui.run()
	return nil
}

// AppID 为 Fyne / 平台应用标识，须与 cmd/try-gui/FyneApp.toml 的 ID 一致。
const AppID = "com.loveloki.try.gui"

func newDesktopGUI(triesPath string, shipPaths []string, themeName string) *desktopGUI {
	a := app.NewWithID(AppID)
	msgs := i18n.Get()
	chrome := NewWindowChrome(a, msgs.GUITitle)

	g := &desktopGUI{
		app:        a,
		window:     chrome.Window(),
		chrome:     chrome,
		service:    newService(triesPath, shipPaths),
		msgs:       msgs,
		themeName:  normalizeTheme(themeName),
		view:       "selector",
		marked:     map[string]bool{},
		fileMarked: map[string]bool{},
	}
	g.watcher = newDirWatcher(func() {
		fyne.Do(func() {
			if g.view != "files" || g.dropBusy {
				return
			}
			g.refreshFilesUI()
		})
	}, filesWatchDebounce)
	g.applyTheme()
	g.setupTray()
	g.setupWindow()
	g.refreshEntries()
	return g
}

func (g *desktopGUI) run() {
	if g.watcher != nil {
		defer g.watcher.Close()
	}
	g.window.ShowAndRun()
}

func (g *desktopGUI) syncFilesWatch() {
	if g.watcher == nil {
		return
	}
	if g.view != "files" || g.filesPath == "" {
		_ = g.watcher.SetPath("")
		return
	}
	_ = g.watcher.SetPath(g.filesPath)
}

func (g *desktopGUI) hideToTray() {
	g.window.Hide()
}

func (g *desktopGUI) setWindowContent(body fyne.CanvasObject) {
	g.window.SetContent(g.chrome.WrapContent(body))
}

func (g *desktopGUI) setupTray() {
	desk, ok := g.app.(desktop.App)
	if !ok {
		g.window.SetCloseIntercept(g.hideToTray)
		return
	}
	menu := fyne.NewMenu(g.msgs.GUITitle,
		fyne.NewMenuItem(g.msgs.GUITrayShow, func() {
			g.window.Show()
			g.window.RequestFocus()
		}),
		fyne.NewMenuItem(g.msgs.GUITrayQuit, func() {
			g.app.Quit()
		}),
	)
	desk.SetSystemTrayMenu(menu)
	desk.SetSystemTrayIcon(theme.FolderIcon())
	desk.SetSystemTrayWindow(g.window)
	g.window.SetCloseIntercept(g.hideToTray)
}

func (g *desktopGUI) bindKeys() {
	c := g.window.Canvas()
	c.SetOnTypedKey(func(e *fyne.KeyEvent) {
		// 搜索框与列表通过各自 TypedKey 处理；此处仅兜底无焦点对象的情况。
		if focused := c.Focused(); focused != nil {
			if focused == g.search || focused == g.list {
				return
			}
		}
		g.handleNavKey(e)
	})
	c.SetOnTypedRune(func(r rune) {
		if g.view != "selector" || r != '/' {
			return
		}
		if g.search != nil && c.Focused() != g.search {
			g.focusSearch()
		}
	})
	for _, sc := range []struct {
		key fyne.KeyName
		fn  func()
	}{
		{fyne.KeyT, g.promptCreate},
		{fyne.KeyD, g.toggleMark},
		{fyne.KeyR, g.promptRename},
		{fyne.KeyG, g.promptShip},
		{fyne.KeyP, func() { g.moveSelection(-1) }},
		{fyne.KeyN, func() { g.moveSelection(1) }},
		{fyne.KeyF, g.focusSearch},
	} {
		c.AddShortcut(ctrlShortcut{key: sc.key}, func(f func()) func(fyne.Shortcut) {
			return func(fyne.Shortcut) { f() }
		}(sc.fn))
	}
}

func (g *desktopGUI) handleNavKey(e *fyne.KeyEvent) {
	switch e.Name {
	case fyne.KeyUp:
		g.moveSelection(-1)
	case fyne.KeyDown:
		g.moveSelection(1)
	case fyne.KeyEscape:
		g.handleEsc()
	case fyne.KeyReturn, fyne.KeyEnter:
		g.openSelected()
	case fyne.KeyDelete:
		g.confirmDelete()
	case fyne.KeySpace:
		g.toggleMark()
	case fyne.KeyTab:
		if g.view != "selector" {
			return
		}
		delta := 1
		if currentKeyModifiers()&fyne.KeyModifierShift != 0 {
			delta = -1
		}
		g.cycleSource(delta)
	}
}

func loadOrInitConfig() (config.Config, error) {
	cfg, err := config.LoadConfig()
	if err == nil {
		return cfg, nil
	}
	if _, initErr := config.InitConfigFile(); initErr != nil {
		return config.Config{}, fmt.Errorf("init config: %w", initErr)
	}
	cfg, err = config.LoadConfig()
	if err != nil {
		return config.Config{}, fmt.Errorf("load config: %w", err)
	}
	return cfg, nil
}

func ensureGUIDirs(triesPath string, shipPaths []string) {
	_ = os.MkdirAll(triesPath, 0o755)
	for _, sp := range shipPaths {
		_ = os.MkdirAll(sp, 0o755)
	}
}
