package i18n

var EN = Messages{
	Title:           "Try Directory Selection",
	SearchPrefix:    "Search: ",
	CreateNew:       "Create new: ",
	CreateHint:      "⌃T create \"%s\"",
	LoadingHint:     "Loading directories...",
	EmptyStateHint:  "No directories yet",
	NoMatchesHint:   "No matches for \"%s\"",
	DeleteMode:      " DELETE MODE  %d marked  |  Ctrl-D: Toggle  Enter: Confirm  Esc: Cancel",
	DeleteModeLabel: " DELETE MODE ",
	MarkedCount:     "%d marked",
	ItemCount:       "%d items",
	HintBar:         "Ctrl-T: New  Ctrl-D: Delete  Ctrl-R: Rename  Ctrl-G: Ship  Tab: Filter  Esc: Quit",
	EmptyInputHint:  "Please enter a directory name first",
	DeleteCancelled: "Delete cancelled",
	FilterAll:       "all",

	DeleteTitle:         "Delete %d directories?",
	DeleteConfirmPrompt: "Confirm: ",
	DeleteOptionNo:      "NO",
	DeleteOptionYes:     "YES",
	DeleteFooter:        "←/→: Select  Enter: Confirm  Esc: Cancel",

	RenameTitle:  "Rename directory",
	RenamePrompt: "New name: ",
	RenameEmpty:  "Name cannot be empty",
	RenameSlash:  "Name cannot contain /",
	RenameExists: "Directory exists: ",
	RenameFooter: "Enter: Confirm  Esc: Cancel",

	ShipTitle:       "Ship try to project",
	ShipMoveLabel:   "Move to: ",
	ShipHint:        "The directory will be moved to the destination",
	ShipEmptyErr:    "Destination cannot be empty",
	ShipExistsErr:   "Destination already exists: ",
	ShipNoParentErr: "Parent directory does not exist: ",
	ShipFooter:      "Tab: Switch  Enter: Confirm  Esc: Cancel",

	HelpText: `try - temporary experiment directory manager

Usage:
  try [query]                Search and select directory
  try clone <url> [name]     Clone Git repository
  try worktree <dir> [name]  Create Git worktree
  try install                Install shell integration
  try . <name>               Create worktree or directory

Options:
  --help, -h           Show help
  --version, -v        Show version

Shortcuts:
  Enter      Select/Confirm
  ↑/↓        Move up/down (wraps at edges)
  Ctrl-P/N   Move up/down (wraps at edges)
  Space      Toggle delete mark
  Delete     Toggle delete mark
  Ctrl-A     Mark all visible
  Ctrl-D     Toggle delete mode
  Ctrl-T     Create new directory
  Ctrl-R     Rename
  Ctrl-G     Ship (publish as project)
  Tab        Switch source filter
  Shift-Tab  Switch source filter backward
  /          Focus search
  Esc        Quit/Cancel`,
	UsageClone:    "Usage: try clone <url> [name]",
	UsageWorktree: "Usage: try worktree <dir> [name]",
	UsageDot:      "Usage: try . <name>",

	TimeJustNow:  "just now",
	TimeMinAgo:   "%dm ago",
	TimeHourAgo:  "%dh ago",
	TimeDayAgo:   "%dd ago",
	TimeWeekAgo:  "%dw ago",
	TimeMonthAgo: "%dmo ago",
	TimeYearAgo:  "%dy ago",

	ErrNotTerminal: "stdin is not a terminal",
	ErrParseGitURI: "Cannot parse Git URI: ",
	ErrParsePath:   "Cannot resolve path: ",
	ErrReadDir:     "Cannot read directory: ",

	MsgCloneFrom:     "Using git clone to create this trial from %s.",
	MsgShipped:       "Shipped: %s → %s",
	ErrMkdir:         "failed to create directory",
	ErrMkdirParent:   "failed to create parent directory",
	ErrClone:         "git clone failed",
	ErrWorktreeAdd:   "git worktree add failed (directory created): %v",
	ErrDeletePartial: "partial delete failed:\n%s",
	ErrRename:        "rename failed",
	ErrWorktreeMove:  "git worktree move failed",
	ErrMove:          "failed to move directory",
	ErrUnknownOp:     "unknown operation type: %d",
	ErrPathDenied:    "path outside allowed directories",
	ErrBadRequest:    "invalid request",

	GUITitle:               "try-gui",
	GUISearchPlace:         "Type to filter or create...",
	GUIFilesTitle:          "Files",
	GUIFilesEmpty:          "This directory is empty",
	GUIFilesHintBar:        "Enter: Open  Delete: Delete  Esc: Back  ·  Drop to copy",
	GUIBack:                "Back",
	GUIOpen:                "Open",
	GUIEdit:                "Edit",
	GUIDelete:              "Delete",
	GUIConfirm:             "Confirm",
	GUICancel:              "Cancel",
	GUIThemeDark:           "Dark",
	GUIThemeLight:          "Light",
	GUIThemeToggle:         "Toggle theme",
	GUITrayShow:            "Show",
	GUITrayQuit:            "Quit",
	GUIToastCreated:        "Created",
	GUIToastDeleted:        "Deleted",
	GUIToastRenamed:        "Renamed",
	GUIToastShipped:        "Shipped",
	GUIToastOpened:         "Opened",
	GUIToastOpenedFolder:   "Opened in file manager",
	GUIToastCopied:         "Copied %d item(s)",
	GUIToastSkipped:        "Skipped %d existing item(s)",
	GUIToastPartial:        "Copied %d, skipped %d",
	GUIToastCopying:        "Copying…",
	GUIDropImporting:       "Copying files…",
	GUIDropProgress:        "Copying %d / %d",
	GUIOpenFolder:          "Reveal",
	GUIDocxPack:            "Pack .docx",
	GUIDocxUnpack:          "Unpack .docx",
	GUIToastDocxPacked:     "Packed %s",
	GUIToastDocxUnpacked:   "Unpacked to %s",
	GUIErrDocxNeedFile:     "Select a .docx file",
	GUIErrDocxNeedDir:      "Select a folder to pack",
	GUIErrDocxMulti:        "Select only one item",
	GUINotImplemented:      "Not implemented yet",
	GUIFilesItemCount:      "%d items",
	GUIColName:             "Name",
	GUIColSize:             "Size",
	GUIColMTime:            "Modified",
	GUIShortcutNav:         "Navigate",
	GUIShortcutOpen:        "Open",
	GUIShortcutNew:         "New",
	GUIShortcutDelete:      "Delete",
	GUIShortcutEscBack:     "Back",
	GUIShortcutDrop:        "Drop to upload",
	GUIContextMenuOpen:     "Open",
	GUIContextMenuOpenWith: "Open With...",
	GUIContextMenuReveal:   "Show in Folder",
	GUIContextMenuRename:   "Rename",
	GUIContextMenuDelete:   "Delete",
	GUIContextMenuNoApps:   "No applications found",
	GUIRenameFileTitle:     "Rename file",
	GUIErrCopy:             "copy failed",
	GUIErrOpen:             "open failed",
	GUIErrOpenFolder:       "open folder failed",
	GUIErrNoSelection:      "no selection",
	GUITitleClone:          "Clone Repository",
	GUICloneURL:            "URL",
	GUICloneName:           "Name (optional)",
	GUICloneURLPlace:       "https://github.com/user/repo",
	GUICloning:             "Cloning…",
	GUIToastCloned:         "Cloned successfully",

	// 设置对话框
	GUISettingsTitle:          "Settings",
	GUISettingsAppearance:     "Appearance",
	GUISettingsTheme:          "Theme",
	GUISettingsThemeDark:      "Dark",
	GUISettingsThemeLight:     "Light",
	GUISettingsThemeAuto:      "Auto",
	GUISettingsLanguage:       "Language",
	GUISettingsLangAuto:       "Auto",
	GUISettingsOpenWith:       "Open With",
	GUISettingsExtension:      "Extension",
	GUISettingsApplication:    "Application",
	GUISettingsAdd:            "Add",
	GUISettingsDelete:         "Delete",
	GUISettingsExtPlaceholder: ".go or *",
	GUISettingsInvalidExt:     "Extension must contain only letters and digits",
	GUISettingsInvalidApp:     "Invalid application name",
	GUISettingsNoMappings:     "No custom mappings",
	GUISettingsClose:          "Close",
	GUISettingsBack:           "Back",
	GUISettingsMappingAdded:   "%s → %s added",
	GUISettingsErrSave:        "Failed to save settings",

	ErrDetectShell:    "Cannot detect shell type. Please use bash, zsh, or fish.",
	ErrUnsupportShell: "Unsupported shell type: %s",
	ErrGetExePath:     "Cannot get try executable path",
	MsgAlreadyInstall: "try shell integration is already installed in %s.",
	MsgReinstallHint:  "To reinstall, first remove the old version (search for \"%s\").",
	ErrCreateDir:      "failed to create directory",
	ErrWriteFile:      "cannot write to %s",
	ErrWrite:          "write failed",
	MsgInstalled:      "try shell integration has been written to %s",
	MsgSourceHint:     "Please run source %s or restart your terminal to activate.",
}
