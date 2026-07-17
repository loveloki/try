//go:build embed

package gui

import (
	"embed"
	"io/fs"
)

//go:embed all:web
var webFS embed.FS

// WebAssets 返回嵌入的 web 静态资源（根为 web/ 内容）。
func WebAssets() fs.FS {
	sub, err := fs.Sub(webFS, "web")
	if err != nil {
		return webFS
	}
	return sub
}
