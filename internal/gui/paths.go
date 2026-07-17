package gui

import (
	"path/filepath"
	"strings"
)

// IsAllowedPath 判断 path 是否落在任一 allowedRoots 子树内（含根目录自身，供只读列目录使用）。
func IsAllowedPath(path string, allowedRoots []string) bool {
	return matchRoots(path, allowedRoots, true)
}

// IsAllowedTarget 判断 path 是否为可安全修改的目标：必须严格位于某个根目录之内，
// 根目录自身（tries / ship 根）不允许被删除、重命名或 ship，避免误删整个实验库。
func IsAllowedTarget(path string, allowedRoots []string) bool {
	return matchRoots(path, allowedRoots, false)
}

func matchRoots(path string, allowedRoots []string, allowRootItself bool) bool {
	if path == "" || len(allowedRoots) == 0 {
		return false
	}
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return false
	}
	for _, root := range allowedRoots {
		if root == "" {
			continue
		}
		rootAbs, err := filepath.Abs(filepath.Clean(root))
		if err != nil {
			continue
		}
		if abs == rootAbs {
			if allowRootItself {
				return true
			}
			continue
		}
		if isUnderRoot(abs, rootAbs) {
			return true
		}
	}
	return false
}

func isUnderRoot(abs, root string) bool {
	if abs == root {
		return true
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
