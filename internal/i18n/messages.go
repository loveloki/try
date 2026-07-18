package i18n

// Messages 定义所有用户可见文本，按功能模块分组
type Messages struct {
	// === selector 主界面 ===
	Title           string
	SearchPrefix    string
	CreateNew       string
	CreateHint      string
	LoadingHint     string
	EmptyStateHint  string
	NoMatchesHint   string // 含 %s 占位符
	DeleteMode      string // 含 %d 占位符
	DeleteModeLabel string
	MarkedCount     string // 含 %d 占位符
	ItemCount       string // 含 %d 占位符
	HintBar         string
	EmptyInputHint  string
	DeleteCancelled string
	FilterAll       string // 来源过滤标签："all" / "全部"

	// === 删除对话框 ===
	DeleteTitle         string // 含 %d 占位符
	DeleteConfirmPrompt string
	DeleteOptionNo      string
	DeleteOptionYes     string
	DeleteFooter        string

	// === 重命名对话框 ===
	RenameTitle  string
	RenamePrompt string
	RenameEmpty  string
	RenameSlash  string
	RenameExists string // 前缀，后接目录名
	RenameFooter string

	// === Ship 对话框 ===
	ShipTitle       string
	ShipMoveLabel   string
	ShipHint        string
	ShipEmptyErr    string
	ShipExistsErr   string // 前缀，后接路径
	ShipNoParentErr string // 前缀，后接路径
	ShipFooter      string

	// === CLI 帮助与用法 ===
	HelpText      string // 完整帮助文本（多行）
	UsageClone    string // "Usage: try clone <url> [name]"
	UsageWorktree string
	UsageDot      string

	// === CLI 错误消息 ===
	ErrNotTerminal string // stdin 非终端时的提示
	ErrParseGitURI string // 前缀
	ErrParsePath   string // 前缀
	ErrReadDir     string // 前缀

	// === 操作执行消息 ===
	MsgCloneFrom     string // 含 %s
	MsgShipped       string // "Shipped: %s → %s"
	ErrMkdir         string
	ErrMkdirParent   string
	ErrClone         string
	ErrWorktreeAdd   string // 含 %v（非致命，附加提示）
	ErrDeletePartial string // 含 %s
	ErrRename        string
	ErrWorktreeMove  string
	ErrMove          string
	ErrUnknownOp     string // 含 %d
	ErrPathDenied    string // 路径越界
	ErrBadRequest    string // 非法请求

	// === GUI 桌面窗口 ===
	GUITitle             string
	GUISearchPlace       string
	GUIFilesTitle        string
	GUIFilesEmpty        string
	GUIFilesHintBar      string // 兼容旧单行；新底栏用下方分段字段
	GUIBack              string
	GUIOpen              string
	GUIEdit              string
	GUIDelete            string
	GUIConfirm           string
	GUICancel            string
	GUIThemeDark         string
	GUIThemeLight        string
	GUIThemeToggle       string // 主题按钮无障碍标签
	GUITrayShow          string
	GUITrayQuit          string
	GUIToastCreated      string
	GUIToastDeleted      string
	GUIToastRenamed      string
	GUIToastShipped      string
	GUIToastOpened       string
	GUIToastOpenedFolder string
	GUIToastCopied       string // 含 %d
	GUIToastSkipped      string // 含 %d
	GUIToastPartial      string // 含 %d %d（copied, skipped）
	GUIToastCopying   string
	GUIDropImporting  string
	GUIDropProgress   string // 含 %d %d（done, total）
	GUIOpenFolder     string
	GUIDocxPack          string
	GUIDocxUnpack        string
	GUINotImplemented    string
	GUIFilesItemCount    string // 含 %d
	GUIColName           string
	GUIColSize           string
	GUIColMTime          string
	GUIShortcutNav       string
	GUIShortcutOpen      string
	GUIShortcutNew       string
	GUIShortcutDelete    string
	GUIShortcutEscBack   string
	GUIShortcutDrop      string
	GUIErrCopy           string
	GUIErrOpen           string
	GUIErrOpenFolder     string
	GUIErrNoSelection    string

	// === 时间格式化 ===
	TimeJustNow  string // "just now" / "刚刚"
	TimeMinAgo   string // 含 %d："%dm ago" / "%d分钟前"
	TimeHourAgo  string // 含 %d："%dh ago" / "%d小时前"
	TimeDayAgo   string // 含 %d："%dd ago" / "%d天前"
	TimeWeekAgo  string // 含 %d："%dw ago" / "%d周前"
	TimeMonthAgo string // 含 %d："%dmo ago" / "%d个月前"
	TimeYearAgo  string // 含 %d："%dy ago" / "%d年前"

	// === Shell 安装消息 ===
	ErrDetectShell    string
	ErrUnsupportShell string // 含 %s
	ErrGetExePath     string
	MsgAlreadyInstall string // 含 %s
	MsgReinstallHint  string // 含 %s
	ErrCreateDir      string
	ErrWriteFile      string // 含 %s
	ErrWrite          string
	MsgInstalled      string // 含 %s
	MsgSourceHint     string // 含 %s
}

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

	GUITitle:             "try-gui",
	GUISearchPlace:       "Type to filter or create...",
	GUIFilesTitle:        "Files",
	GUIFilesEmpty:        "This directory is empty",
	GUIFilesHintBar:      "Enter: Open  Delete: Delete  Esc: Back  ·  Drop to copy",
	GUIBack:              "Back",
	GUIOpen:              "Open",
	GUIEdit:              "Edit",
	GUIDelete:            "Delete",
	GUIConfirm:           "Confirm",
	GUICancel:            "Cancel",
	GUIThemeDark:         "Dark",
	GUIThemeLight:        "Light",
	GUIThemeToggle:       "Toggle theme",
	GUITrayShow:          "Show",
	GUITrayQuit:          "Quit",
	GUIToastCreated:      "Created",
	GUIToastDeleted:      "Deleted",
	GUIToastRenamed:      "Renamed",
	GUIToastShipped:      "Shipped",
	GUIToastOpened:       "Opened",
	GUIToastOpenedFolder: "Opened in file manager",
	GUIToastCopied:       "Copied %d item(s)",
	GUIToastSkipped:      "Skipped %d existing item(s)",
	GUIToastPartial:      "Copied %d, skipped %d",
	GUIToastCopying:      "Copying…",
	GUIDropImporting:     "Copying files…",
	GUIDropProgress:      "Copying %d / %d",
	GUIOpenFolder:        "Reveal",
	GUIDocxPack:          "Pack .docx",
	GUIDocxUnpack:        "Unpack .docx",
	GUINotImplemented:    "Not implemented yet",
	GUIFilesItemCount:    "%d items",
	GUIColName:           "Name",
	GUIColSize:           "Size",
	GUIColMTime:          "Modified",
	GUIShortcutNav:       "Navigate",
	GUIShortcutOpen:      "Open",
	GUIShortcutNew:       "New",
	GUIShortcutDelete:    "Delete",
	GUIShortcutEscBack:   "Back",
	GUIShortcutDrop:      "Drop to upload",
	GUIErrCopy:           "copy failed",
	GUIErrOpen:           "open failed",
	GUIErrOpenFolder:     "open folder failed",
	GUIErrNoSelection:    "no selection",

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

