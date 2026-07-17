package selector

// icon constants for list rows and dialogs.
const (
	iconFolder      = "📁"
	iconMarked      = "✕"
	iconSelected    = "›"
	iconRadioOn     = "●"
	iconRadioOff    = "○"
	iconDeleteDlg   = "🗑️"
	iconRenameDlg   = "✏️"
	iconShipDlg     = "🚀"
	iconSearch      = "🔍"
	iconEmptyFolder = "🔍"
	iconLoading     = "⏳"
)

// rowIconForEntry 返回条目左侧图标：标记删除时显示 ✕，否则显示展开文件夹图标。
func rowIconForEntry(isMarked bool) string {
	if isMarked {
		return iconMarked
	}
	return iconFolder
}
