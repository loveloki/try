package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// DetectShell 检测当前 Shell 类型：优先 $SHELL 环境变量，回退到父进程名
func DetectShell() string {
	shellEnv := os.Getenv("SHELL")
	if strings.Contains(shellEnv, "fish") {
		return "fish"
	}
	if strings.Contains(shellEnv, "zsh") {
		return "zsh"
	}
	if strings.Contains(shellEnv, "bash") {
		return "bash"
	}

	parent := getParentProcessName()
	if strings.Contains(parent, "fish") {
		return "fish"
	}
	if strings.Contains(parent, "zsh") {
		return "zsh"
	}
	if strings.Contains(parent, "bash") {
		return "bash"
	}
	return ""
}

func getParentProcessName() string {
	ppid := os.Getppid()
	// Linux: 读取 /proc/<ppid>/comm
	if data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", ppid)); err == nil {
		return strings.TrimSpace(string(data))
	}
	// macOS/BSD: 回退到 ps 命令
	out, err := exec.Command("ps", "-p", strconv.Itoa(ppid), "-o", "comm=").Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	return ""
}
