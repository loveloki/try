package i18n

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

	GUITitle:               "try-gui",
	GUISearchPlace:         "输入以过滤或创建...",
	GUIFilesTitle:          "文件",
	GUIFilesEmpty:          "目录为空",
	GUIFilesHintBar:        "Enter: 打开  Delete: 删除  Esc: 返回  ·  拖拽上传",
	GUIBack:                "返回",
	GUIOpen:                "打开",
	GUIEdit:                "编辑",
	GUIDelete:              "删除",
	GUIConfirm:             "确认",
	GUICancel:              "取消",
	GUIThemeDark:           "深色",
	GUIThemeLight:          "浅色",
	GUIThemeToggle:         "切换主题",
	GUITrayShow:            "显示",
	GUITrayQuit:            "退出",
	GUIToastCreated:        "已创建",
	GUIToastDeleted:        "已删除",
	GUIToastRenamed:        "已重命名",
	GUIToastShipped:        "已发布",
	GUIToastOpened:         "已打开",
	GUIToastOpenedFolder:   "已在文件管理器中打开",
	GUIToastCopied:         "已复制 %d 项",
	GUIToastSkipped:        "已跳过 %d 个已存在项",
	GUIToastPartial:        "已复制 %d，跳过 %d",
	GUIToastCopying:        "正在复制…",
	GUIDropImporting:       "正在复制文件…",
	GUIDropProgress:        "正在复制 %d / %d",
	GUIOpenFolder:          "在文件管理器中打开",
	GUIDocxPack:            "打包 .docx",
	GUIDocxUnpack:          "解压 .docx",
	GUIToastDocxPacked:     "已打包 %s",
	GUIToastDocxUnpacked:   "已解压到 %s",
	GUIErrDocxNeedFile:     "请选择一个 .docx 文件",
	GUIErrDocxNeedDir:      "请选择一个文件夹进行打包",
	GUIErrDocxMulti:        "请只选择一项",
	GUINotImplemented:      "尚未实现",
	GUIFilesItemCount:      "%d 个项目",
	GUIColName:             "名称",
	GUIColSize:             "大小",
	GUIColMTime:            "修改时间",
	GUIShortcutNav:         "导航",
	GUIShortcutOpen:        "打开",
	GUIShortcutNew:         "新建",
	GUIShortcutDelete:      "删除",
	GUIShortcutEscBack:     "返回",
	GUIShortcutDrop:        "拖拽上传",
	GUIContextMenuOpen:     "打开",
	GUIContextMenuOpenWith: "用...打开",
	GUIContextMenuReveal:   "在文件夹中显示",
	GUIContextMenuRename:   "重命名",
	GUIContextMenuDelete:   "删除",
	GUIContextMenuNoApps:   "未找到可用应用",
	GUIRenameFileTitle:     "重命名文件",
	GUIErrCopy:             "复制失败",
	GUIErrOpen:             "打开失败",
	GUIErrOpenFolder:       "打开文件夹失败",
	GUIErrNoSelection:      "未选择项目",
	GUITitleClone:          "克隆仓库",
	GUICloneURL:            "URL",
	GUICloneName:           "名称（可选）",
	GUICloneURLPlace:       "https://github.com/user/repo",
	GUICloning:             "正在克隆…",
	GUIToastCloned:         "克隆完成",

	// 设置对话框
	GUISettingsTitle:          "设置",
	GUISettingsAppearance:     "外观",
	GUISettingsTheme:          "主题",
	GUISettingsThemeDark:      "深色",
	GUISettingsThemeLight:     "浅色",
	GUISettingsThemeAuto:      "跟随系统",
	GUISettingsLanguage:       "语言",
	GUISettingsLangAuto:       "自动",
	GUISettingsOpenWith:       "打开方式",
	GUISettingsExtension:      "扩展名",
	GUISettingsApplication:    "应用",
	GUISettingsAdd:            "添加",
	GUISettingsDelete:         "删除",
	GUISettingsExtPlaceholder: ".go 或 *",
	GUISettingsInvalidExt:     "扩展名只能包含字母和数字",
	GUISettingsInvalidApp:     "应用名不合法",
	GUISettingsNoMappings:     "暂无自定义映射",
	GUISettingsClose:          "关闭",
	GUISettingsBack:           "返回",
	GUISettingsMappingAdded:   "%s → %s 已添加",
	GUISettingsErrSave:        "保存设置失败",

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