var ZH = Messages{
	Title:           "Try 目录选择",
	SearchPrefix:    "搜索：",
	CreateNew:       "新建：",
	CreateHint:      "⌃T 创建 \"%s\"",
	LoadingHint:     "正在加载目录...",
	EmptyStateHint:  "暂无目录",
	NoMatchesHint:   "无匹配结果 \"%s\"",
	DeleteMode:      " 删除模式  已标记 %d 个  |  Ctrl-D: 切换  Enter: 确认  Esc: 取消",
	DeleteModeLabel: " 删除模式 ",
	MarkedCount:     "已标记 %d 个",
	ItemCount:       "%d 项",
	HintBar:         "Ctrl-T: 新建  Ctrl-D: 删除  Ctrl-R: 重命名  Ctrl-G: 发布  Tab: 过滤  Esc: 退出",
	EmptyInputHint:  "请先输入目录名称",
	DeleteCancelled: "已取消删除",
	FilterAll:       "all",

	DeleteTitle:         "确认删除 %d 个目录？",
	DeleteConfirmPrompt: "确认：",
	DeleteOptionNo:      "NO",
	DeleteOptionYes:     "YES",
	DeleteFooter:        "←/→: 选择  Enter: 确认  Esc: 取消",

	RenameTitle:  "重命名目录",
	RenamePrompt: "新名称：",
	RenameEmpty:  "名称不能为空",
	RenameSlash:  "名称不能包含 /",
	RenameExists: "目录已存在：",
	RenameFooter: "Enter: 确认  Esc: 取消",

	ShipTitle:       "发布为正式项目",
	ShipMoveLabel:   "移动到：",
	ShipHint:        "目录将被移动到目标位置",
	ShipEmptyErr:    "目标路径不能为空",
	ShipExistsErr:   "目标已存在：",
	ShipNoParentErr: "父目录不存在：",
	ShipFooter:      "Tab: 切换  Enter: 确认  Esc: 取消",

	HelpText: `try - 临时实验目录管理工具

用法:
  try [query]                搜索并选择目录
  try clone <url> [name]     克隆 Git 仓库
  try worktree <dir> [name]  创建 Git worktree
  try install                安装 Shell 集成
  try . <name>               创建 worktree 或目录

选项:
  --help, -h           显示帮助
  --version, -v        显示版本

快捷键:
  Enter      选择/确认
  ↑/↓        上下移动（到边界后循环）
  Ctrl-P/N   上下移动（到边界后循环）
  Space      标记/取消删除
  Delete     标记/取消删除
  Ctrl-A     标记全部可见项
  Ctrl-D     切换删除模式
  Ctrl-T     创建新目录
  Ctrl-R     重命名
  Ctrl-G     Ship（发布为正式项目）
  Tab        切换来源过滤
  Shift-Tab  反向切换来源过滤
  /          聚焦搜索
  Esc        退出/取消`,
	UsageClone:    "用法: try clone <url> [名称]",
	UsageWorktree: "用法: try worktree <目录> [名称]",
	UsageDot:      "用法: try . <名称>",

	TimeJustNow:  "刚刚",
	TimeMinAgo:   "%d分钟前",
	TimeHourAgo:  "%d小时前",
	TimeDayAgo:   "%d天前",
	TimeWeekAgo:  "%d周前",
	TimeMonthAgo: "%d个月前",
	TimeYearAgo:  "%d年前",

	ErrNotTerminal: "标准输入不是终端",
	ErrParseGitURI: "无法解析 Git URI: ",
	ErrParsePath:   "无法解析路径: ",
	ErrReadDir:     "无法读取目录: ",

	MsgCloneFrom:     "正在从 %s 克隆仓库。",
	MsgShipped:       "已发布: %s → %s",
	ErrMkdir:         "创建目录失败",
	ErrMkdirParent:   "创建父目录失败",
	ErrClone:         "git clone 失败",
	ErrWorktreeAdd:   "git worktree add 失败（已创建普通目录）: %v",
	ErrDeletePartial: "部分删除失败:\n%s",
	ErrRename:        "重命名失败",
	ErrWorktreeMove:  "git worktree move 失败",
	ErrMove:          "移动目录失败",
	ErrUnknownOp:     "未知的操作类型: %d",
	ErrPathDenied:    "路径超出允许范围",
	ErrBadRequest:    "无效请求",

	GUITitle:             "try-gui",
	GUISearchPlace:       "输入以过滤或创建...",
	GUIFilesTitle:        "文件",
	GUIFilesEmpty:        "目录为空",
	GUIFilesHintBar:      "Enter: 打开  Delete: 删除  Esc: 返回  ·  拖拽上传",
	GUIBack:              "返回",
	GUIOpen:              "打开",
	GUIEdit:              "编辑",
	GUIDelete:            "删除",
	GUIConfirm:           "确认",
	GUICancel:            "取消",
	GUIThemeDark:         "深色",
	GUIThemeLight:        "浅色",
	GUIThemeToggle:       "切换主题",
	GUITrayShow:          "显示",
	GUITrayQuit:          "退出",
	GUIToastCreated:      "已创建",
	GUIToastDeleted:      "已删除",
	GUIToastRenamed:      "已重命名",
	GUIToastShipped:      "已发布",
	GUIToastOpened:       "已打开",
	GUIToastOpenedFolder: "已在文件管理器中打开",
	GUIToastCopied:       "已复制 %d 项",
	GUIToastSkipped:      "已跳过 %d 个已存在项",
	GUIToastPartial:      "已复制 %d，跳过 %d",
	GUIToastCopying:      "正在复制…",
	GUIDropImporting:     "正在复制文件…",
	GUIDropProgress:      "正在复制 %d / %d",
	GUIOpenFolder:        "在文件管理器中打开",
	GUIDocxPack:          "打包 .docx",
	GUIDocxUnpack:        "解压 .docx",
	GUINotImplemented:    "尚未实现",
	GUIFilesItemCount:    "%d 个项目",
	GUIColName:           "名称",
	GUIColSize:           "大小",
	GUIColMTime:          "修改时间",
	GUIShortcutNav:       "导航",
	GUIShortcutOpen:      "打开",
	GUIShortcutNew:       "新建",
	GUIShortcutDelete:    "删除",
	GUIShortcutEscBack:   "返回",
	GUIShortcutDrop:      "拖拽上传",
	GUIErrCopy:           "复制失败",
	GUIErrOpen:           "打开失败",
	GUIErrOpenFolder:     "打开文件夹失败",
	GUIErrNoSelection:    "未选择项目",

	ErrDetectShell:    "无法检测 Shell 类型，请确认使用 bash、zsh 或 fish",
	ErrUnsupportShell: "不支持的 Shell 类型: %s",
	ErrGetExePath:     "无法获取 try 可执行文件路径",
	MsgAlreadyInstall: "try shell integration 已安装在 %s 中。",
	MsgReinstallHint:  "如需重新安装，请先手动移除旧版（搜索 \"%s\"）。",
	ErrCreateDir:      "创建目录失败",
	ErrWriteFile:      "无法写入 %s",
	ErrWrite:          "写入失败",
	MsgInstalled:      "已将 try shell integration 写入 %s",
	MsgSourceHint:     "请运行 source %s 或重启终端以生效。",
}

// 进程级全局语言包，CLI 启动时通过 Init 设置一次
var global *Messages

// Init 设置全局语言包，应在 CLI 入口处调用一次
func Init(locale string) {
	global = ForLocale(locale)
}

// Get 返回全局语言包，未初始化时回退到英文
func Get() *Messages {
	if global == nil {
		return &EN
	}
	return global
}

// ForLocale 根据 locale 返回对应语言包
func ForLocale(locale string) *Messages {
	switch locale {
	case "zh":
		return &ZH
	default:
		return &EN
	}
}
