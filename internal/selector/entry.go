package selector

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"
)

// DateSuffixRe 匹配 name-YYYY-MM-DD 格式的日期后缀
var DateSuffixRe = regexp.MustCompile(`-\d{4}-\d{2}-\d{2}$`)

// SelectionType 选择结果类型
type SelectionType int

const (
	SelectCD SelectionType = iota
	SelectMkdir
	SelectDelete
	SelectRename
	SelectShip
)

// Entry 从文件系统加载的原始目录条目
type Entry struct {
	Basename  string
	Path      string
	Mtime     time.Time
	BaseScore float64
	Source    string // "tries" 或 ship 目录的 basename（如 "ship"、"bug"）
}

// MatchedEntry 匹配后的条目（包含评分和高亮信息），实现 list.Item 接口
type MatchedEntry struct {
	Entry              Entry
	Score              float64
	HighlightPositions []int
}

func (m MatchedEntry) FilterValue() string { return "" }
func (m MatchedEntry) Title() string       { return m.Entry.Basename }
func (m MatchedEntry) Description() string { return "" }

// DeleteItem 删除操作的条目信息
type DeleteItem struct {
	Path     string // 安全检查后的绝对路径
	Basename string
}

// SelectionResult 选择器的最终输出
type SelectionResult struct {
	Type     SelectionType
	Path     string       // :cd, :mkdir
	Paths    []DeleteItem // :delete
	Old, New string       // :rename
	Source   string       // :ship
	Dest     string       // :ship
	Basename string       // :ship
	BasePath string       // :delete, :rename, :ship
}

// --- 工具函数 ---

// DirExists 检查目录是否存在
func DirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// FileExists 检查文件或目录是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsFile 检查路径是否为文件（非目录）
func IsFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// EnvInt 读取环境变量为 int，不存在或解析失败返回 0
func EnvInt(key string) int {
	v, _ := strconv.Atoi(os.Getenv(key))
	return v
}

// FormatTimeAgo 将时间差格式化为本地化的人类可读格式
func FormatTimeAgo(d time.Duration) string {
	m := msgs()
	seconds := int(d.Seconds())
	if seconds < 60 {
		return m.TimeJustNow
	}
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf(m.TimeMinAgo, minutes)
	}
	hours := minutes / 60
	if hours < 24 {
		return fmt.Sprintf(m.TimeHourAgo, hours)
	}
	days := hours / 24
	if days < 7 {
		return fmt.Sprintf(m.TimeDayAgo, days)
	}
	weeks := days / 7
	return fmt.Sprintf(m.TimeWeekAgo, weeks)
}
