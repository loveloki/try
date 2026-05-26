package i18n

// Messages 定义所有用户可见的 TUI 文本
type Messages struct {
	// selector 主界面
	Title           string
	SearchPrefix    string
	CreateNew       string
	DeleteMode      string // 含 %d 占位符
	HintBar         string
	EmptyInputHint  string
	DeleteCancelled string

	// 删除对话框
	DeleteTitle       string // 含 %d 占位符
	DeletePlaceholder string
	DeletePrompt      string
	DeleteFooter      string

	// 重命名对话框
	RenameTitle  string
	RenamePrompt string
	RenameEmpty  string
	RenameSlash  string
	RenameExists string // 前缀，后接目录名
	RenameFooter string

	// Ship 对话框
	ShipTitle       string
	ShipDestLabel   string
	ShipMoveLabel   string
	ShipHint        string
	ShipEmptyErr    string
	ShipExistsErr   string // 前缀，后接路径
	ShipNoParentErr string // 前缀，后接路径
	ShipFooter      string
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
