package git

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// GitURIInfo 从 Git URL 中解析出的结构化信息
type GitURIInfo struct {
	User string
	Repo string
	Host string
}

var (
	httpsRe    = regexp.MustCompile(`^https?://([^/]+)/([^/]+)/([^/]+)`)
	sshRe      = regexp.MustCompile(`^git@([^:]+):([^/]+)/([^/]+)`)
	sshProtoRe = regexp.MustCompile(`^ssh://(?:[^@]+@)?([^/:]+)(?::\d+)?/([^/]+)/([^/]+)`)
)

// ParseGitURI 解析 Git URL（HTTPS 和 SSH 格式），返回 nil 表示无法解析
func ParseGitURI(uri string) *GitURIInfo {
	uri = strings.TrimSuffix(uri, ".git")

	if m := httpsRe.FindStringSubmatch(uri); m != nil {
		return &GitURIInfo{Host: m[1], User: m[2], Repo: m[3]}
	}
	if m := sshRe.FindStringSubmatch(uri); m != nil {
		return &GitURIInfo{Host: m[1], User: m[2], Repo: m[3]}
	}
	if m := sshProtoRe.FindStringSubmatch(uri); m != nil {
		return &GitURIInfo{Host: m[1], User: m[2], Repo: m[3]}
	}
	return nil
}

// IsGitURI 快速判断参数是否看起来像 Git URL（不做完整解析）
func IsGitURI(arg string) bool {
	if arg == "" {
		return false
	}
	return strings.HasPrefix(arg, "https://") ||
		strings.HasPrefix(arg, "http://") ||
		strings.HasPrefix(arg, "ssh://") ||
		strings.HasPrefix(arg, "git@") ||
		strings.Contains(arg, "github.com") ||
		strings.Contains(arg, "gitlab.com") ||
		strings.HasSuffix(arg, ".git")
}

// GenerateCloneDirName 生成 clone 目录名。
// 自定义名称优先；自动命名格式：user-repo-YYYY-MM-DD。
func GenerateCloneDirName(gitURI, customName string) string {
	return generateCloneDirNameWithDate(gitURI, customName, time.Now().Format("2006-01-02"))
}

// generateCloneDirNameWithDate 可注入日期的内部实现，便于测试
func generateCloneDirNameWithDate(gitURI, customName, dateSuffix string) string {
	if customName != "" {
		return customName
	}
	parsed := ParseGitURI(gitURI)
	if parsed == nil {
		return ""
	}
	return parsed.User + "-" + parsed.Repo + "-" + dateSuffix
}

// ResolveUniqueName 处理同名目录冲突，返回不含日期后缀的 base 名称。
// 调用方负责拼接 base + "-" + dateSuffix 得到最终目录名。
func ResolveUniqueName(triesPath, base, dateSuffix string) string {
	initial := base + "-" + dateSuffix
	if !dirExists(filepath.Join(triesPath, initial)) {
		return base
	}

	// 尾部是数字：递增数字（如 v2 → v3）
	trailingNum := regexp.MustCompile(`^(.*?)(\d+)$`)
	if m := trailingNum.FindStringSubmatch(base); m != nil {
		stem := m[1]
		n, _ := strconv.Atoi(m[2])
		for {
			n++
			candidate := stem + strconv.Itoa(n)
			if !dirExists(filepath.Join(triesPath, candidate+"-"+dateSuffix)) {
				return candidate
			}
		}
	}

	// 无数字后缀：追加 -2, -3, ...
	for i := 2; ; i++ {
		candidate := base + "-" + strconv.Itoa(i)
		if !dirExists(filepath.Join(triesPath, candidate+"-"+dateSuffix)) {
			return candidate
		}
	}
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
