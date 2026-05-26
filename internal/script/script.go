package script

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const ScriptWarning = "# if you can read this, you didn't launch try from an alias. run try --help."

// Quote 用单引号包裹路径，处理路径中的单引号
func Quote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// EmitCd 输出 cd 脚本到 stdout，由父 Shell eval 执行
func EmitCd(path string) {
	EmitCdTo(os.Stdout, path)
}

// EmitCdTo 输出 cd 脚本到指定 writer，便于测试
func EmitCdTo(w io.Writer, path string) {
	fmt.Fprintln(w, ScriptWarning)
	fmt.Fprintln(w, "cd "+Quote(path))
}
