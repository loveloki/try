package i18n

// Messages 定义所有用户可见文本，按功能模块分组
type Messages struct {
	// === selector 主界面 ===
	Title           string
	SearchPrefix    string
	CreateNew       string
	DeleteMode      string // 含 %d 占位符
	HintBar         string
	EmptyInputHint  string
	DeleteCancelled string

	// === 删除对话框 ===
	DeleteTitle       string // 含 %d 占位符
	DeletePlaceholder string
	DeletePrompt      string
	DeleteFooter      string

	// === 重命名对话框 ===
	RenameTitle  string
	RenamePrompt string
	RenameEmpty  string
	RenameSlash  string
	RenameExists string // 前缀，后接目录名
	RenameFooter string

	// === Ship 对话框 ===
	ShipTitle       string
	ShipDestLabel   string
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
	Title:           "🏠 Try Directory Selection",
	SearchPrefix:    "Search: ",
	CreateNew:       "📂 Create new: ",
	DeleteMode:      " DELETE MODE  %d marked  |  Ctrl-D: Toggle  Enter: Confirm  Esc: Cancel",
	HintBar:         "Ctrl-T: New  Ctrl-D: Delete  Ctrl-R: Rename  Ctrl-G: Ship  Esc: Quit",
	EmptyInputHint:  "Please enter a directory name first",
	DeleteCancelled: "Delete cancelled",

	DeleteTitle:       "🗑️  Delete %d directories?",
	DeletePlaceholder: "Type YES to confirm",
	DeletePrompt:      "Type YES to confirm: ",
	DeleteFooter:      "Enter: Confirm  Esc: Cancel",

	RenameTitle:  "✏️  Rename directory",
	RenamePrompt: "New name: ",
	RenameEmpty:  "Name cannot be empty",
	RenameSlash:  "Name cannot contain /",
	RenameExists: "Directory exists: ",
	RenameFooter: "Enter: Confirm  Esc: Cancel",

	ShipTitle:       "🚀  Ship try to project",
	ShipDestLabel:   "Destination: ",
	ShipMoveLabel:   "Move to: ",
	ShipHint:        "The directory will be moved to the destination",
	ShipEmptyErr:    "Destination cannot be empty",
	ShipExistsErr:   "Destination already exists: ",
	ShipNoParentErr: "Parent directory does not exist: ",
	ShipFooter:      "Enter: Confirm  Esc: Cancel",

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
  --path PATH          Set tries root directory
  --theme dark|light   Color theme (default: auto detect)
  --locale en|zh       UI language (default: auto detect)
  --no-colors          Disable colors

Shortcuts:
  Enter    Select/Confirm
  Ctrl-T   Create new directory
  Ctrl-D   Toggle delete mark
  Ctrl-R   Rename
  Ctrl-G   Ship (publish as project)
  Ctrl-P/N Move up/down
  Esc      Quit`,
	UsageClone:    "Usage: try clone <url> [name]",
	UsageWorktree: "Usage: try worktree <dir> [name]",
	UsageDot:      "Usage: try . <name>",

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
	Title:           "🏠 Try 目录选择",
	SearchPrefix:    "搜索：",
	CreateNew:       "📂 新建：",
	DeleteMode:      " 删除模式  已标记 %d 个  |  Ctrl-D: 切换  Enter: 确认  Esc: 取消",
	HintBar:         "Ctrl-T: 新建  Ctrl-D: 删除  Ctrl-R: 重命名  Ctrl-G: 发布  Esc: 退出",
	EmptyInputHint:  "请先输入目录名称",
	DeleteCancelled: "已取消删除",

	DeleteTitle:       "🗑️  确认删除 %d 个目录？",
	DeletePlaceholder: "输入 YES 确认",
	DeletePrompt:      "输入 YES 确认：",
	DeleteFooter:      "Enter: 确认  Esc: 取消",

	RenameTitle:  "✏️  重命名目录",
	RenamePrompt: "新名称：",
	RenameEmpty:  "名称不能为空",
	RenameSlash:  "名称不能包含 /",
	RenameExists: "目录已存在：",
	RenameFooter: "Enter: 确认  Esc: 取消",

	ShipTitle:       "🚀  发布为正式项目",
	ShipDestLabel:   "目标目录：",
	ShipMoveLabel:   "移动到：",
	ShipHint:        "目录将被移动到目标位置",
	ShipEmptyErr:    "目标路径不能为空",
	ShipExistsErr:   "目标已存在：",
	ShipNoParentErr: "父目录不存在：",
	ShipFooter:      "Enter: 确认  Esc: 取消",

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
  --path PATH          指定 tries 根目录
  --theme dark|light   配色主题（默认 auto 自动检测）
  --locale en|zh       界面语言（默认 auto 自动检测）
  --no-colors          禁用颜色

快捷键:
  Enter    选择/确认
  Ctrl-T   创建新目录
  Ctrl-D   标记/取消删除
  Ctrl-R   重命名
  Ctrl-G   Ship（发布为正式项目）
  Ctrl-P/N 上下移动
  Esc      退出`,
	UsageClone:    "用法: try clone <url> [名称]",
	UsageWorktree: "用法: try worktree <目录> [名称]",
	UsageDot:      "用法: try . <名称>",

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
