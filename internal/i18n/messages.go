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
	GUITitle               string
	GUISearchPlace         string
	GUIFilesTitle          string
	GUIFilesEmpty          string
	GUIFilesHintBar        string // 兼容旧单行；新底栏用下方分段字段
	GUIBack                string
	GUIOpen                string
	GUIEdit                string
	GUIDelete              string
	GUIConfirm             string
	GUICancel              string
	GUIThemeDark           string
	GUIThemeLight          string
	GUIThemeToggle         string // 主题按钮无障碍标签
	GUITrayShow            string
	GUITrayQuit            string
	GUIToastCreated        string
	GUIToastDeleted        string
	GUIToastRenamed        string
	GUIToastShipped        string
	GUIToastOpened         string
	GUIToastOpenedFolder   string
	GUIToastCopied         string // 含 %d
	GUIToastSkipped        string // 含 %d
	GUIToastPartial        string // 含 %d %d（copied, skipped）
	GUIToastCopying        string
	GUIDropImporting       string
	GUIDropProgress        string // 含 %d %d（done, total）
	GUIOpenFolder          string
	GUIDocxPack            string
	GUIDocxUnpack          string
	GUIToastDocxPacked     string // 含 %s：输出文件名
	GUIToastDocxUnpacked   string // 含 %s：输出目录名
	GUIErrDocxNeedFile     string
	GUIErrDocxNeedDir      string
	GUIErrDocxMulti        string
	GUINotImplemented      string
	GUIFilesItemCount      string // 含 %d
	GUIColName             string
	GUIColSize             string
	GUIColMTime            string
	GUIShortcutNav         string
	GUIShortcutOpen        string
	GUIShortcutNew         string
	GUIShortcutDelete      string
	GUIShortcutEscBack     string
	GUIShortcutDrop        string
	GUIContextMenuOpen     string
	GUIContextMenuOpenWith string
	GUIContextMenuReveal   string
	GUIContextMenuRename   string
	GUIContextMenuDelete   string
	GUIContextMenuNoApps   string
	GUIRenameFileTitle     string
	GUIErrCopy             string
	GUIErrOpen             string
	GUIErrOpenFolder       string
	GUIErrNoSelection      string
	GUITitleClone          string // 克隆对话框标题
	GUICloneURL            string // URL 输入框标签
	GUICloneName           string // 名称输入框标签（可选）
	GUICloneURLPlace       string // URL 输入框占位符
	GUICloning             string // 克隆进行中状态栏文案
	GUIToastCloned         string // 克隆完成提示

	// 设置对话框
	GUISettingsTitle          string
	GUISettingsAppearance     string
	GUISettingsTheme          string
	GUISettingsThemeDark      string
	GUISettingsThemeLight     string
	GUISettingsThemeAuto      string
	GUISettingsLanguage       string
	GUISettingsLangAuto       string
	GUISettingsOpenWith       string
	GUISettingsExtension      string
	GUISettingsApplication    string
	GUISettingsAdd            string
	GUISettingsDelete         string
	GUISettingsExtPlaceholder string
	GUISettingsInvalidExt     string
	GUISettingsNoMappings     string
	GUISettingsClose          string
	GUISettingsErrSave        string

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
