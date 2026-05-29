package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/loveloki/try/internal/i18n"
)

func msgs() *i18n.Messages { return i18n.Get() }

// Install 自动检测 Shell 类型并将包装函数追加到配置文件
func Install() error {
	m := msgs()
	shellType := DetectShell()
	if shellType == "" {
		return fmt.Errorf("%s", m.ErrDetectShell)
	}

	cfg, ok := Shells[shellType]
	if !ok {
		return fmt.Errorf(m.ErrUnsupportShell, shellType) //nolint:govet // 模板来自 i18n 消息
	}

	return installToFile(cfg)
}

func installToFile(cfg ShellConfig) error {
	m := msgs()
	rcFile := cfg.RCFile()

	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("%s: %w", m.ErrGetExePath, err)
	}
	binaryPath, _ = filepath.EvalSymlinks(binaryPath)

	if data, err := os.ReadFile(rcFile); err == nil {
		if strings.Contains(string(data), marker) {
			fmt.Fprintf(os.Stderr, m.MsgAlreadyInstall+"\n", rcFile)
			fmt.Fprintf(os.Stderr, m.MsgReinstallHint+"\n", marker)
			return nil
		}
	}

	if err := os.MkdirAll(filepath.Dir(rcFile), 0o755); err != nil {
		return fmt.Errorf("%s: %w", m.ErrCreateDir, err)
	}

	initContent := cfg.InitFunc(binaryPath)
	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf(m.ErrWriteFile+": %w", rcFile, err)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "\n%s\n", initContent); err != nil {
		return fmt.Errorf("%s: %w", m.ErrWrite, err)
	}

	fmt.Fprintf(os.Stderr, m.MsgInstalled+"\n", rcFile)
	fmt.Fprintf(os.Stderr, m.MsgSourceHint+"\n", rcFile)
	return nil
}
