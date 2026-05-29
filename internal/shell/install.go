package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xleine/try/internal/i18n"
)

// Install 自动检测 Shell 类型并将包装函数追加到配置文件
func Install(msgs *i18n.Messages) error {
	shellType := DetectShell()
	if shellType == "" {
		return fmt.Errorf("%s", msgs.ErrDetectShell)
	}

	cfg, ok := Shells[shellType]
	if !ok {
		return fmt.Errorf(msgs.ErrUnsupportShell, shellType) //nolint:govet // 模板来自 i18n 消息
	}

	return installToFile(cfg, msgs)
}

func installToFile(cfg ShellConfig, msgs *i18n.Messages) error {
	rcFile := cfg.RCFile()

	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("%s: %w", msgs.ErrGetExePath, err)
	}
	binaryPath, _ = filepath.EvalSymlinks(binaryPath)

	// 检查是否已安装
	if data, err := os.ReadFile(rcFile); err == nil {
		if strings.Contains(string(data), marker) {
			fmt.Fprintf(os.Stderr, msgs.MsgAlreadyInstall+"\n", rcFile)
			fmt.Fprintf(os.Stderr, msgs.MsgReinstallHint+"\n", marker)
			return nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(rcFile), 0o755); err != nil {
		return fmt.Errorf("%s: %w", msgs.ErrCreateDir, err)
	}

	initContent := cfg.InitFunc(binaryPath)
	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf(msgs.ErrWriteFile+": %w", rcFile, err)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "\n%s\n", initContent); err != nil {
		return fmt.Errorf("%s: %w", msgs.ErrWrite, err)
	}

	fmt.Fprintf(os.Stderr, msgs.MsgInstalled+"\n", rcFile)
	fmt.Fprintf(os.Stderr, msgs.MsgSourceHint+"\n", rcFile)
	return nil
}
