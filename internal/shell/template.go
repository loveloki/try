package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const marker = "# try shell integration"

// ShellConfig 描述一种 Shell 的配置信息
type ShellConfig struct {
	Name     string
	RCFile   func() string
	InitFunc func(binaryPath string) string
}

// Shells 注册表：支持的 Shell 类型
var Shells = map[string]ShellConfig{
	"bash": {Name: "bash", RCFile: bashRCFile, InitFunc: posixInit},
	"zsh":  {Name: "zsh", RCFile: zshRCFile, InitFunc: posixInit},
	"fish": {Name: "fish", RCFile: fishRCFile, InitFunc: fishInit},
}

func bashRCFile() string {
	home, _ := os.UserHomeDir()
	rc := filepath.Join(home, ".bashrc")
	if _, err := os.Stat(rc); err == nil {
		return rc
	}
	return filepath.Join(home, ".bash_profile")
}

func zshRCFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".zshrc")
}

func fishRCFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "fish", "config.fish")
}

func posixInit(binaryPath string) string {
	quoted := "'" + strings.ReplaceAll(binaryPath, "'", `'"'"'`) + "'"
	return fmt.Sprintf(`%s
try() {
  local out
  out=$(%s exec "$@" 2>/dev/tty)
  if [ $? -eq 0 ]; then
    eval "$out"
  else
    echo "$out"
  fi
}`, marker, quoted)
}

func fishInit(binaryPath string) string {
	quoted := "'" + strings.ReplaceAll(binaryPath, "'", `'"'"'`) + "'"
	return fmt.Sprintf(`%s
function try
  set -l out (%s exec $argv 2>/dev/tty | string collect)
  if test $pipestatus[1] -eq 0
    eval $out
  else
    echo $out
  end
end`, marker, quoted)
}
