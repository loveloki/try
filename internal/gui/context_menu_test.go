package gui

import (
	"testing"

	"fyne.io/fyne/v2"
	"github.com/loveloki/try/internal/i18n"
)

func TestBuildContextMenuItems(t *testing.T) {
	t.Parallel()
	i18n.Init("en")
	g := &desktopGUI{msgs: i18n.Get()}
	actions := contextMenuActions{
		onOpen:   func() {},
		onReveal: func() {},
		onRename: func() {},
		onDelete: func() {},
	}
	// 构造一个带 nil ChildMenu 的 openWith 菜单项
	openWith := buildNoAppsPlaceholder(g.msgs)
	menu := g.buildContextMenu(actions, openWith)
	if len(menu.Items) != 7 {
		t.Fatalf("expected 7 menu items (incl separator), got %d", len(menu.Items))
	}
}

func TestBuildContextMenuLabels(t *testing.T) {
	t.Parallel()
	i18n.Init("zh")
	g := &desktopGUI{msgs: i18n.Get()}
	actions := contextMenuActions{
		onOpen:   func() {},
		onReveal: func() {},
		onRename: func() {},
		onDelete: func() {},
	}
	openWith := buildNoAppsPlaceholder(g.msgs)
	menu := g.buildContextMenu(actions, openWith)
	// 检查关键标签
	labels := map[string]bool{
		"打开": false, "用...打开": false, "在文件夹中显示": false,
		"重命名": false, "删除": false,
	}
	for _, item := range menu.Items {
		if item == nil || item.IsSeparator {
			continue
		}
		if _, ok := labels[item.Label]; ok {
			labels[item.Label] = true
		}
	}
	for label, found := range labels {
		if !found {
			t.Errorf("missing menu label %q", label)
		}
	}
}

func TestBuildContextMenuEmpty(t *testing.T) {
	t.Parallel()
	i18n.Init("en")
	g := &desktopGUI{msgs: i18n.Get()}
	actions := contextMenuActions{}
	openWith := buildNoAppsPlaceholder(g.msgs)
	menu := g.buildContextMenu(actions, openWith)
	// 即使 actions 为空回调，菜单结构应完整
	if len(menu.Items) < 5 {
		t.Fatalf("expected at least 5 menu items, got %d", len(menu.Items))
	}
}

func buildNoAppsPlaceholder(msgs *i18n.Messages) *fyne.MenuItem {
	item := fyne.NewMenuItem(msgs.GUIContextMenuOpenWith, nil)
	item.ChildMenu = fyne.NewMenu("",
		fyne.NewMenuItem(msgs.GUIContextMenuNoApps, func() {}),
	)
	return item
}
