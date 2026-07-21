package gui

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

// appCandidate 表示一个已知应用及其可执行名。
type appCandidate struct {
	Name string
	Path string // macOS: bundle 路径; 其他: 可执行名
}

// builtinApps 返回各平台内置常见编辑器列表。
// macOS 使用 /Applications 下的 bundle 名称，
// Linux 使用可执行文件名（通过 which 检测）。
func builtinApps() []appCandidate {
	switch runtime.GOOS {
	case "darwin":
		return []appCandidate{
			{Name: "Visual Studio Code", Path: "Visual Studio Code.app"},
			{Name: "Sublime Text", Path: "Sublime Text.app"},
			{Name: "Vim", Path: "/Applications/MacVim.app"},
			{Name: "TextMate", Path: "TextMate.app"},
			{Name: "TextEdit", Path: "TextEdit.app"},
			{Name: "Xcode", Path: "Xcode.app"},
		}
	case "windows":
		return []appCandidate{
			{Name: "Notepad", Path: "notepad"},
			{Name: "VS Code", Path: "code"},
			{Name: "Sublime Text", Path: "sublime_text"},
		}
	default:
		return []appCandidate{
			{Name: "VS Code", Path: "code"},
			{Name: "Sublime Text", Path: "subl"},
			{Name: "Vim", Path: "vim"},
			{Name: "Gedit", Path: "gedit"},
			{Name: "Nano", Path: "nano"},
		}
	}
}

// availableApp 带有可用性标志的应用信息。
type availableApp struct {
	Name      string
	Path      string
	Available bool
}

// buildAvailableApps 构建应用列表，按 config > 系统可用 过滤。
// openWithConfig 为用户配置的扩展名→应用名映射，优先级高于内置列表。
func buildAvailableApps(ext string, openWithConfig map[string]string) []availableApp {
	candidates := builtinApps()
	apps := make([]availableApp, 0, len(candidates))
	for _, c := range candidates {
		apps = append(apps, availableApp{
			Name:      c.Name,
			Path:      c.Path,
			Available: isAppAvailable(c.Path),
		})
	}
	return filterAppsByConfig(apps, ext, openWithConfig)
}

// filterAppsByConfig 应用打开方式映射，优先级：精确扩展名 > 通配 `*` > 内置列表。
// 通配 `*` 为通用映射，在无精确命中时生效（含无扩展名文件与目录）。
func filterAppsByConfig(apps []availableApp, ext string, cfg map[string]string) []availableApp {
	if cfg == nil {
		return apps
	}
	if ext != "" {
		if configured, ok := cfg[ext]; ok {
			return resolveConfiguredApp(apps, configured)
		}
	}
	if wildcard, ok := cfg["*"]; ok {
		return resolveConfiguredApp(apps, wildcard)
	}
	return apps
}

// resolveConfiguredApp 将映射的应用名解析为可用应用：内置列表命中时只保留该项；
// 未命中时按自定义应用解析（macOS 应用名 / PATH 可执行名 / 绝对路径）。
func resolveConfiguredApp(apps []availableApp, name string) []availableApp {
	configured := strings.ToLower(strings.TrimSpace(name))
	for _, a := range apps {
		if strings.ToLower(a.Name) == configured || strings.Contains(strings.ToLower(a.Path), configured) {
			a.Available = true
			return []availableApp{a}
		}
	}
	return findCustomApp(name)
}

// findCustomApp 将用户配置的应用名解析为可用应用，解析失败返回空列表。
func findCustomApp(name string) []availableApp {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	for _, p := range customAppPaths(name) {
		if isAppAvailable(p) {
			return []availableApp{{Name: name, Path: p, Available: true}}
		}
	}
	return nil
}

func customAppPaths(name string) []string {
	if filepath.IsAbs(name) {
		return []string{name}
	}
	if runtime.GOOS == "darwin" {
		return bundleCandidates(name)
	}
	return execCandidates(name)
}

// bundleCandidates macOS 下依次尝试原名、.app 后缀与小写形式。
func bundleCandidates(name string) []string {
	if strings.HasSuffix(name, ".app") {
		return []string{name}
	}
	return []string{name, name + ".app", strings.ToLower(name)}
}

// execCandidates Linux/Windows 下依次尝试原名与小写形式。
func execCandidates(name string) []string {
	if lower := strings.ToLower(name); lower != name {
		return []string{name, lower}
	}
	return []string{name}
}

// installedAppNames 返回已安装应用候选：macOS 为 .app 应用名（可用 open -a 打开），
// 其他平台为 PATH 中的可执行名。
func installedAppNames() []string {
	if runtime.GOOS == "darwin" {
		return applicationNames()
	}
	return pathExecutables(0)
}

// applicationNames 扫描 macOS 应用目录，按字典序返回去掉 .app 后缀的应用名。
func applicationNames() []string {
	dirs := []string{
		"/Applications",
		"/Applications/Utilities",
		"/System/Applications",
		"/System/Applications/Utilities",
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, "Applications"))
	}
	seen := make(map[string]bool)
	names := make([]string, 0, 64)
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			name := e.Name()
			if !strings.HasSuffix(name, ".app") || seen[name] {
				continue
			}
			seen[name] = true
			names = append(names, strings.TrimSuffix(name, ".app"))
		}
	}
	sort.Strings(names)
	return names
}

// pathExecutables 扫描 PATH 目录，按字典序返回去重后的可执行文件名；limit > 0 时最多 limit 个。
func pathExecutables(limit int) []string {
	seen := make(map[string]bool)
	names := make([]string, 0, limit)
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if seen[e.Name()] || !isExecutableEntry(e) {
				continue
			}
			seen[e.Name()] = true
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	if limit > 0 && len(names) > limit {
		names = names[:limit]
	}
	return names
}

func isExecutableEntry(e os.DirEntry) bool {
	if e.IsDir() {
		return false
	}
	if runtime.GOOS == "windows" {
		ext := strings.ToLower(filepath.Ext(e.Name()))
		return ext == ".exe" || ext == ".bat" || ext == ".cmd"
	}
	info, err := e.Info()
	if err != nil {
		return false
	}
	return info.Mode().IsRegular() && info.Mode()&0o111 != 0
}

// isAppAvailable 检测应用在当前系统上是否可用。
func isAppAvailable(appPath string) bool {
	if runtime.GOOS == "darwin" {
		return macOSAppExists(appPath)
	}
	return executableExists(appPath)
}

func macOSAppExists(bundleName string) bool {
	if filepath.IsAbs(bundleName) {
		_, err := os.Stat(bundleName)
		return err == nil
	}
	candidates := []string{
		filepath.Join("/Applications", bundleName),
		filepath.Join("/System/Applications", bundleName),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func executableExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
