package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Install 自动检测 Shell 类型并将包装函数追加到配置文件
func Install() error {
	shellType := DetectShell()
	if shellType == "" {
		return fmt.Errorf("无法检测 Shell 类型，请确认使用 bash、zsh 或 fish")
	}

	cfg, ok := Shells[shellType]
	if !ok {
		return fmt.Errorf("不支持的 Shell 类型: %s", shellType)
	}

	return installToFile(cfg)
}

func installToFile(cfg ShellConfig) error {
	rcFile := cfg.RCFile()

	// 获取 try 二进制的绝对路径
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("无法获取 try 可执行文件路径: %w", err)
	}
	binaryPath, _ = filepath.EvalSymlinks(binaryPath)

	// 检查是否已安装
	if data, err := os.ReadFile(rcFile); err == nil {
		if strings.Contains(string(data), marker) {
			fmt.Fprintf(os.Stderr, "try shell integration 已安装在 %s 中。\n", rcFile)
			fmt.Fprintf(os.Stderr, "如需重新安装，请先手动移除旧版（搜索 \"%s\"）。\n", marker)
			return nil
		}
	}

	// 确保父目录存在
	if err := os.MkdirAll(filepath.Dir(rcFile), 0o755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 追加包装函数
	initContent := cfg.InitFunc(binaryPath)
	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("无法写入 %s: %w", rcFile, err)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "\n%s\n", initContent); err != nil {
		return fmt.Errorf("写入失败: %w", err)
	}

	fmt.Fprintf(os.Stderr, "已将 try shell integration 写入 %s\n", rcFile)
	fmt.Fprintf(os.Stderr, "请运行 source %s 或重启终端以生效。\n", rcFile)
	return nil
}
