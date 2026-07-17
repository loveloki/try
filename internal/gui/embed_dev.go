//go:build !embed

package gui

import "io/fs"

// WebAssets 非 embed 构建返回 nil，仅提供 API。
func WebAssets() fs.FS {
	return nil
}
